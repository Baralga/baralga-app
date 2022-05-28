package user

import (
	"context"

	"github.com/baralga/shared"
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
