package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func (a *app) EncryptPassword(password string) string {
	encryptedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(encryptedPassword)
}

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

func (a *app) CreateCookie(tokenAuth *jwtauth.JWTAuth, expiryDuration time.Duration, principal *Principal) http.Cookie {
	claims := mapPrincipalToClaims(principal)
	claims[jwt.ExpirationKey] = expiryDuration

	_, tokenString, _ := tokenAuth.Encode(claims)

	return http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(expiryDuration),
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		Path:     "/",
	}
}

func (a *app) DeleteCookie() http.Cookie {
	return http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now(),
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		Path:     "/",
	}
}
