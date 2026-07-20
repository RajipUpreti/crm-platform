# CRM Platform Roadmap

This roadmap tracks the planned development of the CRM platform from initial project setup through authentication, multi-tenancy, core CRM capabilities, observability, security, and production deployment.

The platform is being built incrementally using:

* Go for backend services
* Next.js for the frontend
* Keycloak for authentication
* PostgreSQL for application data
* Docker Compose for local development
* A modular monolith initially, with a path to future microservices

---

# Roadmap Principles

The project follows these principles:

1. Build secure foundations before business features.
2. Keep authentication separate from CRM authorization.
3. Start with a modular monolith.
4. Extract microservices only when there is a concrete need.
5. Keep development fully containerized.
6. Make every phase independently testable.
7. Document architecture and trade-offs as the system evolves.
8. Avoid premature infrastructure complexity.
9. Enforce tenant isolation in both application logic and database queries.
10. Treat observability, backups, and security as core platform features.

---

# Current Progress

| Phase     | Description                                           | Status   |
| --------- | ----------------------------------------------------- | -------- |
| Phase 01  | Docker-first project foundation                       | Complete |
| Phase 02  | Keycloak and PostgreSQL authentication infrastructure | Complete |
| Phase 03A | Go OIDC discovery and configuration                   | Complete |
| Phase 03B | Secure login initiation with state, nonce, and PKCE   | Complete |
| Phase 03C | OIDC callback and application session creation        | Next     |
| Phase 03D | Current-user endpoint, logout, and session middleware | Planned  |
| Phase 04  | CRM user persistence and database migrations          | Planned  |
| Phase 05  | Organisation onboarding and membership model          | Planned  |
| Phase 06  | Tenant-aware authorization                            | Planned  |
| Phase 07  | Next.js authentication integration                    | Planned  |
| Phase 08  | Contacts module                                       | Planned  |
| Phase 09  | Companies module                                      | Planned  |
| Phase 10  | Deals and pipelines                                   | Planned  |
| Phase 11  | Activities, tasks, and notes                          | Planned  |
| Phase 12  | Team invitations and role management                  | Planned  |
| Phase 13  | Audit logging and security controls                   | Planned  |
| Phase 14  | Background jobs and notifications                     | Planned  |
| Phase 15  | Observability and operational readiness               | Planned  |
| Phase 16  | CI/CD and production deployment                       | Planned  |
| Phase 17  | Scalability and microservice extraction review        | Planned  |

---

# Phase 01 — Docker-First Project Foundation

## Status

Complete

## Goal

Create a maintainable, container-first monorepo that can begin as a modular monolith and later support independently deployable services.

## Deliverables

* Monorepo structure
* `apps/api`
* `apps/web`
* Shared package directories
* Infrastructure directories
* Docker Compose development environment
* Go development container
* Next.js development container
* PostgreSQL containers
* Keycloak container
* Makefile
* Documentation structure
* Git repository initialization

## Completion criteria

* Docker is the only required local runtime.
* API and frontend can start in containers.
* Repository structure supports future services.
* Development commands are standardized.
* Initial documentation exists.

---

# Phase 02 — Authentication Infrastructure

## Status

Complete

## Goal

Create the identity infrastructure required before implementing application authentication code.

## Deliverables

* Keycloak database
* CRM database
* Keycloak `crm` realm
* Confidential OIDC client
* Authorization Code flow
* PKCE support
* Local development user
* OIDC discovery endpoint
* Local issuer configuration
* Client secret configuration

## Completion criteria

* Keycloak is reachable.
* The `crm` realm exists.
* Discovery returns the correct localhost issuer.
* The `crm-backend` client exists.
* Direct Access Grant is disabled.
* Implicit Flow is disabled.
* The development user can sign in.

---

# Phase 03A — Go OIDC Foundation

## Status

Complete

## Goal

Teach the Go API how to discover and communicate with Keycloak securely.

## Deliverables

* Centralized configuration package
* Configuration validation
* OIDC provider discovery
* OAuth2 client configuration
* Reusable ID-token verifier
* Docker-aware HTTP transport
* General health endpoint
* Authentication health endpoint
* Graceful startup and shutdown
* Phase verification script

## Completion criteria

* API starts only with valid authentication configuration.
* OIDC discovery succeeds during startup.
* Issuer validation remains enabled.
* API can reach Keycloak from Docker.
* `/health` returns success.
* `/health/auth` returns the configured issuer and client.

---

# Phase 03B — Secure Login Initiation

## Status

Complete

## Goal

Implement the first half of the OIDC Authorization Code flow.

## Deliverables

