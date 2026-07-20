package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	URL             string
	MaxConnections  int32
	MinConnections  int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func OpenPostgres(
	ctx context.Context,
	cfg PostgresConfig,
) (*pgxpool.Pool, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf(
			"PostgreSQL URL is required",
		)
	}

	poolConfig, err :=
		pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf(
			"parse PostgreSQL configuration: %w",
			err,
		)
	}

	if cfg.MaxConnections > 0 {
		poolConfig.MaxConns =
			cfg.MaxConnections
	}

	if cfg.MinConnections >= 0 {
		poolConfig.MinConns =
			cfg.MinConnections
	}

	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime =
			cfg.MaxConnLifetime
	}

	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime =
			cfg.MaxConnIdleTime
	}

	pool, err := pgxpool.NewWithConfig(
		ctx,
		poolConfig,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create PostgreSQL pool: %w",
			err,
		)
	}

	pingContext, cancel :=
		context.WithTimeout(
			ctx,
			5*time.Second,
		)
	defer cancel()

	if err := pool.Ping(pingContext); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"ping PostgreSQL: %w",
			err,
		)
	}

	return pool, nil
}
