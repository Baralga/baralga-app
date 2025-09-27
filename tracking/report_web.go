package tracking

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/baralga/shared"
	"github.com/baralga/shared/hx"
	"github.com/baralga/shared/paged"
	time_utils "github.com/baralga/tracking/time"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	g "maragu.dev/gomponents"
	ghx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html" //nolint:all
)

type ReportWeb struct {
	config          *shared.Config
	activityService *ActitivityService
}

func NewReportWebHandlers(config *shared.Config, activityService *ActitivityService) *ReportWeb {
	return &ReportWeb{
		config:          config,
		activityService: activityService,
	}
}

func (a *ReportWeb) RegisterProtected(r chi.Router) {
	r.Get("/reports", a.HandleReportPage())
}

func (a *ReportWeb) RegisterOpen(r chi.Router) {
}

func (a *ReportWeb) HandleReportPage() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := shared.MustPrincipalFromContext(r.Context())
		pageContext := &shared.PageContext{
			Ctx:          r.Context(),
			Principal:    principal,
			CurrentPath:  r.URL.Path,
			CurrentQuery: r.URL.Query(),
			Title:        "Report Activities",
		}

		queryParams := r.URL.Query()
		filter, err := filterFromQueryParams(queryParams)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, errors.New("invalid query params"))
			return
		}

		view := reportViewFromQueryParams(queryParams, filter.Timespan)

		reportView, err := a.ReportView(pageContext, view, filter)
		if err != nil {
			shared.RenderProblemHTML(w, isProduction, errors.New("invalid reports"))
			return
		}

		if hx.IsHXTargetRequest(r, "baralga__report_content") {
			shared.RenderHTML(w, reportView)
			return
		}

		shared.RenderHTML(w, a.ReportPage(pageContext, reportView))
	}
}

