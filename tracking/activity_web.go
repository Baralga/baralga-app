package tracking

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/baralga/shared"
	"github.com/baralga/shared/hx"
	"github.com/baralga/shared/paged"
	time_utils "github.com/baralga/shared/time"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	g "github.com/maragudk/gomponents"
	ghx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/html"
	"github.com/pkg/errors"
	"github.com/snabb/isoweek"
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

type ActivityWeb struct {
	config             *shared.Config
	activityService    *ActitivityService
	activityRepository ActivityRepository
	projectRepository  ProjectRepository
}

func NewActivityWeb(config *shared.Config, activityService *ActitivityService, activityRepository ActivityRepository, projectRepository ProjectRepository) *ActivityWeb {
	return &ActivityWeb{
		config:             config,
		activityService:    activityService,
		activityRepository: activityRepository,
		projectRepository:  projectRepository,
	}
}

func (a *ActivityWeb) RegisterProtected(r chi.Router) {
	r.Get("/", a.HandleTrackingPage())
	r.Get("/activities/new", a.HandleActivityAddPage())
	r.Post("/activities/validate-start-time", a.HandleStartTimeValidation())
	r.Post("/activities/validate-end-time", a.HandleEndTimeValidation())
	r.Get("/activities/{activity-id}/edit", a.HandleActivityEditPage())
	r.Post("/activities/new", a.HandleActivityForm())
	r.Post("/activities/{activity-id}", a.HandleActivityForm())
	r.Post("/activities/track", a.HandleActivityTrackForm())
}

func (a *ActivityWeb) RegisterOpen(r chi.Router) {
}

func newActivityFormModel() activityFormModel {
	now := time.Now()
	return activityFormModel{
		Date:      time_utils.FormatDateDE(now),
		StartTime: time_utils.FormatTime(now),
		EndTime:   time_utils.FormatTime(now),
	}
}

func (a *ActivityWeb) HandleTrackingPage() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	activityService := a.activityService
	projectRepository := a.projectRepository
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		wyear, week := isoweek.FromDate(now.Year(), now.Month(), now.Day())
		filter := &ActivityFilter{
			Timespan: TimespanWeek,
			start:    isoweek.StartTime(wyear, week, time.UTC),
		}
		pageParams := &paged.PageParams{
			Page: 0,
			Size: 100,
		}

		principal := r.Context().Value(shared.ContextKeyPrincipal).(*shared.Principal)
		activitiesPage, projectsOfActivities, err := activityService.ReadActivitiesWithProjects(
			r.Context(),
			principal,
			filter,
			pageParams,
		)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		projects, err := projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		if hx.IsHXTargetRequest(r, "baralga__main_content") {
			shared.RenderHTML(w, Div(ActivitiesInWeekView(filter, activitiesPage, projectsOfActivities)))
			return
		}

		pageContext := &shared.PageContext{
			Principal:   principal,
			CurrentPath: r.URL.Path,
		}

		formModel := activityTrackFormModel{Action: "start"}
		formModel.CSRFToken = csrf.Token(r)

		shared.RenderHTML(w, TrackingPage(pageContext, formModel, filter, activitiesPage, projectsOfActivities, projects))
	}
}

func (a *ActivityWeb) HandleActivityAddPage() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	projectRepository := a.projectRepository
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(shared.ContextKeyPrincipal).(*shared.Principal)

		pageParams := &paged.PageParams{
			Page: 0,
			Size: 50,
		}

		projects, err := projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		pageContext := &shared.PageContext{
			Principal:   principal,
			CurrentPath: r.URL.Path,
			Title:       "Add Activity",
		}
		activityFormModel := newActivityFormModel()
		activityFormModel.CSRFToken = csrf.Token(r)

		if !hx.IsHXRequest(r) {
			activityFormModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, ActivityAddPage(pageContext, activityFormModel, projects))
			return
		}

		w.Header().Set("HX-Trigger", "baralga__main_content_modal-show")

		activityFormModel.CSRFToken = csrf.Token(r)
		shared.RenderHTML(w, ActivityForm(activityFormModel, projects, ""))
	}
}

