package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestHandleInviteSignUpPageWithLoggedInUser(t *testing.T) {
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

	// Create a request with a logged-in user context
	req := httptest.NewRequest("GET", "/signup/invite/test-token", nil)

	// Add principal to context to simulate logged-in user
	principal := &shared.Principal{
		Username:       "testuser@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New(),
		Roles:          []string{"ROLE_USER"},
	}
	req = req.WithContext(shared.ToContextWithPrincipal(req.Context(), principal))

	wRec := httptest.NewRecorder()

	// Set up chi router context
	r := chi.NewRouter()
	r.Get("/signup/invite/{token}", handlers.HandleInviteSignUpPage())
	r.ServeHTTP(wRec, req)

	// Assert
	is.Equal(wRec.Code, http.StatusOK)
	body := wRec.Body.String()

	is.True(strings.Contains(body, "You are currently logged in as &#39;testuser@example.com&#39;"))
	is.True(strings.Contains(body, "Please logout first before using an invite link"))
	is.True(strings.Contains(body, "Logout"))
	is.True(strings.Contains(body, "Go to Dashboard"))
}

func TestInviteLogoutRequiredPage(t *testing.T) {
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

	// Act
	pageHTML := handlers.InviteLogoutRequiredPage("testuser@example.com")

	// Convert to string for testing
	wRec := httptest.NewRecorder()
	shared.RenderHTML(wRec, pageHTML)
	body := wRec.Body.String()

	// Assert
	is.True(strings.Contains(body, "You are currently logged in as &#39;testuser@example.com&#39;"))
	is.True(strings.Contains(body, "Please logout first before using an invite link"))
	is.True(strings.Contains(body, "Logout"))
	is.True(strings.Contains(body, "Go to Dashboard"))
	is.True(strings.Contains(body, "/logout"))
	is.True(strings.Contains(body, "/"))
}
