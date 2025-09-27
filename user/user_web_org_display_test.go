package user

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestInviteSignupFormDisplaysOrganizationName(t *testing.T) {
	// Arrange
	is := is.New(t)
	config := &shared.Config{Webroot: "http://localhost:8080"}
	userService := &UserService{}
	userRepository := NewInMemUserRepository()

	handlers := &UserWebHandlers{
		config:         config,
		userService:    userService,
		userRepository: userRepository,
	}

	formModel := signupFormModel{
		CSRFToken:   "test-csrf-token",
		InviteToken: "test-invite-token",
	}

	organization := &Organization{
		ID:    uuid.New(),
		Title: "Acme Corporation",
	}

	// Act
	formHTML := handlers.InviteSignupForm(formModel, "", nil, organization)

	// Convert to string for testing
	wRec := httptest.NewRecorder()
	shared.RenderHTML(wRec, formHTML)
	body := wRec.Body.String()

	// Assert
	is.True(strings.Contains(body, "Join Acme Corporation"))
	is.True(strings.Contains(body, "Acme Corporation"))
}

func TestInviteSignUpPageDisplaysOrganizationName(t *testing.T) {
	// Arrange
	is := is.New(t)
	config := &shared.Config{Webroot: "http://localhost:8080"}
	userService := &UserService{}
	userRepository := NewInMemUserRepository()

	handlers := &UserWebHandlers{
		config:         config,
		userService:    userService,
		userRepository: userRepository,
	}

	formModel := signupFormModel{
		CSRFToken:   "test-csrf-token",
		InviteToken: "test-invite-token",
	}

	organization := &Organization{
		ID:    uuid.New(),
		Title: "Tech Startup Inc",
	}

	// Act
	pageHTML := handlers.InviteSignUpPage("/signup/invite/test-token", formModel, organization)

	// Convert to string for testing
	wRec := httptest.NewRecorder()
	shared.RenderHTML(wRec, pageHTML)
	body := wRec.Body.String()

	// Assert
	is.True(strings.Contains(body, "You&#39;ve been invited to join &#39;Tech Startup Inc&#39;"))
	is.True(strings.Contains(body, "Tech Startup Inc"))
}