func (a *ActivityWeb) HandleActivityEditPage() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	activityRepository := a.activityRepository
	projectRepository := a.projectRepository
	return func(w http.ResponseWriter, r *http.Request) {
		activityIDParam := chi.URLParam(r, "activity-id")
		principal := r.Context().Value(shared.ContextKeyPrincipal).(*shared.Principal)

		activityID, err := uuid.Parse(activityIDParam)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		activity, err := activityRepository.FindActivityByID(r.Context(), activityID, principal.OrganizationID)
		if errors.Is(err, ErrActivityNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		pageParams := &paged.PageParams{
			Page: 0,
			Size: 50,
		}

		projects, err := projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		pageContext := &shared.PageContext{
			Principal:   principal,
			CurrentPath: r.URL.Path,
			Title:       "Edit Activity",
		}
		formModel := mapActivityToForm(*activity)

		if !hx.IsHXRequest(r) {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, ActivityAddPage(pageContext, formModel, projects))
			return
		}

		w.Header().Set("HX-Trigger", "baralga__main_content_modal-show")

		formModel.CSRFToken = csrf.Token(r)
		shared.RenderHTML(w, ActivityForm(formModel, projects, ""))
	}
}

func (a *ActivityWeb) HandleActivityTrackForm() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	activityService := a.activityService
	projectRepository := a.projectRepository
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		var formModel activityTrackFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		actionParam := r.URL.Query().Get("action")

		principal := r.Context().Value(shared.ContextKeyPrincipal).(*shared.Principal)

		pageParams := &paged.PageParams{
			Page: 0,
			Size: 50,
		}

		var projects []*Project

		if formModel.Action == "start" {
			projectsPage, err := projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
			if err != nil {
				shared.RenderProblemHTML(w, isProduction, err)
				return
			}
			projects = projectsPage.Projects

			if actionParam == "reload" {
				shared.RenderHTML(w, TrackPanel(projects, formModel))
				return
			}

			now := time.Now()
			formModel.Action = "running"
			formModel.Date = time_utils.FormatDateDE(now)
			formModel.StartTime = time_utils.FormatTime(now)
			formModel.Duration = "0:00 h"

			for _, project := range projectsPage.Projects {
				if project.ID.String() == formModel.ProjectID {
					formModel.ProjectTitle = project.Title
					break
				}
			}

			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, TrackPanel(projects, formModel))
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
				EndTime:     time_utils.FormatTime(time.Now()),
				ProjectID:   formModel.ProjectID,
				Description: formModel.Description,
			}
			activityToCreate, _ := mapFormToActivity(activityFormModel)
			formModel.Duration = activityToCreate.DurationFormatted()

			if actionParam == "reload" {
				shared.RenderHTML(w, TrackPanel(projects, formModel))
				return
			}

			_, err := activityService.CreateActivity(r.Context(), principal, activityToCreate)
			if err != nil {
				shared.RenderProblemHTML(w, isProduction, err)
				return
			}

			w.Header().Set("HX-Trigger", "baralga__activities-changed")

			projectsPage, err := projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
			if err != nil {
				shared.RenderProblemHTML(w, isProduction, err)
				return
			}
			projects = projectsPage.Projects

			formModel = activityTrackFormModel{Action: "start"}
			formModel.CSRFToken = csrf.Token(r)

			shared.RenderHTML(w, TrackPanel(projects, formModel))
		}
	}
}

func (a *ActivityWeb) HandleActivityForm() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validator := validator.New()
	activityService := a.activityService
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(shared.ContextKeyPrincipal).(*shared.Principal)

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
			_, err = activityService.CreateActivity(r.Context(), principal, activityNew)
		} else {
			_, err = activityService.UpdateActivity(r.Context(), principal, activityNew)
		}
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, err)
			return
		}

		w.Header().Set("HX-Trigger", "{ \"baralga__activities-changed\": true, \"baralga__main_content_modal-hide\": true } ")
	}
}

