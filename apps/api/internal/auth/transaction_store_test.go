package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryLoginTransactionStoreConsume(t *testing.T) {
	t.Parallel()

	store := NewMemoryLoginTransactionStore()
	ctx := context.Background()

	transaction := LoginTransaction{
		State:        "state-value",
		Nonce:        "nonce-value",
		CodeVerifier: "verifier-value",
		ReturnTo:     "/dashboard",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(time.Minute),
	}

	if err := store.Save(ctx, transaction); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	result, err := store.Consume(
		ctx,
		transaction.State,
	)
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}

	if result.State != transaction.State {
		t.Fatalf(
			"Consume() state = %q; expected %q",
			result.State,
			transaction.State,
		)
	}

	_, err = store.Consume(ctx, transaction.State)

	if !errors.Is(
		err,
		ErrLoginTransactionNotFound,
	) {
		t.Fatalf(
			"second Consume() error = %v; expected %v",
			err,
			ErrLoginTransactionNotFound,
		)
	}
}

func TestMemoryLoginTransactionStoreRejectsExpired(
	t *testing.T,
) {
	t.Parallel()

	store := NewMemoryLoginTransactionStore()
	ctx := context.Background()

	transaction := LoginTransaction{
		State:     "expired-state",
		CreatedAt: time.Now().UTC().Add(-20 * time.Minute),
		ExpiresAt: time.Now().UTC().Add(-10 * time.Minute),
	}

	if err := store.Save(ctx, transaction); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	_, err := store.Consume(ctx, transaction.State)

	if !errors.Is(
		err,
		ErrLoginTransactionExpired,
	) {
		t.Fatalf(
			"Consume() error = %v; expected %v",
			err,
			ErrLoginTransactionExpired,
		)
	}
}
