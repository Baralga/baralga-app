package main

import (
	"fmt"
	"net/http"
	"time"

	hx "github.com/baralga/htmx"
	"github.com/baralga/paged"
	"github.com/baralga/util"
	"github.com/google/uuid"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
	"github.com/snabb/isoweek"
)

type pageContext struct {
	principal   *Principal
	title       string
	currentPath string
}

func (a *app) HandleReportPage() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)
		pageContext := &pageContext{
			principal:   principal,
			currentPath: r.URL.Path,
			title:       "Report Activities",
		}

		queryParams := r.URL.Query()
		if len(queryParams["t"]) == 0 {
			queryParams["t"] = []string{"week"}
		}
		filter, err := filterFromQueryParams(queryParams)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		pageParams := &paged.PageParams{
			Page: 0,
			Size: 500,
		}
		activitiesPage, projects, err := a.ReadActivitiesWithProjects(r.Context(), principal, filter, pageParams)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		if hx.IsHXTargetRequest(r, "baralga__report_content") {
			util.RenderHTML(w, ReportView(filter, activitiesPage, projects))
			return
		}

		util.RenderHTML(w, ReportPage(pageContext, filter, activitiesPage, projects))
	}
}

func (a *app) HandleIndexPage() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		wyear, week := isoweek.FromDate(now.Year(), now.Month(), now.Day())
		filter := &ActivityFilter{
			Timespan: TimespanWeek,
			start:    isoweek.StartTime(wyear, week, time.UTC),
		}
		pageParams := &paged.PageParams{
			Page: 0,
			Size: 50,
		}

		principal := r.Context().Value(contextKeyPrincipal).(*Principal)
		activitiesPage, _, err := a.ReadActivitiesWithProjects(
			r.Context(),
			principal,
			filter,
			pageParams,
		)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		projects, err := a.ProjectRepository.FindProjects(r.Context(), principal.OrganizationID, pageParams)
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		if hx.IsHXTargetRequest(r, "baralga__main_content") {
			util.RenderHTML(w, Div(ActivitiesInWeekView(filter.String(), activitiesPage, projects)))
			return
		}

		pageContext := &pageContext{
			principal:   principal,
			currentPath: r.URL.Path,
			title:       "Track Activities",
		}
		util.RenderHTML(w, IndexPage(pageContext, filter.String(), activitiesPage, projects))
	}
}

func ReportPage(pageContext *pageContext, filter *ActivityFilter, activitiesPage *ActivitiesPaged, projects []*Project) g.Node {
	return Page(
		pageContext.title,
		pageContext.currentPath,
		[]g.Node{
			Navbar(pageContext),
			ReportView(filter, activitiesPage, projects),
			ModalView(),
		},
	)
}

