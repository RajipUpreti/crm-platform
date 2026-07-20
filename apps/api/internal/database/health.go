package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthChecker struct {
	pool *pgxpool.Pool
}

func NewHealthChecker(
	pool *pgxpool.Pool,
) *HealthChecker {
	return &HealthChecker{
		pool: pool,
	}
}

func (h *HealthChecker) Check(
	ctx context.Context,
) error {
	checkContext, cancel :=
		context.WithTimeout(
			ctx,
			2*time.Second,
		)
	defer cancel()

	return h.pool.Ping(checkContext)
}
