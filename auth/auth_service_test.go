package auth

import (
	"context"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/baralga/user"
	"github.com/matryer/is"
	"github.com/pkg/errors"
)

func TestAuthenticateTrustedWithExistingUser(t *testing.T) {
	// Arrange
	is := is.New(t)

	a := &AuthService{
		app: &shared.App{
			Config: &shared.Config{},
		},
		userRepository: user.NewInMemUserRepository(),
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
	a := &AuthService{
		app: &shared.App{
			Config: &shared.Config{},
		},
		userRepository: user.NewInMemUserRepository(),
	}
	username := "not.found@baralga.com"

	// Act
	_, err := a.AuthenticateTrusted(context.Background(), username)

	// Assert
	is.True(errors.Is(err, user.ErrUserNotFound))
}

func TestCreateExpiredCookie(t *testing.T) {
	// Arrange
	is := is.New(t)
	a := &AuthService{
		app: &shared.App{
			Config: &shared.Config{},
		},
	}

	// Act
	cookie := a.CreateExpiredCookie()

	// Assert
	is.True(cookie.Expires.Before(time.Now()) || cookie.Expires.Equal(time.Now()))
	is.Equal("jwt", cookie.Name)
	is.Equal("/", cookie.Path)
}
