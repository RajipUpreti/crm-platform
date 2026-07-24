package membership

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const membershipUniqueConstraint = "memberships_tenant_user_unique"

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
) (Membership, error) {
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
			$3,
			$4,
			$5
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

	createdMembership, err := scanMembership(
		r.pool.QueryRow(
			ctx,
			query,
			input.TenantID,
			input.UserID,
			input.Role,
			input.Status,
			input.JoinedAt,
		),
	)
	if err != nil {
		if isConstraintViolation(
			err,
			membershipUniqueConstraint,
		) {
			return Membership{},
				ErrAlreadyExists
		}

		return Membership{}, fmt.Errorf(
			"create membership: %w",
			err,
		)
	}

	return createdMembership, nil
}

func (r *PostgresRepository) FindByID(
	ctx context.Context,
	id string,
) (Membership, error) {
	const query = `
		SELECT
			id::text,
			tenant_id::text,
			user_id::text,
			role,
			status,
			joined_at,
			created_at,
			updated_at
		FROM memberships
		WHERE id = $1
	`

	foundMembership, err := scanMembership(
		r.pool.QueryRow(
			ctx,
			query,
			id,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Membership{}, ErrNotFound
	}

	if err != nil {
		return Membership{}, fmt.Errorf(
			"find membership by ID: %w",
			err,
		)
	}

	return foundMembership, nil
}

func (r *PostgresRepository) FindByTenantAndUser(
	ctx context.Context,
	tenantID string,
	userID string,
) (Membership, error) {
	const query = `
		SELECT
			id::text,
			tenant_id::text,
			user_id::text,
			role,
			status,
			joined_at,
			created_at,
			updated_at
		FROM memberships
		WHERE tenant_id = $1
		  AND user_id = $2
	`

	foundMembership, err := scanMembership(
		r.pool.QueryRow(
			ctx,
			query,
			tenantID,
			userID,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Membership{}, ErrNotFound
	}

	if err != nil {
		return Membership{}, fmt.Errorf(
			"find tenant membership: %w",
			err,
		)
	}

	return foundMembership, nil
}

func (r *PostgresRepository) ListByTenantID(
	ctx context.Context,
	tenantID string,
) ([]Membership, error) {
	const query = `
		SELECT
			id::text,
			tenant_id::text,
			user_id::text,
			role,
			status,
			joined_at,
			created_at,
			updated_at
		FROM memberships
		WHERE tenant_id = $1
		ORDER BY
			created_at ASC,
			id ASC
	`

	return r.list(
		ctx,
		query,
		tenantID,
	)
}

func (r *PostgresRepository) ListByUserID(
	ctx context.Context,
	userID string,
) ([]Membership, error) {
	const query = `
		SELECT
			id::text,
			tenant_id::text,
			user_id::text,
			role,
			status,
			joined_at,
			created_at,
			updated_at
		FROM memberships
		WHERE user_id = $1
		ORDER BY
			created_at ASC,
			id ASC
	`

	return r.list(
		ctx,
		query,
		userID,
	)
}

func (r *PostgresRepository) ListDetailedByTenantID(
	ctx context.Context,
	tenantID string,
) ([]Member, error) {
	const query = `
		SELECT
			m.id::text,
			m.tenant_id::text,
			m.user_id::text,
			m.role,
			m.status,
			m.joined_at,
			u.id::text,
			u.email,
			COALESCE(u.display_name, ''),
			u.status
		FROM memberships m
		JOIN users u ON u.id = m.user_id
		WHERE m.tenant_id = $1
		  AND m.status <> 'LEFT'
		ORDER BY
			CASE m.role
				WHEN 'OWNER' THEN 1
				WHEN 'ADMIN' THEN 2
				ELSE 3
			END,
			LOWER(u.display_name) NULLS LAST,
			LOWER(u.email),
			m.id
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list detailed memberships: %w",
			err,
		)
	}
	defer rows.Close()

	members := make([]Member, 0)
	for rows.Next() {
		foundMember, err := scanMember(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan detailed membership: %w",
				err,
			)
		}

		members = append(members, foundMember)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate detailed memberships: %w",
			err,
		)
	}

	return members, nil
}

func (r *PostgresRepository) FindDetailedByID(
	ctx context.Context,
	id string,
) (Member, error) {
	const query = `
		SELECT
			m.id::text,
			m.tenant_id::text,
			m.user_id::text,
			m.role,
			m.status,
			m.joined_at,
			u.id::text,
			u.email,
			COALESCE(u.display_name, ''),
			u.status
		FROM memberships m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = $1
	`

	foundMember, err := scanMember(
		r.pool.QueryRow(ctx, query, id),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	if err != nil {
		return Member{}, fmt.Errorf(
			"find detailed membership: %w",
			err,
		)
	}

	return foundMember, nil
}

func (r *PostgresRepository) CountActiveOwners(
	ctx context.Context,
	tenantID string,
) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM memberships
		WHERE tenant_id = $1
		  AND role = 'OWNER'
		  AND status = 'ACTIVE'
	`

	var count int
	if err := r.pool.QueryRow(
		ctx,
		query,
		tenantID,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf(
			"count active owners: %w",
			err,
		)
	}

	return count, nil
}

func (r *PostgresRepository) UpdateRoleForTenant(
	ctx context.Context,
	id string,
	tenantID string,
	role Role,
) (Membership, error) {
	const query = `
		UPDATE memberships
		SET
			role = $3,
			updated_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
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

	updatedMembership, err := scanMembership(
		r.pool.QueryRow(
			ctx,
			query,
			id,
			tenantID,
			role,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Membership{}, ErrNotFound
	}
	if err != nil {
		return Membership{}, fmt.Errorf(
			"update tenant membership role: %w",
			err,
		)
	}

	return updatedMembership, nil
}

func (r *PostgresRepository) UpdateStatusForTenant(
	ctx context.Context,
	id string,
	tenantID string,
	status Status,
) (Membership, error) {
	const query = `
		UPDATE memberships
		SET
			status = $3,
			joined_at = CASE
				WHEN $3 = 'ACTIVE'
					THEN COALESCE(joined_at, NOW())
				ELSE joined_at
			END,
			updated_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
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

	updatedMembership, err := scanMembership(
		r.pool.QueryRow(
			ctx,
			query,
			id,
			tenantID,
			status,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Membership{}, ErrNotFound
	}
	if err != nil {
		return Membership{}, fmt.Errorf(
			"update tenant membership status: %w",
			err,
		)
	}

	return updatedMembership, nil
}

func (r *PostgresRepository) list(
	ctx context.Context,
	query string,
	value string,
) ([]Membership, error) {
	rows, err := r.pool.Query(
		ctx,
		query,
		value,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query memberships: %w",
			err,
		)
	}
	defer rows.Close()

	memberships := make([]Membership, 0)

	for rows.Next() {
		foundMembership, err := scanMembership(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan membership: %w",
				err,
			)
		}

		memberships = append(
			memberships,
			foundMembership,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate memberships: %w",
			err,
		)
	}

	return memberships, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMembership(
	row rowScanner,
) (Membership, error) {
	var foundMembership Membership

	err := row.Scan(
		&foundMembership.ID,
		&foundMembership.TenantID,
		&foundMembership.UserID,
		&foundMembership.Role,
		&foundMembership.Status,
		&foundMembership.JoinedAt,
		&foundMembership.CreatedAt,
		&foundMembership.UpdatedAt,
	)
	if err != nil {
		return Membership{}, err
	}

	return foundMembership, nil
}

func scanMember(
	row rowScanner,
) (Member, error) {
	var foundMember Member

	err := row.Scan(
		&foundMember.Membership.ID,
		&foundMember.Membership.TenantID,
		&foundMember.Membership.UserID,
		&foundMember.Membership.Role,
		&foundMember.Membership.Status,
		&foundMember.Membership.JoinedAt,
		&foundMember.User.ID,
		&foundMember.User.Email,
		&foundMember.User.DisplayName,
		&foundMember.User.Status,
	)
	if err != nil {
		return Member{}, err
	}

	return foundMember, nil
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
