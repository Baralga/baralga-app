package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baralga/shared"
	"github.com/baralga/user"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

// MockUserRepository for testing
type MockUserRepository struct {
	users map[string]*user.User
	roles map[uuid.UUID][]string
}

func NewMockUserRepository() *MockUserRepository {
	userID := uuid.New()
	orgID := uuid.New()

	return &MockUserRepository{
		users: map[string]*user.User{
			"test@example.com": {
				ID:             userID,
				Name:           "Test User",
				Username:       "test@example.com",
				EMail:          "test@example.com",
				OrganizationID: orgID,
			},
		},
		roles: map[uuid.UUID][]string{
			userID: {"ROLE_USER"},
		},
	}
}

func (m *MockUserRepository) FindUserByUsername(ctx context.Context, username string) (*user.User, error) {
	if user, exists := m.users[username]; exists {
		return user, nil
	}
	return nil, user.ErrUserNotFound
}

func (m *MockUserRepository) FindRolesByUserID(ctx context.Context, organizationID, userID uuid.UUID) ([]string, error) {
	if roles, exists := m.roles[userID]; exists {
		return roles, nil
	}
	return []string{}, nil
}

func (m *MockUserRepository) ConfirmUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *MockUserRepository) FindUserIDByConfirmationID(ctx context.Context, confirmationID string) (uuid.UUID, error) {
	return uuid.Nil, errors.New("not implemented")
}

func (m *MockUserRepository) InsertUserWithConfirmationID(ctx context.Context, user *user.User, confirmationID uuid.UUID) (*user.User, error) {
	return nil, errors.New("not implemented")
}

func TestMCPAuthenticationMiddleware(t *testing.T) {
	is := is.New(t)

	userRepo := NewMockUserRepository()
	mcpAuthService := NewMCPAuthService(userRepo)

	// Create a test handler that checks if principal is in context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())
		is.Equal(principal.Username, "test@example.com")
		is.Equal(principal.Name, "Test User")
		is.Equal(len(principal.Roles), 1)
		is.Equal(principal.Roles[0], "ROLE_USER")
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with authentication middleware
	authMiddleware := mcpAuthService.AuthenticationMiddleware()
	handler := authMiddleware(testHandler)

	t.Run("Valid API key in X-API-Key header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/mcp/test", nil)
		req.Header.Set("X-API-Key", "test@example.com")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		is.Equal(w.Code, http.StatusOK)
	})

	t.Run("Valid API key in Authorization Bearer", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/mcp/test", nil)
		req.Header.Set("Authorization", "Bearer test@example.com")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		is.Equal(w.Code, http.StatusOK)
	})

	t.Run("Missing API key", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/mcp/test", nil)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		is.Equal(w.Code, http.StatusBadRequest)
	})

	t.Run("Invalid email format", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/mcp/test", nil)
		req.Header.Set("X-API-Key", "invalid-email")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		is.Equal(w.Code, http.StatusBadRequest)
	})

	t.Run("User not found", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/mcp/test", nil)
		req.Header.Set("X-API-Key", "nonexistent@example.com")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		is.Equal(w.Code, http.StatusBadRequest)
	})

	t.Run("OPTIONS request bypasses authentication", func(t *testing.T) {
		// Create a simple handler that doesn't check for principal
		optionsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		authMiddleware := mcpAuthService.AuthenticationMiddleware()
		handler := authMiddleware(optionsHandler)

		req := httptest.NewRequest("OPTIONS", "/mcp/test", nil)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		is.Equal(w.Code, http.StatusOK)
	})
}

func TestExtractAPIKey(t *testing.T) {
	is := is.New(t)
	userRepo := NewMockUserRepository()
	mcpAuthService := NewMCPAuthService(userRepo)

	t.Run("Extract from X-API-Key header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("X-API-Key", "test@example.com")

		apiKey := mcpAuthService.extractAPIKey(req)
		is.Equal(apiKey, "test@example.com")
	})

	t.Run("Extract from Authorization Bearer", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Authorization", "Bearer test@example.com")

		apiKey := mcpAuthService.extractAPIKey(req)
		is.Equal(apiKey, "test@example.com")
	})

	t.Run("No API key present", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)

		apiKey := mcpAuthService.extractAPIKey(req)
		is.Equal(apiKey, "")
	})

	t.Run("X-API-Key takes precedence over Authorization", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("X-API-Key", "xapi@example.com")
		req.Header.Set("Authorization", "Bearer bearer@example.com")

		apiKey := mcpAuthService.extractAPIKey(req)
		is.Equal(apiKey, "xapi@example.com")
	})
}

func TestIsValidEmail(t *testing.T) {
	is := is.New(t)
	userRepo := NewMockUserRepository()
	mcpAuthService := NewMCPAuthService(userRepo)

	t.Run("Valid email", func(t *testing.T) {
		is.True(mcpAuthService.isValidEmail("test@example.com"))
		is.True(mcpAuthService.isValidEmail("user.name+tag@domain.co.uk"))
	})

	t.Run("Invalid email", func(t *testing.T) {
		is.True(!mcpAuthService.isValidEmail("invalid"))
		is.True(!mcpAuthService.isValidEmail("@example.com"))
		is.True(!mcpAuthService.isValidEmail("test@"))
		is.True(!mcpAuthService.isValidEmail(""))
	})
}
