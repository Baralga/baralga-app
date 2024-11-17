package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/baralga/shared"
	"github.com/baralga/shared/hx"
	"github.com/baralga/user"
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/dghubble/gologin/v2/google"
	gologinOauth2 "github.com/dghubble/gologin/v2/oauth2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
	googleOAuth2 "golang.org/x/oauth2/google"
	g "maragu.dev/gomponents"
	ghx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

type loginFormModel struct {
	CSRFToken string
	EMail     string
	Password  string
	Redirect  string
}

type loginParams struct {
	errorMessage string
	infoMessage  string
	redirect     string
}

type AuthWebHandlers struct {
	config      *shared.Config
	authService *AuthService
	userService *user.UserService
	tokenAuth   *jwtauth.JWTAuth
}

func NewAuthWebHandlers(config *shared.Config, authService *AuthService, userService *user.UserService, tokenAuth *jwtauth.JWTAuth) *AuthWebHandlers {
	return &AuthWebHandlers{
		config:      config,
		authService: authService,
		userService: userService,
		tokenAuth:   tokenAuth,
	}

}

func (a *AuthWebHandlers) RegisterProtected(r chi.Router) {
	r.HandleFunc("GET /logout", a.HandleLogoutPage())
}

func (a *AuthWebHandlers) RegisterOpen(r chi.Router) {
	r.Handle("GET /login", a.HandleLoginPage())
	r.Handle("POST /login", a.HandleLoginForm())

	r.Handle("/github/login", a.GithubLoginHandler())
	r.Handle("/github/callback", a.GithubCallbackHandler())

	r.Handle("/google/login", a.GoogleLoginHandler())
	r.Handle("/google/callback", a.GoogleCallbackHandler())
}

func (a *AuthWebHandlers) HandleLoginForm() http.HandlerFunc {
	expiryDuration := a.config.ExpiryDuration()
	authService := a.authService
	tokenAuth := a.tokenAuth
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			formModel := loginFormModel{}
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, a.LoginPage(r.URL.Path, formModel, &loginParams{}))
			return
		}

		var formModel loginFormModel
		err = schema.NewDecoder().Decode(&formModel, r.PostForm)

		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			shared.RenderHTML(w, a.LoginPage(r.URL.Path, formModel, &loginParams{}))
			return
		}

		principal, err := authService.Authenticate(r.Context(), formModel.EMail, formModel.Password)
		if err != nil {
			formModel.CSRFToken = csrf.Token(r)
			loginParams := &loginParams{
				errorMessage: "Login failed. Please check your credentials and try again.",
			}
			shared.RenderHTML(w, a.LoginPage(r.URL.Path, formModel, loginParams))
			return
		}

		cookie := authService.CreateCookie(tokenAuth, expiryDuration, principal)
		http.SetCookie(w, &cookie)

		if formModel.Redirect != "" {
			http.Redirect(w, r, formModel.Redirect, http.StatusFound)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (a *AuthWebHandlers) HandleLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loginParams := loginParamsFromQueryParams(r.URL.Query())

		formModel := loginFormModel{
			Redirect: loginParams.redirect,
		}
		formModel.CSRFToken = csrf.Token(r)
		shared.RenderHTML(w, a.LoginPage(r.URL.Path, formModel, loginParams))
	}
}

func loginParamsFromQueryParams(params url.Values) *loginParams {
	loginParams := &loginParams{}
	if len(params["info"]) == 1 && params["info"][0] == "confirm_successfull" {
		loginParams.infoMessage = "You've been confirmed, so happy time tracking!"
	}
	if len(params["redirect"]) == 1 && strings.HasPrefix(params["redirect"][0], "/") {
		loginParams.redirect = params["redirect"][0]
	}
	return loginParams
}

func (a *AuthWebHandlers) HandleLogoutPage() http.HandlerFunc {
	authService := a.authService
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := authService.CreateExpiredCookie()
		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (a *AuthWebHandlers) GithubLoginHandler() http.Handler {
	stateConfig, oauth2Config := a.githubAuthConfig()
	return github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil))
}

func (a *AuthWebHandlers) GithubCallbackHandler() http.Handler {
	stateConfig, oauth2Config := a.githubAuthConfig()
	return github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, a.IssueCookieForGithub(), HandleTokenFailure()))
}

func (a *AuthWebHandlers) GoogleLoginHandler() http.Handler {
	stateConfig, oauth2Config := a.googleAuthConfig()
	return google.StateHandler(stateConfig, google.LoginHandler(oauth2Config, nil))
}

func (a *AuthWebHandlers) GoogleCallbackHandler() http.Handler {
	stateConfig, oauth2Config := a.googleAuthConfig()
	return google.StateHandler(stateConfig, google.CallbackHandler(oauth2Config, a.IssueCookieForGoogle(), HandleTokenFailure()))
}

func HandleTokenFailure() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		_, err := gologinOauth2.TokenFromContext(r.Context())
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}
	return http.HandlerFunc(fn)
}

