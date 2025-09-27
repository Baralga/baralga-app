package user

import (
	"context"
	"errors"

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

func (r *InMemOrganizationRepository) UpdateOrganization(ctx context.Context, organization *Organization) error {
	for i, org := range r.organizations {
		if org.ID == organization.ID {
			r.organizations[i].Title = organization.Title
			return nil
		}
	}
	return errors.New("organization not found")
}

func (r *InMemOrganizationRepository) FindOrganizationByID(ctx context.Context, organizationID uuid.UUID) (*Organization, error) {
	for _, org := range r.organizations {
		if org.ID == organizationID {
			return org, nil
		}
	}
	return nil, errors.New("organization not found")
}