func (a *ReportWeb) ReportView(pageContext *shared.PageContext, view *reportView, filter *ActivityFilter) (g.Node, error) {
	previousFilter := filter.Previous()
	homeFilter := filter.Home()
	nextFilter := filter.Next()

	var reportGeneralView, reportTimeView, reportProjectView, reportTagView g.Node
	var err error
	if view.main == "general" {
		reportGeneralView, err = a.reportGeneralView(pageContext, filter, view)
		if err != nil {
			return nil, err
		}
	}
	if view.main == "time" {
		reportTimeView, err = a.reportTimeView(pageContext, view, filter)
		if err != nil {
			return nil, err
		}
	}
	if view.main == "project" {
		reportProjectView, err = a.reportProjectView(pageContext, view, filter)
		if err != nil {
			return nil, err
		}
	}
	if view.main == "tag" {
		reportTagView, err = a.reportTagView(pageContext, view, filter)
		if err != nil {
			return nil, err
		}
	}

	return Div(
		ID("baralga__report_content"),
		Class("container mt-lg-2"),

		ghx.Trigger("baralga__activities-changed from:body"),
		ghx.Get(reportHref(filter, view)),
		ghx.Target("#baralga__report_content"),
		ghx.Swap("outerHTML"),

		Div(
			Class("row mb-2"),
			Div(
				Class("col-md-4 col-12 mt-2"),
				Select(
					ghx.Get(fmt.Sprintf("/reports?c=%v", view.asParam())),
					ghx.PushURL("true"),
					ghx.Target("#baralga__report_content"),
					ghx.Swap("outerHTML"),

					Name("t"),
					Class("form-select"),
					Option(
						Value("day"),
						g.Text("Day"),
						g.If(filter.Timespan == "day", Selected()),
					),
					Option(
						Value("week"),
						g.Text("Week"),
						g.If(filter.Timespan == "week", Selected()),
					),
					Option(
						Value("month"),
						g.Text("Month"),
						g.If(filter.Timespan == "month", Selected()),
					),
					Option(
						Value("quarter"),
						g.Text("Quarter"),
						g.If(filter.Timespan == "quarter", Selected()),
					),
					Option(
						Value("year"),
						g.Text("Year"),
						g.If(filter.Timespan == "year", Selected()),
					),
				),
			),
			Div(
				Class("col-md-4 col-6 text-center mt-2"),
				Div(
					Class("btn-group"),
					Role("group"),
					A(
						ghx.Get(reportHref(previousFilter, view)),
						ghx.PushURL("true"),
						ghx.Target("#baralga__report_content"),
						ghx.Trigger("click, keyup[shiftKey && key == 'ArrowLeft'] from:body"),
						ghx.Swap("outerHTML"),

						TitleAttr(fmt.Sprintf("Show previous actvities from %v", previousFilter.String())),
						Class("btn btn-outline-primary"),
						I(Class("bi-arrow-left")),
					),
					A(
						ghx.Get(reportHref(homeFilter, view)),
						ghx.PushURL("true"),
						ghx.Target("#baralga__report_content"),
						ghx.Trigger("click, keyup[shiftKey && key == 'ArrowDown'] from:body"),
						ghx.Swap("outerHTML"),

						TitleAttr(fmt.Sprintf("Show current actvities from %v", homeFilter.String())),
						Class("btn btn-outline-primary"),
						I(Class("bi-house-fill")),
					),
					A(
						ghx.Get(reportHref(nextFilter, view)),
						ghx.PushURL("true"),
						ghx.Target("#baralga__report_content"),
						ghx.Trigger("click, keyup[shiftKey && key == 'ArrowRight'] from:body"),
						ghx.Swap("outerHTML"),

						TitleAttr(fmt.Sprintf("Show next actvities from %v", nextFilter.String())),
						Class("btn btn-outline-primary"),
						I(Class("bi-arrow-right")),
					),
				),
			),
			Div(
				Class("col-md-3 col-3 mt-2"),
				H5(
					Class("text-muted"),
					Span(
						g.Text(filter.String()),
					),
					g.If(filter.Timespan != TimespanDay,
						Span(
							Class("ms-4 d-none d-lg-inline"),
							g.Text(filter.StringFormatted()),
						),
					),
				),
			),
			Div(
				Class("col-1 text-end mt-2"),
				A(
					Href(
						fmt.Sprintf("/api/activities?contentType=application/vnd.ms-excel&t=%v&v=%v", filter.Timespan, filter.String()),
					),
					Class("btn btn-outline-primary"),
					I(Class("bi-file-excel")),
					TitleAttr("Export Activities"),
				),
			),
		),

		Div(
			Class("row mb-lg-4 mb-2"),
			Div(
				Class("col d-flex justify-content-center"),
				Nav(
					Class("nav nav-pills"),
					A(
						g.If(view.main == "general",
							Class("nav-link active"),
						),
						g.If(view.main != "general",
							g.Group([]g.Node{
								Class("btn nav-link"),
								ghx.Get(reportHrefForView(filter, "general", "")),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),
							}),
						),
						I(Class("bi-list me-2")),
						g.Text("General"),
					),
					A(
						g.If(view.main == "time",
							Class("nav-link active"),
						),
						g.If(view.main != "time",
							g.Group([]g.Node{
								Class("btn nav-link"),
								ghx.Get(reportHrefForView(filter, "time", "d")),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),
							}),
						),
						I(Class("bi-clock me-2")),
						g.Text("Time"),
						Class("nav-link"),
					),
					A(
						g.If(view.main == "project",
							Class("nav-link active"),
						),
						g.If(view.main != "project",
							g.Group([]g.Node{
								Class("btn nav-link"),
								ghx.Get(reportHrefForView(filter, "project", "d")),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),
							}),
						),
						I(Class("bi-pie-chart me-2")),
						g.Text("Project"),
						Class("nav-link"),
					),
					A(
						g.If(view.main == "tag",
							Class("nav-link active"),
						),
						g.If(view.main != "tag",
							g.Group([]g.Node{
								Class("btn nav-link"),
								ghx.Get(reportHrefForView(filter, "tag", "")),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),
							}),
						),
						I(Class("bi-tags me-2")),
						g.Text("Tag"),
						Class("nav-link"),
					),
				),
			),
		),
		g.If(view.main == "general",
			reportGeneralView,
		),
		g.If(view.main == "time",
			reportTimeView,
		),
		g.If(view.main == "project",
			reportProjectView,
		),
		g.If(view.main == "tag",
			reportTagView,
		),
	), nil
}

