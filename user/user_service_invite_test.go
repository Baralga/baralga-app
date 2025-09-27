package user

import (
	"context"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestSetUpNewUserWithInvite(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	userRepository := NewInMemUserRepository()
	organizationRepository := NewInMemOrganizationRepository()
	inviteRepository := NewInMemOrganizationInviteRepository()

	// Create an existing organization and invite
	existingOrg := &Organization{
		ID:    uuid.New(),
		Title: "Existing Organization",
	}
	_, err := organizationRepository.InsertOrganization(context.Background(), existingOrg)
	is.NoErr(err)

	invite := &OrganizationInvite{
		ID:        uuid.New(),
		OrgID:     existingOrg.ID,
		Token:     "valid-invite-token",
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Active:    true,
	}
	_, err = inviteRepository.InsertInvite(context.Background(), invite)
	is.NoErr(err)

	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		inviteService:          NewOrganizationInviteService(shared.NewInMemRepositoryTxer(), inviteRepository),
		projectService:         &mockProjectService{},
	}

	user := &User{
		Name:     "Invited User",
		Username: "invited@example.com",
		EMail:    "invited@example.com",
		Password: "securePassword123",
		Origin:   "baralga",
	}

	// Act
	err = userService.SetUpNewUserWithInvite(context.Background(), user, invite.Token)

	// Assert
	is.NoErr(err)

	// Verify user was created with correct organization
	createdUser, err := userRepository.FindUserByUsername(context.Background(), user.EMail)
	is.NoErr(err)
	is.Equal(createdUser.OrganizationID, existingOrg.ID)
	is.Equal(createdUser.Name, user.Name)
	is.Equal(createdUser.EMail, user.EMail)

	// Verify invite was marked as used
	usedInvite, err := inviteRepository.FindInviteByToken(context.Background(), invite.Token)
	is.NoErr(err)
	is.True(usedInvite.UsedAt != nil)
	is.True(usedInvite.UsedBy != nil)
	is.Equal(*usedInvite.UsedBy, createdUser.ID)
}

func TestSetUpNewUserWithInvalidInvite(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	userRepository := NewInMemUserRepository()
	organizationRepository := NewInMemOrganizationRepository()
	inviteRepository := NewInMemOrganizationInviteRepository()

	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		inviteService:          NewOrganizationInviteService(shared.NewInMemRepositoryTxer(), inviteRepository),
		projectService:         &mockProjectService{},
	}

	user := &User{
		Name:     "Test User",
		Username: "test@example.com",
		EMail:    "test@example.com",
		Password: "securePassword123",
		Origin:   "baralga",
	}

	// Act
	err := userService.SetUpNewUserWithInvite(context.Background(), user, "invalid-token")

	// Assert
	is.True(err != nil)
	is.Equal(err, ErrInviteNotFound)
}

func TestSetUpNewUserWithExpiredInvite(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	userRepository := NewInMemUserRepository()
	organizationRepository := NewInMemOrganizationRepository()
	inviteRepository := NewInMemOrganizationInviteRepository()

	// Create an existing organization and expired invite
	existingOrg := &Organization{
		ID:    uuid.New(),
		Title: "Existing Organization",
	}
	_, err := organizationRepository.InsertOrganization(context.Background(), existingOrg)
	is.NoErr(err)

	expiredInvite := &OrganizationInvite{
		ID:        uuid.New(),
		OrgID:     existingOrg.ID,
		Token:     "expired-invite-token",
		CreatedBy: uuid.New(),
		CreatedAt: time.Now().Add(-48 * time.Hour), // 48 hours ago
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired 24 hours ago
		Active:    true,
	}
	_, err = inviteRepository.InsertInvite(context.Background(), expiredInvite)
	is.NoErr(err)

	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		inviteService:          NewOrganizationInviteService(shared.NewInMemRepositoryTxer(), inviteRepository),
		projectService:         &mockProjectService{},
	}

	user := &User{
		Name:     "Test User",
		Username: "test@example.com",
		EMail:    "test@example.com",
		Password: "securePassword123",
		Origin:   "baralga",
	}

	// Act
	err = userService.SetUpNewUserWithInvite(context.Background(), user, expiredInvite.Token)

	// Assert
	is.True(err != nil)
	is.Equal(err, ErrInviteExpired)
}

func TestSetUpNewUserWithUsedInvite(t *testing.T) {
	// Arrange
	is := is.New(t)
	mailResource := shared.NewInMemMailResource()
	userRepository := NewInMemUserRepository()
	organizationRepository := NewInMemOrganizationRepository()
	inviteRepository := NewInMemOrganizationInviteRepository()

	// Create an existing organization and used invite
	existingOrg := &Organization{
		ID:    uuid.New(),
		Title: "Existing Organization",
	}
	_, err := organizationRepository.InsertOrganization(context.Background(), existingOrg)
	is.NoErr(err)

	usedInvite := &OrganizationInvite{
		ID:        uuid.New(),
		OrgID:     existingOrg.ID,
		Token:     "used-invite-token",
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Active:    true,
		UsedAt:    &[]time.Time{time.Now()}[0],
		UsedBy:    &[]uuid.UUID{uuid.New()}[0],
	}
	_, err = inviteRepository.InsertInvite(context.Background(), usedInvite)
	is.NoErr(err)

	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		mailResource:           mailResource,
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		inviteService:          NewOrganizationInviteService(shared.NewInMemRepositoryTxer(), inviteRepository),
		projectService:         &mockProjectService{},
	}

	user := &User{
		Name:     "Test User",
		Username: "test@example.com",
		EMail:    "test@example.com",
		Password: "securePassword123",
		Origin:   "baralga",
	}

	// Act
	err = userService.SetUpNewUserWithInvite(context.Background(), user, usedInvite.Token)

	// Assert
	is.True(err != nil)
	is.Equal(err, ErrInviteAlreadyUsed)
}
