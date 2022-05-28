package user

import (
	"context"
	"testing"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

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
		app: &shared.App{
			Config: &shared.Config{},
		},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		organizationInitializer: func(ctxWithTx context.Context, organizationID uuid.UUID) error {
			organizationInitializerCalled = true
			return nil
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
		app: &shared.App{
			Config: &shared.Config{},
		},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		organizationInitializer: func(ctxWithTx context.Context, organizationID uuid.UUID) error {
			return nil
		},
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
