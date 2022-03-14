package main

import (
	"context"
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-http-utils/etag"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/gorilla/csrf"
	"github.com/hellofresh/health-go/v4"
	healthPgx4 "github.com/hellofresh/health-go/v4/checks/pgx4"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/unrolled/secure"
)

type config struct {
	BindPort string `default:"8080"`
	Webroot  string `default:"http://localhost:8080"`
	Db       string `default:"postgres://postgres:postgres@localhost:5432/baralga"`
	Env      string `default:"dev"`

	JWTSecret  string `default:"secret"`
	JWTExpiry  string `default:"24h"`
	CSRFSecret string `default:"CSRFsecret"`

	SMTPServername string `default:"smtp.server:465"`
	SMTPFrom       string `default:"smtp.from@baralga.com"`
	SMTPUser       string `default:"smtp.user@baralga.com"`
	SMTPPassword   string `default:"SMTPPassword"`

	DataProtectionURL string `default:"#"`

	GithubClientId     string `default:"GithubClientID"`
	GithubClientSecret string `default:"GithubClientSecret"`
	GithubRedirectURL  string `default:"http://localhost:8080/github/callback"`
}

func (c *config) ExpiryDuration() time.Duration {
	expiryDuration, err := time.ParseDuration(c.JWTExpiry)
	if err != nil {
		log.Printf("could not parse jwt expiry %s", c.JWTExpiry)
		expiryDuration = time.Duration(24 * time.Hour)
	}
	return expiryDuration
}

type app struct {
	Router *chi.Mux
	Conn   *pgx.Conn
	Config *config

	MailResource MailResource

	RepositoryTxer         RepositoryTxer
	UserRepository         UserRepository
	OrganizationRepository OrganizationRepository
	ProjectRepository      ProjectRepository
	ActivityRepository     ActivityRepository
}

//go:embed migrations
var migrations embed.FS

//go:embed assets
var assets embed.FS

func newApp() (*app, error) {
	var c config
	err := envconfig.Process("baralga", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	port := os.Getenv("PORT")
	if port != "" {
		c.BindPort = port
	}

	router := chi.NewRouter()

	app := &app{
		Router: router,
		Config: &c,
	}

	app.routes()
	app.healthcheck()

	return app, nil
}

func (a *app) run() error {
	connPool, err := connect(a.Config.Db)
	if err != nil {
		return err
	}
	defer connPool.Close()

	a.MailResource = NewSmtpMailResource(
		a.Config.SMTPServername,
		a.Config.SMTPFrom,
		a.Config.SMTPUser,
		a.Config.SMTPPassword,
	)

	a.RepositoryTxer = NewDbRepositoryTxer(connPool)
	a.UserRepository = NewDbUserRepository(connPool)
	a.OrganizationRepository = NewDbOrganizationRepository(connPool)
	a.ProjectRepository = NewDbProjectRepository(connPool)
	a.ActivityRepository = NewDbActivityRepository(connPool)

	return http.ListenAndServe(":"+a.Config.BindPort, a.Router)
}

func (a *app) healthcheck() {
	h, _ := health.New(health.WithChecks(health.Config{
		Name:      "http",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: func(ctx context.Context) error {
			return nil
		},
	},
		health.Config{
			Name:      "db",
			Timeout:   time.Second * 2,
			SkipOnErr: false,
			Check: healthPgx4.New(healthPgx4.Config{
				DSN: a.Config.Db,
			}),
		},
	))
	a.Router.Handle("/health", h.Handler())
}

func (a *app) routes() {
	tokenAuth := jwtauth.New("HS256", []byte(a.Config.JWTSecret), nil)

	a.Router.Use(middleware.Logger)
	a.Router.Use(middleware.Recoverer)
	a.Router.Use(gziphandler.GzipHandler)

	a.Router.Mount("/api", a.apiRouter(tokenAuth))
	a.webRouter(tokenAuth)
}

func (a *app) apiRouter(tokenAuth *jwtauth.JWTAuth) http.Handler {
	r := chi.NewRouter()

	r.Post("/auth/login", a.HandleLogin(tokenAuth))

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(a.JWTPrincipalHandler())

		r.Get("/projects", a.HandleGetProjects())
		r.Post("/projects", a.HandleCreateProject())
		r.Get("/projects/{project-id}", a.HandleGetProject())
		r.Delete("/projects/{project-id}", a.HandleDeleteProject())
		r.Patch("/projects/{project-id}", a.HandleUpdateProject())

		r.Get("/activities", a.HandleGetActivities())
		r.Post("/activities", a.HandleCreateActivity())
		r.Get("/activities/{activity-id}", a.HandleGetActivity())
		r.Delete("/activities/{activity-id}", a.HandleDeleteActivity())
		r.Patch("/activities/{activity-id}", a.HandleUpdateActivity())
	})

	return r
}

