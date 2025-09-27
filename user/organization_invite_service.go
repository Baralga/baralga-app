package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type OrganizationInviteService struct {
	repositoryTxer   shared.RepositoryTxer
	inviteRepository OrganizationInviteRepository
}

func NewOrganizationInviteService(
	repositoryTxer shared.RepositoryTxer,
	inviteRepository OrganizationInviteRepository,
) *OrganizationInviteService {
	return &OrganizationInviteService{
		repositoryTxer:   repositoryTxer,
		inviteRepository: inviteRepository,
	}
}

// GenerateInvite creates a new organization invite with a secure token
func (s *OrganizationInviteService) GenerateInvite(ctx context.Context, orgID, createdBy uuid.UUID) (*OrganizationInvite, error) {
	// Generate secure token
	token, err := s.generateSecureToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate secure token")
	}

	now := time.Now()
	invite := &OrganizationInvite{
		ID:        uuid.New(),
		OrgID:     orgID,
		Token:     token,
		CreatedBy: createdBy,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour), // 24 hour expiration
		Active:    true,
	}

	var result *OrganizationInvite
	err = s.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.inviteRepository.InsertInvite(ctx, invite)
			return err
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create organization invite")
	}

	return result, nil
}

// ValidateInvite validates an invite token and returns the invite if valid
func (s *OrganizationInviteService) ValidateInvite(ctx context.Context, token string) (*OrganizationInvite, error) {
	var invite *OrganizationInvite
	err := s.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			var err error
			invite, err = s.inviteRepository.FindInviteByToken(ctx, token)
			return err
		},
	)

	if err != nil {
		if err == ErrInviteNotFound {
			return nil, ErrInviteNotFound
		}
		return nil, errors.Wrap(err, "failed to find organization invite")
	}

	// Check if invite is expired
	if time.Now().After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}

	// Check if invite is already used
	if invite.UsedAt != nil {
		return nil, ErrInviteAlreadyUsed
	}

	// Check if invite is active
	if !invite.Active {
		return nil, ErrInviteAlreadyUsed
	}

	return invite, nil
}

// UseInvite marks an invite as used by a specific user
func (s *OrganizationInviteService) UseInvite(ctx context.Context, token string, userID uuid.UUID) error {
	// First validate the invite
	invite, err := s.ValidateInvite(ctx, token)
	if err != nil {
		return err
	}

	// Mark invite as used
	now := time.Now()
	invite.UsedAt = &now
	invite.UsedBy = &userID
	invite.Active = false

	err = s.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			return s.inviteRepository.UpdateInvite(ctx, invite)
		},
	)

	if err != nil {
		return errors.Wrap(err, "failed to mark invite as used")
	}

	return nil
}

// FindInvitesByOrganization returns all invites for an organization
func (s *OrganizationInviteService) FindInvitesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*OrganizationInvite, error) {
	var invites []*OrganizationInvite
	err := s.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			var err error
			invites, err = s.inviteRepository.FindInvitesByOrganizationID(ctx, orgID)
			return err
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to find organization invites")
	}

	return invites, nil
}

// generateSecureToken creates a cryptographically secure random token
func (s *OrganizationInviteService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
