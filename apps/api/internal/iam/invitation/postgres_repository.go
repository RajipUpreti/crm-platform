package invitation

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rajipupreti/crm-platform/apps/api/internal/iam/membership"
)

const pendingInvitationUniqueIndex = "invitations_pending_tenant_email_unique"

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
) (Invitation, error) {
	const query = `
		INSERT INTO invitations (
			tenant_id,
			invited_by_user_id,
			email,
			role,
			token_digest,
			expires_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		)
		RETURNING
			id::text,
			tenant_id::text,
			invited_by_user_id::text,
			email,
			role,
			status,
			expires_at,
			accepted_at,
			created_at,
			updated_at
	`

	created, err := scanInvitation(
		r.pool.QueryRow(
			ctx,
			query,
			input.TenantID,
			input.InvitedByUserID,
			input.Email,
			input.Role,
			input.TokenDigest,
			input.ExpiresAt,
		),
	)
	if err != nil {
		if isConstraintViolation(
			err,
			pendingInvitationUniqueIndex,
		) {
			return Invitation{},
				ErrAlreadyPending
		}

		return Invitation{}, fmt.Errorf(
			"create invitation: %w",
			err,
		)
	}

	return created, nil
}

func (r *PostgresRepository) FindByTokenDigest(
	ctx context.Context,
	digest string,
) (Invitation, error) {
	const query = `
		SELECT
			id::text,
			tenant_id::text,
			invited_by_user_id::text,
			email,
			role,
			status,
			expires_at,
			accepted_at,
			created_at,
			updated_at
		FROM invitations
		WHERE token_digest = $1
	`

	found, err := scanInvitation(
		r.pool.QueryRow(
			ctx,
			query,
			digest,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invitation{}, ErrNotFound
	}

	if err != nil {
		return Invitation{}, fmt.Errorf(
			"find invitation: %w",
			err,
		)
	}

	return found, nil
}

func (r *PostgresRepository) Accept(
	ctx context.Context,
	invitationID string,
	userID string,
) (membership.Membership, error) {
	tx, err := r.pool.BeginTx(
		ctx,
		pgx.TxOptions{},
	)
	if err != nil {
		return membership.Membership{}, fmt.Errorf(
			"begin invitation acceptance: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const invitationQuery = `
		SELECT
			id::text,
			tenant_id::text,
			invited_by_user_id::text,
			email,
			role,
			status,
			expires_at,
			accepted_at,
			created_at,
			updated_at
		FROM invitations
		WHERE id = $1
		FOR UPDATE
	`

	found, err := scanInvitation(
		tx.QueryRow(
			ctx,
			invitationQuery,
			invitationID,
		),
	)
	if err != nil {
		return membership.Membership{},
			fmt.Errorf(
				"lock invitation: %w",
				err,
			)
	}

	const membershipQuery = `
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
			'ACTIVE',
			NOW()
		)
		ON CONFLICT (
			tenant_id,
			user_id
		)
		DO UPDATE SET
			role = EXCLUDED.role,
			status = 'ACTIVE',
			joined_at = COALESCE(
				memberships.joined_at,
				NOW()
			),
			updated_at = NOW()
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

	err = tx.QueryRow(
		ctx,
		membershipQuery,
		found.TenantID,
		userID,
		found.Role,
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
				"create invited membership: %w",
				err,
			)
	}

	const updateInvitation = `
		UPDATE invitations
		SET
			status = 'ACCEPTED',
			accepted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`

	if _, err := tx.Exec(
		ctx,
		updateInvitation,
		invitationID,
	); err != nil {
		return membership.Membership{},
			fmt.Errorf(
				"mark invitation accepted: %w",
				err,
			)
	}

	if err := tx.Commit(ctx); err != nil {
		return membership.Membership{},
			fmt.Errorf(
				"commit invitation acceptance: %w",
				err,
			)
	}

	return created, nil
}

func (r *PostgresRepository) ListByTenantID(
	ctx context.Context,
	tenantID string,
) ([]Invitation, error) {
	const query = `
		SELECT
			id::text,
			tenant_id::text,
			invited_by_user_id::text,
			email,
			role,
			status,
			expires_at,
			accepted_at,
			created_at,
			updated_at
		FROM invitations
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list invitations: %w",
			err,
		)
	}
	defer rows.Close()

	result := make([]Invitation, 0)

	for rows.Next() {
		item, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresRepository) Revoke(
	ctx context.Context,
	invitationID string,
) (Invitation, error) {
	const query = `
		UPDATE invitations
		SET
			status = 'REVOKED',
			updated_at = NOW()
		WHERE id = $1
		  AND status = 'PENDING'
		RETURNING
			id::text,
			tenant_id::text,
			invited_by_user_id::text,
			email,
			role,
			status,
			expires_at,
			accepted_at,
			created_at,
			updated_at
	`

	revoked, err := scanInvitation(
		r.pool.QueryRow(
			ctx,
			query,
			invitationID,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invitation{}, ErrNotFound
	}

	if err != nil {
		return Invitation{}, err
	}

	return revoked, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanInvitation(
	row rowScanner,
) (Invitation, error) {
	var result Invitation

	err := row.Scan(
		&result.ID,
		&result.TenantID,
		&result.InvitedByUserID,
		&result.Email,
		&result.Role,
		&result.Status,
		&result.ExpiresAt,
		&result.AcceptedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return Invitation{}, err
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
