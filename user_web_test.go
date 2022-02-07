package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestHandleSignUpPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config: &config{},
	}

	r, _ := http.NewRequest("GET", "/signup", nil)

	a.HandleSignUpPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Sign Up # Baralga"))
}

func TestHandleSignUpFormWithSuccessfullSignUp(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()
	mailService := NewInMemMailService()

	a := &app{
		Config: &config{},

		MailService: mailService,

		UserRepository:    NewInMemUserRepository(),
		ProjectRepository: NewInMemProjectRepository(),
	}

	data := url.Values{}
	data["Name"] = []string{"Norah Newbie"}
	data["EMail"] = []string{"newbie@baralga.com"}
	data["Password"] = []string{"myPassword?!§!"}

	r, _ := http.NewRequest("POST", "/signup", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleSignUpForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(len(mailService.mails), 1)
}

func TestHandleSignUpFormValidation(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		UserRepository:    NewInMemUserRepository(),
		ProjectRepository: NewInMemProjectRepository(),
	}

	data := url.Values{}
	data["EMail"] = []string{"newbie--no--wmIL"}
	data["Password"] = []string{"myPassword?!§!"}

	r, _ := http.NewRequest("POST", "/signup", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleSignUpFormValidate()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "is-invalid"))
}

func TestHandleSignUpConfirmWithExistingConfirmation(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:         &config{},
		UserRepository: NewInMemUserRepository(),
		MailService:    NewInMemMailService(),
	}

	r, _ := http.NewRequest("GET", fmt.Sprintf("/signup/confirm/%v", confirmationIdSample), nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("confirmation-id", confirmationIdSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleSignUpConfirm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusFound)

	l, err := httpRec.Result().Location()
	is.NoErr(err)
	is.Equal(l.String(), "/login")
}

func TestHandleSignUpConfirmWithoutConfirmation(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:         &config{},
		UserRepository: NewInMemUserRepository(),
	}

	confirmationId := uuid.New()
	r, _ := http.NewRequest("GET", fmt.Sprintf("/signup/confirm/%v", confirmationId), nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("confirmation-id", confirmationId.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleSignUpConfirm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusFound)

	l, err := httpRec.Result().Location()
	is.NoErr(err)
	is.Equal(l.String(), "/signup")
}
