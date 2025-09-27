package user

import (
	"context"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestOrganizationInviteService(t *testing.T) {
	is := is.New(t)

	// Setup test dependencies
	inviteRepository := NewInMemOrganizationInviteRepository()
	repositoryTxer := shared.NewInMemRepositoryTxer()

	service := &OrganizationInviteService{
		repositoryTxer:   repositoryTxer,
		inviteRepository: inviteRepository,
	}

	ctx := context.Background()

	t.Run("GenerateInvite", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		invite, err := service.GenerateInvite(ctx, orgID, createdBy)

		is.NoErr(err)
		is.True(invite.ID != uuid.Nil)
		is.Equal(invite.OrgID, orgID)
		is.Equal(invite.CreatedBy, createdBy)
		is.True(invite.Token != "")
		is.True(invite.Active)
		is.True(invite.UsedAt == nil)
		is.True(invite.UsedBy == nil)

		// Check expiration is 24 hours from now
		expectedExpiry := time.Now().Add(24 * time.Hour)
		timeDiff := invite.ExpiresAt.Sub(expectedExpiry)
		is.True(timeDiff < time.Minute) // Allow 1 minute tolerance
		is.True(timeDiff > -time.Minute)
	})

	t.Run("GenerateInviteWithUniqueTokens", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		invite1, err1 := service.GenerateInvite(ctx, orgID, createdBy)
		invite2, err2 := service.GenerateInvite(ctx, orgID, createdBy)

		is.NoErr(err1)
		is.NoErr(err2)
		is.True(invite1.Token != invite2.Token)
		is.True(invite1.ID != invite2.ID)
	})

	t.Run("ValidateInvite", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		// Generate invite
		invite, err := service.GenerateInvite(ctx, orgID, createdBy)
		is.NoErr(err)

		// Validate invite
		validatedInvite, err := service.ValidateInvite(ctx, invite.Token)

		is.NoErr(err)
		is.Equal(validatedInvite.ID, invite.ID)
		is.Equal(validatedInvite.Token, invite.Token)
		is.Equal(validatedInvite.OrgID, orgID)
	})

	t.Run("ValidateInviteNotFound", func(t *testing.T) {
		_, err := service.ValidateInvite(ctx, "non-existent-token")
		is.Equal(err, ErrInviteNotFound)
	})

	t.Run("ValidateInviteExpired", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		// Create expired invite manually
		expiredInvite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     orgID,
			Token:     "expired-token",
			CreatedBy: createdBy,
			CreatedAt: time.Now().Add(-25 * time.Hour), // 25 hours ago
			ExpiresAt: time.Now().Add(-1 * time.Hour),  // 1 hour ago
			Active:    true,
		}

		_, err := inviteRepository.InsertInvite(ctx, expiredInvite)
		is.NoErr(err)

		// Try to validate expired invite
		_, err = service.ValidateInvite(ctx, expiredInvite.Token)
		is.Equal(err, ErrInviteExpired)
	})

	t.Run("ValidateInviteAlreadyUsed", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		// Create used invite manually
		now := time.Now()
		usedBy := uuid.New()
		usedInvite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     orgID,
			Token:     "used-token",
			CreatedBy: createdBy,
			CreatedAt: time.Now().Add(-1 * time.Hour),
			ExpiresAt: time.Now().Add(23 * time.Hour),
			UsedAt:    &now,
			UsedBy:    &usedBy,
			Active:    false,
		}

		_, err := inviteRepository.InsertInvite(ctx, usedInvite)
		is.NoErr(err)

		// Try to validate used invite
		_, err = service.ValidateInvite(ctx, usedInvite.Token)
		is.Equal(err, ErrInviteAlreadyUsed)
	})

	t.Run("UseInvite", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		// Generate invite
		invite, err := service.GenerateInvite(ctx, orgID, createdBy)
		is.NoErr(err)

		// Use invite
		userID := uuid.New()
		err = service.UseInvite(ctx, invite.Token, userID)

		is.NoErr(err)

		// Verify invite is marked as used
		usedInvite, err := inviteRepository.FindInviteByToken(ctx, invite.Token)
		is.NoErr(err)
		is.True(usedInvite.UsedAt != nil)
		is.True(usedInvite.UsedBy != nil)
		is.Equal(*usedInvite.UsedBy, userID)
		is.Equal(usedInvite.Active, false)
	})

	t.Run("UseInviteNotFound", func(t *testing.T) {
		err := service.UseInvite(ctx, "non-existent-token", uuid.New())
		is.Equal(err, ErrInviteNotFound)
	})

	t.Run("UseInviteExpired", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		// Create expired invite manually
		expiredInvite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     orgID,
			Token:     "expired-token-2",
			CreatedBy: createdBy,
			CreatedAt: time.Now().Add(-25 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Active:    true,
		}

		_, err := inviteRepository.InsertInvite(ctx, expiredInvite)
		is.NoErr(err)

		// Try to use expired invite
		err = service.UseInvite(ctx, expiredInvite.Token, uuid.New())
		is.Equal(err, ErrInviteExpired)
	})

	t.Run("UseInviteAlreadyUsed", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		// Create used invite manually
		now := time.Now()
		usedBy := uuid.New()
		usedInvite := &OrganizationInvite{
			ID:        uuid.New(),
			OrgID:     orgID,
			Token:     "used-token-2",
			CreatedBy: createdBy,
			CreatedAt: time.Now().Add(-1 * time.Hour),
			ExpiresAt: time.Now().Add(23 * time.Hour),
			UsedAt:    &now,
			UsedBy:    &usedBy,
			Active:    false,
		}

		_, err := inviteRepository.InsertInvite(ctx, usedInvite)
		is.NoErr(err)

		// Try to use already used invite
		err = service.UseInvite(ctx, usedInvite.Token, uuid.New())
		is.Equal(err, ErrInviteAlreadyUsed)
	})

	t.Run("FindInvitesByOrganization", func(t *testing.T) {
		orgID := uuid.New()
		createdBy := uuid.New()

		// Generate multiple invites
		invite1, err := service.GenerateInvite(ctx, orgID, createdBy)
		is.NoErr(err)
		invite2, err := service.GenerateInvite(ctx, orgID, createdBy)
		is.NoErr(err)

		// Find invites for organization
		invites, err := service.FindInvitesByOrganization(ctx, orgID)

		is.NoErr(err)
		is.True(len(invites) >= 2)

		// Verify our invites are present
		foundTokens := make(map[string]bool)
		for _, invite := range invites {
			foundTokens[invite.Token] = true
		}
		is.True(foundTokens[invite1.Token])
		is.True(foundTokens[invite2.Token])
	})
}
