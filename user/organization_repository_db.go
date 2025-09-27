package user

import (
	"context"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DbOrganizationRepository is a SQL database repository for users
type DbOrganizationRepository struct {
	connPool *pgxpool.Pool
}

var _ OrganizationRepository = (*DbOrganizationRepository)(nil)

// NewDbOrganizationRepository creates a new SQL database repository for users
func NewDbOrganizationRepository(connPool *pgxpool.Pool) *DbOrganizationRepository {
	return &DbOrganizationRepository{
		connPool: connPool,
	}
}

func (r *DbOrganizationRepository) InsertOrganization(ctx context.Context, organization *Organization) (*Organization, error) {
	tx := shared.MustTxFromContext(ctx)

	_, err := tx.Exec(
		ctx,
		`INSERT INTO organizations 
		   (org_id, title) 
		 VALUES 
		   ($1, $2)`,
		organization.ID,
		organization.Title,
	)
	return organization, err
}

// FindByID retrieves an organization by ID
func (r *DbOrganizationRepository) FindByID(ctx context.Context, orgID uuid.UUID) (*Organization, error) {
	tx := shared.MustTxFromContext(ctx)

	var organization Organization
	err := tx.QueryRow(
		ctx,
		`SELECT org_id, title FROM organizations WHERE org_id = $1`,
		orgID,
	).Scan(&organization.ID, &organization.Title)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	return &organization, nil
}

// Update updates an organization
func (r *DbOrganizationRepository) Update(ctx context.Context, organization *Organization) error {
	tx := shared.MustTxFromContext(ctx)

	_, err := tx.Exec(
		ctx,
		`UPDATE organizations SET title = $1 WHERE org_id = $2`,
		organization.Title,
		organization.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// Exists checks if an organization exists
func (r *DbOrganizationRepository) Exists(ctx context.Context, orgID uuid.UUID) (bool, error) {
	tx := shared.MustTxFromContext(ctx)

	var count int
	err := tx.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM organizations WHERE org_id = $1`,
		orgID,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// FindByName retrieves an organization by name
func (r *DbOrganizationRepository) FindByName(ctx context.Context, name string) (*Organization, error) {
	tx := shared.MustTxFromContext(ctx)

	var organization Organization
	err := tx.QueryRow(
		ctx,
		`SELECT org_id, title FROM organizations WHERE title = $1`,
		name,
	).Scan(&organization.ID, &organization.Title)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	return &organization, nil
}
