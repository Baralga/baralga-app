package main

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestUserRepository(t *testing.T) {
	// skip in short mode
	if testing.Short() {
		return
	}

	is := is.New(t)

	// Setup database
	ctx := context.Background()
	dbContainer, connPool, err := setupDatabase(ctx)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := dbContainer.Terminate(ctx)
		if err != nil {
			t.Log(err)
		}
	}()

	userRepository := NewDbUserRepository(connPool)

	t.Run("FindExistingUserByUsername", func(t *testing.T) {
		adminUser, err := userRepository.FindUserByUsername(
			context.Background(),
			"admin@baralga.com",
		)

		is.NoErr(err)
		is.Equal(adminUser.Username, "admin@baralga.com")
	})

	t.Run("FindNotExistingUserByUsername", func(t *testing.T) {
		_, err := userRepository.FindUserByUsername(
			context.Background(),
			"-not here-",
		)

		is.True(errors.Is(err, ErrUserNotFound))
	})

	t.Run("FindRolesByExistingUserID", func(t *testing.T) {
		roles, err := userRepository.FindRolesByUserID(
			context.Background(),
			organizationIDSample,
			userIDAdminSample,
		)

		is.NoErr(err)
		is.Equal(len(roles), 1)
		is.Equal(roles[0], "ROLE_ADMIN")
	})

	t.Run("FindRolesByMissingUserID", func(t *testing.T) {
		roles, err := userRepository.FindRolesByUserID(
			context.Background(),
			organizationIDSample,
			uuid.MustParse("efa45cae-5dc7-412a-887f-945ddbb0a23f"),
		)

		is.NoErr(err)
		is.Equal(len(roles), 0)
	})

	t.Run("InsertUserWithOrganizationAndConfirmation", func(t *testing.T) {
		user := &User{
			ID:             uuid.New(),
			Name:           "Ned Newbie",
			Username:       "ned.newbie@baralga.com",
			EMail:          "ned.newbie@baralga.com",
			OrganizationID: uuid.New(),
		}
		confirmationID := uuid.New()

		_, err := userRepository.InsertUserWithOrganizationAndConfirmation(
			context.Background(),
			user,
			confirmationID,
		)
		is.NoErr(err)

		userIdByConfId, err := userRepository.FindUserIDByConfirmationID(
			context.Background(),
			confirmationID.String(),
		)
		is.NoErr(err)
		is.Equal(user.ID, userIdByConfId)

		err = userRepository.ConfirmUser(
			context.Background(),
			user.ID,
		)
		is.NoErr(err)
	})
}

type InMemUserRepository struct {
	users []*User
}

var _ UserRepository = (*InMemUserRepository)(nil)

func NewInMemUserRepository() *InMemUserRepository {
	return &InMemUserRepository{
		users: []*User{
			{
				ID:             uuid.MustParse("00000000-0000-0000-1111-000000000001"),
				Username:       "admin@baralga.com",
				EMail:          "admin@baralga.com",
				Password:       "$2a$10$NuzYobDOSTCx/EKBClGwGe0A9c8/yC7D4IP75hwz1jn.RCBfdEtb2",
				OrganizationID: organizationIDSample,
			},
		},
	}
}

func (r *InMemUserRepository) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	for _, a := range r.users {
		if a.Username == username {
			return a, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *InMemUserRepository) FindRolesByUserID(ctx context.Context, organizationID, userID uuid.UUID) ([]string, error) {
	return []string{"ROLE_ADMIN"}, nil
}

func (r *InMemUserRepository) InsertUserWithOrganizationAndConfirmation(ctx context.Context, user *User, confirmationID uuid.UUID) (*User, error) {
	r.users = append(r.users, user)
	return user, nil
}

func (r *InMemUserRepository) FindUserIDByConfirmationID(ctx context.Context, confirmationID string) (uuid.UUID, error) {
	if confirmationID == confirmationIdSample.String() {
		return r.users[0].ID, nil
	}
	return uuid.Nil, ErrUserNotFound
}

func (r *InMemUserRepository) ConfirmUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}
