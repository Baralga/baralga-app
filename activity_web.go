package main

import (
	"fmt"
	"net/http"
	"time"

	hx "github.com/baralga/htmx"
	"github.com/baralga/paged"
	"github.com/baralga/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
	"github.com/pkg/errors"
)

type activityFormModel struct {
	CSRFToken   string
	ID          string
	ProjectID   string `validate:"required"`
	Date        string `validate:"required"`
	StartTime   string `validate:"required,min=5,max=5"`
	EndTime     string `validate:"required,min=5,max=5"`
	Description string `validate:"min=0,max=500"`
}

type activityTrackFormModel struct {
	CSRFToken string

	Action   string
	Duration string

	ProjectID    string
	ProjectTitle string
	Date         string
	StartTime    string
	Description  string
}

func newActivityFormModel() activityFormModel {
	now := time.Now()
	return activityFormModel{
		Date:      util.FormatDateDE(now),
		StartTime: util.FormatTime(now),
		EndTime:   util.FormatTime(now),
	}
}

func (a *app) HandleActivityAddPage() http.HandlerFunc {
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

		pageContext := &pageContext{
			principal:   principal,
			currentPath: r.URL.Path,
			title:       "Add Activity",
		}
		activityFormModel := newActivityFormModel()
		activityFormModel.CSRFToken = csrf.Token(r)

		if !hx.IsHXRequest(r) {
			activityFormModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, ActivityAddPage(pageContext, activityFormModel, projects))
			return
		}

		w.Header().Set("HX-Trigger", "baralga__main_content_modal-show")

		activityFormModel.CSRFToken = csrf.Token(r)
		util.RenderHTML(w, ActivityForm(activityFormModel, projects, ""))
	}
}

func (a *app) HandleActivityEditPage() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		activityIDParam := chi.URLParam(r, "activity-id")
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		activityID, err := uuid.Parse(activityIDParam)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		activity, err := a.ActivityRepository.FindActivityByID(r.Context(), activityID, principal.OrganizationID)
		if errors.Is(err, ErrActivityNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		pageParams := &paged.PageParams{
			Page: 0,
			Size: 50,
		}

		projects, err := a.ProjectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		pageContext := &pageContext{
			principal:   principal,
			currentPath: r.URL.Path,
			title:       "Edit Activity",
		}
		formModel := mapActivityToForm(*activity)

		if !hx.IsHXRequest(r) {
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, ActivityAddPage(pageContext, formModel, projects))
			return
		}

		w.Header().Set("HX-Trigger", "baralga__main_content_modal-show")

		formModel.CSRFToken = csrf.Token(r)
		util.RenderHTML(w, ActivityForm(formModel, projects, ""))
	}
}

func (a *app) HandleActivityTrackForm() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		var formModel activityTrackFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		actionParam := r.URL.Query().Get("action")

		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		pageParams := &paged.PageParams{
			Page: 0,
			Size: 50,
		}

		var projects []*Project

		if formModel.Action == "start" {
			projectsPage, err := a.ProjectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
			if err != nil {
				util.RenderProblemHTML(w, isProduction, err)
				return
			}
			projects = projectsPage.Projects

			if actionParam == "reload" {
				util.RenderHTML(w, TrackPanel(projects, formModel))
				return
			}

			now := time.Now()
			formModel.Action = "running"
			formModel.Date = util.FormatDateDE(now)
			formModel.StartTime = util.FormatTime(now)
			formModel.Duration = "0:00 h"

			for _, project := range projectsPage.Projects {
				if project.ID.String() == formModel.ProjectID {
					formModel.ProjectTitle = project.Title
					break
				}
			}

			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, TrackPanel(projects, formModel))
		} else if formModel.Action == "running" {
			projects = []*Project{
				{
					ID:             uuid.MustParse(formModel.ProjectID),
					OrganizationID: principal.OrganizationID,
					Title:          formModel.ProjectTitle,
				},
			}

			activityFormModel := activityFormModel{
				Date:        formModel.Date,
				StartTime:   formModel.StartTime,
				EndTime:     util.FormatTime(time.Now()),
				ProjectID:   formModel.ProjectID,
				Description: formModel.Description,
			}
			activityToCreate, _ := mapFormToActivity(activityFormModel)
			formModel.Duration = activityToCreate.DurationFormatted()

			if actionParam == "reload" {
				util.RenderHTML(w, TrackPanel(projects, formModel))
				return
			}

			_, err := a.CreateActivity(r.Context(), principal, activityToCreate)
			if err != nil {
				util.RenderProblemHTML(w, isProduction, err)
				return
			}

			w.Header().Set("HX-Trigger", "baralga__activities-changed")

			projectsPage, err := a.ProjectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
			if err != nil {
				util.RenderProblemHTML(w, isProduction, err)
				return
			}
			projects = projectsPage.Projects

			formModel = activityTrackFormModel{Action: "start"}
			formModel.CSRFToken = csrf.Token(r)

			util.RenderHTML(w, TrackPanel(projects, formModel))
		}
	}
}

