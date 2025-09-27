package user

import (
	"context"
	"fmt"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ProjectServiceInterface interface {
	InitializeOrganization(ctx context.Context, organizationID uuid.UUID) error
}

type UserService struct {
	config                 *shared.Config
	repositoryTxer         shared.RepositoryTxer
	mailResource           shared.MailResource
	userRepository         UserRepository
	organizationRepository OrganizationRepository
	inviteService          *OrganizationInviteService
	projectService         ProjectServiceInterface
}

func NewInMemUserService() *UserService {
	return &UserService{
		userRepository: NewInMemUserRepository(),
	}
}

func NewUserService(
	config *shared.Config,
	repositoryTxer shared.RepositoryTxer,
	mailResource shared.MailResource,
	userRepository UserRepository,
	organizationRepository OrganizationRepository,
	inviteRepository OrganizationInviteRepository,
	projectService ProjectServiceInterface,
) *UserService {
	inviteService := NewOrganizationInviteService(repositoryTxer, inviteRepository)
	return &UserService{
		config:                 config,
		repositoryTxer:         repositoryTxer,
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		inviteService:          inviteService,
		projectService:         projectService,
	}
}

func (a *UserService) ConfirmUser(ctx context.Context, userID uuid.UUID) error {
	return a.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			return a.userRepository.ConfirmUser(ctx, userID)
		},
	)
}

func (a *UserService) EncryptPassword(password string) string {
	encryptedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(encryptedPassword)
}

func (a *UserService) SetUpNewUser(ctx context.Context, user *User, confirmationID uuid.UUID) error {
	// Create Organization
	organization := &Organization{
		ID:    uuid.New(),
		Title: user.Name,
	}

	// Create User
	user.ID = uuid.New()
	user.OrganizationID = organization.ID

	// Send email confirmation link
	subject := "Confirm your Email address"
	body := fmt.Sprintf(
		`Confirm your Email address at %v/signup/confirm/%v to activate your account.`,
		a.config.Webroot,
		confirmationID,
	)

	return a.repositoryTxer.InTx(
		ctx,
		// Create Organization
		func(ctx context.Context) error {
			_, err := a.organizationRepository.InsertOrganization(ctx, organization)
			if err != nil {
				return err
			}
			return nil
		},
		// Create User
		func(ctx context.Context) error {
			_, err := a.userRepository.InsertUserWithConfirmationID(ctx, user, confirmationID)
			if err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			return a.projectService.InitializeOrganization(ctx, organization.ID)
		},
		// Send email confirmation link
		func(ctx context.Context) error {
			if user.EMail == "" {
				return nil
			}
			return a.mailResource.SendMail(user.EMail, subject, body)
		},
	)
}

func (a *UserService) UpdateOrganizationName(ctx context.Context, principal *shared.Principal, newName string) error {
	// Check if user has admin role
	if !principal.HasRole("ROLE_ADMIN") {
		return fmt.Errorf("insufficient permissions: only administrators can update organization name")
	}

	// Validate organization name
	if newName == "" {
		return fmt.Errorf("organization name cannot be empty")
	}
	if len(newName) > 255 {
		return fmt.Errorf("organization name cannot exceed 255 characters")
	}

	// Create organization object with updated name
	organization := &Organization{
		ID:    principal.OrganizationID,
		Title: newName,
	}

	// Update organization in database
	return a.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			return a.organizationRepository.UpdateOrganization(ctx, organization)
		},
	)
}

func (a *UserService) FindOrganizationByID(ctx context.Context, organizationID uuid.UUID) (*Organization, error) {
	var organization *Organization
	err := a.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			var err error
			organization, err = a.organizationRepository.FindOrganizationByID(ctx, organizationID)
			return err
		},
	)
	return organization, err
}

// GenerateOrganizationInvite creates a new invite link for the organization
func (a *UserService) GenerateOrganizationInvite(ctx context.Context, principal *shared.Principal) (*OrganizationInvite, error) {
	// Check if user has admin role
	if !principal.HasRole("ROLE_ADMIN") {
		return nil, fmt.Errorf("insufficient permissions: only administrators can generate invite links")
	}

	// Get user ID from username
	user, err := a.userRepository.FindUserByUsername(ctx, principal.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Generate invite using the invite service
	invite, err := a.inviteService.GenerateInvite(ctx, principal.OrganizationID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate organization invite: %w", err)
	}

	return invite, nil
}

// FindOrganizationInvites returns all invites for the user's organization
func (a *UserService) FindOrganizationInvites(ctx context.Context, principal *shared.Principal) ([]*OrganizationInvite, error) {
	// Check if user has admin role
	if !principal.HasRole("ROLE_ADMIN") {
		return nil, fmt.Errorf("insufficient permissions: only administrators can view invite links")
	}

	// Find invites using the invite service
	invites, err := a.inviteService.FindInvitesByOrganization(ctx, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find organization invites: %w", err)
	}

	return invites, nil
}

// ValidateInviteToken validates an invite token and returns the invite
func (a *UserService) ValidateInviteToken(ctx context.Context, token string) (*OrganizationInvite, error) {
	invite, err := a.inviteService.ValidateInvite(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate invite token: %w", err)
	}

	return invite, nil
}
