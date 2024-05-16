package api

import (
	"context"
	"fmt"
	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/17HIERARCH70/SocialManager/internal/services/authService"
	emailService2 "github.com/17HIERARCH70/SocialManager/internal/services/emailService"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/exp/slog"
	"net/http"
)

type App struct {
	psql       *pgxpool.Pool
	cfg        *config.Config
	log        *slog.Logger
	router     *mux.Router
	httpServer *http.Server
	emailSvc   *emailService2.EmailService
}

func NewApp(psql *pgxpool.Pool, cfg *config.Config, log *slog.Logger) *App {
	// Create a new router
	router := mux.NewRouter()

	// Initialize email service
	authServices, _ := authService.NewAuthService(psql, cfg, log)
	emailService, _ := emailService2.NewEmailService(psql, cfg, authServices, log)

	// Create the App instance
	app := &App{
		psql:     psql,
		cfg:      cfg,
		log:      log,
		router:   router,
		emailSvc: emailService,
	}

	app.SetupRoutes()

	return app
}

func (a *App) Run() {
	go a.emailSvc.StartEmailPolling()

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port),
		Handler: a.router,
	}
	a.log.Info("Starting server", "port", a.cfg.Server.Port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port), a.router); err != nil {
		a.log.Error("Failed to start server", "error", err)
	}
}

func (a *App) Shutdown(ctx context.Context) interface{} {
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.log.Error("Failed to shutdown the server properly", "error", err)
			return err
		}
		a.log.Info("HTTP server stopped")
	}
	if a.psql != nil {
		a.psql.Close()
		a.log.Info("Database connection closed")
	}

	return nil
}
