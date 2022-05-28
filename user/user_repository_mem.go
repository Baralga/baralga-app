package user

import (
	"context"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

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
				OrganizationID: shared.OrganizationIDSample,
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

func (r *InMemUserRepository) InsertUserWithConfirmationID(ctx context.Context, user *User, confirmationID uuid.UUID) (*User, error) {
	if confirmationID == shared.ConfirmationIDError {
		return nil, errors.New("error for tests")
	}
	r.users = append(r.users, user)
	return user, nil
}

func (r *InMemUserRepository) FindUserIDByConfirmationID(ctx context.Context, confirmationID string) (uuid.UUID, error) {
	if confirmationID == shared.ConfirmationIdSample.String() {
		return r.users[0].ID, nil
	}
	return uuid.Nil, ErrUserNotFound
}

func (r *InMemUserRepository) ConfirmUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}
