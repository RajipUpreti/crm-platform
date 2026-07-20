package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrLoginTransactionNotFound = errors.New(
		"login transaction not found",
	)

	ErrLoginTransactionExpired = errors.New(
		"login transaction expired",
	)
)

type LoginTransactionStore interface {
	Save(
		ctx context.Context,
		transaction LoginTransaction,
	) error

	Consume(
		ctx context.Context,
		state string,
	) (LoginTransaction, error)

	DeleteExpired(
		ctx context.Context,
		now time.Time,
	) (int, error)
}

type MemoryLoginTransactionStore struct {
	mu           sync.Mutex
	transactions map[string]LoginTransaction
}

func NewMemoryLoginTransactionStore() *MemoryLoginTransactionStore {
	return &MemoryLoginTransactionStore{
		transactions: make(map[string]LoginTransaction),
	}
}

func (s *MemoryLoginTransactionStore) Save(
	ctx context.Context,
	transaction LoginTransaction,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.transactions[transaction.State] = transaction

	return nil
}

func (s *MemoryLoginTransactionStore) Consume(
	ctx context.Context,
	state string,
) (LoginTransaction, error) {
	if err := ctx.Err(); err != nil {
		return LoginTransaction{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, exists := s.transactions[state]
	if !exists {
		return LoginTransaction{}, ErrLoginTransactionNotFound
	}

	// Delete before returning so every state can be used only once.
	delete(s.transactions, state)

	if transaction.Expired(time.Now().UTC()) {
		return LoginTransaction{}, ErrLoginTransactionExpired
	}

	return transaction, nil
}

func (s *MemoryLoginTransactionStore) DeleteExpired(
	ctx context.Context,
	now time.Time,
) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := 0

	for state, transaction := range s.transactions {
		if transaction.Expired(now) {
			delete(s.transactions, state)
			deleted++
		}
	}

	return deleted, nil
}

func StartLoginTransactionCleanup(
	ctx context.Context,
	store LoginTransactionStore,
	interval time.Duration,
	logger func(format string, args ...any),
) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case now := <-ticker.C:
				deleted, err := store.DeleteExpired(
					ctx,
					now.UTC(),
				)
				if err != nil {
					if !errors.Is(
						err,
						context.Canceled,
					) {
						logger(
							"delete expired login transactions: %v",
							err,
						)
					}

					continue
				}

				if deleted > 0 {
					logger(
						"deleted %d expired login transactions",
						deleted,
					)
				}
			}
		}
	}()
}
