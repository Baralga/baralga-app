package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestHandleOrganizationManagementPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:   orgID,
		Name: "Test Organization",
	}

	// Create test user with admin role

	// Create organization service mock
	orgService := &MockOrganizationService{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Create organization web handlers
	config := &shared.Config{}
	webHandlers := &OrganizationWebHandlers{
		config:              config,
		organizationService: orgService,
	}

	// Create request with user context
	r, _ := http.NewRequest("GET", "/profile/organization", nil)
	principal := &shared.Principal{
		Name:           "Test User",
		Username:       "testuser",
		OrganizationID: orgID,
		Roles:          []string{"admin"},
	}
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), principal))

	// Create router and register handler
	router := chi.NewRouter()
	router.Get("/profile/organization", webHandlers.HandleOrganizationManagementPage())

	// Execute request
	router.ServeHTTP(httpRec, r)

	// Verify response
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Organization Settings"))
	is.True(strings.Contains(htmlBody, "Test Organization"))
}

func TestHandleOrganizationManagementPageNonAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	// Create test user without admin role
	orgID := uuid.New()

	// Create organization service mock
	orgService := &MockOrganizationService{}

	// Create organization web handlers
	config := &shared.Config{}
	webHandlers := &OrganizationWebHandlers{
		config:              config,
		organizationService: orgService,
	}

	// Create request with user context (non-admin)
	r, _ := http.NewRequest("GET", "/profile/organization", nil)
	principal := &shared.Principal{
		Name:           "Test User",
		Username:       "testuser",
		OrganizationID: orgID,
		Roles:          []string{"user"}, // Non-admin role
	}
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), principal))

	// Create router and register handler
	router := chi.NewRouter()
	router.Get("/profile/organization", webHandlers.HandleOrganizationManagementPage())

	// Execute request
	router.ServeHTTP(httpRec, r)

	// Verify response
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Access denied"))
}

func TestHandleOrganizationTitleUpdate(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:   orgID,
		Name: "Old Organization Name",
	}

	// Create test user with admin role

	// Create organization service mock
	orgService := &MockOrganizationService{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Create organization web handlers
	config := &shared.Config{}
	webHandlers := &OrganizationWebHandlers{
		config:              config,
		organizationService: orgService,
	}

	// Create form data
	data := url.Values{}
	data["title"] = []string{"New Organization Name"}

	// Create request with user context
	r, _ := http.NewRequest("POST", "/profile/organization", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create router and register handler
	router := chi.NewRouter()
	router.Post("/profile/organization", webHandlers.HandleOrganizationTitleUpdate())

	// Execute request
	router.ServeHTTP(httpRec, r)

	// Verify response
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Organization name updated successfully"))
}

func TestHandleOrganizationTitleUpdateValidationError(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	// Create test user with admin role

	// Create organization service mock
	orgService := &MockOrganizationService{}

	// Create organization web handlers
	config := &shared.Config{}
	webHandlers := &OrganizationWebHandlers{
		config:              config,
		organizationService: orgService,
	}

	// Create form data with empty title
	data := url.Values{}
	data["title"] = []string{""}

	// Create request with user context
	r, _ := http.NewRequest("POST", "/profile/organization", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create router and register handler
	router := chi.NewRouter()
	router.Post("/profile/organization", webHandlers.HandleOrganizationTitleUpdate())

	// Execute request
	router.ServeHTTP(httpRec, r)

	// Verify response
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Organization name is required"))
}

// MockOrganizationService for testing
type MockOrganizationService struct {
	organizations map[uuid.UUID]*Organization
	updateError   error
}

func (m *MockOrganizationService) GetOrganization(ctx context.Context, orgID uuid.UUID) (*Organization, error) {
	if org, exists := m.organizations[orgID]; exists {
		return org, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockOrganizationService) UpdateOrganizationName(ctx context.Context, orgID uuid.UUID, name string) error {
	if m.updateError != nil {
		return m.updateError
	}
	if org, exists := m.organizations[orgID]; exists {
		org.Name = name
		return nil
	}
	return shared.ErrNotFound
}

func (m *MockOrganizationService) IsUserAdmin(ctx context.Context, orgID uuid.UUID) (bool, error) {
	// This mock will use the Principal from context
	principal := shared.MustPrincipalFromContext(ctx)
	return principal.HasRole("admin"), nil
}
