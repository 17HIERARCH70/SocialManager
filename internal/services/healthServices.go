package services

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

// HealthServiceMethods defines the methods that a HealthService must implement.
type HealthServiceMethods interface {
	Check() error
}

// healthService provides health check functionalities.
type healthService struct {
	db *pgxpool.Pool
}

// NewHealthService creates a new instance of HealthService with a database pool.
func NewHealthService(dbPool *pgxpool.Pool) HealthServiceMethods {
	return &healthService{db: dbPool}
}

// Check returns nil if the service is healthy, otherwise an error indicating the issue.
func (s *healthService) Check() error {

	if err := s.db.Ping(context.Background()); err != nil {
		return err
	}
	return nil
}