func (a *ActivityWeb) HandleStartTimeValidation() http.HandlerFunc {
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

		formModel.StartTime = time_utils.CompleteTimeValue(formModel.StartTime)

		shared.RenderHTML(w, StartTimeInputView(formModel))
	}
}

func (a *ActivityWeb) HandleEndTimeValidation() http.HandlerFunc {
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

		formModel.EndTime = time_utils.CompleteTimeValue(formModel.EndTime)

		shared.RenderHTML(w, EndTimeInputView(formModel))
	}
}

func TrackingPage(pageContext *shared.PageContext, formModel activityTrackFormModel, filter *ActivityFilter, activitiesPage *ActivitiesPaged, projectsOfActivities []*Project, projects *ProjectsPaged) g.Node {
	return shared.Page(
		"Track Activities",
		pageContext.CurrentPath,
		[]g.Node{
			shared.Navbar(pageContext),
			Div(
				Class("container"),
				Div(
					Class("row"),
					Div(
						ID("baralga__main_content"),
						Class("col-lg-8 col-sm-12 mb-2 order-2 order-lg-1 mt-lg-4 mt-2"),

						ghx.Target("#baralga__main_content"),
						ghx.Swap("innerHTML"),

						ghx.Trigger("baralga__activities-changed from:body"),
						ghx.Get("/"),

						ActivitiesInWeekView(filter, activitiesPage, projectsOfActivities),
					),
					Div(Class("col-lg-4 col-sm-12 order-1 order-lg-2 mt-lg-4 mt-2"),
						TrackPanel(projects.Projects, formModel),
					),
				),
			),
			shared.ModalView(),
		},
	)
}

func ActivitiesInWeekView(filter *ActivityFilter, activitiesPage *ActivitiesPaged, projects []*Project) g.Node {
	// prepare projects
	projectsById := make(map[uuid.UUID]*Project)
	for _, project := range projects {
		projectsById[project.ID] = project
	}

	var durationWeekTotal float64
	for _, activity := range activitiesPage.Activities {
		durationWeekTotal = durationWeekTotal + float64(activity.DurationMinutesTotal())
	}

	nodes := []g.Node{
		Div(
			Class("mb-2 d-flex"),
			Div(
				Class("flex-fill"),
				H2(
					Span(
						StyleAttr("white-space: nowrap;"),

						g.Text(
							filter.StringFormatted(),
						),
					),
					Br(
						Class("d-block d-md-none"),
					),
					Span(
						Class("ms-4 d-none d-md-inline"),
					),
					Small(
						StyleAttr("white-space: nowrap;"),
						Class("text-muted"),
						g.Text("My Week "),
						g.If(len(activitiesPage.Activities) > 0,
							Span(
								Class("badge rounded-pill bg-secondary fw-normal"),
								g.Text(FormatMinutesAsDuration(durationWeekTotal)),
							),
						),
					),
				),
			),
			Div(
				A(
					ghx.Target("#baralga__main_content_modal_content"),
					ghx.Trigger("click, keyup[altKey && shiftKey && key == 'P'] from:body"),
					ghx.Swap("outerHTML"),
					ghx.Get("/projects"),
					Class("btn btn-outline-primary btn-sm ms-1"),
					I(Class("bi-card-list")),
					TitleAttr("Manage Projects"),
				),
			),
			Div(
				A(
					ghx.Target("#baralga__main_content_modal_content"),
					ghx.Trigger("click, keyup[altKey && shiftKey && key == 'N'] from:body"),
					ghx.Swap("outerHTML"),
					ghx.Get("/activities/new"),
					Class("btn btn-outline-primary btn-sm ms-1"),
					I(Class("bi-plus")),
					TitleAttr("Add Activity"),
				),
			),
		),
		ActivitiesSumByDayView(activitiesPage, projects),
		g.If(
			len(activitiesPage.Activities) == 0,
			Div(
				Class("alert alert-info"),
				Role("alert"),
				g.Text("No activities in current week. Add some "),
				A(
					Href("#"),
					Class("info-link"),
					ghx.Target("#baralga__main_content_modal_content"),
					ghx.Swap("outerHTML"),
					ghx.Get("/activities/new"),
					g.Text("here"),
				),
				g.Text("!"),
			),
		),
	}
	return g.Group(nodes)
}

