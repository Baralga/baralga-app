package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/baralga/util"
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

// HandleLogin handles the authentication request of a user
func (a *app) HandleLogin(tokenAuth *jwtauth.JWTAuth) http.HandlerFunc {
	expiryDuration := a.Config.ExpiryDuration()
	return func(w http.ResponseWriter, r *http.Request) {
		var loginModel loginModel
		err := json.NewDecoder(r.Body).Decode(&loginModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		principal, err := a.Authenticate(r.Context(), loginModel.Username, loginModel.Password)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusForbidden)
			return
		}

		cookie := a.CreateCookie(tokenAuth, expiryDuration, principal)
		http.SetCookie(w, &cookie)

		loginResponseModel := &loginResponseModel{AccessToken: cookie.Value}
		util.RenderJSON(w, loginResponseModel)
	}
}

// JWTPrincipalHandler sets up the user principal from the JWT
func (a *app) JWTPrincipalHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, claims, _ := jwtauth.FromContext(r.Context())
			if token == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			principal := mapPrincipalFromClaims(claims)
			ctx := context.WithValue(r.Context(), contextKeyPrincipal, principal)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func mapPrincipalToClaims(principal *Principal) map[string]interface{} {
	return map[string]interface{}{
		"name":           principal.Name,
		"username":       principal.Username,
		"organizationId": principal.OrganizationID.String(),
		"roles":          strings.Join(principal.Roles, ","),
	}
}

func mapPrincipalFromClaims(claims map[string]interface{}) *Principal {
	return &Principal{
		Name:           claims["name"].(string),
		Username:       claims["username"].(string),
		OrganizationID: uuid.MustParse(claims["organizationId"].(string)),
		Roles:          strings.Split(claims["roles"].(string), ","),
	}
}
