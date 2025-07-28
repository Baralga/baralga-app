package shared

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	g "maragu.dev/gomponents"
	ghx "maragu.dev/gomponents-htmx"
	c "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type PageContext struct {
	Ctx          context.Context
	Principal    *Principal
	Title        string
	CurrentPath  string
	CurrentQuery url.Values
}

func HandleWebManifest() http.HandlerFunc {
	manifest := []byte(`
	{
		"short_name": "Baralga",
		"name": "Baralga Time Tracker",
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
		Script(
			Src("/assets/modal.js"),
			g.Attr("crossorigin", "anonymous"),
			g.Attr("defer", "defer"),
		),
	})
}

func Page(title, currentPath string, body []g.Node) g.Node {
	return HTML5Page(c.HTML5Props{
		Title:    fmt.Sprintf("%s # Baralga", title),
		Language: "en",
		Head: []g.Node{
			Meta(
				g.Attr("name", "description"),
				g.Attr("content", "Simple and lightweight time tracking for individuals and teams, for the cloud in the cloud."),
			),
			Meta(
				g.Attr("color-scheme", "light dark"),
			),
			Link(
				Rel("stylesheet"),
				Href("/assets/bootstrap-5.3.2/bootstrap.min.css"),
				g.Attr("integrity", "sha384-T3c6CoIi6uLrA9TneNEoa7RxnatzjcDSCmG1MXxSR1GAsXEV/Dwwykc2MPK8M2HN"),
				g.Attr("crossorigin", "anonymous"),
			),
			Link(
				Rel("stylesheet"),
				Href("/assets/bootstrap-icons-1.10.5/bootstrap-icons.min.css"),
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
				Src("/assets/bootstrap-5.3.2/bootstrap.bundle.min.js"),
				g.Attr("crossorigin", "anonymous"),
				g.Attr("integrity", "sha384-C6RzsynM9kWDrMNeT87bh95OGNyZPhcTNXj1NW7RuBCsyN/o0jlpcV8Qyq46cDfL"),
				g.Attr("defer", "defer"),
			),
			Script(
				Src("/assets/htmx-2.0.6/htmx.min.js"),
				g.Attr("crossorigin", "anonymous"),
				g.Attr("defer", "defer"),
			),
		},
		Body: body,
	})
}

// HTML5 document template.
func HTML5Page(p c.HTML5Props) g.Node {
	return Doctype(
		HTML(g.If(p.Language != "", Lang(p.Language)),
			g.Attr("data-bs-theme", "dark"),
			Head(
				Meta(Charset("utf-8")),
				Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
				TitleEl(g.Text(p.Title)),
				g.If(p.Description != "", Meta(Name("description"), Content(p.Description))),
				g.Group(p.Head),
			),
			Body(g.Group(p.Body)),
		),
	)
}

func Navbar(pageContext *PageContext) g.Node {
	return Nav(
		Class("container-xxl navbar navbar-expand-lg bg-body-tertiary"),
		ghx.Boost(""),
		A(
			Class("navbar-brand ms-2"), Href("/"),
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
				NavbarLi("/", "Track", pageContext.CurrentPath),
				NavbarLi("/reports", "Report", pageContext.CurrentPath),
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
						TitleAttr(pageContext.Principal.Name),
					),
					Ul(
						Class("dropdown-menu dropdown-menu-end"),
						Li(
							A(
								Href("/logout"),
								ghx.Boost(""),
								Class("dropdown-item"),
								I(Class("bi-box-arrow-right me-2")),
								TitleAttr(fmt.Sprintf("Sign out %v", pageContext.Principal.Name)),
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

func RenderHTML(w http.ResponseWriter, n g.Node) {
	w.Header().Set("Content-Type", "text/html")
	err := n.Render(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RenderProblemHTML(w http.ResponseWriter, isProduction bool, err error) {
	log.Printf("internal server error: %s", err)

	if !isProduction {
		http.Error(w, fmt.Sprintf("internal server error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	http.Error(w, "internal server error", http.StatusInternalServerError)
}
