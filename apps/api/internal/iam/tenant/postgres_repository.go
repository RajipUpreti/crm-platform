package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tenantSlugUniqueConstraint = "tenants_slug_unique"

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

func (r *PostgresRepository) Create(
	ctx context.Context,
	input CreateInput,
) (Tenant, error) {
	const query = `
		INSERT INTO tenants (
			name,
			slug
		)
		VALUES (
			$1,
			$2
		)
		RETURNING
			id::text,
			name,
			slug,
			status,
			created_at,
			updated_at
	`

	createdTenant, err := scanTenant(
		r.pool.QueryRow(
			ctx,
			query,
			input.Name,
			input.Slug,
		),
	)
	if err != nil {
		if isConstraintViolation(
			err,
			tenantSlugUniqueConstraint,
		) {
			return Tenant{}, ErrSlugAlreadyExists
		}

		return Tenant{}, fmt.Errorf(
			"create tenant: %w",
			err,
		)
	}

	return createdTenant, nil
}

func (r *PostgresRepository) Update(
	ctx context.Context,
	id string,
	input UpdateInput,
) (Tenant, error) {
	const query = `
		UPDATE tenants
		SET
			name = COALESCE($2, name),
			status = COALESCE($3, status),
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id::text,
			name,
			slug,
			status,
			created_at,
			updated_at
	`

	updatedTenant, err := scanTenant(
		r.pool.QueryRow(
			ctx,
			query,
			id,
			input.Name,
			input.Status,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrNotFound
	}

	if err != nil {
		return Tenant{}, fmt.Errorf(
			"update tenant: %w",
			err,
		)
	}

	return updatedTenant, nil
}

func (r *PostgresRepository) FindByID(
	ctx context.Context,
	id string,
) (Tenant, error) {
	const query = `
		SELECT
			id::text,
			name,
			slug,
			status,
			created_at,
			updated_at
		FROM tenants
		WHERE id = $1
	`

	foundTenant, err := scanTenant(
		r.pool.QueryRow(
			ctx,
			query,
			id,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrNotFound
	}

	if err != nil {
		return Tenant{}, fmt.Errorf(
			"find tenant by ID: %w",
			err,
		)
	}

	return foundTenant, nil
}

func (r *PostgresRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (Tenant, error) {
	const query = `
		SELECT
			id::text,
			name,
			slug,
			status,
			created_at,
			updated_at
		FROM tenants
		WHERE slug = $1
	`

	foundTenant, err := scanTenant(
		r.pool.QueryRow(
			ctx,
			query,
			slug,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrNotFound
	}

	if err != nil {
		return Tenant{}, fmt.Errorf(
			"find tenant by slug: %w",
			err,
		)
	}

	return foundTenant, nil
}

func (r *PostgresRepository) ListByUserID(
	ctx context.Context,
	userID string,
) ([]Tenant, error) {
	const query = `
		SELECT
			t.id::text,
			t.name,
			t.slug,
			t.status,
			t.created_at,
			t.updated_at
		FROM tenants t
		INNER JOIN memberships m
			ON m.tenant_id = t.id
		WHERE m.user_id = $1
		  AND m.status = 'ACTIVE'
		  AND t.status <> 'DELETED'
		ORDER BY
			t.name ASC,
			t.id ASC
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list user tenants: %w",
			err,
		)
	}
	defer rows.Close()

	tenants := make([]Tenant, 0)

	for rows.Next() {
		foundTenant, err := scanTenant(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan listed tenant: %w",
				err,
			)
		}

		tenants = append(
			tenants,
			foundTenant,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate user tenants: %w",
			err,
		)
	}

	return tenants, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTenant(
	row rowScanner,
) (Tenant, error) {
	var foundTenant Tenant

	err := row.Scan(
		&foundTenant.ID,
		&foundTenant.Name,
		&foundTenant.Slug,
		&foundTenant.Status,
		&foundTenant.CreatedAt,
		&foundTenant.UpdatedAt,
	)
	if err != nil {
		return Tenant{}, err
	}

	return foundTenant, nil
}

func isConstraintViolation(
	err error,
	constraintName string,
) bool {
	var postgresError *pgconn.PgError

	if !errors.As(err, &postgresError) {
		return false
	}

	return postgresError.ConstraintName ==
		constraintName
}

func (r *PostgresRepository) ListAccessByUserID(
	ctx context.Context,
	userID string,
) ([]Access, error) {
	const query = `
		SELECT
			t.id::text,
			t.name,
			t.slug,
			t.status,
			t.created_at,
			t.updated_at,

			m.id::text,
			m.tenant_id::text,
			m.user_id::text,
			m.role,
			m.status,
			m.joined_at,
			m.created_at,
			m.updated_at
		FROM memberships m
		INNER JOIN tenants t
			ON t.id = m.tenant_id
		WHERE m.user_id = $1
		  AND m.status = 'ACTIVE'
		  AND t.status = 'ACTIVE'
		ORDER BY
			t.name ASC,
			t.id ASC
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list tenant access: %w",
			err,
		)
	}
	defer rows.Close()

	accesses := make([]Access, 0)

	for rows.Next() {
		var access Access

		err := rows.Scan(
			&access.Tenant.ID,
			&access.Tenant.Name,
			&access.Tenant.Slug,
			&access.Tenant.Status,
			&access.Tenant.CreatedAt,
			&access.Tenant.UpdatedAt,

			&access.Membership.ID,
			&access.Membership.TenantID,
			&access.Membership.UserID,
			&access.Membership.Role,
			&access.Membership.Status,
			&access.Membership.JoinedAt,
			&access.Membership.CreatedAt,
			&access.Membership.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"scan tenant access: %w",
				err,
			)
		}

		accesses = append(
			accesses,
			access,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate tenant access: %w",
			err,
		)
	}

	return accesses, nil
}
