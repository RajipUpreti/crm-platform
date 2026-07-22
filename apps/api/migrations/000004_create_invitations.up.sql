CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    tenant_id UUID NOT NULL,
    invited_by_user_id UUID NOT NULL,

    email VARCHAR(320) NOT NULL,
    role VARCHAR(30) NOT NULL,

    token_digest CHAR(64) NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',

    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT invitations_tenant_fk
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(id)
        ON DELETE CASCADE,

    CONSTRAINT invitations_invited_by_user_fk
        FOREIGN KEY (invited_by_user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT invitations_role_check
        CHECK (
            role IN (
                'ADMIN',
                'MEMBER'
            )
        ),

    CONSTRAINT invitations_status_check
        CHECK (
            status IN (
                'PENDING',
                'ACCEPTED',
                'REVOKED',
                'EXPIRED'
            )
        ),

    CONSTRAINT invitations_token_digest_unique
        UNIQUE (token_digest)
);

CREATE UNIQUE INDEX invitations_pending_tenant_email_unique
    ON invitations (
        tenant_id,
        LOWER(email)
    )
    WHERE status = 'PENDING';

CREATE INDEX invitations_tenant_idx
    ON invitations (tenant_id);

CREATE INDEX invitations_email_lower_idx
    ON invitations (LOWER(email));

CREATE INDEX invitations_expires_at_idx
    ON invitations (expires_at);
