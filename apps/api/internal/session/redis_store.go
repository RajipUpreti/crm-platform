package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client    redis.UniversalClient
	keyPrefix string
}

func NewRedisStore(
	client redis.UniversalClient,
	keyPrefix string,
) (*RedisStore, error) {
	if client == nil {
		return nil, fmt.Errorf(
			"Redis client is required",
		)
	}

	keyPrefix = strings.TrimSpace(keyPrefix)
	if keyPrefix == "" {
		return nil, fmt.Errorf(
			"Redis session key prefix is required",
		)
	}

	return &RedisStore{
		client:    client,
		keyPrefix: strings.TrimSuffix(keyPrefix, ":"),
	}, nil
}

func (s *RedisStore) Create(
	ctx context.Context,
	digest string,
	storedSession Session,
	ttl time.Duration,
) error {
	if strings.TrimSpace(digest) == "" {
		return ErrInvalid
	}

	if ttl <= 0 {
		return fmt.Errorf(
			"%w: session TTL must be positive",
			ErrInvalid,
		)
	}

	payload, err := json.Marshal(storedSession)
	if err != nil {
		return fmt.Errorf(
			"encode session: %w",
			err,
		)
	}

	created, err := s.client.SetNX(
		ctx,
		s.key(digest),
		payload,
		ttl,
	).Result()
	if err != nil {
		return fmt.Errorf(
			"store session in Redis: %w",
			err,
		)
	}

	if !created {
		return fmt.Errorf(
			"%w: session token collision",
			ErrInvalid,
		)
	}

	return nil
}

func (s *RedisStore) Find(
	ctx context.Context,
	digest string,
) (Session, error) {
	if strings.TrimSpace(digest) == "" {
		return Session{}, ErrInvalid
	}

	payload, err := s.client.Get(
		ctx,
		s.key(digest),
	).Bytes()
	if errors.Is(err, redis.Nil) {
		return Session{}, ErrNotFound
	}

	if err != nil {
		return Session{}, fmt.Errorf(
			"read session from Redis: %w",
			err,
		)
	}

	var storedSession Session

	if err := json.Unmarshal(
		payload,
		&storedSession,
	); err != nil {
		return Session{}, fmt.Errorf(
			"decode session: %w",
			err,
		)
	}

	return storedSession, nil
}

func (s *RedisStore) Delete(
	ctx context.Context,
	digest string,
) error {
	if strings.TrimSpace(digest) == "" {
		return ErrInvalid
	}

	if err := s.client.Del(
		ctx,
		s.key(digest),
	).Err(); err != nil {
		return fmt.Errorf(
			"delete session from Redis: %w",
			err,
		)
	}

	return nil
}

func (s *RedisStore) key(
	digest string,
) string {
	return s.keyPrefix + ":" + digest
}
