package user

import (
	"context"
	"strings"

	"github.com/baralga/shared"
	"github.com/google/uuid"
)

type OrganizationService struct {
	repositoryTxer         shared.RepositoryTxer
	organizationRepository OrganizationRepository
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(repositoryTxer shared.RepositoryTxer, organizationRepository OrganizationRepository) OrganizationService {
	return OrganizationService{
		repositoryTxer:         repositoryTxer,
		organizationRepository: organizationRepository,
	}
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(ctx context.Context, orgID uuid.UUID) (*Organization, error) {
	var org *Organization
	err := s.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			foundOrg, err := s.organizationRepository.FindByID(ctx, orgID)
			if err != nil {
				return err
			}
			org = foundOrg
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// UpdateOrganizationName updates the name of an organization
func (s *OrganizationService) UpdateOrganizationName(ctx context.Context, orgID uuid.UUID, name string) error {
	// Validate input
	if strings.TrimSpace(name) == "" {
		return shared.ErrValidation("Organization name is required")
	}

	if len(name) > 255 {
		return shared.ErrValidation("Organization name must be between 1 and 255 characters")
	}

	// Use transaction for all database operations
	err := s.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			// Check if organization exists
			exists, err := s.organizationRepository.Exists(ctx, orgID)
			if err != nil {
				return err
			}
			if !exists {
				return shared.ErrNotFound
			}

			// Check if name already exists (excluding current organization)
			existingOrg, err := s.organizationRepository.FindByName(ctx, name)
			if err != nil && err != shared.ErrNotFound {
				return err
			}
			if existingOrg != nil && existingOrg.ID != orgID {
				return shared.ErrConflict("Organization name already exists")
			}

			// Get current organization
			organization, err := s.organizationRepository.FindByID(ctx, orgID)
			if err != nil {
				return err
			}

			// Update organization name
			organization.Title = strings.TrimSpace(name)
			return s.organizationRepository.Update(ctx, organization)
		},
	)

	return err
}

// IsUserAdmin checks if a user is an admin of an organization
func (s *OrganizationService) IsUserAdmin(ctx context.Context, orgID uuid.UUID) (bool, error) {
	// Get the principal from context to check roles
	principal := shared.MustPrincipalFromContext(ctx)

	// Check if the user has admin role
	return principal.HasRole("ROLE_ADMIN"), nil
}
