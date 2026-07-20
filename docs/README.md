# CRM Platform Documentation

This directory contains the implementation history, architecture documentation, development guidance, and roadmap for the CRM platform.

The project is being developed incrementally. Each phase documents:

* What was implemented
* Why the approach was selected
* Security considerations
* Alternatives and trade-offs
* Commands used
* Verification steps
* Completion criteria
* Recommended Git commit message

---

# Documentation Structure

```text
docs/
├── README.md
├── roadmap.md
├── architecture/
│   ├── authentication.md
│   ├── authorization.md
│   ├── folder-structure.md
│   └── decisions/
│       └── README.md
├── phase-01/
│   ├── README.md
│   └── images/
├── phase-02/
│   ├── README.md
│   └── images/
├── phase-03a/
│   ├── README.md
│   └── images/
└── phase-03b/
    ├── README.md
    └── images/
```

---

# Project Overview

The CRM platform is a self-hosted, multi-tenant application built with:

| Component               | Technology                      |
| ----------------------- | ------------------------------- |
| Frontend                | Next.js with TypeScript         |
| Backend                 | Go                              |
| Identity provider       | Keycloak                        |
| Application database    | PostgreSQL                      |
| Keycloak database       | PostgreSQL                      |
| Local development       | Docker Compose                  |
| Authentication protocol | OpenID Connect                  |
| Login flow              | Authorization Code with PKCE    |
| Browser session         | Backend-managed HttpOnly cookie |
| Reverse proxy           | Caddy, planned                  |
| Deployment model        | Modular monolith initially      |

The development environment is container-first.

Developers should only need Docker installed locally. Go, Node.js, PostgreSQL, Keycloak, and development tooling run inside containers.

---

# Current Architecture

```text
┌──────────────────────────────┐
│ Browser                      │
│                              │
│ Receives opaque session      │
│ cookie after authentication  │
└───────────────┬──────────────┘
                │
                ▼
┌──────────────────────────────┐
│ Next.js frontend             │
│                              │
│ CRM user interface           │
└───────────────┬──────────────┘
                │
                ▼
┌──────────────────────────────┐
│ Go backend                   │
│                              │
│ OIDC client                  │
│ Session management           │
│ CRM authorization            │
│ Business logic               │
└───────┬───────────────┬──────┘
        │               │
        ▼               ▼
┌───────────────┐  ┌────────────────┐
│ Keycloak      │  │ CRM PostgreSQL │
│               │  │                │
│ Authentication│  │ Organisations  │
│ Passwords     │  │ Memberships    │
│ MFA           │  │ Contacts       │
│ OIDC tokens   │  │ Deals          │
└───────┬───────┘  └────────────────┘
        │
        ▼
┌───────────────────┐
│ Keycloak          │
│ PostgreSQL        │
└───────────────────┘
```

---

# Core Architectural Principles

## Authentication and authorization are separate

Keycloak answers:

```text
Who is the user?
Was authentication successful?
Is the identity token valid?
```

The CRM backend answers:

```text
Which organisation can the user access?
What role does the user have in that organisation?
Can the user view or modify this CRM record?
```

Organisation-specific roles will be stored in the CRM database rather than as global Keycloak realm roles.

A user may therefore have different roles in different organisations.

Example:

```text
Organisation A: OWNER
Organisation B: VIEWER
```

---

## The browser does not manage Keycloak tokens

The browser will not store:

* Access tokens
* Refresh tokens
* Client secrets
* ID tokens

After authentication, the Go backend will create an application session and send the browser an opaque HttpOnly cookie.

This reduces token exposure and keeps the Go backend as the primary application security boundary.

---

## Business records are tenant-scoped

Every organisation-owned CRM table will contain an organisation identifier.

Example:

```text
contacts.organisation_id
companies.organisation_id
deals.organisation_id
activities.organisation_id
```

Backend authorization and database queries must both enforce organisation isolation.

---

## Start as a modular monolith

The platform begins with one Go API application organized by business domain.

Example future modules:

```text
internal/
├── auth/
├── organisation/
├── contact/
├── company/
├── deal/
├── activity/
└── session/
```

This avoids premature microservice complexity while preserving clear boundaries for future extraction.

Possible future services include:

```text
api-gateway
organisation-service
crm-service
activity-service
notification-service
reporting-service
worker-service
```

