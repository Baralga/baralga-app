package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/baralga/util"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
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
	expiryDuration, err := time.ParseDuration("1d")
	if err != nil {
		log.Printf("could not parse jwt expiry %v", a.Config.JWTExpiry)
		expiryDuration = time.Duration(24 * time.Hour)
	}

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

		claims := mapPrincipalToClaims(principal)
		claims[jwt.ExpirationKey] = expiryDuration

		_, tokenString, _ := tokenAuth.Encode(claims)
		loginResponseModel := &loginResponseModel{AccessToken: tokenString}

		cookie := http.Cookie{
			Name:     "jwt",
			Value:    tokenString,
			Expires:  time.Now().Add(expiryDuration),
			SameSite: http.SameSiteLaxMode,
			Secure:   true,
			Path:     "/",
		}
		http.SetCookie(w, &cookie)

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
		"username":       principal.Username,
		"organizationId": principal.OrganizationID.String(),
		"roles":          strings.Join(principal.Roles, ","),
	}
}

func mapPrincipalFromClaims(claims map[string]interface{}) *Principal {
	return &Principal{
		Username:       claims["username"].(string),
		OrganizationID: uuid.MustParse(claims["organizationId"].(string)),
		Roles:          strings.Split(claims["roles"].(string), ","),
	}
}