func ReportView(filter *ActivityFilter, activitiesPage *ActivitiesPaged, projects []*Project) g.Node {
	// prepare projects
	projectsById := make(map[uuid.UUID]*Project)
	for _, project := range projects {
		projectsById[project.ID] = project
	}

	previousFilter := filter.Previous()
	nextFilter := filter.Next()

	return Div(
		ID("baralga__report_content"),
		Class("container mt-4"),

		hx.Trigger("baralga__activities-changed from:body"),
		hx.Get(fmt.Sprintf("/reports?t=%v&v=%v", filter.Timespan, filter.String())),
		hx.Target("#baralga__report_content"),

		Div(
			Class("d-flex justify-content-between mb-4"),
			Div(
				Select(
					hx.Get("/reports"),
					hx.Target("#baralga__report_content"),

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
				Class("btn-group"),
				Role("group"),
				A(
					hx.Get(fmt.Sprintf("/reports?t=%v&v=%v", previousFilter.Timespan, previousFilter.String())),
					hx.Target("#baralga__report_content"),

					TitleAttr(previousFilter.String()),
					Class("btn btn-outline-primary"),
					I(Class("bi-arrow-left")),
				),
				A(
					hx.Get(fmt.Sprintf("/reports?t=%v", filter.Timespan)),
					hx.Target("#baralga__report_content"),

					Class("btn btn-outline-primary"),
					I(Class("bi-house-fill")),
				),
				A(
					hx.Get(fmt.Sprintf("/reports?t=%v&v=%v", nextFilter.Timespan, nextFilter.String())),
					hx.Target("#baralga__report_content"),

					TitleAttr(nextFilter.String()),
					Class("btn btn-outline-primary"),
					I(Class("bi-arrow-right")),
				),
			),
			Div(
				H5(
					StyleAttr("min-width: 10rem"),
					Class("text-muted"),
					g.Text(filter.String()),
				),
			),
			Div(
				A(
					Href(
						fmt.Sprintf("/api/activities?contentType=text/csv&t=%v&v=%v", filter.Timespan, filter.String()),
					),
					Class("btn btn-outline-primary"),
					I(Class("bi-file-excel")),
					TitleAttr("Export Activities"),
				),
			),
		),
		g.If(
			len(activitiesPage.Activities) == 0,
			Div(
				Class("alert alert-info"),
				Role("alert"),
				g.Text(fmt.Sprintf("No activities found in %v.", filter.String())),
			),
		),
		g.If(
			len(activitiesPage.Activities) != 0,
			Div(
				Class("table-responsive-sm d-lg-none"),
				Table(
					Class("table table-borderless table-striped"),
					THead(
						Tr(
							Th(g.Text("Project SM")),
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
		),
		g.If(
			len(activitiesPage.Activities) != 0,
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
		),
	)
}

func IndexPage(pageContext *pageContext, filterTitle string, activitiesPage *ActivitiesPaged, projects *ProjectsPaged) g.Node {
	return Page(
		pageContext.title,
		pageContext.currentPath,
		[]g.Node{
			Navbar(pageContext),
			Div(
				Class("container mt-4"),
				Div(
					Class("row"),
					Div(
						ID("baralga__main_content"),
						Class("col-lg-8 col-sm-12 mb-2"),

						hx.Target("#baralga__main_content"),
						hx.Swap("innerHTML"),

						hx.Trigger("baralga__activities-changed from:body"),
						hx.Get("/"),

						ActivitiesInWeekView(filterTitle, activitiesPage, projects),
					),
					Div(Class("col-lg-4 col-sm-12"),
						TrackPanel(projects.Projects, activityTrackFormModel{Action: "start"}),
					),
				),
			),
			ModalView(),
		},
	)
}

func ModalView() g.Node {
	return g.Group([]g.Node{
		Div(
			ID("baralga__main_content_modal"),
			Class("modal"),
			Div(
				Class("modal-dialog modal-fullscreen-sm-down modal-dialog-centered"),
				Div(
					ID("baralga__main_content_modal_content"),
					Class("modal-content"),
				),
			),
		),
		g.Raw(`<script>
		document.addEventListener('DOMContentLoaded', function() {
			document.body.addEventListener('baralga__main_content_modal-show', function (evt) {
				var modal = bootstrap.Modal.getOrCreateInstance(document.getElementById('baralga__main_content_modal'), { keyboard: true });
				modal.show();
			});
			document.body.addEventListener('baralga__main_content_modal-hide', function (evt) {
				var modal = bootstrap.Modal.getOrCreateInstance(document.getElementById('baralga__main_content_modal'), { keyboard: true });
				modal.hide();
			});
		});
		</script>`),
	})
}

func ActivitiesInWeekView(filterTitle string, activitiesPage *ActivitiesPaged, projects *ProjectsPaged) g.Node {
	// prepare projects
	projectsById := make(map[uuid.UUID]*Project)
	for _, project := range projects.Projects {
		projectsById[project.ID] = project
	}
	nodes := []g.Node{
		Div(
			Class("d-flex justify-content-between"),
			H2(
				Class("mb-4"),
				g.Text("My Week "),
				Small(Class("text-muted"), g.Text(filterTitle)),
			),
			Div(
				A(
					hx.Target("#baralga__main_content_modal_content"),
					hx.Swap("outerHTML"),
					hx.Get("/projects"),
					Class("btn btn-outline-primary btn-sm ms-1"),
					I(Class("bi-card-list")),
					TitleAttr("Manage Projects"),
				),
				A(
					hx.Target("#baralga__main_content_modal_content"),
					hx.Swap("outerHTML"),
					hx.Get("/activities/new"),
					Class("btn btn-outline-primary btn-sm ms-1"),
					I(Class("bi-plus")),
					TitleAttr("Add Activity"),
				),
			),
		),
		g.If(
			len(activitiesPage.Activities) == 0,
			Div(
				Class("alert alert-info"),
				Role("alert"),
				g.Text("No activities in current week. Add some here!"),
			),
		),
		g.Group(g.Map(len(activitiesPage.Activities), func(i int) g.Node {
			activity := activitiesPage.Activities[i]
			activityCardID := fmt.Sprintf("activity-card-%v", activity.ID)
			return Div(
				ID(activityCardID),
				Class("card mb-2"),

				TitleAttr(activity.Description),

				hx.Target("this"),
				hx.Swap("outerHTML"),

				Div(
					Class("card-body"),
					StyleAttr("padding: 0.2rem 1rem"),
					H6(
						Class("card-subtitle mt-2"),
						Div(
							Class("d-flex justify-content-between mb-2"),
							Span(
								Class("flex-grow-1 text-muted"),
								g.Text(activity.Start.Format("Monday")),
							),
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
					),
					H6(
						Class("card-title"),
						Div(
							Class("d-flex justify-content-between"),
							Span(g.Text(util.FormatTime(activity.Start)+" - "+util.FormatTime(activity.End))),
							Span(g.Text(projectsById[activity.ProjectID].Title)),
							Span(g.Text(activity.DurationFormatted())),
						),
					),
				),
			)
		})),
	}
	return g.Group(nodes)
}

func Page(title, currentPath string, body []g.Node) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    "Baralga # " + title,
		Language: "en",
		Head: []g.Node{
			Meta(
				g.Attr("color-scheme", "light dark"),
			),
			Link(
				Rel("stylesheet"), Href("https://cdn.jsdelivr.net/npm/bootstrap-dark-5@1.1.3/dist/css/bootstrap-dark.min.css"),
				//				g.Attr("integrity", "sha384-1BmE4kWBq78iYhFldvKuhfTAU6auU8tT94WrHftjDbrCEXSU1oBoqyl2QvZ6jIW3"),
				g.Attr("crossorigin", "anonymous"),
			),
			Link(
				Rel("stylesheet"), Href("https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css"),
				g.Attr("media", "print"),
				g.Attr("onload", "this.media='all'"),
				g.Attr("crossorigin", "anonymous"),
			),
			Script(
				Src("https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"),
				g.Attr("integrity", "sha384-ka7Sk0Gln4gmtz2MlQnikT1wXgYsOg+OMhuP+IlRH9sENBO0LRn5q+8nbTov4+1p"),
				g.Attr("crossorigin", "anonymous"),
				g.Attr("defer", "defer"),
			),
			Script(
				Src("https://unpkg.com/htmx.org@1.6.1/dist/htmx.min.js"),
				g.Attr("crossorigin", "anonymous"),
				g.Attr("defer", "defer"),
			),
		},
		Body: body,
	})
}

func Navbar(pageContext *pageContext) g.Node {
	return Nav(
		Class("navbar navbar-expand-lg navbar-light bg-dark"),
		hx.Boost(),
		Div(
			Class("container-fluid"),
			A(
				Class("navbar-brand"), Href("/"),
				g.Text("baralga"),
			),
			Button(
				Class("navbar-toggler"), Type("button"),
				g.Attr("data-bs-toggle", "collapse"),
				g.Attr("data-bs-target", "#navbarSupportedContent"),
				Span(Class("navbar-toggler-icon")),
			),
			Div(
				ID("navbarSupportedContent"),
				Class("collapse navbar-collapse"),
				Ul(
					Class("navbar-nav me-auto mb-2 mb-lg-0"),
					NavbarLi("/", "Track", pageContext.currentPath),
					NavbarLi("/reports", "Report", pageContext.currentPath),
				),
			),
		),
	)
}

func NavbarLi(href, name, currentPath string) g.Node {
	return Li(
		Class("nav-item"),
		A(
			Class("nav-link active"),
			Href(href),
			g.Text(name),
		),
	)
}
