package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestOrganizationInviteWebHandlers_HandleInviteList(t *testing.T) {
	is := is.New(t)

	t.Run("AdminUser", func(t *testing.T) {
		// Setup
		config := &shared.Config{Webroot: "https://example.com"}
		userRepository := NewInMemUserRepository()
		userService := &UserService{
			userRepository: userRepository,
			inviteService: &OrganizationInviteService{
				inviteRepository: NewInMemOrganizationInviteRepository(),
				repositoryTxer:   shared.NewInMemRepositoryTxer(),
			},
		}
		handlers := NewOrganizationInviteWeb(config, userService)

		// Create test principal with admin role
		principal := &shared.Principal{
			Username:       "admin@example.com",
			OrganizationID: uuid.New(),
			Roles:          []string{"ROLE_ADMIN"},
		}

		// Add user to repository
		user := &User{
			ID:       uuid.New(),
			Name:     "Admin User",
			Username: "admin@example.com",
		}
		userRepository.InsertUserWithConfirmationID(context.Background(), user, uuid.New())

		// Create request with principal in context
		req := httptest.NewRequest("GET", "/organization/invites", nil)
		req = req.WithContext(shared.ToContextWithPrincipal(req.Context(), principal))

		// Create response recorder
		w := httptest.NewRecorder()

		// Execute
		handlers.HandleInviteList()(w, req)

		// Verify
		is.Equal(w.Code, http.StatusOK)
		body := w.Body.String()
		is.True(strings.Contains(body, "Organization Invites"))
		is.True(strings.Contains(body, "Generate Invite"))
	})

	t.Run("NonAdminUser", func(t *testing.T) {
		// Setup
		config := &shared.Config{Webroot: "https://example.com"}
		userRepository := NewInMemUserRepository()
		userService := &UserService{
			userRepository: userRepository,
			inviteService: &OrganizationInviteService{
				inviteRepository: NewInMemOrganizationInviteRepository(),
				repositoryTxer:   shared.NewInMemRepositoryTxer(),
			},
		}
		handlers := NewOrganizationInviteWeb(config, userService)

		// Create test principal without admin role
		principal := &shared.Principal{
			Username:       "user@example.com",
			OrganizationID: uuid.New(),
			Roles:          []string{"ROLE_USER"},
		}

		// Add user to repository
		user := &User{
			ID:       uuid.New(),
			Name:     "Regular User",
			Username: "user@example.com",
		}
		userRepository.InsertUserWithConfirmationID(context.Background(), user, uuid.New())

		// Create request with principal in context
		req := httptest.NewRequest("GET", "/organization/invites", nil)
		req = req.WithContext(shared.ToContextWithPrincipal(req.Context(), principal))

		// Create response recorder
		w := httptest.NewRecorder()

		// Execute
		handlers.HandleInviteList()(w, req)

		// Verify
		is.Equal(w.Code, http.StatusOK)
		body := w.Body.String()
		is.True(strings.Contains(body, "Organization Invites"))
		is.True(!strings.Contains(body, "Generate Invite"))
	})
}

func TestOrganizationInviteWebHandlers_HandleGenerateInvite(t *testing.T) {
	is := is.New(t)

	t.Run("AdminUser", func(t *testing.T) {
		// Setup
		config := &shared.Config{Webroot: "https://example.com"}

		// Create mock user service
		userRepository := NewInMemUserRepository()
		userService := &UserService{
			userRepository: userRepository,
			inviteService: &OrganizationInviteService{
				inviteRepository: NewInMemOrganizationInviteRepository(),
				repositoryTxer:   shared.NewInMemRepositoryTxer(),
			},
		}
		handlers := NewOrganizationInviteWeb(config, userService)

		// Create test principal with admin role
		principal := &shared.Principal{
			Username:       "admin@example.com",
			OrganizationID: uuid.New(),
			Roles:          []string{"ROLE_ADMIN"},
		}

		// Add user to repository
		user := &User{
			ID:       uuid.New(),
			Name:     "Admin User",
			Username: "admin@example.com",
		}
		userRepository.InsertUserWithConfirmationID(context.Background(), user, uuid.New())

		// Create request with principal in context
		req := httptest.NewRequest("POST", "/organization/invites/generate", nil)
		req = req.WithContext(shared.ToContextWithPrincipal(req.Context(), principal))

		// Create response recorder
		w := httptest.NewRecorder()

		// Execute
		handlers.HandleGenerateInvite()(w, req)

		// Verify
		is.Equal(w.Code, http.StatusOK)
		body := w.Body.String()
		is.True(strings.Contains(body, "Organization Invites"))
	})

	t.Run("NonAdminUser", func(t *testing.T) {
		// Setup
		config := &shared.Config{Webroot: "https://example.com"}

		// Create mock user service
		userRepository := NewInMemUserRepository()
		userService := &UserService{
			userRepository: userRepository,
			inviteService: &OrganizationInviteService{
				inviteRepository: NewInMemOrganizationInviteRepository(),
				repositoryTxer:   shared.NewInMemRepositoryTxer(),
			},
		}
		handlers := NewOrganizationInviteWeb(config, userService)

		// Create test principal without admin role
		principal := &shared.Principal{
			Username:       "user@example.com",
			OrganizationID: uuid.New(),
			Roles:          []string{"ROLE_USER"},
		}

		// Add user to repository
		user := &User{
			ID:       uuid.New(),
			Name:     "Regular User",
			Username: "user@example.com",
		}
		userRepository.InsertUserWithConfirmationID(context.Background(), user, uuid.New())

		// Create request with principal in context
		req := httptest.NewRequest("POST", "/organization/invites/generate", nil)
		req = req.WithContext(shared.ToContextWithPrincipal(req.Context(), principal))

		// Create response recorder
		w := httptest.NewRecorder()

		// Execute
		handlers.HandleGenerateInvite()(w, req)

		// Verify
		is.Equal(w.Code, http.StatusForbidden)
	})
}