func (a *ReportWeb) reportTimeView(pageContext *shared.PageContext, view *reportView, filter *ActivityFilter) (g.Node, error) {
	var aggregateBy string
	switch view.sub {
	case "w":
		aggregateBy = "week"
	case "m":
		aggregateBy = "month"
	case "q":
		aggregateBy = "quarter"
	case "d":
		aggregateBy = "day"
	default:
		aggregateBy = "day"
	}

	timeReports, err := a.activityService.TimeReports(pageContext.Ctx, pageContext.Principal, filter, aggregateBy)
	if err != nil {
		return nil, err
	}

	var reportView g.Node
	var showWeekView, showMonthView, showQuarterView bool

	showWeekView = filter.Timespan == "year" || filter.Timespan == "quarter" || filter.Timespan == "month" || filter.Timespan == "week"
	showMonthView = filter.Timespan == "year" || filter.Timespan == "quarter" || filter.Timespan == "month"
	showQuarterView = filter.Timespan == "year" || filter.Timespan == "quarter"

	switch view.sub {
	case "w":
		reportView = reportByWeekView(timeReports)
	case "m":
		reportView = reportByMonthView(timeReports)
	case "q":
		reportView = reportByQuarterView(timeReports)
	case "d":
		reportView = reportByDayView(timeReports)
	default:
		reportView = reportByDayView(timeReports)
	}
	if err != nil {
		return nil, err
	}

	if len(timeReports) == 0 {
		return Div(
			Class("alert alert-info"),
			Role("alert"),
			g.Text(fmt.Sprintf("No activities found in %v.", filter.String())),
		), nil
	}

	return g.Group([]g.Node{
		Nav(
			Div(
				Class("nav nav-tabs"),
				A(
					g.If(view.sub == "d",
						Class("nav-link active"),
					),
					g.If(view.sub != "d",
						g.Group([]g.Node{
							Class("nav-link"),
							ghx.Get(reportHrefForView(filter, "time", "d")),
							ghx.PushURL("true"),
							ghx.Target("#baralga__report_content"),
							ghx.Swap("outerHTML"),
						}),
					),
					Type("button"),
					g.Text("By Day"),
				),
				g.If(showWeekView,
					A(
						g.If(view.sub == "w",
							Class("nav-link active"),
						),
						g.If(view.sub != "w",
							g.Group([]g.Node{
								Class("nav-link"),
								ghx.Get(reportHrefForView(filter, "time", "w")),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),
							}),
						),
						Type("button"),
						g.Text("By Week"),
					),
				),
				g.If(showMonthView,
					A(
						g.If(view.sub == "m",
							Class("nav-link active"),
						),
						g.If(view.sub != "m",
							g.Group([]g.Node{
								Class("nav-link"),
								ghx.Get(reportHrefForView(filter, "time", "m")),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),
							}),
						),
						Type("button"),
						g.Text("By Month"),
					),
				),
				g.If(showQuarterView,
					A(
						g.If(view.sub == "q",
							Class("nav-link active"),
						),
						g.If(view.sub != "q",
							g.Group([]g.Node{
								Class("nav-link"),
								ghx.Get(reportHrefForView(filter, "time", "q")),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),
							}),
						),
						Type("button"),
						g.Text("By Quarter"),
					),
				),
			),
		),
		Div(
			Class("tab-content"),
			reportView,
		),
	}), nil
}

