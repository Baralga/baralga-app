package tracking

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
	"github.com/matryer/is"
)

func TestHandleProjectsPageAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/projects", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleProjectsPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(!strings.Contains(htmlBody, "<form"))
	is.True(!strings.Contains(htmlBody, "hx-delete"))
}

func TestHandleProjectsPageAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/projects", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
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
	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: repo,
	}

	countBefore := len(repo.projects)

	data := url.Values{}
	data["Title"] = []string{"t"}

	r, _ := http.NewRequest("POST", "/projects/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
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
	w := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: repo,
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: repo,
		},
	}

	countBefore := len(repo.projects)

	data := url.Values{}
	data["Title"] = []string{"My new Title"}

	r, _ := http.NewRequest("POST", "/projects/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	w.HandleProjectForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(countBefore+1, len(repo.projects))

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "My new Title"))
}

func TestHandleCreateProjectWithValidProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: repo,
	}

	countBefore := len(repo.projects)

	data := url.Values{}
	data["Title"] = []string{"My new Title"}

	r, _ := http.NewRequest("POST", "/projects/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleProjectForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
	is.Equal(countBefore, len(repo.projects))
}

func TestHandleCreateProjectWithInvalidProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: repo,
	}

	data := url.Values{}
	data["NothingHere"] = []string{"My new Title"}

	r, _ := http.NewRequest("POST", "/projects/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	a.HandleProjectForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
}

func TestHandleArchiveProjectAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()

	w := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: repo,
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: repo,
		},
	}
	r, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%v/archive", shared.ProjectIDSample), nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", shared.ProjectIDSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w.HandleArchiveProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}

func TestHandleArchiveProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()

	w := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: repo,
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: repo,
		},
	}

	r, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%v/archive", shared.ProjectIDSample), nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_USER"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", shared.ProjectIDSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w.HandleArchiveProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
}

func TestHandleProjectViewAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", fmt.Sprintf("/projects/%s", shared.ProjectIDSample.String()), nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", shared.ProjectIDSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleProjectView()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "card"))
}

func TestHandleProjectEditAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", fmt.Sprintf("/projects/%s/edit", shared.ProjectIDSample.String()), nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", shared.ProjectIDSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleProjectEdit()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
}

func TestHandleProjectEditAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", fmt.Sprintf("/projects/%s/edit", shared.ProjectIDSample.String()), nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "admin",
		Roles:    []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", shared.ProjectIDSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleProjectEdit()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "form"))
}

func TestHandleProjectEditFormAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%s/edit", shared.ProjectIDSample.String()), nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", shared.ProjectIDSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleProjectEditForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
}

func TestHandleProjectEditFormAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	projectRepository := NewInMemProjectRepository()
	a := &ProjectWeb{
		config:            &shared.Config{},
		projectRepository: projectRepository,
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: projectRepository,
		},
	}

	data := url.Values{}
	data["ID"] = []string{shared.ProjectIDSample.String()}
	data["Title"] = []string{"My new Title!"}

	r, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%s/edit", shared.ProjectIDSample.String()), strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "admin",
		Roles:    []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", shared.ProjectIDSample.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleProjectEditForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "card"))
}