---

# Implementation Phases

## Phase 01 — Project Foundation

Documentation:

```text
docs/phase-01/README.md
```

Phase 01 established:

* Docker-first development
* Monorepo structure
* Go backend directory
* Next.js frontend directory
* Infrastructure directories
* Shared contracts directory
* Documentation structure
* Makefile
* Docker Compose workflow
* Microservice-ready application layout

The main architectural decision was to begin with a modular monolith rather than immediately creating multiple distributed services.

Status:

```text
Complete
```

---

## Phase 02 — Authentication Infrastructure

Documentation:

```text
docs/phase-02/README.md
```

Phase 02 established:

* Keycloak container
* Keycloak PostgreSQL database
* CRM PostgreSQL database
* `crm` realm
* `crm-backend` confidential OIDC client
* Authorization Code flow
* PKCE support
* Local development user
* OIDC discovery endpoint
* Local issuer configuration

Keycloak handles identity and authentication.

The CRM database will handle tenant membership and business authorization.

Status:

```text
Complete
```

---

## Phase 03A — Go OIDC Foundation

Documentation:

```text
docs/phase-03a/README.md
```

Phase 03A established the Go backend’s connection to Keycloak.

Implemented:

* Centralized environment configuration
* Configuration validation
* OIDC provider discovery
* OAuth2 client configuration
* Reusable ID-token verifier
* Docker-aware OIDC transport
* API health endpoint
* Authentication health endpoint
* Graceful application startup and shutdown
* Automated phase verification script

Key endpoints:

```text
GET /health
GET /health/auth
```

Status:

```text
Complete
```

---

## Phase 03B — Secure Login Initiation

Documentation:

```text
docs/phase-03b/README.md
```

Phase 03B implemented the first half of the authentication flow.

Implemented:

* `GET /auth/login`
* Cryptographically secure state generation
* OpenID Connect nonce generation
* PKCE verifier generation
* S256 PKCE challenge
* Temporary login transaction storage
* One-time transaction consumption
* Transaction expiration
* Abandoned transaction cleanup
* Safe frontend return-path validation
* Redirect to the Keycloak authorization endpoint

A login request now redirects to Keycloak with unique values for:

```text
state
nonce
code_challenge
```

Status:

```text
Complete
```

---

# Current Authentication Flow

The application currently supports:

```text
Browser
   │
   │ GET /auth/login
   ▼
Go backend
   │
   ├── Generate state
   ├── Generate nonce
   ├── Generate PKCE verifier
   ├── Store login transaction
   └── Build authorization URL
   │
   │ HTTP 302
   ▼
Keycloak
   │
   ├── Display login screen
   └── Authenticate user
   │
   ▼
GET /auth/callback
```

The callback endpoint has not yet been implemented.

A successful Keycloak login may therefore reach `/auth/callback` and return a not-found response until the next phase is completed.

---

# Current Status

| Capability                 | Status      |
| -------------------------- | ----------- |
| Docker-first development   | Complete    |
| Monorepo structure         | Complete    |
| Go API container           | Complete    |
| Next.js container          | Complete    |
| CRM PostgreSQL             | Complete    |
| Keycloak PostgreSQL        | Complete    |
| Keycloak realm             | Complete    |
| Confidential OIDC client   | Complete    |
| OIDC discovery             | Complete    |
| Go OIDC client             | Complete    |
| ID-token verifier setup    | Complete    |
| OIDC health checks         | Complete    |
| Login initiation           | Complete    |
| State generation           | Complete    |
| Nonce generation           | Complete    |
| PKCE S256                  | Complete    |
| Login transaction storage  | Complete    |
| Authorization callback     | Not started |
| Code exchange              | Not started |
| ID-token claims validation | Not started |
| Application sessions       | Not started |
| HttpOnly session cookie    | Not started |
| Current-user endpoint      | Not started |
| Logout                     | Not started |
| CRM user synchronization   | Not started |
| Organisation onboarding    | Not started |
| Tenant authorization       | Not started |
| Contacts module            | Not started |

---

# Next Phase

The next implementation phase is:

```text
Phase 03C — OIDC Callback and Session Creation
```

It will add:

```text
GET /auth/callback
```

The callback will:

