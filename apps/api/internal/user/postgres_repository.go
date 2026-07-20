package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(
	pool *pgxpool.Pool,
) (*PostgresRepository, error) {
	if pool == nil {
		return nil, fmt.Errorf(
			"PostgreSQL pool is required",
		)
	}

	return &PostgresRepository{
		pool: pool,
	}, nil
}

func (r *PostgresRepository) UpsertFromIdentity(
	ctx context.Context,
	identity Identity,
) (User, error) {
	const query = `
		INSERT INTO users (
			identity_provider,
			identity_provider_user_id,
			email,
			email_verified,
			first_name,
			last_name,
			display_name
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			NULLIF($5, ''),
			NULLIF($6, ''),
			NULLIF($7, '')
		)
		ON CONFLICT (
			identity_provider,
			identity_provider_user_id
		)
		DO UPDATE SET
			email = EXCLUDED.email,
			email_verified = EXCLUDED.email_verified,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			display_name = EXCLUDED.display_name,
			updated_at = NOW()
		RETURNING
			id::text,
			identity_provider,
			identity_provider_user_id,
			email,
			email_verified,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(display_name, ''),
			status,
			created_at,
			updated_at
	`

	return scanUser(
		r.pool.QueryRow(
			ctx,
			query,
			identity.Provider,
			identity.ProviderUserID,
			identity.Email,
			identity.EmailVerified,
			identity.FirstName,
			identity.LastName,
			identity.DisplayName,
		),
	)
}

func (r *PostgresRepository) FindByID(
	ctx context.Context,
	id string,
) (User, error) {
	const query = `
		SELECT
			id::text,
			identity_provider,
			identity_provider_user_id,
			email,
			email_verified,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(display_name, ''),
			status,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	storedUser, err := scanUser(
		r.pool.QueryRow(
			ctx,
			query,
			id,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}

	return storedUser, err
}

func (r *PostgresRepository) FindByProviderIdentity(
	ctx context.Context,
	provider string,
	providerUserID string,
) (User, error) {
	const query = `
		SELECT
			id::text,
			identity_provider,
			identity_provider_user_id,
			email,
			email_verified,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(display_name, ''),
			status,
			created_at,
			updated_at
		FROM users
		WHERE identity_provider = $1
		  AND identity_provider_user_id = $2
	`

	storedUser, err := scanUser(
		r.pool.QueryRow(
			ctx,
			query,
			strings.TrimSpace(provider),
			strings.TrimSpace(providerUserID),
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}

	return storedUser, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(
	row rowScanner,
) (User, error) {
	var storedUser User

	err := row.Scan(
		&storedUser.ID,
		&storedUser.IdentityProvider,
		&storedUser.IdentityProviderUserID,
		&storedUser.Email,
		&storedUser.EmailVerified,
		&storedUser.FirstName,
		&storedUser.LastName,
		&storedUser.DisplayName,
		&storedUser.Status,
		&storedUser.CreatedAt,
		&storedUser.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf(
			"scan user: %w",
			err,
		)
	}

	return storedUser, nil
}
