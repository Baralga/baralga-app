package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestHandleProjectsPageAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/projects", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleProjectsPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(!strings.Contains(htmlBody, "<form"))
	is.True(!strings.Contains(htmlBody, "hx-delete"))
}

func TestHandleProjectsPageAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/projects", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "admin",
		Roles:    []string{"ROLE_ADMIN"},
	}))

	a.HandleProjectsPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
	is.True(strings.Contains(htmlBody, "hx-delete"))
}

func TestHandleCreateProjectWithNotValidProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &app{
		Config:            &config{},
		ProjectRepository: repo,
	}

	countBefore := len(repo.projects)

	data := url.Values{}
	data["Title"] = []string{"t"}

	r, _ := http.NewRequest("POST", "/projects/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	a.HandleProjectForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(countBefore, len(repo.projects))

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
}

func TestHandleCreateProjectWithValidProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &app{
		Config:            &config{},
		RepositoryTxer:    NewInMemRepositoryTxer(),
		ProjectRepository: repo,
	}

	countBefore := len(repo.projects)

	data := url.Values{}
	data["Title"] = []string{"My new Title"}

	r, _ := http.NewRequest("POST", "/projects/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	a.HandleProjectForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(countBefore+1, len(repo.projects))

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "My new Title"))
}

func TestHandleCreateProjectWithValidProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &app{
		Config:            &config{},
		ProjectRepository: repo,
	}

	countBefore := len(repo.projects)

	data := url.Values{}
	data["Title"] = []string{"My new Title"}

	r, _ := http.NewRequest("POST", "/projects/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleProjectForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
	is.Equal(countBefore, len(repo.projects))
}

func TestHandleCreateProjectWithInvalidProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &app{
		Config:            &config{},
		ProjectRepository: repo,
	}

	data := url.Values{}
	data["NothingHere"] = []string{"My new Title"}

	r, _ := http.NewRequest("POST", "/projects/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	a.HandleProjectForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
}
