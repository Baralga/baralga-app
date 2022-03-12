package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	hx "github.com/baralga/htmx"
	"github.com/baralga/paged"
	"github.com/baralga/util"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
	"github.com/snabb/isoweek"
)

type pageContext struct {
	ctx          context.Context
	principal    *Principal
	title        string
	currentPath  string
	currentQuery url.Values
}

func (a *app) HandleWebManifest() http.HandlerFunc {
	manifest := []byte(`
	{
		"short_name": "Baralga",
		"name": "Baralga time tracker",
		"icons": [
		  {
			"src": "assets/favicon.png",
			"type": "image/x-icon",
			"sizes": "64x64 32x32 24x24 16x16"
		  },
		  {
			"src": "assets/baralga_192.png",
			"type": "image/png",
			"sizes": "192x192"
		  },
		  {
			"src": "assets/baralga_512.png",
			"type": "image/png",
			"sizes": "512x512"
		  }
		],
		"start_url": ".",
		"display": "standalone",
		"theme_color": "#000000",
		"background_color": "#ffffff"
	  }
	`)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		_, _ = w.Write(manifest)
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
			Size: 100,
		}

		principal := r.Context().Value(contextKeyPrincipal).(*Principal)
		activitiesPage, projectsOfActivities, err := a.ReadActivitiesWithProjects(
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
			util.RenderHTML(w, Div(ActivitiesInWeekView(filter, activitiesPage, projectsOfActivities)))
			return
		}

		pageContext := &pageContext{
			principal:   principal,
			currentPath: r.URL.Path,
		}

		formModel := activityTrackFormModel{Action: "start"}
		formModel.CSRFToken = csrf.Token(r)

		util.RenderHTML(w, IndexPage(pageContext, formModel, filter, activitiesPage, projectsOfActivities, projects))
	}
}

