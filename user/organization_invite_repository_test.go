package user

import (
	"context"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestOrganizationInviteRepository(t *testing.T) {
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

	repository := NewDbOrganizationInviteRepository(connPool)
	repositoryTxer := shared.NewDbRepositoryTxer(connPool)

	t.Run("InsertInvite", func(t *testing.T) {
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     shared.OrganizationIDSample,
			Token:     "test-token-123",
			CreatedBy: uuid.MustParse("eeeeeb80-33f3-4d3f-befe-58694d2ac841"), // Use existing user ID
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		err := repositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				_, err := repository.InsertInvite(ctx, invite)
				return err
			},
		)

		is.NoErr(err)
	})

	t.Run("FindInviteByToken", func(t *testing.T) {
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     shared.OrganizationIDSample,
			Token:     "find-test-token",
			CreatedBy: uuid.MustParse("eeeeeb80-33f3-4d3f-befe-58694d2ac841"), // Use existing user ID
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		err := repositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				// Insert invite
				_, err := repository.InsertInvite(ctx, invite)
				if err != nil {
					return err
				}

				// Find invite by token
				foundInvite, err := repository.FindInviteByToken(ctx, invite.Token)
				if err != nil {
					return err
				}

				is.Equal(foundInvite.ID, invite.ID)
				is.Equal(foundInvite.Token, invite.Token)
				is.Equal(foundInvite.OrgID, invite.OrgID)

				return nil
			},
		)

		is.NoErr(err)
	})

	t.Run("FindInviteByTokenNotFound", func(t *testing.T) {
		err := repositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				_, err := repository.FindInviteByToken(ctx, "non-existent-token")
				is.Equal(err, ErrInviteNotFound)
				return nil
			},
		)

		is.NoErr(err)
	})

	t.Run("FindInvitesByOrganizationID", func(t *testing.T) {
		orgID := shared.OrganizationIDSample // Use existing organization ID
		invite1 := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     orgID,
			Token:     "org-invite-1",
			CreatedBy: uuid.MustParse("eeeeeb80-33f3-4d3f-befe-58694d2ac841"), // Use existing user ID
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}
		invite2 := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     orgID,
			Token:     "org-invite-2",
			CreatedBy: uuid.MustParse("eeeeeb80-33f3-4d3f-befe-58694d2ac841"), // Use existing user ID
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		err := repositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				// Insert invites
				_, err := repository.InsertInvite(ctx, invite1)
				if err != nil {
					return err
				}
				_, err = repository.InsertInvite(ctx, invite2)
				if err != nil {
					return err
				}

				// Find invites by organization ID
				invites, err := repository.FindInvitesByOrganizationID(ctx, orgID)
				if err != nil {
					return err
				}

				// Check that we have at least our 2 invites (there might be others from previous tests)
				is.True(len(invites) >= 2)

				// Verify our specific invites are present
				foundTokens := make(map[string]bool)
				for _, invite := range invites {
					foundTokens[invite.Token] = true
				}
				is.True(foundTokens["org-invite-1"])
				is.True(foundTokens["org-invite-2"])

				return nil
			},
		)

		is.NoErr(err)
	})

	t.Run("UpdateInvite", func(t *testing.T) {
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     shared.OrganizationIDSample,
			Token:     "update-test-token",
			CreatedBy: uuid.MustParse("eeeeeb80-33f3-4d3f-befe-58694d2ac841"), // Use existing user ID
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		err := repositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				// Insert invite
				_, err := repository.InsertInvite(ctx, invite)
				if err != nil {
					return err
				}

				// Update invite
				now := time.Now()
				userID := uuid.MustParse("04b4adc8-2b7f-4ec0-aeb8-407ce164484e") // Use existing user ID
				invite.UsedAt = &now
				invite.UsedBy = &userID
				invite.Active = false

				err = repository.UpdateInvite(ctx, invite)
				if err != nil {
					return err
				}

				// Verify update
				foundInvite, err := repository.FindInviteByToken(ctx, invite.Token)
				if err != nil {
					return err
				}

				is.True(foundInvite.UsedAt != nil)
				is.True(foundInvite.UsedBy != nil)
				is.Equal(*foundInvite.UsedBy, userID)
				is.Equal(foundInvite.Active, false)

				return nil
			},
		)

		is.NoErr(err)
	})

}

func TestInMemOrganizationInviteRepository(t *testing.T) {
	is := is.New(t)

	repository := NewInMemOrganizationInviteRepository()
	ctx := context.Background()

	t.Run("InsertInvite", func(t *testing.T) {
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     uuid.New(),
			Token:     "test-token",
			CreatedBy: uuid.New(),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		result, err := repository.InsertInvite(ctx, invite)
		is.NoErr(err)
		is.Equal(result.ID, invite.ID)
	})

	t.Run("FindInviteByToken", func(t *testing.T) {
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     uuid.New(),
			Token:     "find-token",
			CreatedBy: uuid.New(),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		_, err := repository.InsertInvite(ctx, invite)
		is.NoErr(err)

		found, err := repository.FindInviteByToken(ctx, invite.Token)
		is.NoErr(err)
		is.Equal(found.ID, invite.ID)
	})

	t.Run("FindInviteByTokenNotFound", func(t *testing.T) {
		_, err := repository.FindInviteByToken(ctx, "non-existent")
		is.Equal(err, ErrInviteNotFound)
	})

	t.Run("FindInvitesByOrganizationID", func(t *testing.T) {
		orgID := uuid.New()
		invite1 := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     orgID,
			Token:     "org-token-1",
			CreatedBy: uuid.New(),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}
		invite2 := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     orgID,
			Token:     "org-token-2",
			CreatedBy: uuid.New(),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		_, err := repository.InsertInvite(ctx, invite1)
		is.NoErr(err)
		_, err = repository.InsertInvite(ctx, invite2)
		is.NoErr(err)

		invites, err := repository.FindInvitesByOrganizationID(ctx, orgID)
		is.NoErr(err)
		is.Equal(len(invites), 2)
	})

	t.Run("UpdateInvite", func(t *testing.T) {
		invite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     uuid.New(),
			Token:     "update-token",
			CreatedBy: uuid.New(),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Active:    true,
		}

		_, err := repository.InsertInvite(ctx, invite)
		is.NoErr(err)

		now := time.Now()
		userID := uuid.New()
		invite.UsedAt = &now
		invite.UsedBy = &userID
		invite.Active = false

		err = repository.UpdateInvite(ctx, invite)
		is.NoErr(err)

		found, err := repository.FindInviteByToken(ctx, invite.Token)
		is.NoErr(err)
		is.True(found.UsedAt != nil)
		is.True(found.UsedBy != nil)
		is.Equal(found.Active, false)
	})

}
