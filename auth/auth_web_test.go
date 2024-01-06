package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/baralga/user"
	"github.com/go-chi/jwtauth/v5"
	"github.com/matryer/is"
)

func TestHandleLoginPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &AuthWebHandlers{
		config: &shared.Config{},
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

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	config := &shared.Config{}

	userRepository := user.NewInMemUserRepository()
	a := &AuthWebHandlers{
		config:    config,
		tokenAuth: tokenAuth,
		authService: &AuthService{
			config:         config,
			userRepository: userRepository,
		},
	}

	data := url.Values{}
	data["EMail"] = []string{"admin@baralga.com"}
	data["Password"] = []string{"adm1n"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusFound)
	is.Equal(httpRec.Header()["Location"][0], "/")
}

func TestHandleLoginFormWithSuccessfullLoginAndRedirect(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	config := &shared.Config{}

	a := &AuthWebHandlers{
		config:    config,
		tokenAuth: tokenAuth,
		authService: &AuthService{
			config:         config,
			userRepository: user.NewInMemUserRepository(),
		},
	}

	data := url.Values{}
	data["EMail"] = []string{"admin@baralga.com"}
	data["Password"] = []string{"adm1n"}
	data["Redirect"] = []string{"/report"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusFound)
	is.Equal(httpRec.Header()["Location"][0], "/report")
}

func TestHandleLoginFormWithInvalidLogin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	config := &shared.Config{}

	userRepository := user.NewInMemUserRepository()
	a := &AuthWebHandlers{
		config:    config,
		tokenAuth: tokenAuth,
		authService: &AuthService{
			config:         config,
			userRepository: userRepository,
		},
	}

	data := url.Values{}
	data["EMail"] = []string{"admin@baralga.com"}
	data["Password"] = []string{"-just-wrong-"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Sign In # Baralga"))
}

func TestHandleLoginFormWithInvalidFormData(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	a := &AuthWebHandlers{
		config:      &shared.Config{},
		tokenAuth:   tokenAuth,
		userService: user.NewInMemUserService(),
	}

	data := url.Values{}
	data["NotMatching"] = []string{"admin"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Sign In # Baralga"))
}

func TestHandleLoginFormWithInvalidBodyData(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	a := &AuthWebHandlers{
		config:      &shared.Config{},
		tokenAuth:   tokenAuth,
		userService: user.NewInMemUserService(),
	}

	data := url.Values{}
	data["NotMatching"] = []string{"admin"}

	r, _ := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()+";;;;"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	a.HandleLoginForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Sign In # Baralga"))
}

func TestLoginParamsFromQueryParams(t *testing.T) {
	is := is.New(t)

	t.Run("login params without any query params", func(t *testing.T) {
		params := make(url.Values)

		filter := loginParamsFromQueryParams(params)

		is.Equal(filter.errorMessage, "")
		is.Equal(filter.infoMessage, "")
	})

	t.Run("login params with info query param 'confirm_successfull'", func(t *testing.T) {
		params := make(url.Values)
		params.Add("info", "confirm_successfull")

		filter := loginParamsFromQueryParams(params)

		is.Equal(filter.errorMessage, "")
		is.Equal(filter.infoMessage, "You've been confirmed, so happy time tracking!")
	})

	t.Run("login params with invalid info query param '-not-valid-'", func(t *testing.T) {
		params := make(url.Values)
		params.Add("info", "-not-valid-")

		filter := loginParamsFromQueryParams(params)

		is.Equal(filter.errorMessage, "")
		is.Equal(filter.infoMessage, "")
	})

	t.Run("login params with invalid redirect", func(t *testing.T) {
		params := make(url.Values)
		params.Add("redirect", "https://malicious-site.de")

		filter := loginParamsFromQueryParams(params)

		is.Equal(filter.errorMessage, "")
		is.Equal(filter.infoMessage, "")
		is.Equal(filter.redirect, "")
	})

	t.Run("login params with valid redirect", func(t *testing.T) {
		params := make(url.Values)
		params.Add("redirect", "/reports")

		filter := loginParamsFromQueryParams(params)

		is.Equal(filter.errorMessage, "")
		is.Equal(filter.infoMessage, "")
		is.Equal(filter.redirect, "/reports")
	})
}
