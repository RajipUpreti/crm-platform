package onboarding

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/tenant"
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

func (r *PostgresRepository) FindPrimaryAccess(
	ctx context.Context,
	userID string,
) (Result, error) {
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
			m.created_at ASC,
			m.id ASC
		LIMIT 1
	`

	result, err := scanResult(
		r.pool.QueryRow(
			ctx,
			query,
			userID,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{},
			membership.ErrNotFound
	}

	if err != nil {
		return Result{}, fmt.Errorf(
			"find primary tenant access: %w",
			err,
		)
	}

	result.Created = false

	return result, nil
}

func (r *PostgresRepository) ProvisionPersonalTenant(
	ctx context.Context,
	userID string,
	tenantName string,
	tenantSlug string,
) (Result, error) {
	tx, err := r.pool.BeginTx(
		ctx,
		pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		},
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"begin onboarding transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	existing, err := findPrimaryAccessWithQuerier(
		ctx,
		tx,
		userID,
	)
	if err == nil {
		existing.Created = false

		if commitErr := tx.Commit(ctx); commitErr != nil {
			return Result{}, fmt.Errorf(
				"commit existing onboarding access: %w",
				commitErr,
			)
		}

		return existing, nil
	}

	if !errors.Is(
		err,
		membership.ErrNotFound,
	) {
		return Result{}, err
	}

	createdTenant, err := createTenantWithQuerier(
		ctx,
		tx,
		tenantName,
		tenantSlug,
	)
	if err != nil {
		return Result{}, err
	}

	createdMembership, err := createOwnerMembershipWithQuerier(
		ctx,
		tx,
		createdTenant.ID,
		userID,
	)
	if err != nil {
		return Result{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf(
			"commit onboarding transaction: %w",
			err,
		)
	}

	return Result{
		Tenant:     createdTenant,
		Membership: createdMembership,
		Created:    true,
	}, nil
}

type querier interface {
	QueryRow(
		ctx context.Context,
		sql string,
		args ...any,
	) pgx.Row
}

func findPrimaryAccessWithQuerier(
	ctx context.Context,
	db querier,
	userID string,
) (Result, error) {
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
			m.created_at ASC,
			m.id ASC
		LIMIT 1
		FOR UPDATE OF m
	`

	result, err := scanResult(
		db.QueryRow(
			ctx,
			query,
			userID,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{},
			membership.ErrNotFound
	}

	if err != nil {
		return Result{}, fmt.Errorf(
			"find onboarding access in transaction: %w",
			err,
		)
	}

	return result, nil
}

func createTenantWithQuerier(
	ctx context.Context,
	db querier,
	name string,
	slug string,
) (tenant.Tenant, error) {
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

	var created tenant.Tenant

	err := db.QueryRow(
		ctx,
		query,
		name,
		slug,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Slug,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		if isConstraintViolation(
			err,
			tenantSlugUniqueConstraint,
		) {
			return tenant.Tenant{},
				tenant.ErrSlugAlreadyExists
		}

		return tenant.Tenant{}, fmt.Errorf(
			"create personal tenant: %w",
			err,
		)
	}

	return created, nil
}

func createOwnerMembershipWithQuerier(
	ctx context.Context,
	db querier,
	tenantID string,
	userID string,
) (membership.Membership, error) {
	const query = `
		INSERT INTO memberships (
			tenant_id,
			user_id,
			role,
			status,
			joined_at
		)
		VALUES (
			$1,
			$2,
			'OWNER',
			'ACTIVE',
			NOW()
		)
		RETURNING
			id::text,
			tenant_id::text,
			user_id::text,
			role,
			status,
			joined_at,
			created_at,
			updated_at
	`

	var created membership.Membership

	err := db.QueryRow(
		ctx,
		query,
		tenantID,
		userID,
	).Scan(
		&created.ID,
		&created.TenantID,
		&created.UserID,
		&created.Role,
		&created.Status,
		&created.JoinedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return membership.Membership{},
			fmt.Errorf(
				"create owner membership: %w",
				err,
			)
	}

	return created, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanResult(
	row rowScanner,
) (Result, error) {
	var result Result

	err := row.Scan(
		&result.Tenant.ID,
		&result.Tenant.Name,
		&result.Tenant.Slug,
		&result.Tenant.Status,
		&result.Tenant.CreatedAt,
		&result.Tenant.UpdatedAt,

		&result.Membership.ID,
		&result.Membership.TenantID,
		&result.Membership.UserID,
		&result.Membership.Role,
		&result.Membership.Status,
		&result.Membership.JoinedAt,
		&result.Membership.CreatedAt,
		&result.Membership.UpdatedAt,
	)
	if err != nil {
		return Result{}, err
	}

	return result, nil
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
