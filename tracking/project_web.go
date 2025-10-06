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
	"github.com/pkg/errors"
	g "maragu.dev/gomponents"
	ghx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html" //nolint:all
	"schneider.vip/problem"
)

type projectFormModel struct {
	CSRFToken string
	ID        string
	Title     string `validate:"required,min=3,max=50"`
	Billable  bool
}

type ProjectWeb struct {
	config            *shared.Config
	projectService    *ProjectService
	projectRepository ProjectRepository
}

func NewProjectWebHandlers(config *shared.Config, projectService *ProjectService, projectRepository ProjectRepository) *ProjectWeb {
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
	r.Get("/projects/{project-id}", a.HandleProjectView())
	r.Get("/projects/{project-id}/edit", a.HandleProjectEdit())
	r.Post("/projects/{project-id}/edit", a.HandleProjectEditForm())
}

func (a *ProjectWeb) RegisterOpen(r chi.Router) {
}

func (a *ProjectWeb) HandleProjectsPage() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

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

func (a *ProjectWeb) HandleProjectView() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	projectRepository := a.projectRepository
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := chi.URLParam(r, "project-id")
		principal := shared.MustPrincipalFromContext(r.Context())

		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		project, err := projectRepository.FindProjectByID(r.Context(), principal.OrganizationID, projectID)
		if errors.Is(err, ErrProjectNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		shared.RenderHTML(w, ProjectRow(principal, project))

		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}
	}
}

func (a *ProjectWeb) HandleProjectEdit() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	projectRepository := a.projectRepository
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := chi.URLParam(r, "project-id")
		principal := shared.MustPrincipalFromContext(r.Context())

		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		if !principal.HasRole("ROLE_ADMIN") {
			http.Error(w, "No permission.", http.StatusForbidden)
			return
		}

		project, err := projectRepository.FindProjectByID(r.Context(), principal.OrganizationID, projectID)
		if errors.Is(err, ErrProjectNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		formModel := mapProjectToForm(*project)
		formModel.CSRFToken = csrf.Token(r)

		shared.RenderHTML(w, ProjectEditForm(formModel))

		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}
	}
}

func (a *ProjectWeb) HandleProjectEditForm() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validator := validator.New()
	projectService := a.projectService
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDParam := chi.URLParam(r, "project-id")
		principal := shared.MustPrincipalFromContext(r.Context())

		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		if !principal.HasRole("ROLE_ADMIN") {
			http.Error(w, "No permission.", http.StatusForbidden)
			return
		}

		err = r.ParseForm()
		if err != nil {
			shared.RenderHTML(w, ProjectEditForm(projectFormModel{}))
			return
		}

		var formModel projectFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			shared.RenderHTML(w, ProjectEditForm(formModel))
			return
		}

		err = validator.Struct(formModel)
		if err != nil {
			shared.RenderHTML(w, ProjectEditForm(formModel))
			return
		}

		projectToUpdate := mapFormToProject(formModel)
		projectToUpdate.ID = projectID
		_, err = projectService.UpdateProject(r.Context(), principal.OrganizationID, &projectToUpdate)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		shared.RenderHTML(w, ProjectRow(principal, &projectToUpdate))

		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}
	}
}

func (a *ProjectWeb) HandleProjectForm() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validator := validator.New()
	projectService := a.projectService
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())

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
		principal := shared.MustPrincipalFromContext(r.Context())

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
				ProjectNewForm(formModel, ""),
			),
			g.Group(
				g.Map(projects.Projects, func(project *Project) g.Node {
					return ProjectRow(principal, project)
				}),
			),
		),
	)
}

func ProjectRow(principal *shared.Principal, project *Project) g.Node {
	return Div(
		Class("card mt-2"),

		ghx.Target("this"),
		ghx.Swap("outerHTML"),

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
							ghx.Get(fmt.Sprintf("/projects/%v/edit", project.ID)),
							Class("btn btn-outline-secondary btn-sm ms-1"),
							I(Class("bi-pen")),
						),
					),
					g.If(
						principal.HasRole("ROLE_ADMIN"),
						A(
							ghx.Confirm(fmt.Sprintf("Do you really want to delete project %v?", project.Title)),
							ghx.Delete(fmt.Sprintf("/api/projects/%v", project.ID)),
							Class("btn btn-outline-secondary btn-sm ms-1"),
							I(Class("bi-trash2")),
						),
					),
					g.If(
						principal.HasRole("ROLE_ADMIN"),
						A(
							ghx.Confirm(fmt.Sprintf("Do you really want to archive project %v?", project.Title)),
							ghx.Get(fmt.Sprintf("/projects/%v/archive", project.ID)),
							Class("btn btn-outline-secondary btn-sm ms-1"),
							I(Class("bi-archive")),
						),
					),
				),
			),
		),
	)
}

func ProjectEditForm(formModel projectFormModel) g.Node {
	return ProjectForm(formModel, true, "")
}

func ProjectNewForm(formModel projectFormModel, errorMessage string) g.Node {
	return ProjectForm(formModel, false, errorMessage)
}

func ProjectForm(formModel projectFormModel, editMode bool, errorMessage string) g.Node {
	return FormEl(
		Class("mb-4 mt-2"),
		g.If(
			!editMode,
			g.Group(
				[]g.Node{
					ID("project_form_new"),
					ghx.Post("/projects/new"),
					ghx.Target("#baralga__main_content_modal_content"),
				},
			),
		),
		g.If(
			editMode,
			g.Group(
				[]g.Node{
					ID(fmt.Sprintf("project_form_edit_%s", formModel.ID)),
					ghx.Post(fmt.Sprintf("/projects/%s/edit", formModel.ID)),
					ghx.Target("this"),
				},
			),
		),
		ghx.Swap("outerHTML"),

		g.If(formModel.ID != "",
			Input(
				Type("hidden"),
				Name("ID"),
				Value(formModel.ID),
			),
		),
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
			g.If(
				editMode,
				g.Group(
					[]g.Node{
						Button(
							Class("btn btn-outline-primary"),
							g.Attr("for", "ProjectTitle"),
							TitleAttr("Update Project"),
							I(Class("bi-save")),
						),
						Button(
							Class("btn btn-outline-secondary"),
							g.Attr("for", "ProjectTitle"),
							TitleAttr("Cancel Edit"),
							ghx.Get(fmt.Sprintf("/projects/%s", formModel.ID)),
							I(Class("bi-x")),
						),
					},
				),
			),
			g.If(
				!editMode,
				Button(
					Class("btn btn-outline-primary"),
					g.Attr("for", "ProjectTitle"),
					TitleAttr("Add Project"),
					I(Class("bi-plus")),
				),
			),
		),
		Div(
			Class("form-check mb-3"),
			Input(
				ID("ProjectBillable"),
				Type("checkbox"),
				Name("Billable"),
				Value("true"),
				Class("form-check-input"),
				g.If(formModel.Billable, Checked()),
			),
			Label(
				Class("form-check-label text-muted"),
				g.Attr("for", "ProjectBillable"),
				g.Text("Billable project"),
			),
		),
	)
}

func mapFormToProject(projectFormModel projectFormModel) Project {
	return Project{
		Title:    projectFormModel.Title,
		Active:   true,
		Billable: projectFormModel.Billable,
	}
}

func mapProjectToForm(project Project) projectFormModel {
	return projectFormModel{
		ID:       project.ID.String(),
		Title:    project.Title,
		Billable: project.Billable,
	}
}
