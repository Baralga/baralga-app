package tracking

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/baralga/shared"
	"github.com/baralga/shared/hal"
	"github.com/baralga/shared/paged"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"schneider.vip/problem"
)

type projectModel struct {
	ID          string     `json:"id"`
	Title       string     `json:"title" validate:"required,min=3,max=100"`
	Description string     `json:"description" validate:"max=500"`
	Active      bool       `json:"active"`
	Links       *hal.Links `json:"_links"`
}

type EmbeddedProjects struct {
	ProjectModels []*projectModel `json:"projects"`
}

type projectsModel struct {
	*EmbeddedProjects `json:"_embedded"`
	*paged.Page       `json:"page"`
	Links             *hal.Links `json:"_links"`
}

type ProjectRestHandlers struct {
	config            *shared.Config
	projectRepository ProjectRepository
	projectService    *ProjectService
}

func NewProjectController(config *shared.Config, projectRepository ProjectRepository, projectService *ProjectService) *ProjectRestHandlers {
	return &ProjectRestHandlers{
		config:            config,
		projectRepository: projectRepository,
		projectService:    projectService,
	}
}

func (a *ProjectRestHandlers) RegisterProtected(r chi.Router) {
	r.Handle("GET /projects", a.HandleGetProjects())
	r.Handle("POST /projects", a.HandleCreateProject())
	r.Handle("GET /projects/{project-id}", a.HandleGetProject())
	r.Handle("DELETE /projects/{project-id}", a.HandleDeleteProject())
	r.Handle("PATCH /projects/{project-id}", a.HandleUpdateProject())
}

func (a *ProjectRestHandlers) RegisterOpen(r chi.Router) {
}

// HandleGetProjects reads projects
func (a *ProjectRestHandlers) HandleGetProjects() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	projectRepository := a.projectRepository
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())
		pageParams := paged.PageParamsOf(r)

		projectsPaged, err := projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		var projectModels []*projectModel
		for _, project := range projectsPaged.Projects {
			projectModel := mapToProjectModel(principal, project)
			projectModels = append(projectModels, projectModel)
		}

		projectsModel := &projectsModel{
			EmbeddedProjects: &EmbeddedProjects{
				ProjectModels: projectModels,
			},
			Page: projectsPaged.Page,
		}

		selfLink := hal.NewSelfLink(r.RequestURI)
		if principal.HasRole("ROLE_ADMIN") {
			projectsModel.Links = hal.NewLinks(
				selfLink,
				hal.NewLink("create", "/api/projects"),
			)
		} else {
			projectsModel.Links = hal.NewLinks(
				selfLink,
			)
		}

		shared.RenderJSON(w, projectsModel)
	}
}

// HandleGetProject reads a project
func (a *ProjectRestHandlers) HandleGetProject() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	projectRepository := a.projectRepository
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := r.PathValue("project-id")
		principal := shared.MustPrincipalFromContext(r.Context())

		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		project, err := projectRepository.FindProjectByID(r.Context(), principal.OrganizationID, projectID)
		if errors.Is(err, ErrProjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		projectModel := mapToProjectModel(principal, project)

		shared.RenderJSON(w, projectModel)
	}
}

// HandleCreateProject creates a project
func (a *ProjectRestHandlers) HandleCreateProject() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validator := validator.New()
	projectService := a.projectService
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

		var projectModel projectModel
		err := json.NewDecoder(r.Body).Decode(&projectModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		if !principal.HasRole("ROLE_ADMIN") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err = validator.Struct(projectModel)
		if err != nil {
			http.Error(w, problem.New(problem.Title("project not valid")).JSONString(), http.StatusBadRequest)
			return
		}

		projectToCreate, err := mapToProject(&projectModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		project, err := projectService.CreateProject(r.Context(), principal, projectToCreate)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		projectModelCreated := mapToProjectModel(principal, project)

		w.WriteHeader(http.StatusCreated)
		shared.RenderJSON(w, projectModelCreated)
	}
}

// HandleUpdateProject updates a project
func (a *ProjectRestHandlers) HandleUpdateProject() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validator := validator.New()
	projectService := a.projectService
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := r.PathValue("project-id")
		principal := shared.MustPrincipalFromContext(r.Context())

		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		var projectModel projectModel
		err = json.NewDecoder(r.Body).Decode(&projectModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		if !principal.HasRole("ROLE_ADMIN") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err = validator.Struct(projectModel)
		if err != nil {
			http.Error(w, problem.New(problem.Title("project not valid")).JSONString(), http.StatusBadRequest)
			return
		}

		project, err := mapToProject(&projectModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		project.ID = projectID

		projectUpdate, err := projectService.UpdateProject(r.Context(), principal.OrganizationID, project)
		if errors.Is(err, ErrProjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		projectModelUpdate := mapToProjectModel(principal, projectUpdate)
		shared.RenderJSON(w, projectModelUpdate)
	}
}

// HandleDeleteProject deletes a project
func (a *ProjectRestHandlers) HandleDeleteProject() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	projectService := a.projectService
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := r.PathValue("project-id")
		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		principal := shared.MustPrincipalFromContext(r.Context())

		if !principal.HasRole("ROLE_ADMIN") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err = projectService.DeleteProjectByID(r.Context(), principal, projectID)
		if errors.Is(err, ErrProjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		w.Header().Set("HX-Trigger", "{ \"baralga__activities-changed\": true, \"baralga__projects-changed\": true } ")
	}
}

func mapToProject(projectModel *projectModel) (*Project, error) {
	var projectID uuid.UUID

	if projectModel.ID != "" {
		pID, err := uuid.Parse(projectModel.ID)
		if err != nil {
			return nil, err
		}
		projectID = pID
	}

	return &Project{
		ID:          projectID,
		Title:       projectModel.Title,
		Description: projectModel.Description,
		Active:      projectModel.Active,
	}, nil
}

func mapToProjectModel(principal *shared.Principal, project *Project) *projectModel {
	projectModel := &projectModel{
		ID:          project.ID.String(),
		Title:       project.Title,
		Description: project.Description,
		Active:      project.Active,
	}
	selfLink := hal.NewSelfLink(fmt.Sprintf("/api/projects/%s", projectModel.ID))
	if principal.HasRole("ROLE_ADMIN") {
		projectModel.Links = hal.NewLinks(
			selfLink,
			hal.NewLink("create", selfLink.Href()),
			hal.NewLink("delete", selfLink.Href()),
			hal.NewLink("edit", selfLink.Href()),
		)
	} else {
		projectModel.Links = hal.NewLinks(
			selfLink,
		)
	}
	return projectModel
}
