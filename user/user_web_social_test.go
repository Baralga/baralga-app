package user

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/matryer/is"
)

func TestInviteSignupFormWithSocialLogins(t *testing.T) {
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

	// Act
	formHTML := handlers.InviteSignupForm(formModel, "", nil)

	// Convert to string for testing
	wRec := httptest.NewRecorder()
	shared.RenderHTML(wRec, formHTML)
	body := wRec.Body.String()

	// Assert
	is.True(strings.Contains(body, "Continue with GitHub"))
	is.True(strings.Contains(body, "Continue with Google"))
	is.True(strings.Contains(body, "/github/login/invite/test-invite-token"))
	is.True(strings.Contains(body, "/google/login/invite/test-invite-token"))
	is.True(strings.Contains(body, "Or sign up with your social account"))
}
