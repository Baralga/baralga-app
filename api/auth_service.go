package main

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func (a *app) Authenticate(ctx context.Context, username, password string) (*Principal, error) {
	user, err := a.UserRepository.FindUserByUsername(ctx, username)
	if errors.Is(err, ErrUserNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	passwdErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if passwdErr != nil {
		return nil, errors.New("password invalid")
	}

	roles, err := a.UserRepository.FindRolesByUserID(ctx, user.OrganizationID, user.ID)
	if err != nil {
		return nil, err
	}

	principal := &Principal{
		Username:       user.Username,
		OrganizationID: user.OrganizationID,
		Roles:          roles,
	}

	return principal, nil
}
