package main

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestSetUpNewUser(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := NewInMemMailResource()
	mailCount := len(mailResource.mails)

	userRepository := NewInMemUserRepository()
	userCount := len(userRepository.users)

	organizationRepository := NewInMemOrganizationRepository()
	organizationCount := len(organizationRepository.organizations)

	projectRepository := NewInMemProjectRepository()
	projectCount := len(projectRepository.projects)

	a := &app{
		Config: &config{},

		MailResource: mailResource,

		RepositoryTxer:         NewInMemRepositoryTxer(),
		UserRepository:         userRepository,
		OrganizationRepository: organizationRepository,
		ProjectRepository:      projectRepository,
	}

	user := &User{
		Name:     "Norah Newbie",
		EMail:    "newbie@baralga.com",
		Password: "myPassword?!§!",
	}

	// Act
	err := a.SetUpNewUser(context.Background(), user)

	// Assert
	is.NoErr(err)
	is.Equal(len(mailResource.mails), mailCount+1)
	is.Equal(len(organizationRepository.organizations), organizationCount+1)
	is.Equal(len(userRepository.users), userCount+1)
	is.Equal(len(projectRepository.projects), projectCount+1)
}
