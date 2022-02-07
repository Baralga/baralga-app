package main

import (
	"net/http"

	hx "github.com/baralga/htmx"
	"github.com/baralga/util"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

type loginFormModel struct {
	CSRFToken string
	EMail     string
	Password  string
}

func (a *app) HandleLoginForm(tokenAuth *jwtauth.JWTAuth) http.HandlerFunc {
	expiryDuration := a.Config.ExpiryDuration()
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			formModel := loginFormModel{}
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, LoginPage(r.URL.Path, formModel))
			return
		}

		var formModel loginFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)

		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, LoginPage(r.URL.Path, formModel))
			return
		}

		principal, err := a.Authenticate(r.Context(), formModel.EMail, formModel.Password)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			util.RenderHTML(w, LoginPage(r.URL.Path, formModel))
			return
		}

		cookie := a.CreateCookie(tokenAuth, expiryDuration, principal)
		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (a *app) HandleLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		formModel := loginFormModel{}
		formModel.CSRFToken = csrf.Token(r)
		util.RenderHTML(w, LoginPage(r.URL.Path, formModel))
	}
}

func (a *app) HandleLogoutPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := a.DeleteCookie()
		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func LoginPage(currentPath string, formModel loginFormModel) g.Node {
	return Page(
		"Sign In",
		currentPath,
		[]g.Node{
			Section(
				Class("full-center"),
				Div(
					Class("container"),
					Div(
						Class("d-flex justify-content-center align-items-center mt-2 mb-3"),
						Img(
							Alt("Baralga"),
							Class("img-responsive"),
							Src("/assets/baralga_192.png"),
						),
						Div(
							Class("ms-4"),
							H2(
								g.Text("Baralga"),
								Small(
									Class("text-muted"),
									StyleAttr("display: block; font-size: 70%;"),
									g.Text("project time tracking"),
								),
							),
						),
					),
					LoginForm(formModel, ""),
				),
			),
		},
	)
}

func LoginForm(formModel loginFormModel, errorMessage string) g.Node {
	return FormEl(
		ID("login_form"),
		Action("/login"),
		Method("POST"),
		g.If(
			errorMessage != "",
			Div(
				Class("alert alert-danger text-center"),
				Role("alert"),
				Span(g.Text(errorMessage)),
			),
		),
		Input(
			Type("hidden"),
			Name("CSRFToken"),
			Value(formModel.CSRFToken),
		),
		Div(
			Class("form-floating mb-3"),
			Input(
				ID("email"),
				Type("text"),
				Name("EMail"),
				Class("form-control"),
				g.Attr("placeholder", "john.doe"),
				Value(formModel.EMail),
			),
			Label(
				g.Attr("for", "email"),
				g.Text("E-Mail"),
			),
		),
		Div(
			Class("form-floating mb-3"),
			Input(
				ID("password"),
				Type("password"),
				Name("Password"),
				Class("form-control"),
				g.Attr("placeholder", "***"),
			),
			Label(
				g.Attr("for", "password"),
				g.Text("Password"),
			),
		),
		Div(
			Class("container-fluid text-center"),
			Button(
				Type("submit"),
				Class("btn btn-primary w-100"),
				g.Text("Sign in"),
			),
		),
		Div(
			Class("row justify-content-around mt-2"),
			Div(
				Class("col-4 text-center"),
				//A(
				//	Class("link-secondary"),
				//	g.Text("Forgot Password?"),
				//),
			),
			Div(
				Class("col-4 text-center"),
				A(
					Href("/signup"),
					hx.Boost(),
					Class("link-secondary"),
					g.Text("Sign up here"),
				),
			),
		),
	)
}

func WebVerifier(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := jwtauth.VerifyRequest(ja, r, jwtauth.TokenFromCookie)
			if err != nil {
				w.Header().Set("HX-Redirect", "/login")

				if !hx.IsHXRequest(r) {
					http.Redirect(w, r, "/login", http.StatusFound)
				}
				return
			}

			ctx := jwtauth.NewContext(r.Context(), token, err)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
