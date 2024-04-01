package main

import (
	"context"
	"github.com/17HIERARCH70/SocialManager/internal/api"
	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/17HIERARCH70/SocialManager/internal/lib/logger/slogpretty"
	"github.com/17HIERARCH70/SocialManager/internal/services"
	"github.com/17HIERARCH70/SocialManager/internal/storage/postgresql"
	"golang.org/x/exp/slog"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Initialize the config
	cfg := config.MustLoad()

	// Initialize the logger
	log := setupLogger(cfg.Env)

	// Initialize the database connection
	dbPool, err := postgresql.InitializeDB(cfg)
	if err != nil {
		log.Error("Failed to initialize database:", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Initialize the services
	userService := services.NewUserService(dbPool)
	healthService := services.NewHealthService(dbPool)
	gmailService := services.NewGmailService(context.Background(), cfg.Gmail.SecretPath, dbPool, log)

	// Initialize the application
	app := api.NewApp(cfg, dbPool, log, userService, healthService, gmailService)

	// Graceful stop
	stopChan := make(chan os.Signal, 1)

	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stopChan

		log.Info("Shutting down server...")

		dbPool.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := app.Shutdown(ctx); err != nil {
			log.Error("Server Shutdown Failed:", err)
		}

		os.Exit(0)
	}()

	// Run HTTP Servers
	app.Run()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
