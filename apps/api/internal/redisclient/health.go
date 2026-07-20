package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type HealthChecker struct {
	client *redis.Client
}

func NewHealthChecker(
	client *redis.Client,
) *HealthChecker {
	return &HealthChecker{
		client: client,
	}
}

func (h *HealthChecker) Check(
	ctx context.Context,
) error {
	checkContext, cancel := context.WithTimeout(
		ctx,
		2*time.Second,
	)
	defer cancel()

	return h.client.Ping(checkContext).Err()
}
