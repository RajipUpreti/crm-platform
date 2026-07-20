CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    identity_provider VARCHAR(50) NOT NULL,
    identity_provider_user_id VARCHAR(255) NOT NULL,

    email VARCHAR(320) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,

    first_name VARCHAR(100),
    last_name VARCHAR(100),
    display_name VARCHAR(255),

    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_provider_identity_unique
        UNIQUE (
            identity_provider,
            identity_provider_user_id
        ),

    CONSTRAINT users_status_check
        CHECK (
            status IN (
                'ACTIVE',
                'SUSPENDED',
                'DELETED'
            )
        )
);

CREATE INDEX users_email_lower_idx
    ON users (LOWER(email));

CREATE INDEX users_status_idx
    ON users (status);