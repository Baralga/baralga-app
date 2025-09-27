package user

import (
	"context"
	"testing"
	"time"

	"github.com/baralga/shared"
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
	cleanupFunc, connPool, err := shared.SetupTestDatabase(ctx)
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err := cleanupFunc()
		if err != nil {
			t.Log(err)
		}
	}()

	organizationRepository := NewDbOrganizationRepository(connPool)
	repositoryTxer := shared.NewDbRepositoryTxer(connPool)

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

	t.Run("UpdateOrganization", func(t *testing.T) {
		organization := &Organization{
			ID:    shared.OrganizationIDSample,
			Title: "Updated Test Organization" + time.Now().String(),
		}

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				err := organizationRepository.UpdateOrganization(
					ctx,
					organization,
				)
				return err
			},
		)
		is.NoErr(err)
	})
}
