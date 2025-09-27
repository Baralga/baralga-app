package user

import (
	"context"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestOrganizationService_GetOrganization(t *testing.T) {
	is := is.New(t)

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:   orgID,
		Name: "Test Organization",
	}

	// Create mock repository
	repo := &MockOrganizationRepository{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Create organization service
	service := &OrganizationService{
		organizationRepository: repo,
	}

	// Test GetOrganization
	ctx := context.Background()
	result, err := service.GetOrganization(ctx, orgID)

	// Verify result
	is.NoErr(err)
	is.Equal(result.ID, orgID)
	is.Equal(result.Name, "Test Organization")
}

func TestOrganizationService_UpdateOrganizationName(t *testing.T) {
	is := is.New(t)

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:   orgID,
		Name: "Old Title",
	}

	// Create mock repository
	repo := &MockOrganizationRepository{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Create organization service
	service := &OrganizationService{
		organizationRepository: repo,
	}

	// Test UpdateOrganizationName
	ctx := context.Background()
	err := service.UpdateOrganizationName(ctx, orgID, "New Title")

	// Verify result
	is.NoErr(err)
	is.Equal(organization.Name, "New Title")
}

func TestOrganizationService_UpdateOrganizationNameValidationError(t *testing.T) {
	is := is.New(t)

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:   orgID,
		Name: "Old Title",
	}

	// Create mock repository
	repo := &MockOrganizationRepository{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Create organization service
	service := &OrganizationService{
		organizationRepository: repo,
	}

	// Test UpdateOrganizationName with empty title
	ctx := context.Background()
	err := service.UpdateOrganizationName(ctx, orgID, "")

	// Verify error
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "Organization name is required"))
}

func TestOrganizationService_UpdateOrganizationNameDuplicateError(t *testing.T) {
	is := is.New(t)

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:   orgID,
		Name: "Old Title",
	}

	// Create mock repository with existing organization with same title
	existingOrgID := uuid.New()
	existingOrg := &Organization{
		ID:   existingOrgID,
		Name: "Duplicate Title",
	}

	repo := &MockOrganizationRepository{
		organizations: map[uuid.UUID]*Organization{
			orgID:         organization,
			existingOrgID: existingOrg,
		},
	}

	// Create organization service
	service := &OrganizationService{
		organizationRepository: repo,
	}

	// Test UpdateOrganizationName with duplicate title
	ctx := context.Background()
	err := service.UpdateOrganizationName(ctx, orgID, "Duplicate Title")

	// Verify error
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "Organization name already exists"))
}

func TestOrganizationService_IsUserAdmin(t *testing.T) {
	is := is.New(t)

	// Create test data
	orgID := uuid.New()

	// Create mock repository
	repo := &MockOrganizationRepository{}

	// Create organization service
	service := &OrganizationService{
		organizationRepository: repo,
	}

	// Create context with admin principal
	principal := &shared.Principal{
		Name:           "Test User",
		Username:       "testuser",
		OrganizationID: orgID,
		Roles:          []string{"admin"},
	}
	ctx := shared.ToContextWithPrincipal(context.Background(), principal)

	// Test IsUserAdmin
	isAdmin, err := service.IsUserAdmin(ctx, orgID)

	// Verify result
	is.NoErr(err)
	is.True(isAdmin)
}

func TestOrganizationService_IsUserAdminNonAdmin(t *testing.T) {
	is := is.New(t)

	// Create test data
	orgID := uuid.New()

	// Create mock repository
	repo := &MockOrganizationRepository{}

	// Create organization service
	service := &OrganizationService{
		organizationRepository: repo,
	}

	// Create context with non-admin principal
	principal := &shared.Principal{
		Name:           "Test User",
		Username:       "testuser",
		OrganizationID: orgID,
		Roles:          []string{"user"},
	}
	ctx := shared.ToContextWithPrincipal(context.Background(), principal)

	// Test IsUserAdmin
	isAdmin, err := service.IsUserAdmin(ctx, orgID)

	// Verify result
	is.NoErr(err)
	is.True(!isAdmin)
}

// MockOrganizationRepository for testing
type MockOrganizationRepository struct {
	organizations map[uuid.UUID]*Organization
	updateError   error
}

func (m *MockOrganizationRepository) InsertOrganization(ctx context.Context, organization *Organization) (*Organization, error) {
	if m.organizations == nil {
		m.organizations = make(map[uuid.UUID]*Organization)
	}
	m.organizations[organization.ID] = organization
	return organization, nil
}

func (m *MockOrganizationRepository) FindByID(ctx context.Context, orgID uuid.UUID) (*Organization, error) {
	if org, exists := m.organizations[orgID]; exists {
		return org, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockOrganizationRepository) Update(ctx context.Context, organization *Organization) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, exists := m.organizations[organization.ID]; exists {
		m.organizations[organization.ID] = organization
		return nil
	}
	return shared.ErrNotFound
}

func (m *MockOrganizationRepository) Exists(ctx context.Context, orgID uuid.UUID) (bool, error) {
	_, exists := m.organizations[orgID]
	return exists, nil
}

func (m *MockOrganizationRepository) FindByName(ctx context.Context, name string) (*Organization, error) {
	for _, org := range m.organizations {
		if org.Name == name {
			return org, nil
		}
	}
	return nil, shared.ErrNotFound
}
