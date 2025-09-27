package user

import (
	"context"
	"testing"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestOrganizationRepository_FindByID(t *testing.T) {
	is := is.New(t)

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:    orgID,
		Title: "Test Organization",
	}

	// Create mock repository
	repo := &MockOrganizationRepository{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Test FindByID
	ctx := context.Background()
	result, err := repo.FindByID(ctx, orgID)

	// Verify result
	is.NoErr(err)
	is.Equal(result.ID, orgID)
	is.Equal(result.Title, "Test Organization")
}

func TestOrganizationRepository_FindByIDNotFound(t *testing.T) {
	is := is.New(t)

	// Create mock repository with no organizations
	repo := &MockOrganizationRepository{
		organizations: make(map[uuid.UUID]*Organization),
	}

	// Test FindByID with non-existent ID
	ctx := context.Background()
	nonExistentID := uuid.New()
	_, err := repo.FindByID(ctx, nonExistentID)

	// Verify error
	is.True(err != nil)
	is.Equal(err, shared.ErrNotFound)
}

func TestOrganizationRepository_Update(t *testing.T) {
	is := is.New(t)

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:    orgID,
		Title: "Old Title",
	}

	// Create mock repository
	repo := &MockOrganizationRepository{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Test Update
	ctx := context.Background()
	updatedOrg := &Organization{
		ID:    orgID,
		Title: "New Title",
	}
	err := repo.Update(ctx, updatedOrg)

	// Verify result
	is.NoErr(err)
	is.Equal(repo.organizations[orgID].Title, "New Title")
}

func TestOrganizationRepository_UpdateNotFound(t *testing.T) {
	is := is.New(t)

	// Create mock repository with no organizations
	repo := &MockOrganizationRepository{
		organizations: make(map[uuid.UUID]*Organization),
	}

	// Test Update with non-existent organization
	ctx := context.Background()
	nonExistentID := uuid.New()
	updatedOrg := &Organization{
		ID:    nonExistentID,
		Title: "New Title",
	}
	err := repo.Update(ctx, updatedOrg)

	// Verify error
	is.True(err != nil)
	is.Equal(err, shared.ErrNotFound)
}

func TestOrganizationRepository_Exists(t *testing.T) {
	is := is.New(t)

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:    orgID,
		Title: "Test Organization",
	}

	// Create mock repository
	repo := &MockOrganizationRepository{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Test Exists with existing organization
	ctx := context.Background()
	exists, err := repo.Exists(ctx, orgID)

	// Verify result
	is.NoErr(err)
	is.True(exists)
}

func TestOrganizationRepository_ExistsNotFound(t *testing.T) {
	is := is.New(t)

	// Create mock repository with no organizations
	repo := &MockOrganizationRepository{
		organizations: make(map[uuid.UUID]*Organization),
	}

	// Test Exists with non-existent organization
	ctx := context.Background()
	nonExistentID := uuid.New()
	exists, err := repo.Exists(ctx, nonExistentID)

	// Verify result
	is.NoErr(err)
	is.True(!exists)
}

func TestOrganizationRepository_FindByTitle(t *testing.T) {
	is := is.New(t)

	// Create test organization
	orgID := uuid.New()
	organization := &Organization{
		ID:    orgID,
		Title: "Test Organization",
	}

	// Create mock repository
	repo := &MockOrganizationRepository{
		organizations: map[uuid.UUID]*Organization{orgID: organization},
	}

	// Test FindByTitle
	ctx := context.Background()
	result, err := repo.FindByTitle(ctx, "Test Organization")

	// Verify result
	is.NoErr(err)
	is.Equal(result.ID, orgID)
	is.Equal(result.Title, "Test Organization")
}

func TestOrganizationRepository_FindByTitleNotFound(t *testing.T) {
	is := is.New(t)

	// Create mock repository with no organizations
	repo := &MockOrganizationRepository{
		organizations: make(map[uuid.UUID]*Organization),
	}

	// Test FindByTitle with non-existent title
	ctx := context.Background()
	_, err := repo.FindByTitle(ctx, "Non-existent Organization")

	// Verify error
	is.True(err != nil)
	is.Equal(err, shared.ErrNotFound)
}
