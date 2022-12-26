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

// DbUserRepository is a SQL database repository for users
type DbUserRepository struct {
	connPool *pgxpool.Pool
}

var _ UserRepository = (*DbUserRepository)(nil)

// NewDbUserRepository creates a new SQL database repository for users
func NewDbUserRepository(connPool *pgxpool.Pool) *DbUserRepository {
	return &DbUserRepository{
		connPool: connPool,
	}
}

func (r *DbUserRepository) insertConfirmation(ctx context.Context, tx pgx.Tx, user *User, confirmationID uuid.UUID) (uuid.UUID, error) {
	_, err := tx.Exec(
		ctx,
		`INSERT INTO user_confirmations 
		   (user_confirmation_id, user_id, created_at) 
		 VALUES 
		   ($1, $2, $3)`,
		confirmationID,
		user.ID,
		time.Now(),
	)
	return confirmationID, err
}

func (r *DbUserRepository) InsertUserWithConfirmationID(ctx context.Context, user *User, confirmationID uuid.UUID) (*User, error) {
	tx := ctx.Value(shared.ContextKeyTx).(pgx.Tx)

	enabled := 0
	if confirmationID == uuid.Nil {
		enabled = 1
	}

	_, err := tx.Exec(
		ctx,
		`INSERT INTO users 
		   (user_id, username, email, name, password, enabled, org_id, origin) 
		 VALUES 
		   ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID,
		user.Username,
		user.EMail,
		user.Name,
		user.Password,
		enabled,
		user.OrganizationID,
		user.Origin,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO roles 
		   (user_id, role, org_id) 
		 VALUES 
		   ($1, 'ROLE_ADMIN', $2)`,
		user.ID,
		user.OrganizationID,
	)
	if err != nil {
		return nil, err
	}

	if confirmationID == uuid.Nil {
		return user, nil
	}

	_, err = r.insertConfirmation(ctx, tx, user, confirmationID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *DbUserRepository) FindUserIDByConfirmationID(ctx context.Context, confirmationID string) (uuid.UUID, error) {
	row := r.connPool.QueryRow(
		ctx,
		`SELECT user_id 
		 FROM user_confirmations 
		 WHERE user_confirmation_id = $1`, confirmationID,
	)

	var (
		userID string
	)

	err := row.Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrUserNotFound
		}

		return uuid.Nil, err
	}

	return uuid.MustParse(userID), nil
}

func (r *DbUserRepository) ConfirmUser(ctx context.Context, userID uuid.UUID) error {
	tx := ctx.Value(shared.ContextKeyTx).(pgx.Tx)

	_, err := tx.Exec(
		ctx,
		`DELETE FROM user_confirmations 
		 WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE users
		 SET enabled = 1 
		 WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *DbUserRepository) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	row := r.connPool.QueryRow(
		ctx,
		`SELECT user_id, name, password, org_id 
		 FROM users 
		 WHERE username = $1 AND enabled = 1`, username,
	)

	var (
		id             string
		name           string
		password       string
		organizationID string
	)

	err := row.Scan(&id, &name, &password, &organizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	user := &User{
		ID:             uuid.MustParse(id),
		Name:           name,
		Username:       username,
		Password:       password,
		OrganizationID: uuid.MustParse(organizationID),
	}
	return user, nil
}

func (r *DbUserRepository) FindRolesByUserID(ctx context.Context, organizationID, userID uuid.UUID) ([]string, error) {
	rows, err := r.connPool.Query(
		ctx,
		`SELECT role 
		 FROM roles 
		 WHERE user_id = $1 AND org_id = $2`, userID, organizationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string

		err = rows.Scan(&role)
		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
}
