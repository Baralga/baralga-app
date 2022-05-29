package user

import (
	"context"

	"github.com/baralga/shared"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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
	tx := ctx.Value(shared.ContextKeyTx).(pgx.Tx)

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