func TestOrganizationInviteWebHandlers_InviteCard(t *testing.T) {
	is := is.New(t)

	config := &shared.Config{Webroot: "https://example.com"}
	handlers := NewOrganizationInviteWeb(config, &UserService{})

	t.Run("ActiveInvite", func(t *testing.T) {
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			Token:     "test-token",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		card := handlers.InviteCard(invite)
		w := httptest.NewRecorder()
		shared.RenderHTML(w, card)
		html := w.Body.String()

		is.True(strings.Contains(html, "Active"))
		is.True(strings.Contains(html, "bg-primary"))
		is.True(strings.Contains(html, "test-token"))
	})

	t.Run("UsedInvite", func(t *testing.T) {
		usedAt := time.Now().Add(-1 * time.Hour)
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			Token:     "test-token",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			UsedAt:    &usedAt,
			Active:    true,
		}

		card := handlers.InviteCard(invite)
		w := httptest.NewRecorder()
		shared.RenderHTML(w, card)
		html := w.Body.String()

		is.True(strings.Contains(html, "Used"))
		is.True(strings.Contains(html, "bg-success"))
		is.True(strings.Contains(html, "Used on"))
	})

	t.Run("ExpiredInvite", func(t *testing.T) {
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			Token:     "test-token",
			CreatedAt: time.Now().Add(-25 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Active:    true,
		}

		card := handlers.InviteCard(invite)
		w := httptest.NewRecorder()
		shared.RenderHTML(w, card)
		html := w.Body.String()

		is.True(strings.Contains(html, "Expired"))
		is.True(strings.Contains(html, "bg-danger"))
		is.True(strings.Contains(html, "Expired on"))
	})
}

func TestOrganizationInviteWebHandlers_GetInviteStatus(t *testing.T) {
	is := is.New(t)

	handlers := NewOrganizationInviteWeb(&shared.Config{}, &UserService{})

	t.Run("ActiveInvite", func(t *testing.T) {
		invite := &OrganizationInvite{
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		status := handlers.getInviteStatus(invite)
		is.Equal(status, "Active")
	})

	t.Run("UsedInvite", func(t *testing.T) {
		usedAt := time.Now()
		invite := &OrganizationInvite{
			UsedAt:    &usedAt,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		status := handlers.getInviteStatus(invite)
		is.Equal(status, "Used")
	})

	t.Run("ExpiredInvite", func(t *testing.T) {
		invite := &OrganizationInvite{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		status := handlers.getInviteStatus(invite)
		is.Equal(status, "Expired")
	})
}

func TestOrganizationInviteWebHandlers_GetStatusClass(t *testing.T) {
	is := is.New(t)

	handlers := NewOrganizationInviteWeb(&shared.Config{}, &UserService{})

	is.Equal(handlers.getStatusClass("Used"), "bg-success")
	is.Equal(handlers.getStatusClass("Expired"), "bg-danger")
	is.Equal(handlers.getStatusClass("Active"), "bg-primary")
	is.Equal(handlers.getStatusClass("Unknown"), "bg-secondary")
}

func TestOrganizationInviteWebHandlers_GetInviteURL(t *testing.T) {
	is := is.New(t)

	config := &shared.Config{Webroot: "https://example.com"}
	handlers := NewOrganizationInviteWeb(config, &UserService{})

	url := handlers.getInviteURL("test-token")
	expected := "https://example.com/signup/invite/test-token"
	is.Equal(url, expected)
}
