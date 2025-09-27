package user

import (
	"context"

	"github.com/baralga/shared"
	"github.com/google/uuid"
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

func (r *DbOrganizationRepository) UpdateOrganization(ctx context.Context, organization *Organization) error {
	tx := shared.MustTxFromContext(ctx)

	_, err := tx.Exec(
		ctx,
		`UPDATE organizations 
		   SET title = $1 
		 WHERE org_id = $2`,
		organization.Title,
		organization.ID,
	)
	return err
}

func (r *DbOrganizationRepository) FindOrganizationByID(ctx context.Context, organizationID uuid.UUID) (*Organization, error) {
	tx := shared.MustTxFromContext(ctx)

	var organization Organization
	err := tx.QueryRow(
		ctx,
		`SELECT org_id, title FROM organizations WHERE org_id = $1`,
		organizationID,
	).Scan(&organization.ID, &organization.Title)

	if err != nil {
		return nil, err
	}

	return &organization, nil
}
