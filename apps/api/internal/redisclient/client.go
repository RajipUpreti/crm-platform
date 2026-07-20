package redisclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Address  string
	Password string
	Database int
}

func New(
	ctx context.Context,
	cfg Config,
) (*redis.Client, error) {
	if strings.TrimSpace(cfg.Address) == "" {
		return nil, fmt.Errorf(
			"Redis address is required",
		)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.Database,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		PoolSize:     10,
		MinIdleConns: 1,
	})

	pingContext, cancel := context.WithTimeout(
		ctx,
		5*time.Second,
	)
	defer cancel()

	if err := client.Ping(pingContext).Err(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf(
			"ping Redis: %w",
			err,
		)
	}

	return client, nil
}