func ActivitiesSumByDayView(activitiesPage *ActivitiesPaged, projects []*Project) g.Node {
	// prepare projects
	projectsById := make(map[uuid.UUID]*Project)
	for _, project := range projects {
		projectsById[project.ID] = project
	}

	// prepare activities
	activitySumByDay := make(map[int]float64)
	activitiesByDay := make(map[int][]*Activity)
	dayFormattedByDay := make(map[int][]string)
	for _, activity := range activitiesPage.Activities {
		day := activity.Start.Day()
		dayFormattedByDay[day] = []string{
			activity.Start.Format("Monday"),
			time_utils.FormatDateDEShort(activity.Start),
			time_utils.FormatDate(activity.Start),
		}
		activitySumByDay[day] = activitySumByDay[day] + float64(activity.DurationMinutesTotal())
		activitiesByDay[day] = append(activitiesByDay[day], activity)
	}

	var dayNodes []int
	for day := range activitySumByDay {
		dayNodes = append(dayNodes, day)
	}

	sort.Slice(dayNodes, func(i, j int) bool { return dayNodes[i] > dayNodes[j] })

	today := time.Now().Day()

	return g.Group(g.Map(dayNodes, func(i int) g.Node {
		activities := activitiesByDay[i]
		activityCardID := fmt.Sprintf("baralga__activity_card_%v", dayFormattedByDay[i][2])

		sum := activitySumByDay[i]
		durationFormatted := FormatMinutesAsDuration(sum)

		return Div(
			ID(activityCardID),
			Class("card mb-4 me-1"),

			ghx.Target("this"),
			ghx.Swap("outerHTML"),

			Div(
				Class("card-body position-relative p-2 pt-1"),
				g.If(today == i,
					StyleAttr("background-color: rgba(255, 255,255, 0.05);"),
				),
				H6(
					Class("card-subtitle mt-2"),
					Div(
						Class("d-flex justify-content-between mb-2"),
						Div(
							Class("text-muted"),
							Span(
								g.Text(dayFormattedByDay[i][0]),
							),
							Span(
								Class("ms-2"),
								StyleAttr("opacity: .45; font-size: 80%;"),
								g.Text(dayFormattedByDay[i][1]),
							),
						),
					),
					Span(
						Class("position-absolute top-0 start-100 translate-middle badge rounded-pill bg-secondary"),
						g.Text(durationFormatted),
					),
				),
				g.Group(g.Map(activities, func(activity *Activity) g.Node {
					return Div(
						Class("d-flex justify-content-between mb-2"),
						ghx.Target(fmt.Sprintf("#%v", activityCardID)),
						TitleAttr(activity.Description),
						Span(
							Class("flex-fill"),
							g.Text(time_utils.FormatTime(activity.Start)+" - "+time_utils.FormatTime(activity.End)),
						),
						Span(
							Class("flex-fill"),
							g.Text(projectsById[activity.ProjectID].Title),
						),
						Span(
							Class("flex-fill text-end pe-3"),
							g.Text(activity.DurationFormatted()),
						),
						Div(
							A(
								ghx.Get(fmt.Sprintf("/activities/%v/edit", activity.ID)),
								ghx.Target("#baralga__main_content_modal_content"),
								ghx.Swap("outerHTML"),

								Class("btn btn-outline-secondary btn-sm"),
								I(Class("bi-pen")),
							),
							A(
								ghx.Confirm(
									fmt.Sprintf(
										"Do you really want to delete the activity from %v on %v?",
										time_utils.FormatTime(activity.Start),
										activity.Start.Format("Monday"),
									),
								),
								ghx.Delete(fmt.Sprintf("/api/activities/%v", activity.ID)),
								Class("btn btn-outline-secondary btn-sm ms-1"),
								I(Class("bi-trash2")),
							),
						),
					)
				})),
			),
		)
	}))
}