* `GET /auth/login`
* Cryptographically secure state
* OpenID Connect nonce
* PKCE verifier
* S256 PKCE challenge
* Temporary login transaction
* In-memory transaction store
* One-time transaction consumption
* Transaction expiration
* Expired transaction cleanup
* Safe frontend return-path validation
* Redirect to Keycloak

## Completion criteria

* Login endpoint returns HTTP 302.
* Redirect uses `response_type=code`.
* State changes for every login.
* Nonce changes for every login.
* PKCE challenge changes for every login.
* PKCE uses S256.
* Return paths are restricted to local frontend paths.
* Keycloak login page opens.
* Successful authentication reaches the callback URI.

---

# Phase 03C — OIDC Callback and Session Creation

## Status

Next

## Goal

Complete the OIDC callback and establish the first end-to-end login flow.

## Planned deliverables

* `GET /auth/callback`
* Provider error handling
* Authorization code validation
* State validation
* Atomic login transaction consumption
* PKCE verifier submission
* Authorization code exchange
* ID-token extraction
* Token signature verification
* Issuer validation
* Audience validation
* Expiration validation
* Nonce validation
* Identity claims parsing
* Backend application-session creation
* Opaque session ID
* HttpOnly cookie
* Safe frontend redirect

## Planned session fields

```text
session ID
identity provider subject
email
display name
access token
refresh token
created time
expiry time
```

Initial sessions may be stored in memory for development.

Before horizontal scaling, session storage should move to Redis or PostgreSQL.

## Completion criteria

* A browser can start login.
* A user can authenticate through Keycloak.
* The callback exchanges the authorization code.
* State is consumed once.
* PKCE is validated.
* ID Token is verified.
* Nonce is validated.
* A backend session is created.
* An HttpOnly cookie is set.
* The user is redirected to the frontend.

---

# Phase 03D — Session Middleware, Current User, and Logout

## Status

Planned

## Goal

Make backend sessions usable across protected endpoints.

## Planned deliverables

* Authentication middleware
* Session lookup
* Request-context user identity
* `GET /auth/me`
* `POST /auth/logout`
* Local session deletion
* Cookie deletion
* Keycloak end-session redirect
* Expired session handling
* Unauthorized response format

## Completion criteria

* Protected routes reject unauthenticated requests.
* Valid sessions attach identity to request context.
* `/auth/me` returns the current authenticated identity.
* Logout deletes local and Keycloak sessions.
* Expired sessions are rejected.

---

# Phase 04 — CRM User Persistence and Database Migrations

## Status

Planned

## Goal

Create the local CRM user model and persist authenticated identities in PostgreSQL.

## Planned deliverables

* Migration tooling
* `users` table
* External identity subject field
* Email and profile fields
* User status
* Created and updated timestamps
* Just-in-time user provisioning
* PostgreSQL repository
* User synchronization after login
* Database health checks

## Planned user model

```text
id
identity_provider
identity_provider_user_id
email
first_name
last_name
status
created_at
updated_at
```

## Completion criteria

* First login creates a local CRM user.
* Subsequent login updates profile fields.
* Keycloak subject is used as the stable identity reference.
* Email is not used as the primary external identity key.
* Database migrations are reproducible.

---

# Phase 05 — Organisation Onboarding and Membership Model

## Status

Planned

## Goal

Introduce multi-tenancy and organisation membership.

## Planned deliverables

* `organisations` table
* `organisation_members` table
* Organisation creation
* Owner membership creation
* Organisation slug
* Membership status
* Organisation onboarding flow
* Organisation list endpoint
* Organisation selector
* First-login onboarding redirect

## Planned roles

```text
OWNER
ADMIN
MANAGER
SALES_REP
VIEWER
```

## Completion criteria

* A user without membership is sent to onboarding.
* Organisation and owner membership are created transactionally.
* A user can belong to multiple organisations.
* Membership role is stored in the CRM database.
* Keycloak realm roles are not used for tenant-specific CRM access.

---

# Phase 06 — Tenant-Aware Authorization

## Status

Planned

## Goal

Enforce organisation access and role permissions across the API.

## Planned deliverables

* Organisation membership middleware
* Permission definitions
* Role-to-permission mapping
* Access-denied responses
* Tenant context in request context
* Tenant-filtered repositories
* Cross-tenant access tests
* Defence-in-depth query filters

## Planned permissions

```text
organisation:manage
members:invite
members:update
contacts:read
contacts:write
contacts:archive
companies:read
companies:write
deals:read
deals:write
activities:read
activities:write
```

## Completion criteria

* Every organisation route verifies active membership.
* Every repository query filters by organisation.
* Direct object access cannot bypass tenant checks.
* Users may hold different roles in different organisations.

---

# Phase 07 — Next.js Authentication Integration

## Status

