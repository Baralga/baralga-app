package tracking

import (
	"fmt"
	"net/http"

	"github.com/baralga/shared"
	"github.com/baralga/shared/hx"
	"github.com/baralga/shared/paged"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
	"github.com/pkg/errors"
	"schneider.vip/problem"
)

type projectFormModel struct {
	CSRFToken string
	ID        string
	Title     string ` validate:"required,min=3,max=50"`
}

type ProjectWeb struct {
	config            *shared.Config
	projectService    *ProjectService
	projectRepository ProjectRepository
}

func NewProjectWeb(config *shared.Config, projectService *ProjectService, projectRepository ProjectRepository) *ProjectWeb {
	return &ProjectWeb{
		config:            config,
		projectService:    projectService,
		projectRepository: projectRepository,
	}
}

func (a *ProjectWeb) RegisterProtected(r chi.Router) {
	r.Get("/projects", a.HandleProjectsPage())
	r.Post("/projects/new", a.HandleProjectForm())
	r.Get("/projects/{project-id}/archive", a.HandleArchiveProject())
}

func (a *ProjectWeb) RegisterOpen(r chi.Router) {
}

func (a *ProjectWeb) HandleProjectsPage() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(shared.ContextKeyPrincipal).(*shared.Principal)

		pageParams := &paged.PageParams{
			Page: 0,
			Size: 50,
		}

		projects, err := a.projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		if !hx.IsHXRequest(r) {
			pageContext := &shared.PageContext{
				Principal:   principal,
				CurrentPath: r.URL.Path,
				Title:       "Projects",
			}

			formModel := projectFormModel{}
			formModel.CSRFToken = csrf.Token(r)

			shared.RenderHTML(w, ProjectsPage(pageContext, formModel, projects))
			return
		}

		w.Header().Set("HX-Trigger", "baralga__main_content_modal-show")

		formModel := projectFormModel{}
		formModel.CSRFToken = csrf.Token(r)

		shared.RenderHTML(w, ProjectsView(principal, formModel, projects))
	}
}

func (a *ProjectWeb) HandleProjectForm() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validator := validator.New()
	projectService := a.projectService
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(shared.ContextKeyPrincipal).(*shared.Principal)

		if !principal.HasRole("ROLE_ADMIN") {
			http.Error(w, "No permission.", http.StatusForbidden)
			return
		}

		err := r.ParseForm()
		if err != nil {
			_ = a.renderProjectsView(
				w,
				r,
				principal,
				isProduction,
				projectFormModel{},
			)
			return
		}

		var formModel projectFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			_ = a.renderProjectsView(
				w,
				r,
				principal,
				isProduction,
				projectFormModel{},
			)
			return
		}

		err = validator.Struct(formModel)
		if err != nil {
			_ = a.renderProjectsView(
				w,
				r,
				principal,
				isProduction,
				formModel,
			)
			return
		}

		projectToCreate := mapFormToProject(formModel)
		_, err = projectService.CreateProject(r.Context(), principal, &projectToCreate)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		w.Header().Set("HX-Trigger", "baralga__projects-changed")
		err = a.renderProjectsView(
			w,
			r,
			principal,
			isProduction,
			projectFormModel{},
		)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}
	}
}

// HandleArchiveProject archives a project
func (a *ProjectWeb) HandleArchiveProject() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	projectService := a.projectService
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := chi.URLParam(r, "project-id")
		principal := r.Context().Value(shared.ContextKeyPrincipal).(*shared.Principal)

		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusNotAcceptable)
			return
		}

		if !principal.HasRole("ROLE_ADMIN") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err = projectService.ArchiveProject(r.Context(), principal.OrganizationID, projectID)
		if errors.Is(err, ErrProjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		w.Header().Set("HX-Trigger", "baralga__projects-changed")
	}
}

