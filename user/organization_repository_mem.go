package user

import (
	"context"

	"github.com/baralga/shared"
	"github.com/google/uuid"
)

type InMemOrganizationRepository struct {
	organizations []*Organization
}

var _ OrganizationRepository = (*InMemOrganizationRepository)(nil)

func NewInMemOrganizationRepository() *InMemOrganizationRepository {
	return &InMemOrganizationRepository{
		organizations: []*Organization{
			{
				ID:    shared.OrganizationIDSample,
				Title: "Test Organization",
			},
		},
	}
}

func (r *InMemOrganizationRepository) InsertOrganization(ctx context.Context, organization *Organization) (*Organization, error) {
	r.organizations = append(r.organizations, organization)
	return organization, nil
}

// FindByID retrieves an organization by ID
func (r *InMemOrganizationRepository) FindByID(ctx context.Context, orgID uuid.UUID) (*Organization, error) {
	for _, org := range r.organizations {
		if org.ID == orgID {
			return org, nil
		}
	}
	return nil, shared.ErrNotFound
}

// Update updates an organization
func (r *InMemOrganizationRepository) Update(ctx context.Context, organization *Organization) error {
	for i, org := range r.organizations {
		if org.ID == organization.ID {
			r.organizations[i] = organization
			return nil
		}
	}
	return shared.ErrNotFound
}

// Exists checks if an organization exists
func (r *InMemOrganizationRepository) Exists(ctx context.Context, orgID uuid.UUID) (bool, error) {
	for _, org := range r.organizations {
		if org.ID == orgID {
			return true, nil
		}
	}
	return false, nil
}

// FindByName retrieves an organization by name
func (r *InMemOrganizationRepository) FindByName(ctx context.Context, name string) (*Organization, error) {
	for _, org := range r.organizations {
		if org.Title == name {
			return org, nil
		}
	}
	return nil, shared.ErrNotFound
}