Planned

## Goal

Connect the Next.js frontend to the Go authentication flow.

## Planned deliverables

* Login page
* Sign-in action
* Protected layout
* Server-side current-user lookup
* Cookie forwarding from Server Components
* Authenticated navigation
* Logout button
* Onboarding routing
* Organisation-aware URL structure
* Loading and error states

## Planned route structure

```text
/login
/dashboard
/onboarding/create-organisation
/app/{organisationSlug}/dashboard
/app/{organisationSlug}/contacts
/app/{organisationSlug}/deals
```

## Completion criteria

* Unauthenticated users are sent to login.
* Login redirects through the Go backend.
* Authenticated users can access protected pages.
* Server Components forward application cookies correctly.
* Frontend checks improve navigation but do not replace backend authorization.

---

# Phase 08 — Contacts Module

## Status

Planned

## Goal

Deliver the first complete CRM business module.

## Planned deliverables

* Contacts table
* Contact owner
* Contact status
* Create contact
* List contacts
* Read contact
* Update contact
* Archive contact
* Pagination
* Search
* Sorting
* Validation
* Tenant isolation
* Audit events

## Completion criteria

* Contacts are isolated by organisation.
* Detail, update, and archive operations filter by organisation.
* List endpoints support pagination.
* Unauthorized users cannot modify contacts.
* Archive is preferred over immediate hard deletion.

---

# Phase 09 — Companies Module

## Status

Planned

## Goal

Introduce company accounts and contact relationships.

## Planned deliverables

* Companies table
* Company profile
* Company owner
* Contact-to-company relationship
* Company activity feed
* Company search
* Company archive
* Tenant-safe API endpoints

## Completion criteria

* Contacts can be associated with companies.
* Companies are organisation-scoped.
* Archived companies remain auditable.
* Cross-tenant associations are prevented.

---

# Phase 10 — Deals and Pipelines

## Status

Planned

## Goal

Add sales pipelines and opportunity tracking.

## Planned deliverables

* Pipelines
* Pipeline stages
* Deals
* Deal owner
* Deal value
* Currency
* Expected close date
* Stage transitions
* Stage history
* Win/loss state
* Basic forecasting

## Completion criteria

* Deals belong to an organisation.
* Every deal belongs to a valid pipeline stage.
* Stage changes are recorded.
* Permission rules distinguish read and write access.
* Reporting queries remain tenant-isolated.

---

# Phase 11 — Activities, Tasks, and Notes

## Status

Planned

## Goal

Build the CRM activity timeline and follow-up workflow.

## Planned deliverables

* Notes
* Tasks
* Calls
* Meetings
* Reminders
* Due dates
* Activity ownership
* Contact/company/deal associations
* Activity timeline
* Completion state

## Completion criteria

* Activities can attach to CRM records.
* Access follows the associated organisation.
* Tasks support assignees and due dates.
* Activity history is ordered and auditable.

---

# Phase 12 — Team Invitations and Role Management

## Status

Planned

## Goal

Allow organisation administrators to manage members securely.

## Planned deliverables

* Invitation table
* Invitation tokens
* Email invitation workflow
* Accept invitation
* Expiration
* Resend invitation
* Cancel invitation
* Change role
* Suspend membership
* Remove membership
* Ownership-transfer rules

## Completion criteria

* Invitations are scoped to one organisation.
* Invitation tokens are random and short-lived.
* Only authorized roles can invite members.
* The last owner cannot be removed accidentally.
* Role changes are audited.

---

# Phase 13 — Audit Logging and Security Controls

## Status

Planned

## Goal

Add traceability and platform security controls.

## Planned deliverables

* Audit logs
* Login events
* Membership changes
* Role changes
* Contact and deal changes
* Request IDs
* CSRF protection
* Rate limiting
* Security headers
* Generic authentication errors
* Brute-force protections
* Sensitive-data log filtering

## Completion criteria

* Administrative actions are auditable.
* State-changing cookie-authenticated requests are CSRF-protected.
* Sensitive values never appear in logs.
* Login and invitation endpoints are rate-limited.
* Security headers are applied consistently.

---

# Phase 14 — Background Jobs and Notifications

## Status

Planned

## Goal

Introduce reliable asynchronous processing.

## Planned deliverables

* Worker process
* Job queue
* Email notifications
* Invitation emails
* Task reminders
* Retry policy
* Dead-letter handling
* Idempotency
* Delivery audit trail

## Completion criteria

* Email sending does not block API requests.
* Failed jobs retry safely.
* Duplicate job processing does not create duplicate side effects.
* Permanently failed jobs are visible for investigation.

---

# Phase 15 — Observability and Operational Readiness

## Status

Planned

## Goal

