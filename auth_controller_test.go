package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/jwtauth/v5"
	"github.com/matryer/is"
)

func TestHandleLogin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config: &config{
			JWTExpiry: "1h",
		},
		UserRepository: NewInMemUserRepository(),
	}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	body := `
	{
		"username": "admin@baralga.com",
		"password": "adm1n"
	 }
	`

	r, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))

	a.HandleLogin(tokenAuth)(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	loginResponse := make(map[string]string)
	err := json.NewDecoder(httpRec.Body).Decode(&loginResponse)
	is.NoErr(err)
	is.True(len(loginResponse["access_token"]) > 10)
}

func TestHandleInvalidLogin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:         &config{},
		UserRepository: NewInMemUserRepository(),
	}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	body := `
	{
		"username": "admin",
		"password": "-invalid-"
	 }
	`

	r, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))

	a.HandleLogin(tokenAuth)(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
}

func TestHandleLoginWithInvalidDuration(t *testing.T) {
	a := &app{
		Config: &config{
			JWTExpiry: "invalid",
		},
		UserRepository: NewInMemUserRepository(),
	}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	a.HandleLogin(tokenAuth)
}

func TestMapPrincipalFromClaims(t *testing.T) {
	is := is.New(t)

	name := "Ado Admin"
	username := "admin"

	claims := make(map[string]interface{})
	claims["name"] = name
	claims["username"] = username
	claims["organizationId"] = organizationIDSample.String()
	claims["roles"] = "ROLE_ADMIN"

	p := mapPrincipalFromClaims(claims)

	is.Equal(name, p.Name)
	is.Equal(username, p.Username)
	is.Equal(organizationIDSample, p.OrganizationID)
	is.Equal(1, len(p.Roles))
	is.Equal("ROLE_ADMIN", p.Roles[0])
}

func TestJWTPrincipalHandlerWithoutJWT(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config: &config{},
	}

	r, _ := http.NewRequest("GET", "/api/projects", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.JWTPrincipalHandler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusIMUsed)
	})).ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusUnauthorized)
}
