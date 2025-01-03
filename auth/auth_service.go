package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/baralga/shared"
	"github.com/baralga/user"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	config         *shared.Config
	userRepository user.UserRepository
}

func NewAuthService(config *shared.Config, UserRepository user.UserRepository) *AuthService {
	return &AuthService{
		config:         config,
		userRepository: UserRepository,
	}
}

func (a *AuthService) Authenticate(ctx context.Context, username, password string) (*shared.Principal, error) {
	u, err := a.userRepository.FindUserByUsername(ctx, username)
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	passwdErr := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if passwdErr != nil {
		return nil, errors.New("password invalid")
	}

	roles, err := a.userRepository.FindRolesByUserID(ctx, u.OrganizationID, u.ID)
	if err != nil {
		return nil, err
	}

	principal := mapUserToPrincipal(u, roles)
	return principal, nil
}

func mapUserToPrincipal(user *user.User, roles []string) *shared.Principal {
	principal := &shared.Principal{
		Name:           user.Name,
		Username:       user.Username,
		OrganizationID: user.OrganizationID,
		Roles:          roles,
	}
	if principal.Name == "" {
		principal.Name = user.Username
	}
	return principal
}

func (a *AuthService) AuthenticateTrusted(ctx context.Context, username string) (*shared.Principal, error) {
	u, err := a.userRepository.FindUserByUsername(ctx, username)
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	roles, err := a.userRepository.FindRolesByUserID(ctx, u.OrganizationID, u.ID)
	if err != nil {
		return nil, err
	}

	principal := mapUserToPrincipal(u, roles)
	return principal, nil
}

func (a *AuthService) CreateCookie(tokenAuth *jwtauth.JWTAuth, expiryDuration time.Duration, principal *shared.Principal) http.Cookie {
	claims := mapPrincipalToClaims(principal)
	claims[jwt.ExpirationKey] = jwtauth.ExpireIn(expiryDuration)

	_, tokenString, _ := tokenAuth.Encode(claims)

	return http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(expiryDuration),
		SameSite: http.SameSiteLaxMode,
		Secure:   a.config.IsProduction(),
		Path:     "/",
	}
}

func (a *AuthService) CreateExpiredCookie() http.Cookie {
	return http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now(),
		SameSite: http.SameSiteLaxMode,
		Secure:   a.config.IsProduction(),
		Path:     "/",
	}
}
