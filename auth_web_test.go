package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/jwtauth/v5"
	"github.com/matryer/is"
)

func TestHandleLoginPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config: &config{},
	}

	r, _ := http.NewRequest("GET", "/login", nil)

	a.HandleLoginPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Sign In # Baralga"))
}

func TestHandleLoginFormWithSuccessfullLogin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:         &config{},
		UserRepository: NewInMemUserRepository(),
	}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	data := url.Values{}
	data["EMail"] = []string{"admin@baralga.com"}
	data["Password"] = []string{"adm1n"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm(tokenAuth)(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusFound)
}

func TestHandleLoginFormWithInvalidLogin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:         &config{},
		UserRepository: NewInMemUserRepository(),
	}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	data := url.Values{}
	data["EMail"] = []string{"admin@baralga.com"}
	data["Password"] = []string{"-just-wrong-"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm(tokenAuth)(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Sign In # Baralga"))
}

func TestHandleLoginFormWithInvalidFormData(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:         &config{},
		UserRepository: NewInMemUserRepository(),
	}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	data := url.Values{}
	data["NotMatching"] = []string{"admin"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm(tokenAuth)(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Sign In # Baralga"))
}

func TestHandleLoginFormWithInvalidBodyData(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:         &config{},
		UserRepository: NewInMemUserRepository(),
	}
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	data := url.Values{}
	data["NotMatching"] = []string{"admin"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()+";;;;"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm(tokenAuth)(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Sign In # Baralga"))
}
