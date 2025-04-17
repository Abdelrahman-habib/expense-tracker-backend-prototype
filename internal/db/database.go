package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service interface {
	Health() map[string]string
	Close() error
	Queries() *Queries
}

type service struct {
	cfg     config.DatabaseConfig
	db      *pgxpool.Pool
	queries *Queries
}

func NewService(cfg config.DatabaseConfig) Service {
	config, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		log.Fatal(err)
	}

	// Apply configuration
	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxLifetime
	config.MaxConnIdleTime = cfg.MaxIdleTime
	config.HealthCheckPeriod = cfg.HealthCheck

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}

	queries := New(pool)

	return &service{
		cfg:     cfg,
		db:      pool,
		queries: queries,
	}
}

// Health check
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.db.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	poolStats := s.db.Stat()
	stats["total_connections"] = fmt.Sprintf("%d", poolStats.TotalConns())
	stats["acquired_connections"] = fmt.Sprintf("%d", poolStats.AcquiredConns())
	stats["idle_connections"] = fmt.Sprintf("%d", poolStats.IdleConns())

	return stats
}

func (s *service) Close() error {
	s.db.Close()
	return nil
}

func (s *service) Queries() *Queries {
	return s.queries
}
