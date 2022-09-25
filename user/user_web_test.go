package user

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestHandleSignUpPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &UserWeb{
		config: &shared.Config{},
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
	mailService := shared.NewInMemMailResource()
	userRepository := NewInMemUserRepository()
	organizationInitializerCalled := false

	config := &shared.Config{}
	w := &UserWeb{
		config: config,
		userService: &UserService{
			config:                 config,
			repositoryTxer:         shared.NewInMemRepositoryTxer(),
			mailResource:           mailService,
			organizationRepository: NewInMemOrganizationRepository(),
			organizationInitializer: func(ctxWithTx context.Context, organizationID uuid.UUID) error {
				organizationInitializerCalled = true
				return nil
			},
			userRepository: userRepository,
		},
		userRepository: userRepository,
	}

	data := url.Values{}
	data["Name"] = []string{"Norah Newbie"}
	data["EMail"] = []string{"newbie@baralga.com"}
	data["Password"] = []string{"myPassword?!§!"}

	r, _ := http.NewRequest("POST", "/signup", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w.HandleSignUpForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(len(mailService.Mails), 1)
	is.True(organizationInitializerCalled)
}

func TestHandleSignUpFormWithInvalidData(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &UserWeb{
		config: &shared.Config{},
	}

	data := url.Values{}
	data["EMail"] = []string{"newbie--no--wmIL"}
	data["Password"] = []string{"myPassword?!§!"}

	r, _ := http.NewRequest("POST", "/signup", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleSignUpForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}

func TestHandleSignUpFormWithInvalidFormData(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &UserWeb{
		config: &shared.Config{},
	}

	r, _ := http.NewRequest("POST", "/signup", strings.NewReader("Not a form!!"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleSignUpForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}

func TestHandleSignUpFormValidation(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	userRepository := NewInMemUserRepository()
	a := &UserWeb{
		config: &shared.Config{},
		userService: &UserService{
			repositoryTxer:         shared.NewInMemRepositoryTxer(),
			organizationRepository: NewInMemOrganizationRepository(),
			userRepository:         userRepository,
		},
		userRepository: userRepository,
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

	userRepository := NewInMemUserRepository()

	config := &shared.Config{}

	w := &UserWeb{
		config: config,
		userService: &UserService{
			config:                 config,
			repositoryTxer:         shared.NewInMemRepositoryTxer(),
			organizationRepository: NewInMemOrganizationRepository(),
			userRepository:         userRepository,
		},
		userRepository: userRepository,
	}

	r, _ := http.NewRequest("GET", fmt.Sprintf("/signup/confirm/%v", shared.ConfirmationIdSample), nil)
	r = r.WithContext(context.WithValue(r.Context(), shared.ContextKeyPrincipal, &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("confirmation-id", shared.ConfirmationIdSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w.HandleSignUpConfirm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusFound)

	l, err := httpRec.Result().Location()
	is.NoErr(err)
	is.Equal(l.String(), "/login?info=confirm_successfull")
}

func TestHandleSignUpConfirmWithoutConfirmation(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	userRepository := NewInMemUserRepository()

	a := &UserWeb{
		config: &shared.Config{},
		userService: &UserService{
			repositoryTxer:         shared.NewInMemRepositoryTxer(),
			organizationRepository: NewInMemOrganizationRepository(),
			userRepository:         userRepository,
		},
		userRepository: userRepository,
	}

	confirmationId := uuid.New()
	r, _ := http.NewRequest("GET", fmt.Sprintf("/signup/confirm/%v", confirmationId), nil)
	r = r.WithContext(context.WithValue(r.Context(), shared.ContextKeyPrincipal, &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("confirmation-id", confirmationId.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleSignUpConfirm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusFound)

	l, err := httpRec.Result().Location()
	is.NoErr(err)
	is.Equal(l.String(), "/signup")
}
