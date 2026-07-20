package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLoginTransactionStore struct {
	client    *redis.Client
	keyPrefix string
}

func NewRedisLoginTransactionStore(
	client *redis.Client,
	keyPrefix string,
) (*RedisLoginTransactionStore, error) {
	if client == nil {
		return nil, fmt.Errorf(
			"Redis client is required",
		)
	}

	keyPrefix = strings.TrimSpace(keyPrefix)

	if keyPrefix == "" {
		return nil, fmt.Errorf(
			"Redis key prefix is required",
		)
	}

	return &RedisLoginTransactionStore{
		client: client,
		keyPrefix: strings.TrimRight(
			keyPrefix,
			":",
		),
	}, nil
}

func (s *RedisLoginTransactionStore) Save(
	ctx context.Context,
	transaction LoginTransaction,
) error {
	if strings.TrimSpace(transaction.State) == "" {
		return fmt.Errorf(
			"login transaction state is required",
		)
	}

	ttl := time.Until(transaction.ExpiresAt)

	if ttl <= 0 {
		return ErrLoginTransactionExpired
	}

	payload, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf(
			"encode login transaction: %w",
			err,
		)
	}

	key := s.transactionKey(transaction.State)

	created, err := s.client.SetNX(
		ctx,
		key,
		payload,
		ttl,
	).Result()
	if err != nil {
		return fmt.Errorf(
			"store login transaction in Redis: %w",
			err,
		)
	}

	if !created {
		return fmt.Errorf(
			"login transaction state already exists",
		)
	}

	return nil
}

func (s *RedisLoginTransactionStore) Consume(
	ctx context.Context,
	state string,
) (LoginTransaction, error) {
	state = strings.TrimSpace(state)

	if state == "" {
		return LoginTransaction{},
			ErrLoginTransactionNotFound
	}

	key := s.transactionKey(state)

	payload, err := s.client.GetDel(
		ctx,
		key,
	).Bytes()

	if errors.Is(err, redis.Nil) {
		return LoginTransaction{},
			ErrLoginTransactionNotFound
	}

	if err != nil {
		return LoginTransaction{}, fmt.Errorf(
			"consume login transaction from Redis: %w",
			err,
		)
	}

	var transaction LoginTransaction

	if err := json.Unmarshal(
		payload,
		&transaction,
	); err != nil {
		return LoginTransaction{}, fmt.Errorf(
			"decode login transaction: %w",
			err,
		)
	}

	if transaction.State != state {
		return LoginTransaction{}, fmt.Errorf(
			"stored login transaction state does not match key",
		)
	}

	if transaction.Expired(time.Now().UTC()) {
		return LoginTransaction{},
			ErrLoginTransactionExpired
	}

	return transaction, nil
}

func (s *RedisLoginTransactionStore) DeleteExpired(
	ctx context.Context,
	now time.Time,
) (int, error) {
	// Redis removes these keys automatically using their TTL.
	return 0, nil
}

func (s *RedisLoginTransactionStore) transactionKey(
	state string,
) string {
	return fmt.Sprintf(
		"%s:oidc:login:%s",
		s.keyPrefix,
		state,
	)
}
