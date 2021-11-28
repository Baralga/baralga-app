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
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/hellofresh/health-go/v4"
	healthPgx4 "github.com/hellofresh/health-go/v4/checks/pgx4"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	BindPort string `default:"8080"`
	Db       string `default:"postgres://postgres:postgres@localhost:5432/baralga"`
	Env      string `default:"dev"`

	JWTSecret string `default:"secret"`
	JWTExpiry string `default:"1d"`
}

type app struct {
	Router *chi.Mux
	Conn   *pgx.Conn
	Config *config

	UserRepository     UserRepository
	ProjectRepository  ProjectRepository
	ActivityRepository ActivityRepository
}

//go:embed migrations
var migrations embed.FS

func newApp() (*app, error) {
	var c config
	err := envconfig.Process("baralga", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	port := os.Getenv("PORT")
	if port == "" {
		c.BindPort = "8080"
	}

	router := chi.NewRouter()

	app := &app{
		Router: router,
		Config: &c,
	}

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(gziphandler.GzipHandler)

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

	a.UserRepository = NewDbUserRepository(connPool)
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
	a.Router.Mount("/api", a.apiRouter())
}

func (a *app) apiRouter() http.Handler {
	r := chi.NewRouter()

	tokenAuth := jwtauth.New("HS256", []byte(a.Config.JWTSecret), nil)
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

func (a *app) isProduction() bool {
	return strings.ToLower(a.Config.Env) == "production"
}

func connect(dbURL string) (*pgxpool.Pool, error) {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, strings.Replace(dbURL, "postgres://", "pgx://", 1))
	if err != nil {
		return nil, err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	version, dirty, err := m.Version()
	if err != nil {
		return nil, err
	}
	log.Printf("running database version %v (dirty: %v)", version, dirty)

	conn, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