func (a *app) HandleActivityForm() http.HandlerFunc {
	isProduction := a.isProduction()
	validator := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		err := r.ParseForm()
		if err != nil {
			a.renderActivityAddView(
				w,
				r,
				principal,
				isProduction,
				activityFormModel{},
			)
			return
		}

		var formModel activityFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			a.renderActivityAddView(
				w,
				r,
				principal,
				isProduction,
				activityFormModel{},
			)
			return
		}

		err = validator.Struct(formModel)
		if err != nil {
			a.renderActivityAddView(
				w,
				r,
				principal,
				isProduction,
				formModel,
			)
			return
		}

		activityNew, err := mapFormToActivity(formModel)
		if err != nil {
			a.renderActivityAddView(
				w,
				r,
				principal,
				isProduction,
				formModel,
			)
			return
		}

		if uuid.Nil == activityNew.ID {
			_, err = a.CreateActivity(r.Context(), principal, activityNew)
		} else {
			_, err = a.UpdateActivity(r.Context(), principal, activityNew)
		}
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		w.Header().Set("HX-Trigger", "{ \"baralga__activities-changed\": true, \"baralga__main_content_modal-hide\": true } ")
	}
}

func (a *app) HandleStartTimeValidation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			return
		}

		var formModel activityFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			return
		}

		formModel.StartTime = util.CompleteTimeValue(formModel.StartTime)

		util.RenderHTML(w, StartTimeInputView(formModel))
	}
}

func (a *app) HandleEndTimeValidation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			return
		}

		var formModel activityFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			return
		}

		formModel.EndTime = util.CompleteTimeValue(formModel.EndTime)

		util.RenderHTML(w, EndTimeInputView(formModel))
	}
}

func ActivityAddPage(pageContext *pageContext, activityFormModel activityFormModel, projects *ProjectsPaged) g.Node {
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
					ActivityForm(activityFormModel, projects, ""),
				),
			),
		},
	)
}

func StartTimeInputView(formModel activityFormModel) g.Node {
	return Div(
		ID("activity_start_time"),
		Class("mb-3"),
		Label(
			Class("form-label"),
			g.Attr("for", "StartTime"),
			g.Text("Start Time"),
		),
		Input(
			ID("StartTime"),
			Type("text"),
			Name("StartTime"),
			hx.Target("#activity_start_time"),
			hx.Post("/activities/validate-start-time"),
			Value(formModel.StartTime),
			Pattern("[0-9]{2}:[0-5][0-9]"),
			MinLength("5"),
			MaxLength("5"),
			g.Attr("required", "required"),
			Class("form-control"),
			g.Attr("placeholder", "10:00"),
		),
	)
}

func EndTimeInputView(formModel activityFormModel) g.Node {
	return Div(
		ID("activity_end_time"),
		Class("mb-3"),
		Label(
			Class("form-label"),
			g.Attr("for", "EndTime"),
			g.Text("End Time"),
		),
		Input(
			ID("EndTime"),
			Type("text"),
			Name("EndTime"),
			hx.Target("#activity_end_time"),
			hx.Post("/activities/validate-end-time"),
			Value(formModel.EndTime),
			Pattern("[0-9]{2}:[0-5][0-9]"),
			MinLength("5"),
			MaxLength("5"),
			g.Attr("required", "required"),
			Class("form-control"),
			g.Attr("placeholder", "10:00"),
		),
	)
}

