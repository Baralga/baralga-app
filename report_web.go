package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	hx "github.com/baralga/htmx"
	"github.com/baralga/paged"
	"github.com/baralga/util"
	"github.com/google/uuid"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
	"github.com/pkg/errors"
)

func (a *app) HandleReportPage() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)
		pageContext := &pageContext{
			ctx:          r.Context(),
			principal:    principal,
			currentPath:  r.URL.Path,
			currentQuery: r.URL.Query(),
			title:        "Report Activities",
		}

		queryParams := r.URL.Query()
		filter, err := filterFromQueryParams(queryParams)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, errors.New("invalid query params"))
			return
		}

		view := reportViewFromQueryParams(queryParams, filter.Timespan)

		reportView, err := a.ReportView(pageContext, view, filter)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, errors.New("invalid reports"))
			return
		}

		if hx.IsHXTargetRequest(r, "baralga__report_content") {
			util.RenderHTML(w, reportView)
			return
		}

		util.RenderHTML(w, a.ReportPage(pageContext, reportView))
	}
}

func (a *app) ReportView(pageContext *pageContext, view *reportView, filter *ActivityFilter) (g.Node, error) {
	previousFilter := filter.Previous()
	homeFilter := filter.Home()
	nextFilter := filter.Next()

	var reportGeneralView, reportTimeView, reportProjectView g.Node
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

	return Div(
		ID("baralga__report_content"),
		Class("container mt-lg-2"),

		hx.Trigger("baralga__activities-changed from:body"),
		hx.Get(reportHref(filter, view)),
		hx.Target("#baralga__report_content"),
		hx.Swap("outerHTML"),

		Div(
			Class("row mb-2"),
			Div(
				Class("col-md-4 col-12 mt-2"),
				Select(
					hx.Get(fmt.Sprintf("/reports?c=%v", view.asParam())),
					hx.PushURLTrue(),
					hx.Target("#baralga__report_content"),
					hx.Swap("outerHTML"),

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
						hx.Get(reportHref(previousFilter, view)),
						hx.PushURLTrue(),
						hx.Target("#baralga__report_content"),
						hx.Trigger("click, keyup[shiftKey && key == 'ArrowLeft'] from:body"),
						hx.Swap("outerHTML"),

						TitleAttr(fmt.Sprintf("Show previous actvities from %v", previousFilter.String())),
						Class("btn btn-outline-primary"),
						I(Class("bi-arrow-left")),
					),
					A(
						hx.Get(reportHref(homeFilter, view)),
						hx.PushURLTrue(),
						hx.Target("#baralga__report_content"),
						hx.Trigger("click, keyup[shiftKey && key == 'ArrowDown'] from:body"),
						hx.Swap("outerHTML"),

						TitleAttr(fmt.Sprintf("Show current actvities from %v", homeFilter.String())),
						Class("btn btn-outline-primary"),
						I(Class("bi-house-fill")),
					),
					A(
						hx.Get(reportHref(nextFilter, view)),
						hx.PushURLTrue(),
						hx.Target("#baralga__report_content"),
						hx.Trigger("click, keyup[shiftKey && key == 'ArrowRight'] from:body"),
						hx.Swap("outerHTML"),

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
								hx.Get(reportHrefForView(filter, "general", "")),
								hx.PushURLTrue(),
								hx.Target("#baralga__report_content"),
								hx.Swap("outerHTML"),
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
								hx.Get(reportHrefForView(filter, "time", "d")),
								hx.PushURLTrue(),
								hx.Target("#baralga__report_content"),
								hx.Swap("outerHTML"),
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
								hx.Get(reportHrefForView(filter, "project", "d")),
								hx.PushURLTrue(),
								hx.Target("#baralga__report_content"),
								hx.Swap("outerHTML"),
							}),
						),
						I(Class("bi-clock me-2")),
						g.Text("Project"),
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
	), nil
}