Make the platform measurable and supportable.

## Planned deliverables

* Structured logs
* Metrics
* Distributed trace readiness
* Health endpoints
* Readiness endpoints
* Database connection metrics
* Authentication metrics
* Request latency
* Error-rate dashboards
* Alerting
* Backup procedures
* Restore testing

## Completion criteria

* Operators can detect failed dependencies.
* API latency and error rates are measurable.
* Authentication failures are observable without leaking secrets.
* Database backups are automated.
* Restore procedures are tested.

---

# Phase 16 — CI/CD and Production Deployment

## Status

Planned

## Goal

Create repeatable, secure build and deployment workflows.

## Planned deliverables

* Go test workflow
* Frontend lint and test workflow
* Container builds
* Vulnerability scans
* Secret scanning
* Migration checks
* Versioned container tags
* Deployment environments
* Production Keycloak mode
* HTTPS
* Production session store
* Production secret management
* Rollback strategy

## Completion criteria

* Pull requests run automated checks.
* Container images are reproducible.
* Floating production image tags are avoided.
* Deployments use secure secrets.
* Production cookies use `Secure`.
* Keycloak no longer uses `start-dev`.
* Database migrations are controlled and observable.

---

# Phase 17 — Scalability and Microservice Extraction Review

## Status

Planned

## Goal

Determine whether any modular-monolith domains should become separate services.

## Extraction criteria

A module should be considered for extraction only when there is evidence of:

* Independent scaling requirements
* Independent deployment frequency
* Separate team ownership
* Stronger security isolation needs
* Different availability requirements
* Heavy asynchronous workloads
* Database contention
* Clear domain boundaries
* Operational benefit greater than distributed-system cost

## Possible future services

```text
api-gateway
organisation-service
crm-service
activity-service
notification-service
reporting-service
worker-service
```

## Completion criteria

* Service boundaries are based on business capabilities.
* Each extracted service owns its data.
* Service-to-service authentication is defined.
* Contracts are versioned.
* Observability exists before distribution.
* Extraction is justified by measurable requirements.

---

# Cross-Cutting Technical Roadmap

## Authentication

```text
Keycloak
→ Authorization Code
→ State
→ Nonce
→ PKCE
→ Callback
→ Backend session
→ HttpOnly cookie
→ Logout
→ MFA
```

## Authorization

```text
CRM user
→ Organisation membership
→ Role
→ Permission
→ Tenant-aware repository
```

## Data

```text
Migrations
→ Users
→ Organisations
→ Contacts
→ Companies
→ Deals
→ Activities
→ Audit logs
```

## Frontend

```text
Login
→ Protected layout
→ Onboarding
→ Organisation selector
→ CRM modules
→ Error and loading states
```

## Operations

```text
Docker
→ CI
→ Metrics
→ Logs
→ Backups
→ HTTPS
→ Deployment
→ Scaling
```

---

# Near-Term Priorities

The immediate implementation order is:

```text
1. Phase 03C — OIDC callback and application session
2. Phase 03D — Session middleware, current user, logout
3. Phase 04 — CRM user persistence
4. Phase 05 — Organisation onboarding
5. Phase 06 — Tenant authorization
6. Phase 07 — Next.js authentication integration
7. Phase 08 — Contacts module
```

Do not begin contacts, deals, or reporting before authentication, sessions, user persistence, organisation membership, and tenant authorization are working and tested.

---

# Definition of Done for Every Phase

A phase is complete only when:

* Implementation is finished.
* Code is formatted.
* Static analysis passes.
* Automated tests pass.
* Docker development still starts correctly.
* Verification commands succeed.
* Security implications are documented.
* Known limitations are recorded.
* Phase README is updated.
* Main documentation index is updated.
* Roadmap status is updated.
* Changes are committed with a scoped commit message.

---

# Roadmap Maintenance

When a phase is completed:

1. Change its status to `Complete`.
2. Mark the next phase as `Next`.
3. Update the current-progress table.
4. Add newly discovered follow-up work.
5. Do not remove postponed items; move them to the appropriate future phase.
6. Record significant architecture changes in an ADR.
7. Keep the roadmap aligned with the actual repository, not the intended repository.

---

# Long-Term Product Direction

The long-term CRM may support:

* Multiple organisations per user
* Contact and company management
* Deal pipelines
* Sales activities
* Team collaboration
* Invitations and delegated administration
* Reporting and dashboards
* Email and calendar integrations
* Audit history
* Import and export
* API access
* Webhooks
* Automation
* Custom fields
* Custom roles
* Enterprise SSO
* MFA enforcement
* Microservice extraction where justified

These capabilities should be added only after the platform foundations remain secure, testable, observable, and maintainable.