func IndexPage(pageContext *pageContext, formModel activityTrackFormModel, filter *ActivityFilter, activitiesPage *ActivitiesPaged, projectsOfActivities []*Project, projects *ProjectsPaged) g.Node {
	return Page(
		"Track Activities",
		pageContext.currentPath,
		[]g.Node{
			Navbar(pageContext),
			Div(
				Class("container"),
				Div(
					Class("row"),
					Div(
						ID("baralga__main_content"),
						Class("col-lg-8 col-sm-12 mb-2 order-2 order-lg-1 mt-lg-4 mt-2"),

						hx.Target("#baralga__main_content"),
						hx.Swap("innerHTML"),

						hx.Trigger("baralga__activities-changed from:body"),
						hx.Get("/"),

						ActivitiesInWeekView(filter, activitiesPage, projectsOfActivities),
					),
					Div(Class("col-lg-4 col-sm-12 order-1 order-lg-2 mt-lg-4 mt-2"),
						TrackPanel(projects.Projects, formModel),
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
					hx.Target("#baralga__main_content_modal_content"),
					hx.Trigger("click, keyup[altKey && shiftKey && key == 'P'] from:body"),
					hx.Swap("outerHTML"),
					hx.Get("/projects"),
					Class("btn btn-outline-primary btn-sm ms-1"),
					I(Class("bi-card-list")),
					TitleAttr("Manage Projects"),
				),
			),
			Div(
				A(
					hx.Target("#baralga__main_content_modal_content"),
					hx.Trigger("click, keyup[altKey && shiftKey && key == 'N'] from:body"),
					hx.Swap("outerHTML"),
					hx.Get("/activities/new"),
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
					hx.Target("#baralga__main_content_modal_content"),
					hx.Swap("outerHTML"),
					hx.Get("/activities/new"),
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
			util.FormatDateDEShort(activity.Start),
			util.FormatDate(activity.Start),
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

	return g.Group(g.Map(len(activitiesByDay), func(i int) g.Node {
		activities := activitiesByDay[dayNodes[i]]
		activityCardID := fmt.Sprintf("baralga__activity_card_%v", dayFormattedByDay[dayNodes[i]][2])

		sum := activitySumByDay[dayNodes[i]]
		durationFormatted := FormatMinutesAsDuration(sum)

		return Div(
			ID(activityCardID),
			Class("card mb-4 me-1"),

			hx.Target("this"),
			hx.Swap("outerHTML"),

			Div(
				Class("card-body position-relative p-2 pt-1"),
				g.If(today == dayNodes[i],
					StyleAttr("background-color: rgba(255, 255,255, 0.05);"),
				),
				H6(
					Class("card-subtitle mt-2"),
					Div(
						Class("d-flex justify-content-between mb-2"),
						Div(
							Class("text-muted"),
							Span(
								g.Text(dayFormattedByDay[dayNodes[i]][0]),
							),
							Span(
								Class("ms-2"),
								StyleAttr("opacity: .45; font-size: 80%;"),
								g.Text(dayFormattedByDay[dayNodes[i]][1]),
							),
						),
					),
					Span(
						Class("position-absolute top-0 start-100 translate-middle badge rounded-pill bg-secondary"),
						g.Text(durationFormatted),
					),
				),
				g.Group(g.Map(len(activities), func(i int) g.Node {
					activity := activities[i]
					return Div(
						Class("d-flex justify-content-between mb-2"),
						hx.Target(fmt.Sprintf("#%v", activityCardID)),
						TitleAttr(activity.Description),
						Span(
							Class("flex-fill"),
							g.Text(util.FormatTime(activity.Start)+" - "+util.FormatTime(activity.End)),
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
				})),
			),
		)
	}))
}

func Page(title, currentPath string, body []g.Node) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    fmt.Sprintf("%s # Baralga", title),
		Language: "en",
		Head: []g.Node{
			Meta(
				g.Attr("color-scheme", "light dark"),
			),
			Link(
				Rel("stylesheet"),
				Href("/assets/bootstrap-dark-5@1.1.3/bootstrap-dark.min.css"),
				g.Attr("crossorigin", "anonymous"),
			),
			Link(
				Rel("stylesheet"),
				Href("/assets/bootstrap-icons-1.8.0/bootstrap-icons.css"),
				g.Attr("media", "print"),
				g.Attr("onload", "this.media='all'"),
				g.Attr("crossorigin", "anonymous"),
			),
			Link(
				Rel("shortcut icon"),
				Href("/assets/favicon.png"),
			),
			Link(
				Rel("apple-touch-icon"),
				Href("/assets/baralga_192.png"),
			),
			Link(
				Rel("manifest"),
				Href("manifest.webmanifest"),
			),
			Script(
				Src("/assets/bootstrap-5.1.3/bootstrap.bundle.min.js"),
				g.Attr("integrity", "sha384-ka7Sk0Gln4gmtz2MlQnikT1wXgYsOg+OMhuP+IlRH9sENBO0LRn5q+8nbTov4+1p"),
				g.Attr("crossorigin", "anonymous"),
				g.Attr("defer", "defer"),
			),
			Script(
				Src("/assets/htmx-1.7.0/htmx.min.js"),
				g.Attr("crossorigin", "anonymous"),
				g.Attr("defer", "defer"),
			),
		},
		Body: body,
	})
}

func Navbar(pageContext *pageContext) g.Node {
	return Nav(
		Class("container-xxl flex-wrap flex-md-nowrap navbar navbar-expand-lg navbar-light bg-dark"),
		hx.Boost(),
		A(
			Class("navbar-brand p-0 me-2"), Href("/"),
			Img(
				Src("assets/baralga_48.png"),
			),
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
				Class("navbar-nav flex-row flex-wrap bd-navbar-nav pt-2 py-md-0"),
				NavbarLi("/", "Track", pageContext.currentPath),
				NavbarLi("/reports", "Report", pageContext.currentPath),
			),
			Hr(
				Class("d-md-none text-white-50"),
			),
			Ul(
				Class("navbar-nav flex-row flex-wrap ms-md-auto"),
				Li(
					Class("nav-item dropdown col-6 col-md-auto"),
					A(
						Class("nav-link dropdown-toggle"),
						Href("#"),
						ID("navbarDropdown"),
						Role("button"),
						g.Attr("data-bs-toggle", "dropdown"),
						I(Class("bi-person-fill")),
						TitleAttr(pageContext.principal.Name),
					),
					Ul(
						Class("dropdown-menu dropdown-menu-end"),
						Li(
							A(
								Href("/logout"),
								hx.Boost(),
								Class("dropdown-item"),
								I(Class("bi-box-arrow-right me-2")),
								TitleAttr(fmt.Sprintf("Sign out %v", pageContext.principal.Name)),
								g.Text("Sign out"),
							),
						),
					),
				),
			),
		),
	)
}

func NavbarLi(href, name, currentPath string) g.Node {
	return Li(
		Class("nav-item col-6 col-md-auto"),
		A(
			Class("nav-link active"),
			Href(href),
			g.Text(name),
		),
	)
}
