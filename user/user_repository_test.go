package user

import (
	"context"
	"errors"
	"testing"

	"github.com/baralga/shared"
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
	cleanupFunc, connPool, err := shared.SetupTestDatabase(ctx)
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err := cleanupFunc()
		if err != nil {
			t.Log(err)
		}
	}()

	userRepository := NewDbUserRepository(connPool)
	repositoryTxer := shared.NewDbRepositoryTxer(connPool)

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
			shared.OrganizationIDSample,
			shared.UserIDAdminSample,
		)

		is.NoErr(err)
		is.Equal(len(roles), 1)
		is.Equal(roles[0], "ROLE_ADMIN")
	})

	t.Run("FindRolesByMissingUserID", func(t *testing.T) {
		roles, err := userRepository.FindRolesByUserID(
			context.Background(),
			shared.OrganizationIDSample,
			uuid.MustParse("efa45cae-5dc7-412a-887f-945ddbb0a23f"),
		)

		is.NoErr(err)
		is.Equal(len(roles), 0)
	})

	t.Run("InsertUserWithConfirmationID", func(t *testing.T) {
		user := &User{
			ID:             uuid.New(),
			Name:           "Ned Newbie",
			Username:       "ned.newbie@baralga.com",
			EMail:          "ned.newbie@baralga.com",
			OrganizationID: shared.OrganizationIDSample,
			Origin:         "baralga",
		}
		confirmationID := uuid.New()

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := userRepository.InsertUserWithConfirmationID(
					ctx,
					user,
					confirmationID,
				)
				return err
			},
		)
		is.NoErr(err)

		userIdByConfId, err := userRepository.FindUserIDByConfirmationID(
			context.Background(),
			confirmationID.String(),
		)
		is.NoErr(err)
		is.Equal(user.ID, userIdByConfId)

		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				return userRepository.ConfirmUser(
					ctx,
					user.ID,
				)
			},
		)

		is.NoErr(err)
	})

	t.Run("InsertUserWithoutConfirmation", func(t *testing.T) {
		user := &User{
			ID:             uuid.New(),
			Name:           "Minny Manners",
			Username:       "minny.manners@baralga.com",
			EMail:          "minny.manners@baralga.com",
			OrganizationID: shared.OrganizationIDSample,
			Origin:         "github",
		}
		confirmationID := uuid.Nil

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := userRepository.InsertUserWithConfirmationID(
					ctx,
					user,
					confirmationID,
				)
				return err
			},
		)
		is.NoErr(err)

		_, err = userRepository.FindUserIDByConfirmationID(
			context.Background(),
			confirmationID.String(),
		)
		is.True(errors.Is(err, ErrUserNotFound))
	})
}