func (a *ReportWeb) reportProjectView(pageContext *shared.PageContext, view *reportView, filter *ActivityFilter) (g.Node, error) {
	projectReports, err := a.activityService.ProjectReports(pageContext.Ctx, pageContext.Principal, filter)
	if err != nil {
		return nil, err
	}

	if len(projectReports) == 0 {
		return Div(
			Class("alert alert-info"),
			Role("alert"),
			g.Text(fmt.Sprintf("No activities found in %v.", filter.String())),
		), nil
	}

	return g.Group([]g.Node{
		Div(
			Class("table-responsive"),
			Table(
				ID("project-report"),
				Class("table table-striped"),
				THead(
					Tr(
						Th(g.Text("Project")),
						Th(
							Class("text-end"),
							g.Text("Duration"),
						),
					),
				),
				TBody(
					g.Group(g.Map(projectReports, func(activity *ActivityProjectReportItem) g.Node {
						return Tr(
							ghx.Target("this"),
							ghx.Swap("outerHTML"),

							Td(g.Text(activity.ProjectTitle)),
							Td(
								Class("text-end"),
								g.Text(activity.DurationFormatted()),
							),
						)
					}),
					),
				),
			),
		),
	}), nil
}

func reportByDayView(timeReports []*ActivityTimeReportItem) g.Node {
	return Table(
		ID("time-report-by-day"),
		Class("table table-striped"),
		THead(
			Tr(
				Th(g.Text("Day")),
				Th(
					Class("text-end"),
					g.Text("Duration"),
				),
			),
		),
		TBody(
			g.Group(g.Map(timeReports, func(reportItem *ActivityTimeReportItem) g.Node {
				return Tr(
					Td(
						g.Text(reportItem.AsTime().Format("02.01.2006 Monday")),
					),
					Td(
						Class("text-end"),
						g.Text(reportItem.DurationFormatted()),
					),
				)
			}),
			),
		),
	)
}

func reportByWeekView(timeReports []*ActivityTimeReportItem) g.Node {
	return Table(
		ID("time-report-by-week"),
		Class("table table-striped"),
		THead(
			Tr(
				Th(g.Text("Week")),
				Th(g.Text("Year")),
				Th(
					Class("text-end"),
					g.Text("Duration"),
				),
			),
		),
		TBody(
			g.Group(g.Map(timeReports, func(reportItem *ActivityTimeReportItem) g.Node {
				return Tr(
					Td(
						g.Text(fmt.Sprintf("%v", reportItem.Week)),
					),
					Td(
						g.Text(fmt.Sprintf("%v", reportItem.Year)),
					),
					Td(
						Class("text-end"),
						g.Text(reportItem.DurationFormatted()),
					),
				)
			}),
			),
		),
	)
}

func reportByMonthView(timeReports []*ActivityTimeReportItem) g.Node {
	return Table(
		ID("time-report-by-month"),
		Class("table table-striped"),
		THead(
			Tr(
				Th(g.Text("Month")),
				Th(g.Text("Year")),
				Th(
					Class("text-end"),
					g.Text("Duration"),
				),
			),
		),
		TBody(
			g.Group(g.Map(timeReports, func(reportItem *ActivityTimeReportItem) g.Node {
				return Tr(
					Td(
						g.Text(reportItem.AsTime().Format("01 January")),
					),
					Td(
						g.Text(fmt.Sprintf("%v", reportItem.Year)),
					),
					Td(
						Class("text-end"),
						g.Text(reportItem.DurationFormatted()),
					),
				)
			}),
			),
		),
	)
}

func reportByQuarterView(timeReports []*ActivityTimeReportItem) g.Node {
	return Table(
		ID("time-report-by-quarter"),
		Class("table table-striped"),
		THead(
			Tr(
				Th(g.Text("Quarter")),
				Th(g.Text("Year")),
				Th(
					Class("text-end"),
					g.Text("Duration"),
				),
			),
		),
		TBody(
			g.Group(g.Map(timeReports, func(reportItem *ActivityTimeReportItem) g.Node {
				return Tr(
					Td(
						g.Text(fmt.Sprintf("Q%v", reportItem.Quarter)),
					),
					Td(
						g.Text(fmt.Sprintf("%v", reportItem.Year)),
					),
					Td(
						Class("text-end"),
						g.Text(reportItem.DurationFormatted()),
					),
				)
			}),
			),
		),
	)
}

