package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/baralga/hal"
	"github.com/baralga/paged"
	"github.com/baralga/util"
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

// HandleGetProjects reads projects
func (a *app) HandleGetProjects() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)
		pageParams := paged.PageParamsOf(r)

		projectsPaged, err := a.ProjectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
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

		util.RenderJSON(w, projectsModel)
	}
}

// HandleGetProject reads a project
func (a *app) HandleGetProject() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := chi.URLParam(r, "project-id")
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		project, err := a.ProjectRepository.FindProjectByID(r.Context(), principal.OrganizationID, projectID)
		if errors.Is(err, ErrProjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}

		projectModel := mapToProjectModel(principal, project)

		util.RenderJSON(w, projectModel)
	}
}

// HandleCreateProject creates a project
func (a *app) HandleCreateProject() http.HandlerFunc {
	isProduction := a.isProduction()
	validator := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

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

		project, err := a.CreateProject(r.Context(), principal, projectToCreate)
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}

		projectModelCreated := mapToProjectModel(principal, project)

		w.WriteHeader(http.StatusCreated)
		util.RenderJSON(w, projectModelCreated)
	}
}

// HandleUpdateProject updates a project
func (a *app) HandleUpdateProject() http.HandlerFunc {
	isProduction := a.isProduction()
	validator := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := chi.URLParam(r, "project-id")
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

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

		projectUpdate, err := a.ProjectRepository.UpdateProject(r.Context(), principal.OrganizationID, project)
		if errors.Is(err, ErrProjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}

		projectModelUpdate := mapToProjectModel(principal, projectUpdate)
		util.RenderJSON(w, projectModelUpdate)
	}
}

// HandleDeleteProject deletes a project
func (a *app) HandleDeleteProject() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := chi.URLParam(r, "project-id")
		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		if !principal.HasRole("ROLE_ADMIN") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err = a.DeleteProjectByID(r.Context(), principal, projectID)
		if errors.Is(err, ErrProjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
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

func mapToProjectModel(principal *Principal, project *Project) *projectModel {
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
