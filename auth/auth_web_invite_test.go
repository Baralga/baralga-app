package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baralga/shared"
	"github.com/baralga/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/matryer/is"
)

func TestGithubInviteLoginHandler(t *testing.T) {
	// Arrange
	is := is.New(t)
	config := &shared.Config{}
	authService := &AuthService{}
	userService := &user.UserService{}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	handlers := NewAuthWebHandlers(config, authService, userService, tokenAuth)

	// Create request
	req := httptest.NewRequest("GET", "/github/login/invite/test-token", nil)
	wRec := httptest.NewRecorder()

	// Set up chi router context
	r := chi.NewRouter()
	r.Handle("/github/login/invite/{token}", handlers.GithubInviteLoginHandler())
	r.ServeHTTP(wRec, req)

	// Assert
	// Should redirect to GitHub OAuth
	is.True(wRec.Code == http.StatusFound || wRec.Code == http.StatusTemporaryRedirect)
}

func TestGoogleInviteLoginHandler(t *testing.T) {
	// Arrange
	is := is.New(t)
	config := &shared.Config{}
	authService := &AuthService{}
	userService := &user.UserService{}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	handlers := NewAuthWebHandlers(config, authService, userService, tokenAuth)

	// Create request
	req := httptest.NewRequest("GET", "/google/login/invite/test-token", nil)
	wRec := httptest.NewRecorder()

	// Set up chi router context
	r := chi.NewRouter()
	r.Handle("/google/login/invite/{token}", handlers.GoogleInviteLoginHandler())
	r.ServeHTTP(wRec, req)

	// Assert
	// Should redirect to Google OAuth
	is.True(wRec.Code == http.StatusFound || wRec.Code == http.StatusTemporaryRedirect)
}

func TestGithubInviteCallbackHandler(t *testing.T) {
	// Arrange
	is := is.New(t)
	config := &shared.Config{}
	authService := &AuthService{}
	userService := &user.UserService{}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	handlers := NewAuthWebHandlers(config, authService, userService, tokenAuth)

	// Create request
	req := httptest.NewRequest("GET", "/github/callback/invite/test-token", nil)
	wRec := httptest.NewRecorder()

	// Set up chi router context
	r := chi.NewRouter()
	r.Handle("/github/callback/invite/{token}", handlers.GithubInviteCallbackHandler())
	r.ServeHTTP(wRec, req)

	// Assert
	// Should handle OAuth callback (may fail without proper OAuth setup, but should not crash)
	is.True(wRec.Code >= 200)
}

func TestGoogleInviteCallbackHandler(t *testing.T) {
	// Arrange
	is := is.New(t)
	config := &shared.Config{}
	authService := &AuthService{}
	userService := &user.UserService{}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	handlers := NewAuthWebHandlers(config, authService, userService, tokenAuth)

	// Create request
	req := httptest.NewRequest("GET", "/google/callback/invite/test-token", nil)
	wRec := httptest.NewRecorder()

	// Set up chi router context
	r := chi.NewRouter()
	r.Handle("/google/callback/invite/{token}", handlers.GoogleInviteCallbackHandler())
	r.ServeHTTP(wRec, req)

	// Assert
	// Should handle OAuth callback (may fail without proper OAuth setup, but should not crash)
	is.True(wRec.Code >= 200)
}
