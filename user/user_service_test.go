package user

import (
	"context"
	"testing"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

type mockProjectService struct {
	initializeOrganization func(ctx context.Context, organizationID uuid.UUID) error
}

func (m *mockProjectService) InitializeOrganization(ctx context.Context, organizationID uuid.UUID) error {
	if m.initializeOrganization != nil {
		return m.initializeOrganization(ctx, organizationID)
	}
	return nil
}

func TestSetUpNewUser(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	mailCount := len(mailResource.Mails)

	userRepository := NewInMemUserRepository()
	userCount := len(userRepository.users)

	organizationRepository := NewInMemOrganizationRepository()
	organizationCount := len(organizationRepository.organizations)
	organizationInitializerCalled := false

	a := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		projectService: &mockProjectService{
			initializeOrganization: func(ctx context.Context, organizationID uuid.UUID) error {
				organizationInitializerCalled = true
				return nil
			},
		},
	}

	user := &User{
		Name:     "Norah Newbie",
		EMail:    "newbie@baralga.com",
		Password: "myPassword?!§!",
	}
	confirmationID := uuid.New()

	// Act
	err := a.SetUpNewUser(context.Background(), user, confirmationID)

	// Assert
	is.NoErr(err)
	is.Equal(len(mailResource.Mails), mailCount+1)
	is.Equal(len(organizationRepository.organizations), organizationCount+1)
	is.True(organizationInitializerCalled)
	is.Equal(len(userRepository.users), userCount+1)
}

func TestSetUpNewUserWithUserRepositoryError(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	mailCount := len(mailResource.Mails)

	userRepository := NewInMemUserRepository()
	organizationRepository := NewInMemOrganizationRepository()

	a := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		projectService:         &mockProjectService{},
	}

	user := &User{
		Name:     "Norah Newbie",
		EMail:    "newbie@baralga.com",
		Password: "myPassword?!§!",
	}

	// Act
	err := a.SetUpNewUser(context.Background(), user, shared.ConfirmationIDError)

	// Assert
	is.True(err != nil)
	is.Equal(len(mailResource.Mails), mailCount)
}

func TestUpdateOrganizationName(t *testing.T) {
	// Arrange
	is := is.New(t)
	organizationRepository := NewInMemOrganizationRepository()

	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		organizationRepository: organizationRepository,
	}

	adminPrincipal := &shared.Principal{
		Name:           "Admin User",
		Username:       "admin",
		OrganizationID: shared.OrganizationIDSample,
		Roles:          []string{"ROLE_ADMIN"},
	}

	userPrincipal := &shared.Principal{
		Name:           "Regular User",
		Username:       "user",
		OrganizationID: shared.OrganizationIDSample,
		Roles:          []string{"ROLE_USER"},
	}

	// Act & Assert - Admin can update organization name
	err := userService.UpdateOrganizationName(context.Background(), adminPrincipal, "New Organization Name")
	is.NoErr(err)

	// Act & Assert - Regular user cannot update organization name
	err = userService.UpdateOrganizationName(context.Background(), userPrincipal, "New Organization Name")
	is.True(err != nil)
	is.Equal(err.Error(), "insufficient permissions: only administrators can update organization name")

	// Act & Assert - Empty organization name validation
	err = userService.UpdateOrganizationName(context.Background(), adminPrincipal, "")
	is.True(err != nil)
	is.Equal(err.Error(), "organization name cannot be empty")

	// Act & Assert - Organization name too long validation
	longName := string(make([]byte, 256)) // 256 characters
	err = userService.UpdateOrganizationName(context.Background(), adminPrincipal, longName)
	is.True(err != nil)
	is.Equal(err.Error(), "organization name cannot exceed 255 characters")
}
