package postgresql

import (
	"context"
	"fmt"
	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

// InitializeDB initializes and returns a connection pool to the PostgreSQL database.
func InitializeDB(cfg *config.Config) (*pgxpool.Pool, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("error configuring the database: %w", err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return pool, nil
}
