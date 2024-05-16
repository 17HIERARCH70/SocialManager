package main

import (
	_ "github.com/17HIERARCH70/SocialManager/docs"
	"github.com/17HIERARCH70/SocialManager/internal/api"
	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/17HIERARCH70/SocialManager/internal/logger"
	"github.com/17HIERARCH70/SocialManager/internal/storage/postgresql"
	_ "net/http/pprof"
	"os"
)

// @title SocialManager API
// @version 1.0
// @description API for managing social accounts and emails.
// @host localhost:8080
// @BasePath /api

func main() {
	// Initialize the config
	cfg := config.MustLoad()

	// Initialize the logger
	log := logger.SetupLogger(cfg.Env)

	// Initialize the Psql database connection
	psqlPool, err := postgresql.InitializeDB(cfg)
	if err != nil {
		log.Error("Failed to initialize database:", err)
		os.Exit(-1)
	}
	defer psqlPool.Close()

	// Initialize the application
	app := api.NewApp(psqlPool, cfg, log)

	// Run the application
	app.Run()
}
