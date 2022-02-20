package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestOrganizationRepository(t *testing.T) {
	// skip in short mode
	if testing.Short() {
		return
	}

	is := is.New(t)

	// Setup database
	ctx := context.Background()
	dbContainer, connPool, err := setupDatabase(ctx)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := dbContainer.Terminate(ctx)
		if err != nil {
			t.Log(err)
		}
	}()

	organizationRepository := NewDbOrganizationRepository(connPool)
	repositoryTxer := NewDbRepositoryTxer(connPool)

	t.Run("InsertOrganization", func(t *testing.T) {
		organization := &Organization{
			ID:    uuid.New(),
			Title: "My Test Organization" + time.Now().String(),
		}

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := organizationRepository.InsertOrganization(
					ctx,
					organization,
				)
				return err
			},
		)
		is.NoErr(err)
	})
}

type InMemOrganizationRepository struct {
	organizations []*Organization
}

var _ OrganizationRepository = (*InMemOrganizationRepository)(nil)

func NewInMemOrganizationRepository() *InMemOrganizationRepository {
	return &InMemOrganizationRepository{
		organizations: []*Organization{
			{
				ID:    organizationIDSample,
				Title: "Test Organization",
			},
		},
	}
}

func (r *InMemOrganizationRepository) InsertOrganization(ctx context.Context, organization *Organization) (*Organization, error) {
	r.organizations = append(r.organizations, organization)
	return organization, nil
}
