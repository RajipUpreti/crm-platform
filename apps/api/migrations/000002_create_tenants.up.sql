CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(255) NOT NULL,
    slug VARCHAR(120) NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT tenants_name_not_blank
        CHECK (LENGTH(TRIM(name)) > 0),

    CONSTRAINT tenants_slug_not_blank
        CHECK (LENGTH(TRIM(slug)) > 0),

    CONSTRAINT tenants_slug_format
        CHECK (
            slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
        ),

    CONSTRAINT tenants_slug_unique
        UNIQUE (slug),

    CONSTRAINT tenants_status_check
        CHECK (
            status IN (
                'ACTIVE',
                'SUSPENDED',
                'DELETED'
            )
        )
);

CREATE INDEX tenants_status_idx
    ON tenants (status);

CREATE INDEX tenants_name_lower_idx
    ON tenants (LOWER(name));
