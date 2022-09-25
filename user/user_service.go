package user

import (
	"context"
	"fmt"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	config                  *shared.Config
	repositoryTxer          shared.RepositoryTxer
	mailResource            shared.MailResource
	userRepository          UserRepository
	organizationRepository  OrganizationRepository
	organizationInitializer func(ctxWithTx context.Context, organizationID uuid.UUID) error
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
	organizationInitializer func(ctxWithTx context.Context, organizationID uuid.UUID) error,
) *UserService {
	return &UserService{
		config:                  config,
		repositoryTxer:          repositoryTxer,
		mailResource:            mailResource,
		userRepository:          userRepository,
		organizationRepository:  organizationRepository,
		organizationInitializer: organizationInitializer,
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
			return a.organizationInitializer(ctx, organization.ID)
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
