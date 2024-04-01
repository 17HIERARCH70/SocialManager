package api

import (
	"context"
	"fmt"
	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/17HIERARCH70/SocialManager/internal/handlers/authHandlers"
	"github.com/17HIERARCH70/SocialManager/internal/handlers/gmailHandlers"
	"github.com/17HIERARCH70/SocialManager/internal/handlers/healthHandlers"
	"github.com/17HIERARCH70/SocialManager/internal/handlers/userHandlers"
	"github.com/17HIERARCH70/SocialManager/internal/services"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/exp/slog"
	"net/http"
)

// App encapsulates all the dependencies for running our server
type App struct {
	dbPool        *pgxpool.Pool
	cfg           *config.Config
	log           *slog.Logger
	router        *mux.Router
	userService   services.UserServiceMethods
	healthService services.HealthServiceMethods
	gmailService  services.GmailServiceMethods
	httpServer    *http.Server
}

// NewApp sets up a new instance of App including the router
func NewApp(
	cfg *config.Config,
	dbPool *pgxpool.Pool,
	log *slog.Logger,
	userService services.UserServiceMethods,
	healthService services.HealthServiceMethods,
	gmailHandler services.GmailServiceMethods,
) *App {
	app := &App{
		dbPool:        dbPool,
		cfg:           cfg,
		log:           log,
		userService:   userService,
		healthService: healthService,
		gmailService:  gmailHandler,
	}
	app.router = app.NewRouter()
	return app
}

// NewRouter configures and returns a new gorilla/mux router instance with our API routes
func (a *App) NewRouter() *mux.Router {
	router := mux.NewRouter()

	authHandler := authHandlers.AuthHandler{
		UserService: a.userService,
		Log:         a.log,
	}

	userHandler := userHandlers.UserHandler{
		UserService: a.userService,
		Log:         a.log,
	}

	healthCheckHandler := healthHandlers.HealthCheckHandler{
		HealthService: a.healthService,
		Log:           a.log,
	}

	gmailHandler := gmailHandlers.GmailHandler{
		GmailService: a.gmailService,
		Log:          a.log,
	}

	// Health routes
	router.HandleFunc("/api/health", healthCheckHandler.CheckHealth).Methods(http.MethodGet)
	// Auth routes
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods(http.MethodPost)
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods(http.MethodPost)
	// User routes
	router.HandleFunc("/api/users/list", userHandler.ListUsers).Methods(http.MethodGet)
	router.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods(http.MethodGet)
	router.HandleFunc("/api/users/{id}", userHandler.UpdateUser).Methods(http.MethodPut)
	router.HandleFunc("/api/users/{id}", userHandler.DeleteUser).Methods(http.MethodDelete)
	// Gmail routes
	router.HandleFunc("/api/gmail/fetch", gmailHandler.FetchEmailsHandler).Methods(http.MethodPost)
	return router
}

// Run starts the HTTP server with the configured routes
func (a *App) Run() {
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port),
		Handler: a.router,
	}

	a.log.Info("Starting server", "port", a.cfg.Server.Port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port), a.router); err != nil {
		a.log.Error("Failed to start server", "error", err)
	}
}

// Shutdown turns off the HTTP server
func (a *App) Shutdown(ctx context.Context) interface{} {
	if a.httpServer != nil {
		// Останавливаем HTTP сервер
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.log.Error("Failed to shutdown the server properly", "error", err)
			return err
		}
		a.log.Info("HTTP server stopped")
	}

	// Закрываем соединение с базой данных
	if a.dbPool != nil {
		a.dbPool.Close()
		a.log.Info("Database connection closed")
	}
	return nil
}
