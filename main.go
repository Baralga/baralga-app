package main

import (
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
	"github.com/hellofresh/health-go/v5"
	healthPgx "github.com/hellofresh/health-go/v5/checks/pgx5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/unrolled/secure"
)

//go:embed shared/assets
var assets embed.FS

func main() {
	config, connPool, router, err := newApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	defer connPool.Close()

	err = http.ListenAndServe(":"+config.BindPort, router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func newApp() (*shared.Config, *pgxpool.Pool, *chi.Mux, error) {
	var config shared.Config
	err := envconfig.Process("baralga", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
	port := os.Getenv("PORT")
	if port != "" {
		config.BindPort = port
	}

	connPool, err := shared.Connect(config.Db, config.DbMaxConns)
	if err != nil {
		return nil, nil, nil, err
	}

	repositoryTxer := shared.NewDbRepositoryTxer(connPool)
	mailResource := shared.NewSmtpMailResource(
		config.SMTPServername,
		config.SMTPFrom,
		config.SMTPUser,
		config.SMTPPassword,
	)

	// Tracking
	projectRepository := tracking.NewDbProjectRepository(connPool)
	projectService := tracking.NewProjectService(repositoryTxer, projectRepository)
	projectWeb := tracking.NewProjectWeb(&config, projectService, projectRepository)
	projectController := tracking.NewProjectController(&config, projectRepository, projectService)

	activityRepository := tracking.NewDbActivityRepository(connPool)
	activityService := tracking.NewActitivityService(repositoryTxer, activityRepository)
	activityController := tracking.NewActivityController(&config, activityService, activityRepository)
	activityWeb := tracking.NewActivityWeb(&config, activityService, activityRepository, projectRepository)

	reportWeb := tracking.NewReportWeb(&config, activityService)

	// User
	userRepository := user.NewDbUserRepository(connPool)
	organizationRepository := user.NewDbOrganizationRepository(connPool)
	userService := user.NewUserService(&config, repositoryTxer, mailResource, userRepository, organizationRepository, projectService.OrganizationInitializer())
	userWeb := user.NewUserWeb(&config, userService, userRepository)

	// Auth
	tokenAuth := jwtauth.New("HS256", []byte(config.JWTSecret), nil)
	authService := auth.NewAuthService(&config, userRepository)
	authController := auth.NewAuthController(&config, authService, tokenAuth)
	authWeb := auth.NewAuthWeb(&config, authService, userService, tokenAuth)

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
	registerRoutes(&config, router, authController, authWeb, apiHandlers, webHandlers)
	registerHealthcheck(&config, router)

	return &config, connPool, router, nil
}

func registerHealthcheck(config *shared.Config, router *chi.Mux) {
	h, _ := health.New(health.WithChecks(
		health.Config{
			Name:      "db",
			Timeout:   time.Second * 2,
			SkipOnErr: false,
			Check: healthPgx.New(healthPgx.Config{
				DSN: config.Db,
			}),
		},
	))
	router.Get("/health", h.HandlerFunc)
}

func registerRoutes(config *shared.Config, router *chi.Mux, authController *auth.AuthController, authWeb *auth.AuthWeb, apiHandlers []shared.DomainHandler, webHandlers []shared.DomainHandler) {
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	router.Mount("/api", apiRouteHandler(authController, apiHandlers))
	registerWebRoutes(config, router, authController, authWeb, webHandlers)
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

func registerWebRoutes(config *shared.Config, router *chi.Mux, authController *auth.AuthController, authWeb *auth.AuthWeb, webHandlers []shared.DomainHandler) {
	assetsDir, _ := fs.Sub(assets, "shared")
	router.Mount("/assets/", etag.Handler(http.FileServer(http.FS(assetsDir)), true))
	router.Get("/manifest.webmanifest", shared.HandleWebManifest())

	secureMiddleware := secure.New(secure.Options{
		HostsProxyHeaders:     []string{"X-Forwarded-Host"},
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		ForceSTSHeader:        true,
		IsDevelopment:         !config.IsProduction(),
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
	if !config.IsProduction() {
		cookieName = "__Insecure-csrf"
	}

	CSRF := csrf.Protect([]byte(config.CSRFSecret), csrf.CookieName(cookieName), csrf.FieldName("CSRFToken"), csrf.Secure(config.IsProduction()))
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