```text
1. Detect errors returned by Keycloak.
2. Require the authorization code.
3. Require the state parameter.
4. Consume the one-time login transaction.
5. Reject missing, expired, or reused state.
6. Exchange the authorization code.
7. Submit the PKCE verifier.
8. Extract the ID Token.
9. Verify token signature.
10. Verify issuer.
11. Verify audience.
12. Verify token expiration.
13. Validate the nonce.
14. Read the user identity claims.
15. Create a backend application session.
16. Set an HttpOnly session cookie.
17. Redirect to the stored frontend destination.
```

After Phase 03C, the project will have its first complete end-to-end login flow.

---

# Architecture Documentation

The following documents describe the current system design independently of the implementation timeline.

## Authentication

```text
docs/architecture/authentication.md
```

Should document:

* Keycloak responsibilities
* OIDC flow
* State
* Nonce
* PKCE
* Application sessions
* Token handling
* Login and logout sequence

## Authorization

```text
docs/architecture/authorization.md
```

Should document:

* Organisation membership
* CRM roles
* Permission checks
* Tenant isolation
* Database query requirements
* Platform-level roles

## Folder Structure

```text
docs/architecture/folder-structure.md
```

Should document:

* Deployable applications
* Shared contracts
* Go modules
* Infrastructure configuration
* Domain package conventions
* Future service extraction

## Architecture Decisions

```text
docs/architecture/decisions/
```

Architecture Decision Records should capture important choices such as:

```text
Use Keycloak
Use a confidential backend client
Use backend-managed sessions
Use a modular monolith initially
Store tenant roles in PostgreSQL
Use a Docker-first development environment
```

---

# Development Commands

Start the development environment:

```bash
make up
```

Or:

```bash
docker compose -f compose.dev.yaml up -d --build
```

Show service status:

```bash
make ps
```

Show logs:

```bash
make logs
```

Open a shell in the Go container:

```bash
make api-shell
```

Open a shell in the Next.js container:

```bash
make web-shell
```

Run Go formatting:

```bash
docker compose -f compose.dev.yaml exec api \
  gofmt -w .
```

Run Go static analysis:

```bash
docker compose -f compose.dev.yaml exec api \
  go vet ./...
```

Run Go tests:

```bash
docker compose -f compose.dev.yaml exec api \
  go test ./...
```

Verify the API:

```bash
curl http://localhost:8080/health
```

Verify OIDC initialization:

```bash
curl http://localhost:8080/health/auth
```

Test login initiation:

```bash
curl -sS -D - -o /dev/null \
  http://localhost:8080/auth/login
```

---

# Local URLs

| Service               | URL                                                                 |
| --------------------- | ------------------------------------------------------------------- |
| Next.js frontend      | `http://localhost:3000`                                             |
| Go API                | `http://localhost:8080`                                             |
| API health            | `http://localhost:8080/health`                                      |
| Authentication health | `http://localhost:8080/health/auth`                                 |
| Login initiation      | `http://localhost:8080/auth/login`                                  |
| Keycloak              | `http://localhost:8081`                                             |
| OIDC discovery        | `http://localhost:8081/realms/crm/.well-known/openid-configuration` |

---

# Documentation Guidelines

When completing a phase:

1. Update the phase README.
2. Update the status table in this document.
3. Update `docs/roadmap.md`.
4. Add or update architecture documentation when the system design changes.
5. Create an ADR for significant architectural decisions.
6. Record verification commands.
7. Record known limitations.
8. Add the phase commit message.

Phase documentation should describe the implementation as it existed when the phase was completed.

Architecture documentation should describe the current intended system design.

---

# Security Rules

The following rules apply throughout development:

* Never commit `.env`.
* Never commit OIDC client secrets.
* Never store tokens in browser local storage.
* Never disable issuer verification to work around networking.
* Never log access tokens, refresh tokens, ID tokens, authorization codes, nonces, PKCE verifiers, or client secrets.
* Use cryptographically secure randomness for authentication state.
* Use S256 for PKCE.
* Treat login transactions as short-lived and single use.
* Keep frontend return paths local.
* Require organisation membership checks in the backend.
* Filter every tenant-owned database query by organisation.
* Use exact redirect URIs.
* Use HTTPS and secure cookies in production.

---

# Repository

```text
https://github.com/rajipupreti/crm-platform
```

Go module:

```text
github.com/rajipupreti/crm-platform/apps/api
```
