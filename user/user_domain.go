package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID             uuid.UUID
	Name           string
	Username       string
	EMail          string
	Password       string
	Origin         string
	OrganizationID uuid.UUID
}

type Organization struct {
	ID    uuid.UUID
	Title string
}

type UserRepository interface {
	ConfirmUser(ctx context.Context, userID uuid.UUID) error
	FindUserIDByConfirmationID(ctx context.Context, confirmationID string) (uuid.UUID, error)
	InsertUserWithConfirmationID(ctx context.Context, user *User, confirmationID uuid.UUID) (*User, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	FindRolesByUserID(ctx context.Context, organizationID, userID uuid.UUID) ([]string, error)
}

type OrganizationRepository interface {
	InsertOrganization(ctx context.Context, organization *Organization) (*Organization, error)
}