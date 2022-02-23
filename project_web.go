package main

import (
	"fmt"
	"net/http"

	hx "github.com/baralga/htmx"
	"github.com/baralga/paged"
	"github.com/baralga/util"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

type projectFormModel struct {
	CSRFToken string
	ID        string
	Title     string ` validate:"required,min=3,max=50"`
}

func (a *app) HandleProjectsPage() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		pageParams := &paged.PageParams{
			Page: 0,
			Size: 50,
		}

		projects, err := a.ProjectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		if !hx.IsHXRequest(r) {
			pageContext := &pageContext{
				principal:   principal,
				currentPath: r.URL.Path,
				title:       "Projects",
			}

			formModel := projectFormModel{}
			formModel.CSRFToken = csrf.Token(r)

			util.RenderHTML(w, ProjectsPage(pageContext, formModel, projects))
			return
		}

		w.Header().Set("HX-Trigger", "baralga__main_content_modal-show")

		formModel := projectFormModel{}
		formModel.CSRFToken = csrf.Token(r)

		util.RenderHTML(w, ProjectsView(principal, formModel, projects))
	}
}

func (a *app) HandleProjectForm() http.HandlerFunc {
	isProduction := a.isProduction()
	validator := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

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
		_, err = a.CreateProject(r.Context(), principal, &projectToCreate)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		err = a.renderProjectsView(
			w,
			r,
			principal,
			isProduction,
			projectFormModel{},
		)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		w.Header().Set("HX-Trigger", "baralga__projects-changed")
	}
}

func (a *app) renderProjectsView(w http.ResponseWriter, r *http.Request, principal *Principal, isProduction bool, formModel projectFormModel) error {
	pageParams := &paged.PageParams{
		Page: 0,
		Size: 50,
	}

	projects, err := a.ProjectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
	if err != nil {
		return err
	}

	formModel.CSRFToken = csrf.Token(r)

	util.RenderHTML(w, ProjectsView(principal, formModel, projects))

	return nil
}

func ProjectsPage(pageContext *pageContext, formModel projectFormModel, projects *ProjectsPaged) g.Node {
	return Page(
		pageContext.title,
		pageContext.currentPath,
		[]g.Node{
			Navbar(pageContext),
			Section(
				Class("full-center"),
				Div(
					Class("container"),
					Div(
						Class("mt-4 mb-4"),
					),
					ProjectsView(pageContext.principal, formModel, projects),
				),
			),
		},
	)
}

func ProjectsView(principal *Principal, formModel projectFormModel, projects *ProjectsPaged) g.Node {
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
				g.Map(len(projects.Projects), func(i int) g.Node {
					project := projects.Projects[i]
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

		//hx.Swap("innerHTML"),

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
		Title: projectFormModel.Title,
	}
}
