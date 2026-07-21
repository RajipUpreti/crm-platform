CREATE TABLE memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,

    role VARCHAR(30) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',

    joined_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT memberships_tenant_user_unique
        UNIQUE (
            tenant_id,
            user_id
        ),

    CONSTRAINT memberships_role_check
        CHECK (
            role IN (
                'OWNER',
                'ADMIN',
                'MEMBER'
            )
        ),

    CONSTRAINT memberships_status_check
        CHECK (
            status IN (
                'ACTIVE',
                'INVITED',
                'SUSPENDED',
                'LEFT'
            )
        ),

    CONSTRAINT memberships_active_joined_at_check
        CHECK (
            status <> 'ACTIVE'
            OR joined_at IS NOT NULL
        ),

    CONSTRAINT memberships_tenant_fk
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(id)
        ON DELETE CASCADE,

    CONSTRAINT memberships_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX memberships_user_idx
    ON memberships (user_id);

CREATE INDEX memberships_tenant_idx
    ON memberships (tenant_id);

CREATE INDEX memberships_tenant_status_idx
    ON memberships (
        tenant_id,
        status
    );

CREATE INDEX memberships_user_status_idx
    ON memberships (
        user_id,
        status
    );