func (a *app) reportTimeView(pageContext *pageContext, view *reportView, filter *ActivityFilter) (g.Node, error) {
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

	timeReports, err := a.TimeReports(pageContext.ctx, pageContext.principal, filter, aggregateBy)
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
							hx.Get(reportHrefForView(filter, "time", "d")),
							hx.PushURLTrue(),
							hx.Target("#baralga__report_content"),
							hx.Swap("outerHTML"),
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
								hx.Get(reportHrefForView(filter, "time", "w")),
								hx.PushURLTrue(),
								hx.Target("#baralga__report_content"),
								hx.Swap("outerHTML"),
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
								hx.Get(reportHrefForView(filter, "time", "m")),
								hx.PushURLTrue(),
								hx.Target("#baralga__report_content"),
								hx.Swap("outerHTML"),
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
								hx.Get(reportHrefForView(filter, "time", "q")),
								hx.PushURLTrue(),
								hx.Target("#baralga__report_content"),
								hx.Swap("outerHTML"),
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

func (a *app) reportProjectView(pageContext *pageContext, view *reportView, filter *ActivityFilter) (g.Node, error) {
	projectReports, err := a.ProjectReports(pageContext.ctx, pageContext.principal, filter)
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
				Class("table table-borderless table-striped"),
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
					g.Group(g.Map(len(projectReports), func(i int) g.Node {
						activity := projectReports[i]
						return Tr(
							hx.Target("this"),
							hx.Swap("outerHTML"),

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
		Class("table table-borderless table-striped"),
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
			g.Group(g.Map(len(timeReports), func(i int) g.Node {
				reportItem := timeReports[i]
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
		Class("table table-borderless table-striped"),
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
			g.Group(g.Map(len(timeReports), func(i int) g.Node {
				reportItem := timeReports[i]
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
		Class("table table-borderless table-striped"),
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
			g.Group(g.Map(len(timeReports), func(i int) g.Node {
				reportItem := timeReports[i]
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
		Class("table table-borderless table-striped"),
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
			g.Group(g.Map(len(timeReports), func(i int) g.Node {
				reportItem := timeReports[i]
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

func (a *app) reportGeneralView(pageContext *pageContext, filter *ActivityFilter, view *reportView) (g.Node, error) {
	pageParams := paged.PageParamsFromQuery(pageContext.currentQuery, 50)

	activitiesPage, projects, err := a.ReadActivitiesWithProjects(pageContext.ctx, pageContext.principal, filter, &pageParams)
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

	return g.Group([]g.Node{
		Div(
			Class("table-responsive-sm d-lg-none"),
			Table(
				Class("table table-borderless table-striped"),
				THead(
					Tr(
						Th(g.Text("Project")),
						Th(g.Text("Date")),
						Th(
							Class("text-end"),
							g.Text("Duration"),
						),
						Th(),
					),
				),
				TBody(
					g.Group(g.Map(len(activitiesPage.Activities), func(i int) g.Node {
						activity := activitiesPage.Activities[i]
						return Tr(
							hx.Target("this"),
							hx.Swap("outerHTML"),

							Td(g.Text(projectsById[activity.ProjectID].Title)),
							Td(g.Text(util.FormatDateDEShort(activity.Start))),
							Td(
								Class("text-end"),
								g.Text(activity.DurationFormatted()),
							),
							Td(
								Class("text-end"),
								A(
									hx.Get(fmt.Sprintf("/activities/%v/edit", activity.ID)),
									hx.Target("#baralga__main_content_modal_content"),
									hx.Swap("outerHTML"),

									Class("btn btn-outline-secondary btn-sm"),
									I(Class("bi-pen")),
								),
								A(
									hx.Confirm(
										fmt.Sprintf(
											"Do you really want to delete the activity from %v on %v?",
											util.FormatTime(activity.Start),
											activity.Start.Format("Monday"),
										),
									),
									hx.Delete(fmt.Sprintf("/api/activities/%v", activity.ID)),
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
				Class("table table-borderless table-striped"),
				THead(
					Tr(
						Th(g.Text("Project")),
						Th(g.Text("Date")),
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
					g.Group(g.Map(len(activitiesPage.Activities), func(i int) g.Node {
						activity := activitiesPage.Activities[i]
						return Tr(
							hx.Target("this"),
							hx.Swap("outerHTML"),

							Td(g.Text(projectsById[activity.ProjectID].Title)),
							Td(g.Text(util.FormatDateDE(activity.Start))),
							Td(g.Text(util.FormatTime(activity.Start))),
							Td(g.Text(util.FormatTime(activity.End))),
							Td(
								Class("text-end"),
								g.Text(activity.DurationFormatted()),
							),
							Td(
								Class("text-end"),
								A(
									hx.Get(fmt.Sprintf("/activities/%v/edit", activity.ID)),
									hx.Target("#baralga__main_content_modal_content"),
									hx.Swap("outerHTML"),

									Class("btn btn-outline-secondary btn-sm"),
									I(Class("bi-pen")),
								),
								A(
									hx.Confirm(
										fmt.Sprintf(
											"Do you really want to delete the activity from %v on %v?",
											util.FormatTime(activity.Start),
											activity.Start.Format("Monday"),
										),
									),
									hx.Delete(fmt.Sprintf("/api/activities/%v", activity.ID)),
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

								hx.Get(reportHrefForPage(filter, view, activitiesPage.Page.Number-1)),
								hx.PushURLTrue(),
								hx.Target("#baralga__report_content"),
								hx.Swap("outerHTML"),

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
					g.Group(g.Map(activitiesPage.Page.TotalPages, func(pageIndex int) g.Node {
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
								!(pageIndex == activitiesPage.Page.Number),
								g.Group([]g.Node{
									Class("page-item"),
									A(
										Class("page-link"),
										Href(""),

										hx.Get(reportHrefForPage(filter, view, pageIndex)),
										hx.PushURLTrue(),
										hx.Target("#baralga__report_content"),
										hx.Swap("outerHTML"),

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

								hx.Get(reportHrefForPage(filter, view, activitiesPage.Page.Number+1)),
								hx.PushURLTrue(),
								hx.Target("#baralga__report_content"),
								hx.Swap("outerHTML"),

								g.Raw("&raquo;"),
							),
						),
						g.If(
							!(activitiesPage.Page.TotalPages-1 > activitiesPage.Page.Number),
							Span(
								Class("page-link active"),
								g.Raw("&raquo;"),
							),
						),
					),
				),
			),
		),
		/**
		<nav aria-label="Page navigation example">
		  <ul class="pagination">
		    <li class="page-item">
		      <a class="page-link" href="#" aria-label="Previous">
		        <span aria-hidden="true">&laquo;</span>
		      </a>
		    </li>
		    <li class="page-item"><a class="page-link" href="#">1</a></li>
		    <li class="page-item"><a class="page-link" href="#">2</a></li>
		    <li class="page-item"><a class="page-link" href="#">3</a></li>
		    <li class="page-item">
		      <a class="page-link" href="#" aria-label="Next">
		        <span aria-hidden="true">&raquo;</span>
		      </a>
		    </li>
		  </ul>
		</nav>
		*/
		//H2(g.Text("BAM!!!")),
	}), nil
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
	return fmt.Sprintf("/reports?t=%v&v=%v&c=%v", filter.Timespan, filter.String(), view.asParam())
}

func (a *app) ReportPage(pageContext *pageContext, reportView g.Node) g.Node {
	return Page(
		pageContext.title,
		pageContext.currentPath,
		[]g.Node{
			Navbar(pageContext),
			reportView,
			ModalView(),
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

	return reportView
}