func ActivityForm(formModel activityFormModel, projects *ProjectsPaged, errorMessage string) g.Node {
	isEditMode := formModel.ID != ""
	return FormEl(
		ID("baralga__main_content_modal_content"),
		Class("modal-content"),

		g.If(formModel.ID == "",
			hx.Post("/activities/new"),
		),
		g.If(formModel.ID != "",
			hx.Post(fmt.Sprintf("/activities/%v", formModel.ID)),
		),

		Div(
			Class("modal-header"),
			H2(
				Class("modal-title"),
				g.If(isEditMode, g.Text("Edit Activity")),
				g.If(!isEditMode, g.Text("Add Activity")),
			),
			A(
				g.Attr("data-bs-dismiss", "modal"),
				Class("btn-close"),
			),
		),
		Div(
			Class("modal-body"),
			g.If(
				errorMessage != "",
				Div(
					Class("alert alert-danger text-center"),
					Role("alert"),
					Span(g.Text(errorMessage)),
				),
			),
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
				Class("mb-3"),
				Label(
					Class("form-label"),
					g.Attr("for", "ProjectID"),
					g.Text("Project"),
				),
				Select(
					Class("form-select"),
					ID("ProjectID"),
					Name("ProjectID"),
					g.Group(
						g.Map(len(projects.Projects), func(i int) g.Node {
							project := projects.Projects[i]
							return Option(
								Value(project.ID.String()),
								g.Text(project.Title),
								g.If(formModel.ProjectID == project.ID.String(), Selected()),
							)
						}),
					),
					Value(formModel.ProjectID),
				),
			),
			Div(
				Class("mb-3"),
				Label(
					Class("form-label"),
					g.Attr("for", "Date"),
					g.Text("Date"),
				),
				Input(
					ID("Date"),
					Type("text"),
					Name("Date"),
					Value(formModel.Date),
					Pattern("[0-3][0-9]\\.[0-1][0-9]\\.20[0-9]{2}"),
					MinLength("10"),
					MaxLength("10"),
					g.Attr("required", "required"),
					Class("form-control"),
					g.Attr("placeholder", "16.11.2021"),
				),
			),
			StartTimeInputView(formModel),
			EndTimeInputView(formModel),
			Div(
				Class("mb-3"),
				Label(
					Class("form-label"),
					g.Attr("for", "Description"),
					g.Text("Description"),
				),
				Textarea(
					ID("Description"),
					Type("text"),
					Name("Description"),
					Class("form-control"),
					g.Attr("placeholder", "Describe what you do ..."),
					g.Text(formModel.Description),
				),
			),
		),
		Div(
			Class("modal-footer"),
			Button(
				Type("submit"),
				Class("text-center btn btn-primary"),

				g.If(isEditMode, I(Class("bi-save me-2"))),
				g.If(!isEditMode, I(Class("bi-plus me-2"))),

				g.If(isEditMode, g.Text("Update")),
				g.If(!isEditMode, g.Text("Add")),
			),
			A(
				g.Attr("data-bs-dismiss", "modal"),
				Class("text-center btn btn-secondary"),
				I(Class("bi-x me-2")),
				g.Text("Cancel"),
			),
		),
	)
}