func (a *ReportWeb) reportGeneralView(pageContext *shared.PageContext, filter *ActivityFilter, view *reportView) (g.Node, error) {
	pageParams := paged.PageParamsFromQuery(pageContext.CurrentQuery, 50)

	activitiesPage, projects, err := a.activityService.ReadActivitiesWithProjects(pageContext.Ctx, pageContext.Principal, filter, &pageParams)
	if err != nil {
		return nil, err
	}

	// prepare projects
	projectsById := make(map[uuid.UUID]*Project)
	for _, project := range projects {
		projectsById[project.ID] = project
	}

	if len(activitiesPage.Activities) == 0 {
		return Div(
			Class("alert alert-info"),
			Role("alert"),
			g.Text(fmt.Sprintf("No activities found in %v.", filter.String())),
		), nil
	}

	// pages to iterate over
	var pageIndices []int
	for i := 1; i <= activitiesPage.Page.TotalPages; i++ {
		pageIndices = append(pageIndices, i)
	}

	return g.Group([]g.Node{
		Div(
			Class("table-responsive-sm d-lg-none"),
			Table(
				Class("table table-striped"),
				THead(
					Tr(
						Th(
							A(
								ghx.Get(reportHref(filter.WithSortToggle("project"), view)),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),

								g.Text("Project"),
							),
						),
						Th(
							A(
								ghx.Get(reportHref(filter.WithSortToggle("start"), view)),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),

								g.Text("Date"),
							),
						),
						Th(
							Class("text-end"),
							g.Text("Duration"),
						),
						Th(),
					),
				),
				TBody(
					g.Group(g.Map(activitiesPage.Activities, func(activity *Activity) g.Node {
						return Tr(
							ghx.Target("this"),
							ghx.Swap("outerHTML"),

							TitleAttr(activity.Description),

							Td(g.Text(projectsById[activity.ProjectID].Title)),
							Td(g.Text(time_utils.FormatDateDEShort(activity.Start))),
							Td(
								Class("text-end"),
								g.Text(activity.DurationFormatted()),
							),
							Td(
								Class("text-end"),
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
					}),
					),
				),
			),
		),
		Div(
			Class("table-responsive-lg d-none d-lg-block"),
			Table(
				Class("table table-striped"),
				THead(
					Tr(
						Th(
							A(
								ghx.Get(reportHref(filter.WithSortToggle("project"), view)),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),

								g.Text("Project"),
							),
						),
						Th(
							A(
								ghx.Get(reportHref(filter.WithSortToggle("start"), view)),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),

								g.Text("Date"),
							),
						),
						Th(g.Text("Start")),
						Th(g.Text("End")),
						Th(
							Class("text-end"),
							g.Text("Duration"),
						),
						Th(),
					),
				),
				TBody(
					g.Group(g.Map(activitiesPage.Activities, func(activity *Activity) g.Node {
						return Tr(
							ghx.Target("this"),
							ghx.Swap("outerHTML"),

							TitleAttr(activity.Description),

							Td(g.Text(projectsById[activity.ProjectID].Title)),
							Td(g.Text(time_utils.FormatDateDE(activity.Start))),
							Td(g.Text(time_utils.FormatTime(activity.Start))),
							Td(g.Text(time_utils.FormatTime(activity.End))),
							Td(
								Class("text-end"),
								g.Text(activity.DurationFormatted()),
							),
							Td(
								Class("text-end"),
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
					}),
					),
				),
			),
		),
		g.If(
			activitiesPage.Page.TotalPages > 1,
			Nav(
				Class("d-flex justify-content-center"),
				Ul(
					Class("pagination"),
					Li(
						Class("page-item"),
						g.If(
							activitiesPage.Page.Number > 0,
							A(
								Class("page-link"),
								Href(""),

								ghx.Get(reportHrefForPage(filter, view, activitiesPage.Page.Number-1)),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),

								g.Raw("&laquo;"),
							),
						),
						g.If(
							activitiesPage.Page.Number <= 0,
							Span(
								Class("page-link active"),
								g.Raw("&laquo;"),
							),
						),
					),
					g.Group(g.Map(pageIndices, func(pageIndex int) g.Node {
						return Li(
							g.If(
								pageIndex == activitiesPage.Page.Number,
								g.Group([]g.Node{
									Class("page-item active"),
									Span(
										Class("page-link"),
										g.Textf("%v", pageIndex+1),
									),
								}),
							),
							g.If(
								pageIndex != activitiesPage.Page.Number,
								g.Group([]g.Node{
									Class("page-item"),
									A(
										Class("page-link"),
										Href(""),

										ghx.Get(reportHrefForPage(filter, view, pageIndex)),
										ghx.PushURL("true"),
										ghx.Target("#baralga__report_content"),
										ghx.Swap("outerHTML"),

										g.Textf("%v", pageIndex+1),
									),
								}),
							),
						)
					})),
					Li(
						Class("page-item"),
						g.If(
							activitiesPage.Page.TotalPages-1 > activitiesPage.Page.Number,
							A(
								Class("page-link"),
								Href(""),

								ghx.Get(reportHrefForPage(filter, view, activitiesPage.Page.Number+1)),
								ghx.PushURL("true"),
								ghx.Target("#baralga__report_content"),
								ghx.Swap("outerHTML"),

								g.Raw("&raquo;"),
							),
						),
						g.If(
							!(activitiesPage.Page.TotalPages-1 > activitiesPage.Page.Number), //nolint:all
							Span(
								Class("page-link active"),
								g.Raw("&raquo;"),
							),
						),
					),
				),
			),
		),
	}), nil
}

