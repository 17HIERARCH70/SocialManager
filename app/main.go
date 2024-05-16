package main

import (
	"github.com/17HIERARCH70/SocialManager/internal/api"
	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/17HIERARCH70/SocialManager/internal/logger"
	"github.com/17HIERARCH70/SocialManager/internal/storage/postgresql"
	_ "net/http/pprof"
	"os"
)

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

	app.Run()
}
