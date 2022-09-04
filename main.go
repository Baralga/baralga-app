package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/baralga/auth"
	"github.com/baralga/shared"
	"github.com/baralga/tracking"
	"github.com/baralga/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-http-utils/etag"
	"github.com/gorilla/csrf"
	"github.com/hellofresh/health-go/v4"
	healthPgx4 "github.com/hellofresh/health-go/v4/checks/pgx4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/unrolled/secure"
)

//go:embed shared/assets
var assets embed.FS

func main() {
	app, connPool, router, err := newApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	defer connPool.Close()

	err = http.ListenAndServe(":"+app.Config.BindPort, router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func newApp() (*shared.App, *pgxpool.Pool, *chi.Mux, error) {
	var c shared.Config
	err := envconfig.Process("baralga", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	port := os.Getenv("PORT")
	if port != "" {
		c.BindPort = port
	}

	connPool, err := shared.Connect(c.Db, c.DbMaxConns)
	if err != nil {
		return nil, nil, nil, err
	}

	a := &shared.App{
		Config: &c,
	}
	repositoryTxer := shared.NewDbRepositoryTxer(connPool)
	mailResource := shared.NewSmtpMailResource(
		c.SMTPServername,
		c.SMTPFrom,
		c.SMTPUser,
		c.SMTPPassword,
	)

	// Tracking
	projectRepository := tracking.NewDbProjectRepository(connPool)
	projectService := tracking.NewProjectService(a, repositoryTxer, projectRepository)
	projectWeb := tracking.NewProjectWeb(a, projectService, projectRepository)
	projectController := tracking.NewProjectController(a, projectRepository, projectService)

	activityRepository := tracking.NewDbActivityRepository(connPool)
	activityService := tracking.NewActitivityService(a, repositoryTxer, activityRepository)
	activityController := tracking.NewActivityController(a, activityService, activityRepository)
	activityWeb := tracking.NewActivityWeb(a, activityService, activityRepository, projectRepository)

	reportWeb := tracking.NewReportWeb(a, activityService)

	// User
	userRepository := user.NewDbUserRepository(connPool)
	organizationRepository := user.NewDbOrganizationRepository(connPool)
	userService := user.NewUserService(a, repositoryTxer, mailResource, userRepository, organizationRepository, projectService.OrganizationInitializer())
	userWeb := user.NewUserWeb(a, userService, userRepository)

	// Auth
	tokenAuth := jwtauth.New("HS256", []byte(a.Config.JWTSecret), nil)
	authService := auth.NewAuthService(a, userRepository)
	authController := auth.NewAuthController(a, authService, tokenAuth)
	authWeb := auth.NewAuthWeb(a, authService, userService, tokenAuth)

	apiHandlers := []shared.DomainHandler{
		authController,
		activityController,
		projectController,
	}
	webHandlers := []shared.DomainHandler{
		userWeb,
		activityWeb,
		authWeb,
		projectWeb,
		reportWeb,
	}

	router := chi.NewRouter()
	registerRoutes(a, router, authController, authWeb, apiHandlers, webHandlers)
	registerHealthcheck(a, router)

	return a, connPool, router, nil
}

func registerHealthcheck(a *shared.App, router *chi.Mux) {
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
	router.Handle("/health", h.Handler())
}

func registerRoutes(a *shared.App, router *chi.Mux, authController *auth.AuthController, authWeb *auth.AuthWeb, apiHandlers []shared.DomainHandler, webHandlers []shared.DomainHandler) {
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	router.Mount("/api", apiRouteHandler(authController, apiHandlers))
	registerWebRoutes(a, router, authController, authWeb, webHandlers)
}

func apiRouteHandler(authController *auth.AuthController, apiHandlers []shared.DomainHandler) http.Handler {
	r := chi.NewRouter()

	for _, apiHandler := range apiHandlers {
		apiHandler.RegisterOpen(r)
	}

	r.Group(func(r chi.Router) {
		r.Use(authController.JWTVerifier())
		r.Use(authController.JWTPrincipalHandler())

		for _, apiHandler := range apiHandlers {
			apiHandler.RegisterProtected(r)
		}
	})

	return r
}

func registerWebRoutes(a *shared.App, router *chi.Mux, authController *auth.AuthController, authWeb *auth.AuthWeb, webHandlers []shared.DomainHandler) {
	assetsDir, _ := fs.Sub(assets, "shared")
	router.Mount("/assets/", etag.Handler(http.FileServer(http.FS(assetsDir)), true))
	router.Get("/manifest.webmanifest", a.HandleWebManifest())

	secureMiddleware := secure.New(secure.Options{
		HostsProxyHeaders:     []string{"X-Forwarded-Host"},
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		ForceSTSHeader:        true,
		IsDevelopment:         !a.IsProduction(),
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
	if !a.IsProduction() {
		cookieName = "__Insecure-csrf"
	}

	CSRF := csrf.Protect([]byte(a.Config.CSRFSecret), csrf.CookieName(cookieName), csrf.FieldName("CSRFToken"), csrf.Secure(a.IsProduction()))
	router.Group(func(r chi.Router) {
		r.Use(authWeb.WebVerifier())
		r.Use(authController.JWTPrincipalHandler())
		r.Use(CSRF)
		r.Use(secureMiddleware.Handler)

		for _, apiHandler := range webHandlers {
			apiHandler.RegisterProtected(r)
		}
	})

	router.Group(func(r chi.Router) {
		r.Use(CSRF)
		r.Use(secureMiddleware.Handler)

		for _, apiHandler := range webHandlers {
			apiHandler.RegisterOpen(r)
		}
	})
}
