package main

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID             uuid.UUID
	Username       string
	Password       string
	OrganizationID uuid.UUID
}

type UserRepository interface {
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	FindRolesByUserID(ctx context.Context, organizationID, userID uuid.UUID) ([]string, error)
}

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

func (r *DbUserRepository) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	row := r.connPool.QueryRow(
		ctx,
		`SELECT user_id, password, org_id 
		 FROM users 
		 WHERE username = $1 AND enabled = 1`, username,
	)

	var (
		id             string
		password       string
		organizationID string
	)

	err := row.Scan(&id, &password, &organizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	user := &User{
		ID:             uuid.MustParse(id),
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
