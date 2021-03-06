package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/baralga/shared"
	"github.com/baralga/shared/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"schneider.vip/problem"
)

type loginModel struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponseModel struct {
	AccessToken string `json:"access_token"`
}

type AuthController struct {
	app         *shared.App
	authService *AuthService
	tokenAuth   *jwtauth.JWTAuth
}

func NewAuthController(app *shared.App, authService *AuthService, tokenAuth *jwtauth.JWTAuth) *AuthController {
	return &AuthController{
		app:         app,
		authService: authService,
		tokenAuth:   tokenAuth,
	}
}

func (a *AuthController) RegisterProtected(r chi.Router) {
}

func (a *AuthController) RegisterOpen(r chi.Router) {
	r.Post("/auth/login", a.HandleLogin())
}

// HandleLogin handles the authentication request of a user
func (a *AuthController) HandleLogin() http.HandlerFunc {
	tokenAuth := a.tokenAuth
	expiryDuration := a.app.Config.ExpiryDuration()
	authService := a.authService
	return func(w http.ResponseWriter, r *http.Request) {
		var loginModel loginModel
		err := json.NewDecoder(r.Body).Decode(&loginModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		principal, err := authService.Authenticate(r.Context(), loginModel.Username, loginModel.Password)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusForbidden)
			return
		}

		cookie := authService.CreateCookie(tokenAuth, expiryDuration, principal)
		http.SetCookie(w, &cookie)

		loginResponseModel := &loginResponseModel{AccessToken: cookie.Value}
		util.RenderJSON(w, loginResponseModel)
	}
}

func (a *AuthController) JWTVerifier() func(next http.Handler) http.Handler {
	return jwtauth.Verifier(a.tokenAuth)
}

// JWTPrincipalHandler sets up the user principal from the JWT
func (a *AuthController) JWTPrincipalHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, claims, _ := jwtauth.FromContext(r.Context())
			if token == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			principal := mapPrincipalFromClaims(claims)
			ctx := context.WithValue(r.Context(), shared.ContextKeyPrincipal, principal)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func mapPrincipalToClaims(principal *shared.Principal) map[string]interface{} {
	return map[string]interface{}{
		"name":           principal.Name,
		"username":       principal.Username,
		"organizationId": principal.OrganizationID.String(),
		"roles":          strings.Join(principal.Roles, ","),
	}
}

func mapPrincipalFromClaims(claims map[string]interface{}) *shared.Principal {
	return &shared.Principal{
		Name:           claims["name"].(string),
		Username:       claims["username"].(string),
		OrganizationID: uuid.MustParse(claims["organizationId"].(string)),
		Roles:          strings.Split(claims["roles"].(string), ","),
	}
}
