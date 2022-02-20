package main

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/pkg/errors"
)

func TestAuthenticateTrustedWithExistingUser(t *testing.T) {
	// Arrange
	is := is.New(t)
	a := &app{
		Config: &config{},

		RepositoryTxer: NewInMemRepositoryTxer(),
		UserRepository: NewInMemUserRepository(),
	}
	username := "admin@baralga.com"

	// Act
	principal, err := a.AuthenticateTrusted(context.Background(), username)

	// Assert
	is.NoErr(err)
	is.Equal(principal.Username, username)
}

func TestAuthenticateTrustedWithMissingUser(t *testing.T) {
	// Arrange
	is := is.New(t)
	a := &app{
		Config: &config{},

		RepositoryTxer: NewInMemRepositoryTxer(),
		UserRepository: NewInMemUserRepository(),
	}
	username := "not.found@baralga.com"

	// Act
	_, err := a.AuthenticateTrusted(context.Background(), username)

	// Assert
	is.True(errors.Is(err, ErrUserNotFound))
}
