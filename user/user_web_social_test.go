package user

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestInviteSignupFormWithSocialLogins(t *testing.T) {
	// Arrange
	is := is.New(t)
	config := &shared.Config{
		Webroot:        "http://localhost:8080",
		GithubClientId: "test-github-client-id",
		GoogleClientId: "test-google-client-id",
	}
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

	// Act
	organization := &Organization{
		ID:    uuid.New(),
		Title: "Test Organization",
	}
	formHTML := handlers.InviteSignupForm(formModel, "", nil, organization)

	// Convert to string for testing
	wRec := httptest.NewRecorder()
	shared.RenderHTML(wRec, formHTML)
	body := wRec.Body.String()

	// Assert
	is.True(strings.Contains(body, "GitHub"))
	is.True(strings.Contains(body, "Google"))
	is.True(strings.Contains(body, "/github/login/invite/test-invite-token"))
	is.True(strings.Contains(body, "/google/login/invite/test-invite-token"))
	is.True(strings.Contains(body, "btn btn-secondary btn-sm"))
}