func (a *ProjectWeb) renderProjectsView(w http.ResponseWriter, r *http.Request, principal *shared.Principal, isProduction bool, formModel projectFormModel) error {
	pageParams := &paged.PageParams{
		Page: 0,
		Size: 50,
	}

	projects, err := a.projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
	if err != nil {
		return err
	}

	formModel.CSRFToken = csrf.Token(r)

	shared.RenderHTML(w, ProjectsView(principal, formModel, projects))

	return nil
}

func ProjectsPage(pageContext *shared.PageContext, formModel projectFormModel, projects *ProjectsPaged) g.Node {
	return shared.Page(
		pageContext.Title,
		pageContext.CurrentPath,
		[]g.Node{
			shared.Navbar(pageContext),
			Section(
				Class("full-center"),
				Div(
					Class("container"),
					Div(
						Class("mt-4 mb-4"),
					),
					ProjectsView(pageContext.Principal, formModel, projects),
				),
			),
		},
	)
}

func ProjectsView(principal *shared.Principal, formModel projectFormModel, projects *ProjectsPaged) g.Node {
	return Div(
		ID("baralga__main_content_modal_content"),
		Class("modal-content"),
		Div(
			Class("modal-header"),
			H2(
				Class("modal-title"),
				g.Text("Projects"),
			),
			Button(
				Type("type"),
				Class("btn-close"),
				g.Attr("data-bs-dismiss", "modal"),
			),
		),
		Div(
			Class("modal-body"),
			g.If(
				principal.HasRole("ROLE_ADMIN"),
				ProjectForm(formModel, ""),
			),
			g.Group(
				g.Map(projects.Projects, func(project *Project) g.Node {
					return Div(
						Class("card mt-2"),

						hx.Target("this"),
						hx.Swap("outerHTML"),

						Div(
							Class("card-body"),
							H5(
								Class("card-title mt-2"),
								Div(
									Class("d-flex justify-content-between mb-2"),
									Span(
										Class("flex-grow-1"),
										g.Text(project.Title),
									),
									g.If(
										principal.HasRole("ROLE_ADMIN"),
										A(
											hx.Confirm(fmt.Sprintf("Do you really want to delete project %v?", project.Title)),
											hx.Delete(fmt.Sprintf("/api/projects/%v", project.ID)),
											Class("btn btn-outline-secondary btn-sm ms-1"),
											I(Class("bi-trash2")),
										),
									),
									g.If(
										principal.HasRole("ROLE_ADMIN"),
										A(
											hx.Confirm(fmt.Sprintf("Do you really want to archive project %v?", project.Title)),
											hx.Get(fmt.Sprintf("/projects/%v/archive", project.ID)),
											Class("btn btn-outline-secondary btn-sm ms-1"),
											I(Class("bi-archive")),
										),
									),
								),
							),
						),
					)
				}),
			),
		),
	)
}

func ProjectForm(formModel projectFormModel, errorMessage string) g.Node {
	return FormEl(
		ID("project_form"),
		Class("mb-4 mt-2"),
		hx.Post("/projects/new"),
		hx.Target("#baralga__main_content_modal_content"),
		hx.Swap("outerHTML"),

		Input(
			Type("hidden"),
			Name("CSRFToken"),
			Value(formModel.CSRFToken),
		),

		Div(
			Class("input-group mb-3"),
			Input(
				ID("ProjectTitle"),
				Type("text"),
				Name("Title"),
				MinLength("3"),
				MaxLength("100"),
				Value(formModel.Title),
				g.Attr("required", "required"),
				Class("form-control"),
				g.Attr("placeholder", "My new Project"),
			),
			Button(
				Class("btn btn-outline-primary"),
				g.Attr("for", "ProjectTitle"),
				TitleAttr("Add Project"),
				I(Class("bi-plus")),
			),
		),
	)
}

func mapFormToProject(projectFormModel projectFormModel) Project {
	return Project{
		Title:  projectFormModel.Title,
		Active: true,
	}
}
