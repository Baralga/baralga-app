package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestHandleInviteSignUpPage(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	userRepository := NewInMemUserRepository()
	organizationRepository := NewInMemOrganizationRepository()
	inviteRepository := NewInMemOrganizationInviteRepository()

	// Create an existing organization and invite
	existingOrg := &Organization{
		ID:    uuid.New(),
		Title: "Existing Organization",
	}
	_, err := organizationRepository.InsertOrganization(context.Background(), existingOrg)
	is.NoErr(err)

	invite := &OrganizationInvite{
		ID:        uuid.New(),
		OrgID:     existingOrg.ID,
		Token:     "valid-invite-token",
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Active:    true,
	}
	_, err = inviteRepository.InsertInvite(context.Background(), invite)
	is.NoErr(err)

	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		inviteService:          NewOrganizationInviteService(shared.NewInMemRepositoryTxer(), inviteRepository),
		projectService:         &mockProjectService{},
	}

	config := &shared.Config{}
	w := &UserWebHandlers{
		config:         config,
		userService:    userService,
		userRepository: userRepository,
	}

	// Create request with valid token
	req := httptest.NewRequest("GET", "/signup/invite/valid-invite-token", nil)
	wRec := httptest.NewRecorder()

	// Set up chi router context
	r := chi.NewRouter()
	r.Get("/signup/invite/{token}", w.HandleInviteSignUpPage())
	r.ServeHTTP(wRec, req)

	// Assert
	is.Equal(wRec.Code, http.StatusOK)
	body := wRec.Body.String()
	is.True(len(body) > 0)
	is.True(strings.Contains(body, "Join Organization"))
	is.True(strings.Contains(body, "valid-invite-token"))
}

func TestHandleInviteSignUpPageWithInvalidToken(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	userRepository := NewInMemUserRepository()
	organizationRepository := NewInMemOrganizationRepository()
	inviteRepository := NewInMemOrganizationInviteRepository()

	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		inviteService:          NewOrganizationInviteService(shared.NewInMemRepositoryTxer(), inviteRepository),
		projectService:         &mockProjectService{},
	}

	config := &shared.Config{}
	w := &UserWebHandlers{
		config:         config,
		userService:    userService,
		userRepository: userRepository,
	}

	// Create request with invalid token
	req := httptest.NewRequest("GET", "/signup/invite/invalid-token", nil)
	wRec := httptest.NewRecorder()

	// Set up chi router context
	r := chi.NewRouter()
	r.Get("/signup/invite/{token}", w.HandleInviteSignUpPage())
	r.ServeHTTP(wRec, req)

	// Assert
	is.Equal(wRec.Code, http.StatusOK)
	body := wRec.Body.String()
	is.True(len(body) > 0)
	is.True(strings.Contains(body, "Invalid or expired invite link"))
}

func TestHandleInviteSignUpPageWithExpiredToken(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	userRepository := NewInMemUserRepository()
	organizationRepository := NewInMemOrganizationRepository()
	inviteRepository := NewInMemOrganizationInviteRepository()

	// Create an existing organization and expired invite
	existingOrg := &Organization{
		ID:    uuid.New(),
		Title: "Existing Organization",
	}
	_, err := organizationRepository.InsertOrganization(context.Background(), existingOrg)
	is.NoErr(err)

	expiredInvite := &OrganizationInvite{
		ID:        uuid.New(),
		OrgID:     existingOrg.ID,
		Token:     "expired-invite-token",
		CreatedBy: uuid.New(),
		CreatedAt: time.Now().Add(-48 * time.Hour), // 48 hours ago
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired 24 hours ago
		Active:    true,
	}
	_, err = inviteRepository.InsertInvite(context.Background(), expiredInvite)
	is.NoErr(err)

	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		inviteService:          NewOrganizationInviteService(shared.NewInMemRepositoryTxer(), inviteRepository),
		projectService:         &mockProjectService{},
	}

	config := &shared.Config{}
	w := &UserWebHandlers{
		config:         config,
		userService:    userService,
		userRepository: userRepository,
	}

	// Create request with expired token
	req := httptest.NewRequest("GET", "/signup/invite/expired-invite-token", nil)
	wRec := httptest.NewRecorder()

	// Set up chi router context
	r := chi.NewRouter()
	r.Get("/signup/invite/{token}", w.HandleInviteSignUpPage())
	r.ServeHTTP(wRec, req)

	// Assert
	is.Equal(wRec.Code, http.StatusOK)
	body := wRec.Body.String()
	is.True(len(body) > 0)
	is.True(strings.Contains(body, "Invalid or expired invite link"))
}
