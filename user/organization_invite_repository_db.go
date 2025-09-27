package user

import (
	"context"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// DbOrganizationInviteRepository is a SQL database repository for organization invites
type DbOrganizationInviteRepository struct {
	connPool *pgxpool.Pool
}

var _ OrganizationInviteRepository = (*DbOrganizationInviteRepository)(nil)

// NewDbOrganizationInviteRepository creates a new SQL database repository for organization invites
func NewDbOrganizationInviteRepository(connPool *pgxpool.Pool) *DbOrganizationInviteRepository {
	return &DbOrganizationInviteRepository{
		connPool: connPool,
	}
}

func (r *DbOrganizationInviteRepository) InsertInvite(ctx context.Context, invite *OrganizationInvite) (*OrganizationInvite, error) {
	tx := shared.MustTxFromContext(ctx)

	_, err := tx.Exec(
		ctx,
		`INSERT INTO organization_invites 
		   (invite_id, org_id, token, created_by, created_at, expires_at, used_at, used_by, active) 
		 VALUES 
		   ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		invite.ID,
		invite.OrgID,
		invite.Token,
		invite.CreatedBy,
		invite.CreatedAt,
		invite.ExpiresAt,
		invite.UsedAt,
		invite.UsedBy,
		invite.Active,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert organization invite")
	}

	return invite, nil
}

func (r *DbOrganizationInviteRepository) FindInviteByToken(ctx context.Context, token string) (*OrganizationInvite, error) {
	tx := shared.MustTxFromContext(ctx)

	var invite OrganizationInvite
	var usedAt *time.Time
	var usedBy *uuid.UUID

	err := tx.QueryRow(
		ctx,
		`SELECT invite_id, org_id, token, created_by, created_at, expires_at, used_at, used_by, active
		 FROM organization_invites 
		 WHERE token = $1`,
		token,
	).Scan(
		&invite.ID,
		&invite.OrgID,
		&invite.Token,
		&invite.CreatedBy,
		&invite.CreatedAt,
		&invite.ExpiresAt,
		&usedAt,
		&usedBy,
		&invite.Active,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInviteNotFound
		}
		return nil, errors.Wrap(err, "failed to find organization invite by token")
	}

	invite.UsedAt = usedAt
	invite.UsedBy = usedBy

	return &invite, nil
}

func (r *DbOrganizationInviteRepository) FindInvitesByOrganizationID(ctx context.Context, organizationID uuid.UUID) ([]*OrganizationInvite, error) {
	tx := shared.MustTxFromContext(ctx)

	rows, err := tx.Query(
		ctx,
		`SELECT invite_id, org_id, token, created_by, created_at, expires_at, used_at, used_by, active
		 FROM organization_invites 
		 WHERE org_id = $1
		 ORDER BY created_at DESC`,
		organizationID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query organization invites")
	}
	defer rows.Close()

	var invites []*OrganizationInvite
	for rows.Next() {
		var invite OrganizationInvite
		var usedAt *time.Time
		var usedBy *uuid.UUID

		err := rows.Scan(
			&invite.ID,
			&invite.OrgID,
			&invite.Token,
			&invite.CreatedBy,
			&invite.CreatedAt,
			&invite.ExpiresAt,
			&usedAt,
			&usedBy,
			&invite.Active,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan organization invite")
		}

		invite.UsedAt = usedAt
		invite.UsedBy = usedBy
		invites = append(invites, &invite)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating organization invites")
	}

	return invites, nil
}

func (r *DbOrganizationInviteRepository) UpdateInvite(ctx context.Context, invite *OrganizationInvite) error {
	tx := shared.MustTxFromContext(ctx)

	_, err := tx.Exec(
		ctx,
		`UPDATE organization_invites 
		 SET used_at = $2, used_by = $3, active = $4
		 WHERE invite_id = $1`,
		invite.ID,
		invite.UsedAt,
		invite.UsedBy,
		invite.Active,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update organization invite")
	}

	return nil
}

func (r *DbOrganizationInviteRepository) DeleteInvite(ctx context.Context, inviteID uuid.UUID) error {
	tx := shared.MustTxFromContext(ctx)

	_, err := tx.Exec(
		ctx,
		`DELETE FROM organization_invites WHERE invite_id = $1`,
		inviteID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete organization invite")
	}

	return nil
}