func (a *ReportWeb) reportTagView(pageContext *shared.PageContext, view *reportView, filter *ActivityFilter) (g.Node, error) {
	// Get all tags for the organization for the tag selection interface
	allTags, err := a.activityService.GetTagsForAutocomplete(pageContext.Ctx, pageContext.Principal, "")
	if err != nil {
		return nil, err
	}

	// Get selected tags from filter
	selectedTags := filter.Tags()

	// Generate tag reports with selected tags
	tagReportData, err := a.activityService.GenerateTagReports(pageContext.Ctx, pageContext.Principal, filter, "day", selectedTags)
	if err != nil {
		return nil, err
	}

	if len(tagReportData.Items) == 0 {
		return Div(
			Class("alert alert-info"),
			Role("alert"),
			g.Text(fmt.Sprintf("No tagged activities found in %v.", filter.String())),
		), nil
	}

	return g.Group([]g.Node{
		// Tag selection interface
		Div(
			Class("row mb-3"),
			Div(
				Class("col-12"),
				H6(g.Text("Filter by Tags")),
				Div(
					Class("d-flex flex-wrap gap-2 mb-2"),
					g.Group(g.Map(allTags, func(tag *Tag) g.Node {
						isSelected := reportContainsTag(selectedTags, tag.Name)
						return Span(
							Class("badge"),
							g.If(isSelected,
								g.Group([]g.Node{
									Class("bg-primary"),
									Style(fmt.Sprintf("background-color: %s !important;", tag.Color)),
								}),
							),
							g.If(!isSelected,
								g.Group([]g.Node{
									Class("bg-light text-dark border"),
									ghx.Get(reportHref(filter.WithTags(reportToggleTag(selectedTags, tag.Name)), view)),
									ghx.PushURL("true"),
									ghx.Target("#baralga__report_content"),
									ghx.Swap("outerHTML"),
									Style("cursor: pointer;"),
								}),
							),
							g.Text(tag.Name),
							g.If(isSelected,
								Span(
									Class("ms-1"),
									ghx.Get(reportHref(filter.WithTags(reportRemoveTag(selectedTags, tag.Name)), view)),
									ghx.PushURL("true"),
									ghx.Target("#baralga__report_content"),
									ghx.Swap("outerHTML"),
									Style("cursor: pointer;"),
									g.Text("×"),
								),
							),
						)
					})),
				),
				g.If(len(selectedTags) > 0,
					A(
						Class("btn btn-sm btn-outline-secondary"),
						ghx.Get(reportHref(filter.WithTags([]string{}), view)),
						ghx.PushURL("true"),
						ghx.Target("#baralga__report_content"),
						ghx.Swap("outerHTML"),
						g.Text("Clear All"),
					),
				),
			),
		),
		// Tag report table
		Div(
			Class("table-responsive"),
			Table(
				ID("tag-report"),
				Class("table table-striped"),
				THead(
					Tr(
						Th(g.Text("Tag")),
						Th(g.Text("Activities")),
						Th(
							Class("text-end"),
							g.Text("Duration"),
						),
					),
				),
				TBody(
					g.Group(g.Map(tagReportData.Items, func(item *TagReportItem) g.Node {
						return Tr(
							Td(
								Span(
									Class("badge me-2"),
									Style(fmt.Sprintf("background-color: %s;", item.TagColor)),
									g.Text(item.TagName),
								),
							),
							Td(g.Text(fmt.Sprintf("%d", item.ActivityCount))),
							Td(
								Class("text-end"),
								g.Text(item.DurationFormatted()),
							),
						)
					})),
				),
			),
		),
		// Summary information
		g.If(len(selectedTags) > 0 || tagReportData.TotalTags > 0,
			Div(
				Class("mt-3 text-muted small"),
				g.If(len(selectedTags) > 0,
					g.Text(fmt.Sprintf("Showing %d activities for %d selected tags. ", len(tagReportData.Items), len(selectedTags))),
				),
				g.Text(fmt.Sprintf("Total: %d tags, %s", tagReportData.TotalTags, time_utils.FormatMinutesAsDuration(float64(tagReportData.TotalTime)))),
			),
		),
	}), nil
}

