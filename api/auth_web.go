package main

import (
	"net/http"

	hx "github.com/baralga/htmx"
	"github.com/baralga/util"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gorilla/schema"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

type loginFormModel struct {
	Username string
	Password string
}

func (a *app) HandleLoginForm(tokenAuth *jwtauth.JWTAuth) http.HandlerFunc {
	expiryDuration := a.Config.ExpiryDuration()
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
		}

		var loginFormModel loginFormModel
		err = schema.NewDecoder().Decode(&loginFormModel, r.PostForm)

		if err != nil {
			util.RenderProblemHTML(w, isProduction, err)
			return
		}

		principal, err := a.Authenticate(r.Context(), loginFormModel.Username, loginFormModel.Password)
		if err != nil {
			_ = LoginPage("Sign In", r.URL.Path).Render(w)
			return
		}

		cookie := a.CreateCookie(tokenAuth, expiryDuration, principal)
		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (a *app) HandleLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		util.RenderHTML(w, LoginPage("Sign In", r.URL.Path))
	}
}

func LoginPage(title, currentPath string) g.Node {
	return Page(
		title,
		currentPath,
		[]g.Node{
			Section(
				Class("full-center"),
				Div(Class("container"),
					LoginForm(loginFormModel{}, ""),
				),
			),
		},
	)
}

func LoginForm(loginFormModel loginFormModel, errorMessage string) g.Node {
	return FormEl(
		ID("login_form"),
		Action("/login"),
		Method("POST"),
		Div(
			Class("mt-4 mb-4"),
			Img(
				Class("img-responsive center-block d-block mx-auto"),
			),
		),
		g.If(
			errorMessage != "",
			Div(
				Class("alert alert-danger text-center"),
				Role("alert"),
				Span(g.Text(errorMessage)),
			),
		),
		Div(
			Class("form-floating mb-3"),
			Input(
				ID("username"),
				Type("text"),
				Name("Username"),
				Class("form-control"),
				g.Attr("placeholder", "john.doe"),
				Value(loginFormModel.Username),
			),
			Label(
				g.Attr("for", "username"),
				g.Text("Username"),
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
				A(
					Class("link-secondary"),
					g.Text("Forgot Password?"),
				),
			),
			Div(
				Class("col-4 text-center"),
				A(
					Class("link-secondary"),
					g.Text("Sign Up?"),
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