func (a *app) webRouter(tokenAuth *jwtauth.JWTAuth) {
	a.Router.Mount("/assets/", etag.Handler(http.FileServer(http.FS(assets)), true))
	a.Router.Get("/manifest.webmanifest", a.HandleWebManifest())

	secureMiddleware := secure.New(secure.Options{
		HostsProxyHeaders:     []string{"X-Forwarded-Host"},
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		ForceSTSHeader:        true,
		IsDevelopment:         !a.isProduction(),
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "base-uri 'self'",
		ReferrerPolicy:        "same-origin",
		PermissionsPolicy:     "fullscreen=*",
	})

	cookieName := "__Secure-csrf"
	if !a.isProduction() {
		cookieName = "__Insecure-csrf"
	}

	CSRF := csrf.Protect([]byte(a.Config.CSRFSecret), csrf.CookieName(cookieName), csrf.FieldName("CSRFToken"), csrf.Secure(a.isProduction()))
	a.Router.Group(func(r chi.Router) {
		r.Use(WebVerifier(tokenAuth))
		r.Use(a.JWTPrincipalHandler())
		r.Use(CSRF)
		r.Use(secureMiddleware.Handler)

		r.Get("/", a.HandleIndexPage())
		r.Get("/reports", a.HandleReportPage())
		r.Get("/projects", a.HandleProjectsPage())
		r.Post("/projects/new", a.HandleProjectForm())
		r.Get("/projects/{project-id}/archive", a.HandleArchiveProject())
		r.Get("/activities/new", a.HandleActivityAddPage())
		r.Post("/activities/validate-start-time", a.HandleStartTimeValidation())
		r.Post("/activities/validate-end-time", a.HandleEndTimeValidation())
		r.Get("/activities/{activity-id}/edit", a.HandleActivityEditPage())
		r.Post("/activities/new", a.HandleActivityForm())
		r.Post("/activities/{activity-id}", a.HandleActivityForm())
		r.Post("/activities/track", a.HandleActivityTrackForm())
		r.Get("/logout", a.HandleLogoutPage())
	})

	a.Router.Group(func(r chi.Router) {
		r.Use(CSRF)
		r.Use(secureMiddleware.Handler)

		r.Get("/login", a.HandleLoginPage())
		r.Post("/login", a.HandleLoginForm(tokenAuth))
		r.Get("/signup", a.HandleSignUpPage())
		r.Post("/signup", a.HandleSignUpForm())
		r.Post("/signup/validate", a.HandleSignUpFormValidate())
		r.Get("/signup/confirm/{confirmation-id}", a.HandleSignUpConfirm())

		r.Handle("/github/login", a.GithubLoginHandler())
		r.Handle("/github/callback", a.GithubCallbackHandler(tokenAuth))
	})
}

func (a *app) isProduction() bool {
	return strings.ToLower(a.Config.Env) == "production"
}

func migrateDb(dbURL string) error {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, strings.Replace(dbURL, "postgres://", "pgx://", 1))
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		return err
	}

	log.Printf("running database version %v (dirty: %v)", version, dirty)
	return nil
}

func connect(dbURL string) (*pgxpool.Pool, error) {
	err := migrateDb(dbURL)
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