func ActivityAddPage(pageContext *shared.PageContext, activityFormModel activityFormModel, projects *ProjectsPaged) g.Node {
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
			ghx.Target("#activity_start_time"),
			ghx.Post("/activities/validate-start-time"),
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
			ghx.Target("#activity_end_time"),
			ghx.Post("/activities/validate-end-time"),
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
			ghx.Post("/activities/new"),
		),
		g.If(formModel.ID != "",
			ghx.Post(fmt.Sprintf("/activities/%v", formModel.ID)),
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
						g.Map(projects.Projects, func(project *Project) g.Node {
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

		ghx.Target("#baralga__track_panel"),
		ghx.Swap("outerHTML"),
		ghx.Post("/activities/track"),

		Input(
			Type("hidden"),

			ghx.Target("#baralga__track_panel"),
			ghx.Swap("outerHTML"),
			ghx.Trigger("baralga__projects-changed from:body"),
			ghx.Post("/activities/track?action=reload"),
		),

		Input(
			Type("hidden"),
			Name("CSRFToken"),
			Value(formModel.CSRFToken),
		),

		g.If(formModel.Action == "running",
			Input(
				Type("hidden"),

				ghx.Target("#baralga__track_panel"),
				ghx.Swap("outerHTML"),
				ghx.Trigger("every 60s"),
				ghx.Post("/activities/track?action=reload"),
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
			Class("row g-2"),
			Div(
				Class("col-lg-3 col-sm-12"),
				g.If(formModel.Action != "running",
					Button(
						Type("submit"),
						Class("btn btn-primary btn-lg"),
						Alt("Start tracking"),
						StyleAttr("width: 100%"),
						I(Class("bi-play")),
					),
				),
				g.If(formModel.Action == "running",
					Button(
						Type("submit"),
						Class("btn btn-danger btn-lg bg-danger progress-bar-striped progress-bar-animated"),
						StyleAttr("width: 100%"),
						Alt("Stop tracking"),
						I(Class("bi-stop")),
					),
				),
			),
			Div(
				Class("col-lg-9 col-sm-12"),
				Select(
					Class("form-select form-select-lg"),
					Name("ProjectID"),
					g.If(formModel.Action == "running",
						Disabled(),
					),
					g.Group(
						g.Map(projects, func(project *Project) g.Node {
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

func (a *ActivityWeb) renderActivityAddView(w http.ResponseWriter, r *http.Request, principal *shared.Principal, isProduction bool, formModel activityFormModel) {
	pageParams := &paged.PageParams{
		Page: 0,
		Size: 50,
	}

	projects, err := a.projectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
	if err != nil {
		shared.RenderProblemHTML(w, isProduction, err)
		return
	}

	if hx.IsHXRequest(r) {
		formModel.CSRFToken = csrf.Token(r)
		shared.RenderHTML(w, ActivityForm(formModel, projects, ""))
		return
	}

	pageContext := &shared.PageContext{
		Principal:   principal,
		CurrentPath: r.URL.Path,
		Title:       "Add Activity",
	}

	activityFormModel := newActivityFormModel()
	activityFormModel.CSRFToken = csrf.Token(r)

	shared.RenderHTML(w, ActivityAddPage(pageContext, activityFormModel, projects))
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

	start, err := time_utils.ParseDateTimeForm(fmt.Sprintf("%v %v", formModel.Date, formModel.StartTime))
	if err != nil {
		return nil, err
	}

	end, err := time_utils.ParseDateTimeForm(fmt.Sprintf("%v %v", formModel.Date, formModel.EndTime))
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
		Date:        time_utils.FormatDateDE(activity.Start),
		StartTime:   time_utils.FormatTime(activity.Start),
		EndTime:     time_utils.FormatTime(activity.End),
		ProjectID:   activity.ProjectID.String(),
		Description: activity.Description,
	}
}
