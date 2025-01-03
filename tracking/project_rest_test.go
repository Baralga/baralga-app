package tracking

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestMapToProjectModelWithoutClaims(t *testing.T) {
	is := is.New(t)

	principal := &shared.Principal{}

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

	principal := &shared.Principal{
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

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/projects", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

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

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/projects/not-a-uuid", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleGetProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", fmt.Sprintf("/projects/%v", shared.ProjectIDSample), nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	mux := &http.ServeMux{}
	mux.Handle("GET /projects/{projectID}", a.HandleGetProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	projectModel := &projectModel{}
	err := json.NewDecoder(httpRec.Body).Decode(projectModel)
	is.NoErr(err)
	is.Equal(shared.ProjectIDSample.String(), projectModel.ID)
}

func TestHandleGetNonExistingProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/projects/897b7f44-1f31-4c95-80cb-bbb43e4dcf05", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	mux := &http.ServeMux{}
	mux.Handle("GET /projects/{projectID}", a.HandleGetProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
}

func TestHandleUpdateProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	projectRepository := NewInMemProjectRepository()
	c := &ProjectRestHandlers{
		config: &shared.Config{},
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: projectRepository,
		},
		projectRepository: projectRepository,
	}

	body := `
	{
		"id":null,
		"title": "My updated Title",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", fmt.Sprintf("/projects/%v", shared.ProjectIDSample), strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	mux := &http.ServeMux{}
	mux.Handle("PATCH /projects/{projectID}", c.HandleUpdateProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	projectsModel := &projectsModel{}
	err := json.NewDecoder(httpRec.Body).Decode(projectsModel)
	is.NoErr(err)
}

func TestHandleUpdateInvalidProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		"id":null,
		"title": "",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", "/projects/00000000-0000-0000-1111-000000000001", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	mux := &http.ServeMux{}
	mux.Handle("PATCH /projects/{projectID}", a.HandleUpdateProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleUpdateProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		"id":null,
		"title": "My updated Title",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", "/projects/00000000-0000-0000-1111-000000000001", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	mux := &http.ServeMux{}
	mux.Handle("PATCH /projects/{projectID}", a.HandleUpdateProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
}

func TestHandleUpdateNonExistingProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	projectRepository := NewInMemProjectRepository()
	c := &ProjectRestHandlers{
		config: &shared.Config{},
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: projectRepository,
		},
		projectRepository: projectRepository,
	}

	body := `
	{
		"id":null,
		"title": "My updated Title",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", "/api/projects/897b7f44-1f31-4c95-80cb-bbb43e4dcf05", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	mux := &http.ServeMux{}
	mux.Handle("PATCH /projects/{projectID}", c.HandleUpdateProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
}

func TestHandleUpdateProjectWithInvalidBody(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		INVALID!!
	 }
	`
	r, _ := http.NewRequest("PATCH", "/api/projects/00000000-0000-0000-1111-000000000001", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	mux := &http.ServeMux{}
	mux.Handle("PATCH /api/projects/{projectID}", a.HandleUpdateProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleUpdateWithIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}
	body := `
	{
		"id":null,
		"title": "My updated Title",
		"description": "My updated Description"
	 }
	`

	r, _ := http.NewRequest("PATCH", "/api/projects/not-a-uuid", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	mux := &http.ServeMux{}
	mux.Handle("PATCH /api/projects/{projectID}", a.HandleUpdateProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleCreateProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()

	c := &ProjectRestHandlers{
		config: &shared.Config{},
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: repo,
		},
		projectRepository: repo,
	}

	countBefore := len(repo.projects)
	body := `
	{
		"title": "My new Title",
		"description": "My new Description"
	}
	`

	r, _ := http.NewRequest("POST", "/api/projects", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	c.HandleCreateProject()(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusCreated)
	is.Equal(countBefore+1, len(repo.projects))
}

func TestHandleInvalidCreateProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()
	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: repo,
	}

	body := `
	{
		"title": "",
		"description": "My new Description"
	}
	`

	r, _ := http.NewRequest("POST", "/api/projects", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	a.HandleCreateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleCreateProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		"title": "My new Title",
		"description": "My new Description"
	 }
	`

	r, _ := http.NewRequest("POST", "/api/projects", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleCreateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
}

func TestHandleCreateProjectWithInvalidBody(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	body := `
	{
		INVALID!!
	 }
	`

	r, _ := http.NewRequest("POST", "/api/projects", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleCreateProject()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleDeleteProjectAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()

	c := &ProjectRestHandlers{
		config: &shared.Config{},
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: repo,
		},
		projectRepository: repo,
	}

	r, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/projects/%v", shared.ProjectIDSample), nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "admin",
		Roles:    []string{"ROLE_ADMIN"},
	}))

	mux := &http.ServeMux{}
	mux.Handle("DELETE /api/projects/{projectID}", c.HandleDeleteProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(0, len(repo.projects))
}

func TestHandleDeleteProjectAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemProjectRepository()

	c := &ProjectRestHandlers{
		config: &shared.Config{},
		projectService: &ProjectService{
			repositoryTxer:    shared.NewInMemRepositoryTxer(),
			projectRepository: repo,
		},
		projectRepository: repo,
	}

	r, _ := http.NewRequest("DELETE", "/api/projects/00000000-0000-0000-1111-000000000001", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "user1",
	}))

	mux := &http.ServeMux{}
	mux.Handle("DELETE /api/projects/{projectID}", c.HandleDeleteProject())
	mux.ServeHTTP(httpRec, r)

	is.Equal(httpRec.Result().StatusCode, http.StatusForbidden)
	is.Equal(1, len(repo.projects))
}

func TestHandleDeleteProjectIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	c := &ProjectRestHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("DELETE", "/api/projects/not-a-uuid", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "user1",
	}))

	mux := &http.ServeMux{}
	mux.Handle("DELETE /api/projects/{projectID}", c.HandleDeleteProject())
	mux.ServeHTTP(httpRec, r)
	
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}