func (a *AuthWebHandlers) IssueCookieForGithub() http.Handler {
	tokenAuth := a.tokenAuth
	expiryDuration := a.config.ExpiryDuration()
	authService := a.authService
	userService := a.userService
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		githubUser, err := github.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		principal, err := authService.AuthenticateTrusted(ctx, fmt.Sprintf("%v", *githubUser.ID))
		if errors.Is(err, user.ErrUserNotFound) {
			user := &user.User{
				Username: fmt.Sprintf("%v", *githubUser.ID),
				Name:     *githubUser.Login,
				Origin:   "github",
			}
			err := userService.SetUpNewUser(r.Context(), user, uuid.Nil)
			if err != nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			principal, err = authService.AuthenticateTrusted(ctx, fmt.Sprintf("%v", user.Username))
			if err != nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}

		cookie := authService.CreateCookie(tokenAuth, expiryDuration, principal)
		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func (a *AuthWebHandlers) IssueCookieForGoogle() http.Handler {
	tokenAuth := a.tokenAuth
	expiryDuration := a.config.ExpiryDuration()
	authService := a.authService
	userService := a.userService
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		googleUser, err := google.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		principal, err := authService.AuthenticateTrusted(ctx, fmt.Sprintf("%v", googleUser.Id))
		if errors.Is(err, user.ErrUserNotFound) {
			user := &user.User{
				Username: fmt.Sprintf("%v", googleUser.Id),
				Name:     googleUser.Name,
				EMail:    googleUser.Email,
				Origin:   "google",
			}
			err := userService.SetUpNewUser(r.Context(), user, uuid.Nil)
			if err != nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			principal, err = authService.AuthenticateTrusted(ctx, fmt.Sprintf("%v", user.Username))
			if err != nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}

		cookie := authService.CreateCookie(tokenAuth, expiryDuration, principal)
		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func (a *AuthWebHandlers) LoginPage(currentPath string, formModel loginFormModel, loginParams *loginParams) g.Node {
	return shared.Page(
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
					LoginForm(formModel, loginParams),
					Div(
						Class("d-flex justify-content-center align-items-center mt-4 mb-3"),
						g.If(
							a.config.GithubClientId != "",
							A(
								Class("btn btn-secondary"),
								Href("/github/login"),
								I(Class("bi-github")),
								g.Text(" Sign in with Github"),
							),
						),
						g.If(
							a.config.GoogleClientId != "",
							A(
								Class("btn btn-secondary ms-2"),
								Href("/google/login"),
								I(Class("bi-google")),
								g.Text(" Sign in with Google"),
							),
						),
					),
				),
			),
		},
	)
}

func LoginForm(formModel loginFormModel, loginParams *loginParams) g.Node {
	return FormEl(
		ID("login_form"),
		Action("/login"),
		Method("POST"),
		g.If(
			loginParams.errorMessage != "",
			Div(
				Class("alert alert-warning text-center"),
				Role("alert"),
				Span(g.Text(loginParams.errorMessage)),
			),
		),
		g.If(
			loginParams.infoMessage != "",
			Div(
				Class("alert alert-success text-center"),
				Role("alert"),
				Span(g.Text(loginParams.infoMessage)),
			),
		),
		Input(
			Type("hidden"),
			Name("CSRFToken"),
			Value(formModel.CSRFToken),
		),
		g.If(
			formModel.Redirect != "",
			Input(
				Type("hidden"),
				Name("Redirect"),
				Value(formModel.Redirect),
			),
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
					ghx.Boost(""),
					Class("link-secondary"),
					g.Text("Sign up here"),
				),
			),
		),
	)
}

func (a *AuthWebHandlers) WebVerifier() func(http.Handler) http.Handler {
	tokenAuth := a.tokenAuth
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := jwtauth.VerifyRequest(tokenAuth, r, jwtauth.TokenFromCookie)
			if err != nil {
				loginUri := "/login"

				if r.RequestURI != "/" {
					loginUri = fmt.Sprintf("/login?redirect=%v", url.QueryEscape(r.RequestURI))
				}

				w.Header().Set("HX-Redirect", loginUri)
				if !hx.IsHXRequest(r) {
					http.Redirect(w, r, loginUri, http.StatusFound)
				}
				return
			}

			ctx := jwtauth.NewContext(r.Context(), token, err)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *AuthWebHandlers) githubAuthConfig() (gologin.CookieConfig, *oauth2.Config) {
	stateConfig := gologin.DefaultCookieConfig
	if !a.config.IsProduction() {
		stateConfig = gologin.DebugOnlyCookieConfig
	}
	oauth2Config := &oauth2.Config{
		ClientID:     a.config.GithubClientId,
		ClientSecret: a.config.GithubClientSecret,
		RedirectURL:  a.config.GithubRedirectURL,
		Endpoint:     githubOAuth2.Endpoint,
	}
	return stateConfig, oauth2Config
}

func (a *AuthWebHandlers) googleAuthConfig() (gologin.CookieConfig, *oauth2.Config) {
	stateConfig := gologin.DefaultCookieConfig
	if !a.config.IsProduction() {
		stateConfig = gologin.DebugOnlyCookieConfig
	}
	oauth2Config := &oauth2.Config{
		ClientID:     a.config.GoogleClientId,
		ClientSecret: a.config.GoogleClientSecret,
		RedirectURL:  a.config.GoogleRedirectURL,
		Endpoint:     googleOAuth2.Endpoint,
		Scopes:       []string{"profile", "email"},
	}
	return stateConfig, oauth2Config
}
