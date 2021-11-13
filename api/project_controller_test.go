package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestMapToProjectModelWithoutClaims(t *testing.T) {
	is := is.New(t)

	principal := &Principal{}

	project := &Project{
		ID:          uuid.New(),
		Title:       "My Title",
		Description: "My Description",
	}

	projectModel := mapToProjectModel(principal, project)

	is.Equal(project.ID.String(), projectModel.ID)
	is.Equal(project.Title, projectModel.Title)
	is.Equal(project.Description, projectModel.Description)
	is.Equal(1, projectModel.Links.Size())
}

func TestMapToProjectModelWithAdminClaim(t *testing.T) {
	is := is.New(t)

	principal := &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}

	project := &Project{
		ID:          uuid.New(),
		Title:       "My Title",
		Description: "My Description",
	}

	projectModel := mapToProjectModel(principal, project)

	is.Equal(project.ID.String(), projectModel.ID)
	is.Equal(project.Title, projectModel.Title)
	is.Equal(project.Description, projectModel.Description)
	is.Equal(4, projectModel.Links.Size())
}

func TestMapToProject(t *testing.T) {
	is := is.New(t)

	projectModel := &projectModel{
		ID:          "00000000-0000-0000-1111-000000000001",
		Title:       "Title",
		Description: "Description",
		Active:      true,
	}

	project, err := mapToProject(projectModel)

	is.NoErr(err)
	is.Equal(projectModel.ID, project.ID.String())
	is.Equal(projectModel.Title, project.Title)
	is.Equal(projectModel.Description, project.Description)
	is.Equal(projectModel.Active, project.Active)
}

func TestMapToProjectWithInvalidId(t *testing.T) {
	is := is.New(t)

	projectModel := &projectModel{
		ID:          "not-a-uuid",
		Title:       "Title",
		Description: "Description",
		Active:      true,
	}

	_, err := mapToProject(projectModel)

	is.True(err != nil)
}

func TestHandleGetProjects(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/projects", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleGetProjects()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	projectsModel := &projectsModel{}
	err := json.NewDecoder(httpRec.Body).Decode(projectsModel)
	is.NoErr(err)
	is.Equal(1, len(projectsModel.EmbeddedProjects.ProjectModels))
}

func TestHandleGetProjectWithInvalidId(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/projects/not-a-uuid", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleGetProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/projects/00000000-0000-0000-1111-000000000001", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "00000000-0000-0000-1111-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	projectModel := &projectModel{}
	err := json.NewDecoder(httpRec.Body).Decode(projectModel)
	is.NoErr(err)
	is.Equal("00000000-0000-0000-1111-000000000001", projectModel.ID)
}

func TestHandleGetNonExistingProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/projects/897b7f44-1f31-4c95-80cb-bbb43e4dcf05", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "897b7f44-1f31-4c95-80cb-bbb43e4dcf05")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
}

func TestHandleUpdateProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		"id":null,
		"title": "My updated Title",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", "/api/projects/00000000-0000-0000-1111-000000000001", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "00000000-0000-0000-1111-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	projectsModel := &projectsModel{}
	err := json.NewDecoder(httpRec.Body).Decode(projectsModel)
	is.NoErr(err)
}

func TestHandleUpdateProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		"id":null,
		"title": "My updated Title",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", "/api/projects/00000000-0000-0000-1111-000000000001", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "00000000-0000-0000-1111-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
}

func TestHandleUpdateNonExistingProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		"id":null,
		"title": "My updated Title",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", "/api/projects/897b7f44-1f31-4c95-80cb-bbb43e4dcf05", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "897b7f44-1f31-4c95-80cb-bbb43e4dcf05")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
}

func TestHandleUpdateProjectWithInvalidBody(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		INVALID!!
	 }
	`
	r, _ := http.NewRequest("PATCH", "/api/projects/00000000-0000-0000-1111-000000000001", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "00000000-0000-0000-1111-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleUpdateWithIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		"id":null,
		"title": "My updated Title",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", "/api/projects/not-a-uuid", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleCreateProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &app{
		Config:            &config{},
		ProjectRepository: repo,
	}

	countBefore := len(repo.projects)
	body := `
	{
		"title": "My new Title",
		"description": "My new Description"
	}
	`

	r, _ := http.NewRequest("POST", "/api/projects", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	a.HandleCreateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusCreated)
	is.Equal(countBefore+1, len(repo.projects))
}

func TestHandleCreateProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		"title": "My new Title",
		"description": "My new Description"
	 }
	`

	r, _ := http.NewRequest("POST", "/api/projects", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleCreateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
}

func TestHandleCreateProjectWithInvalidBody(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		INVALID!!
	 }
	`

	r, _ := http.NewRequest("POST", "/api/projects", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleCreateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleDeleteProjectAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &app{
		Config:            &config{},
		ProjectRepository: repo,
	}

	r, _ := http.NewRequest("DELETE", "/api/projects/00000000-0000-0000-1111-000000000001", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "admin",
		Roles:    []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "00000000-0000-0000-1111-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleDeleteProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(0, len(repo.projects))
}

func TestHandleDeleteProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &app{
		Config:            &config{},
		ProjectRepository: repo,
	}

	r, _ := http.NewRequest("DELETE", "/api/projects/00000000-0000-0000-1111-000000000001", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "user1",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "00000000-0000-0000-1111-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleDeleteProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
	is.Equal(1, len(repo.projects))
}

func TestHandleDeleteProjectIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("DELETE", "/api/projects/not-a-uuid", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "user1",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleDeleteProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}