// Helper functions for tag filtering in reports
func reportContainsTag(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func reportToggleTag(tags []string, tag string) []string {
	if reportContainsTag(tags, tag) {
		return reportRemoveTag(tags, tag)
	}
	return append(tags, tag)
}

func reportRemoveTag(tags []string, tag string) []string {
	var result []string
	for _, t := range tags {
		if t != tag {
			result = append(result, t)
		}
	}
	return result
}

type reportView struct {
	main string
	sub  string
}

func (v *reportView) asParam() string {
	if v.sub == "" {
		return v.main
	}
	return fmt.Sprintf("%v:%v", v.main, v.sub)
}

func reportHrefForView(filter *ActivityFilter, viewMain, viewSub string) string {
	return reportHref(filter, &reportView{main: viewMain, sub: viewSub})
}

func reportHrefForPage(filter *ActivityFilter, view *reportView, page int) string {
	return fmt.Sprintf("%v&p=%v", reportHref(filter, view), page)
}

func reportHref(filter *ActivityFilter, view *reportView) string {
	reportHref := fmt.Sprintf("/reports?t=%v&v=%v&c=%v", filter.Timespan, filter.String(), view.asParam())

	if filter.sortBy != "" && filter.sortOrder != "" {
		reportHref += fmt.Sprintf("&sort=%v", fmt.Sprintf("%v:%v", filter.sortBy, filter.sortOrder))
	}

	// Add tag filters to URL
	if len(filter.Tags()) > 0 {
		reportHref += "&tags=" + strings.Join(filter.Tags(), ",")
	}

	return reportHref
}

func (a *ReportWeb) ReportPage(pageContext *shared.PageContext, reportView g.Node) g.Node {
	return shared.Page(
		pageContext.Title,
		pageContext.CurrentPath,
		[]g.Node{
			shared.Navbar(pageContext),
			reportView,
			shared.ModalView(),
		},
	)
}

func reportViewFromQueryParams(params url.Values, timespan string) *reportView {
	if len(params["c"]) == 0 {
		params["c"] = []string{"general"}
	}

	cParts := strings.Split(params["c"][0], ":")
	reportView := &reportView{
		main: cParts[0],
	}

	if reportView.main == "time" {
		if len(cParts) > 1 {
			reportView.sub = cParts[1]
			if timespan == "month" && reportView.sub == "q" {
				reportView.sub = "m"
			} else if timespan == "week" && (reportView.sub == "q" || reportView.sub == "m") {
				reportView.sub = "w"
			} else if timespan == "day" && (reportView.sub == "q" || reportView.sub == "m" || reportView.sub == "w") {
				reportView.sub = "d"
			}
		} else {
			reportView.sub = "d"
		}
	}

	// Tag view doesn't need sub-views for now
	if reportView.main == "tag" {
		reportView.sub = ""
	}

	return reportView
}
