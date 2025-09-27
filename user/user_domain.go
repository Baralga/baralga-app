package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInviteNotFound = errors.New("invite not found")
var ErrInviteExpired = errors.New("invite expired")
var ErrInviteAlreadyUsed = errors.New("invite already used")

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

type OrganizationInvite struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	Token     string
	CreatedBy uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
	UsedAt    *time.Time
	UsedBy    *uuid.UUID
	Active    bool
}

type UserRepository interface {
	ConfirmUser(ctx context.Context, userID uuid.UUID) error
	FindUserIDByConfirmationID(ctx context.Context, confirmationID string) (uuid.UUID, error)
	InsertUserWithConfirmationID(ctx context.Context, user *User, confirmationID uuid.UUID) (*User, error)
	InsertUserWithConfirmationIDAndRole(ctx context.Context, user *User, confirmationID uuid.UUID, role string) (*User, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	FindRolesByUserID(ctx context.Context, organizationID, userID uuid.UUID) ([]string, error)
}

type OrganizationRepository interface {
	InsertOrganization(ctx context.Context, organization *Organization) (*Organization, error)
	UpdateOrganization(ctx context.Context, organization *Organization) error
	FindOrganizationByID(ctx context.Context, organizationID uuid.UUID) (*Organization, error)
}

type OrganizationInviteRepository interface {
	InsertInvite(ctx context.Context, invite *OrganizationInvite) (*OrganizationInvite, error)
	FindInviteByToken(ctx context.Context, token string) (*OrganizationInvite, error)
	FindInvitesByOrganizationID(ctx context.Context, organizationID uuid.UUID) ([]*OrganizationInvite, error)
	UpdateInvite(ctx context.Context, invite *OrganizationInvite) error
}
