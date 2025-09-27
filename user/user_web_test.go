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

	a := &UserWebHandlers{
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
	w := &UserWebHandlers{
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

func TestHandleOrganizationDialog(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	adminPrincipal := &shared.Principal{
		Name:           "Admin User",
		Username:       "admin",
		OrganizationID: shared.OrganizationIDSample,
		Roles:          []string{"ROLE_ADMIN"},
	}

	userPrincipal := &shared.Principal{
		Name:           "Regular User",
		Username:       "user",
		OrganizationID: shared.OrganizationIDSample,
		Roles:          []string{"ROLE_USER"},
	}

	// Test admin user
	ctx := shared.ToContextWithPrincipal(context.Background(), adminPrincipal)
	r, _ := http.NewRequest("GET", "/organization/dialog", nil)
	r = r.WithContext(ctx)

	organizationRepository := NewInMemOrganizationRepository()
	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		organizationRepository: organizationRepository,
	}

	userWebHandlers := &UserWebHandlers{
		config:      &shared.Config{},
		userService: userService,
	}

	userWebHandlers.HandleOrganizationDialog()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Organization Settings"))
	is.True(strings.Contains(htmlBody, "Update")) // Admin should see update button

	// Test regular user
	httpRec = httptest.NewRecorder()
	ctx = shared.ToContextWithPrincipal(context.Background(), userPrincipal)
	r, _ = http.NewRequest("GET", "/organization/dialog", nil)
	r = r.WithContext(ctx)

	userWebHandlers.HandleOrganizationDialog()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody = httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Organization Settings"))
	is.True(strings.Contains(htmlBody, "readonly")) // Regular user should see readonly field
	is.True(!strings.Contains(htmlBody, "Update"))  // Regular user should not see update button
}

func TestHandleOrganizationUpdate(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	adminPrincipal := &shared.Principal{
		Name:           "Admin User",
		Username:       "admin",
		OrganizationID: shared.OrganizationIDSample,
		Roles:          []string{"ROLE_ADMIN"},
	}

	userPrincipal := &shared.Principal{
		Name:           "Regular User",
		Username:       "user",
		OrganizationID: shared.OrganizationIDSample,
		Roles:          []string{"ROLE_USER"},
	}

	organizationRepository := NewInMemOrganizationRepository()
	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		organizationRepository: organizationRepository,
	}

	userWebHandlers := &UserWebHandlers{
		config:      &shared.Config{},
		userService: userService,
	}

	// Test successful update by admin
	data := url.Values{}
	data["Name"] = []string{"Updated Organization Name"}
	data["CSRFToken"] = []string{"test-token"}

	ctx := shared.ToContextWithPrincipal(context.Background(), adminPrincipal)
	r, _ := http.NewRequest("POST", "/organization/update", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = r.WithContext(ctx)

	userWebHandlers.HandleOrganizationUpdate()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	// Check for HTMX headers indicating success
	headers := httpRec.Header()
	is.True(strings.Contains(headers.Get("HX-Trigger"), "baralga__main_content_modal-hide"))
	is.True(strings.Contains(headers.Get("HX-Refresh"), "true"))

	// Test regular user (should fail)
	httpRec = httptest.NewRecorder()
	ctx = shared.ToContextWithPrincipal(context.Background(), userPrincipal)
	r, _ = http.NewRequest("POST", "/organization/update", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = r.WithContext(ctx)

	userWebHandlers.HandleOrganizationUpdate()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	// Should not have success headers
	headers = httpRec.Header()
	is.True(!strings.Contains(headers.Get("HX-Trigger"), "baralga__main_content_modal-hide"))
}

func TestHandleOrganizationName(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	principal := &shared.Principal{
		Name:           "Test User",
		Username:       "test",
		OrganizationID: shared.OrganizationIDSample,
		Roles:          []string{"ROLE_USER"},
	}

	organizationRepository := NewInMemOrganizationRepository()
	userService := &UserService{
		config:                 &shared.Config{},
		repositoryTxer:         shared.NewInMemRepositoryTxer(),
		organizationRepository: organizationRepository,
	}

	userWebHandlers := &UserWebHandlers{
		config:      &shared.Config{},
		userService: userService,
	}

	ctx := shared.ToContextWithPrincipal(context.Background(), principal)
	r, _ := http.NewRequest("GET", "/organization/name", nil)
	r = r.WithContext(ctx)

	userWebHandlers.HandleOrganizationName()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Test Organization"))
	is.True(strings.Contains(htmlBody, "organization-name"))
}

func TestHandleSignUpFormWithInvalidData(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &UserWebHandlers{
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

	a := &UserWebHandlers{
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
	a := &UserWebHandlers{
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

	w := &UserWebHandlers{
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
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

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

	a := &UserWebHandlers{
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
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("confirmation-id", confirmationId.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleSignUpConfirm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusFound)

	l, err := httpRec.Result().Location()
	is.NoErr(err)
	is.Equal(l.String(), "/signup")
}
