package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Organization represents the organizational entity that users belong to
type Organization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// OrganizationService defines the business logic interface for organization management
type OrganizationService interface {
	GetOrganization(ctx context.Context, orgID uuid.UUID) (*Organization, error)
	UpdateOrganizationName(ctx context.Context, orgID uuid.UUID, name string) error
	IsUserAdmin(ctx context.Context, userID uuid.UUID, orgID uuid.UUID) (bool, error)
}

// OrganizationRepository defines the data access interface for organizations
type OrganizationRepository interface {
	FindByID(ctx context.Context, orgID uuid.UUID) (*Organization, error)
	Update(ctx context.Context, organization *Organization) error
	Exists(ctx context.Context, orgID uuid.UUID) (bool, error)
	FindByName(ctx context.Context, name string) (*Organization, error)
}

// ValidateOrganizationName validates the organization name according to business rules
func ValidateOrganizationName(name string) error {
	if name == "" {
		return errors.New("organization name is required")
	}

	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return errors.New("organization name cannot be only whitespace")
	}

	if len(trimmed) > 255 {
		return errors.New("organization name must be 255 characters or less")
	}

	if len(trimmed) < 1 {
		return errors.New("organization name must be at least 1 character")
	}

	return nil
}

// IsValidRole checks if the provided role is valid
func IsValidRole(role string) bool {
	return role == "admin" || role == "user"
}

// UserRole represents the role of a user within an organization
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// String returns the string representation of the role
func (r UserRole) String() string {
	return string(r)
}

// IsAdmin checks if the role is admin
func (r UserRole) IsAdmin() bool {
	return r == RoleAdmin
}