func TrackPanel(projects []*Project, formModel activityTrackFormModel) g.Node {
	return FormEl(
		ID("baralga__track_panel"),
		Class("container p-3 rounded-3"),
		StyleAttr("background-color: #14142B"),

		hx.Target("#baralga__track_panel"),
		hx.Swap("outerHTML"),
		hx.Post("/activities/track"),

		Input(
			Type("hidden"),

			hx.Target("#baralga__track_panel"),
			hx.Swap("outerHTML"),
			hx.Trigger("baralga__projects-changed from:body"),
			hx.Post("/activities/track?action=reload"),
		),

		Input(
			Type("hidden"),
			Name("CSRFToken"),
			Value(formModel.CSRFToken),
		),

		g.If(formModel.Action == "running",
			Input(
				Type("hidden"),

				hx.Target("#baralga__track_panel"),
				hx.Swap("outerHTML"),
				hx.Trigger("every 60s"),
				hx.Post("/activities/track?action=reload"),
			),
		),

		g.If(formModel.Action == "running",
			Input(
				Type("hidden"),
				Name("Date"),
				Value(formModel.Date),
			),
		),
		g.If(formModel.Action == "running",
			Input(
				Type("hidden"),
				Name("StartTime"),
				Value(formModel.StartTime),
			),
		),
		g.If(formModel.Action == "running",
			Input(
				Type("hidden"),
				Name("ProjectID"),
				Value(formModel.ProjectID),
			),
		),
		g.If(formModel.Action == "running",
			Input(
				Type("hidden"),
				Name("ProjectTitle"),
				Value(formModel.ProjectTitle),
			),
		),

		Input(
			Type("hidden"),
			Name("Action"),
			Value(formModel.Action),
		),

		Div(
			Class("row"),
			Div(
				Class("col-sm-3"),
				g.If(formModel.Action != "running",
					Button(
						Type("submit"),
						Class("btn btn-primary btn-lg"),
						StyleAttr("width: 100%"),
						I(Class("bi-play")),
					),
				),
				g.If(formModel.Action == "running",
					Button(
						Type("submit"),
						Class("btn btn-danger btn-lg bg-danger progress-bar progress-bar-striped progress-bar-animated"),
						StyleAttr("width: 100%"),
						I(Class("bi-stop")),
					),
				),
			),
			Div(
				Class("col-sm-9"),
				StyleAttr("padding-left: 0"),
				Select(
					Class("form-select form-select-lg"),
					Name("ProjectID"),
					g.If(formModel.Action == "running",
						Disabled(),
					),
					g.Group(
						g.Map(len(projects), func(i int) g.Node {
							project := projects[i]
							return Option(
								Value(project.ID.String()),
								g.Text(project.Title),
								g.If(formModel.ProjectID == project.ID.String(), Selected()),
							)
						}),
					),
				),
			),
		),
		g.If(formModel.Action == "running",
			Div(
				Class("row mt-2"),
				g.If(formModel.Action == "running",
					Div(
						Class("col-sm-3 text-center text-muted"),
						Span(g.Text(formModel.StartTime)),
					),
				),
				g.If(formModel.Action == "running",
					Div(
						Class("col-sm-9 ps-3 text-muted"),
						Span(g.Text(formModel.Duration)),
					),
				),
			),
		),
		g.If(formModel.Action == "running",
			Div(
				Class("row mt-2"),
				Div(
					Class("col-sm-12"),
					Textarea(
						Name("Description"),
						Class("form-control"),
						g.Attr("placeholder", "I work on ..."),
						g.Text(formModel.Description),
					),
				),
			),
		),
	)
}

func (a *app) renderActivityAddView(w http.ResponseWriter, r *http.Request, principal *Principal, isProduction bool, formModel activityFormModel) {
	pageParams := &paged.PageParams{
		Page: 0,
		Size: 50,
	}

	projects, err := a.ProjectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
	if err != nil {
		util.RenderProblemHTML(w, isProduction, err)
		return
	}

	if hx.IsHXRequest(r) {
		formModel.CSRFToken = csrf.Token(r)
		util.RenderHTML(w, ActivityForm(formModel, projects, ""))
		return
	}

	pageContext := &pageContext{
		principal:   principal,
		currentPath: r.URL.Path,
		title:       "Add Activity",
	}

	activityFormModel := newActivityFormModel()
	activityFormModel.CSRFToken = csrf.Token(r)

	util.RenderHTML(w, ActivityAddPage(pageContext, activityFormModel, projects))
}

func mapFormToActivity(formModel activityFormModel) (*Activity, error) {
	var activityID uuid.UUID

	if formModel.ID != "" {
		aID, err := uuid.Parse(formModel.ID)
		if err != nil {
			return nil, err
		}
		activityID = aID
	}

	start, err := util.ParseDateTimeForm(fmt.Sprintf("%v %v", formModel.Date, formModel.StartTime))
	if err != nil {
		return nil, err
	}

	end, err := util.ParseDateTimeForm(fmt.Sprintf("%v %v", formModel.Date, formModel.EndTime))
	if err != nil {
		return nil, err
	}

	projectID, err := uuid.Parse(formModel.ProjectID)
	if err != nil {
		return nil, err
	}

	activity := &Activity{
		ID:          activityID,
		Start:       *start,
		End:         *end,
		ProjectID:   projectID,
		Description: formModel.Description,
	}

	return activity, nil
}

func mapActivityToForm(activity Activity) activityFormModel {
	return activityFormModel{
		ID:          activity.ID.String(),
		Date:        util.FormatDateDE(activity.Start),
		StartTime:   util.FormatTime(activity.Start),
		EndTime:     util.FormatTime(activity.End),
		ProjectID:   activity.ProjectID.String(),
		Description: activity.Description,
	}
}
