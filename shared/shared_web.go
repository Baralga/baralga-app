package shared

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/baralga/shared/util/hx"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

type PageContext struct {
	Ctx          context.Context
	Principal    *Principal
	Title        string
	CurrentPath  string
	CurrentQuery url.Values
}

func (a *App) HandleWebManifest() http.HandlerFunc {
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

func Navbar(pageContext *PageContext) g.Node {
	return Nav(
		Class("container-xxl flex-wrap flex-md-nowrap navbar navbar-expand-lg navbar-dark bg-dark"),
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
								hx.Boost(),
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
