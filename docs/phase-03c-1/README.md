# Phase 03C.1 -- Redis-Backed OIDC Login Transactions

> Goal: Replace the in-memory OIDC login transaction store with Redis to
> support reliable, scalable authentication.

## Overview

In Phase 03B, the Go backend generated secure login state (`state`,
`nonce`, and PKCE verifier) and stored it in memory before redirecting
the browser to Keycloak.

That implementation worked for local development but had important
limitations:

-   Restarting the API deleted active login transactions.
-   Multiple API instances could not share login state.
-   Authentication depended on a single process.
-   Cleanup had to be implemented manually.

This phase replaces the in-memory transaction store with Redis while
keeping the same `LoginTransactionStore` interface.

------------------------------------------------------------------------

# Objectives

By the end of this phase:

-   Redis runs as part of the Docker development environment.
-   The Go API connects to Redis during startup.
-   Login transactions are stored in Redis.
-   Every transaction has a TTL.
-   Transactions are consumed atomically.
-   API restarts do not invalidate login attempts.
-   The in-memory implementation remains available for unit tests.

------------------------------------------------------------------------

# Architecture Before

``` text
Browser
   │
   ▼
Go API
   │
   └── Memory
        ├── state
        ├── nonce
        └── PKCE verifier
```

Problems:

-   Lost on restart
-   Not shared across replicas
-   Manual cleanup

------------------------------------------------------------------------

# Architecture After

``` text
Browser
   │
   ▼
Go API
   │
   ▼
Redis
   ├── state
   ├── nonce
   ├── verifier
   └── return_to
```

Redis becomes the shared source of temporary authentication state.

------------------------------------------------------------------------

# Why Redis?

Redis is an excellent fit because login transactions are:

-   Temporary
-   Short-lived
-   Read once
-   Automatically expired
-   Shared between API instances

Redis also supports future features:

-   Application sessions
-   Rate limiting
-   Invitation tokens
-   Password reset tokens
-   Caching
-   Distributed locks

------------------------------------------------------------------------

# Docker Changes

A Redis service was added to `compose.dev.yaml`.

The API now depends on Redis:

``` yaml
depends_on:
  redis:
    condition: service_healthy
```

The Redis health check ensures the API does not start until Redis
responds successfully.

------------------------------------------------------------------------

# Redis Configuration

Environment variables:

``` dotenv
REDIS_ADDRESS=redis:6379
REDIS_PASSWORD=
REDIS_DATABASE=0
REDIS_KEY_PREFIX=crm:development
```

The key prefix isolates environments.

Example key:

``` text
crm:development:oidc:login:<state>
```

------------------------------------------------------------------------

# Redis Client

The backend now creates a Redis client during startup.

Startup sequence:

1.  Load configuration
2.  Connect to PostgreSQL
3.  Connect to Redis
4.  Ping Redis
5.  Discover Keycloak
6.  Start API

The application fails immediately if Redis cannot be reached.

------------------------------------------------------------------------

# Redis Login Transaction Store

The existing interface was preserved:

``` go
type LoginTransactionStore interface {
    Save(...)
    Consume(...)
    DeleteExpired(...)
}
```

Only the implementation changed.

This avoided changes to the login handler.

------------------------------------------------------------------------

# Why Preserve the Interface?

The handler depends only on the behaviour:

-   Save transaction
-   Consume transaction

It does not need to know whether storage is:

-   Memory
-   Redis
-   PostgreSQL

This keeps the authentication flow independent of storage technology.

------------------------------------------------------------------------

# SETNX

Transactions are written using `SETNX`.

Advantages:

-   Prevents accidental overwrite
-   Creates key only if it does not already exist
-   Stores TTL at creation

------------------------------------------------------------------------

# GETDEL

Transactions are consumed using Redis `GETDEL`.

Instead of:

    GET

    DELETE

the application performs:

    GETDEL

Benefits:

-   Atomic
-   Prevents replay
-   Prevents race conditions

------------------------------------------------------------------------

# TTL

Each login transaction receives a ten-minute TTL.

Redis removes expired transactions automatically.

The previous cleanup goroutine is no longer required.

------------------------------------------------------------------------

# Why Keep the Memory Store?

Although Redis is now the production implementation, the in-memory
implementation is retained for unit testing.

Advantages:

-   Faster tests
-   No external dependency
-   Simple mocking

------------------------------------------------------------------------

# Verification

Verify Redis:

``` bash
docker compose -f compose.dev.yaml exec redis redis-cli ping
```

Expected:

``` text
PONG
```

Verify login transaction:

``` bash
curl -sS -D - -o /dev/null \
http://localhost:8080/auth/login
```

List keys:

``` bash
docker compose -f compose.dev.yaml exec redis \
redis-cli KEYS 'crm:development:oidc:login:*'
```

Inspect TTL:

``` bash
redis-cli TTL <key>
```

The TTL should be approximately 600 seconds.

------------------------------------------------------------------------

# Security Considerations

-   Transactions are one-time use.
-   PKCE verifier never leaves the backend.
-   State is unpredictable.
-   Nonce is unpredictable.
-   Redis keys expire automatically.
-   Authorization codes are never stored permanently.

------------------------------------------------------------------------

# Lessons Learned

Authentication state should never depend on one process.

Redis provides a scalable solution that supports:

-   Restarts
-   Multiple replicas
-   Automatic expiration
-   Atomic operations

Adding Redis now avoids future refactoring of the authentication system.

------------------------------------------------------------------------

# Completion Checklist

-   [x] Redis container added
-   [x] Redis health checks
-   [x] Go Redis client
-   [x] Redis configuration
-   [x] Redis login transaction store
-   [x] Atomic GETDEL
-   [x] TTL support
-   [x] API startup validation
-   [x] Login survives API restart
-   [x] Memory store retained for tests

------------------------------------------------------------------------

# Git Commit

``` text
feat(auth): persist OIDC login transactions in Redis
```

------------------------------------------------------------------------

# Next Phase

The next phase implements the OIDC callback:

-   `/auth/callback`
-   Authorization code exchange
-   PKCE verification
-   ID Token validation
-   Nonce validation
-   User synchronization
-   Redis-backed application sessions
-   HttpOnly session cookie
