# Phase 02 – Authentication Infrastructure (Keycloak & PostgreSQL)

> Goal: Build the authentication infrastructure required by the CRM before writing any authentication code.

---

# Table of Contents

- Overview
- Objectives
- Authentication Architecture
- Why Keycloak?
- Why OpenID Connect?
- Why Authorization Code Flow?
- Why PostgreSQL?
- Infrastructure Components
- Keycloak Realm
- OIDC Client
- Development User
- Docker Networking
- OIDC Discovery
- Commands Used
- Verification
- Security Considerations
- Alternatives Considered
- Lessons Learned
- Completion Checklist
- Git Commit
- Next Phase

---

# Overview

Authentication should never be treated as "just another feature."

A CRM stores sensitive business data including:

- Customers
- Contacts
- Deals
- Notes
- Internal communications
- User permissions

Building a secure authentication system from scratch introduces significant risk.

Instead of implementing:

- Password hashing
- Login
- Password reset
- MFA
- OAuth
- Email verification

ourselves, we delegate authentication responsibilities to **Keycloak**.

The CRM backend will **trust Keycloak for identity**, while **retaining complete control over authorization**.

---

# Objectives

By the end of this phase we should have:

- PostgreSQL running
- Keycloak running
- CRM Realm
- OIDC Client
- Development User
- Working OIDC Discovery endpoint
- Backend configuration
- Docker networking

No authentication code exists yet.

Go still cannot log users in.

---

# Authentication Architecture

```
                   Browser
                       │
                       ▼
                Next.js Frontend
                       │
                       ▼
                  Go Backend
                       │
      OpenID Connect Authorization Flow
                       │
                       ▼
                  Keycloak
                       │
                 PostgreSQL
```

Notice that **Next.js never communicates directly with PostgreSQL**.

Likewise, **Keycloak never accesses CRM business tables**.

Each component has a clearly defined responsibility.

---

# Why Keycloak?

Keycloak provides a mature, open-source Identity and Access Management (IAM) platform.

It supports:

- OAuth 2.0
- OpenID Connect
- SAML
- MFA
- Social Login
- LDAP
- Active Directory
- Password Reset
- Email Verification
- Session Management

Instead of reinventing authentication, we rely on a battle-tested solution.

---

# Responsibilities

## Keycloak owns

- Identity
- Passwords
- MFA
- Sessions
- Login
- OAuth
- Tokens

## CRM owns

- Organisations
- Contacts
- Deals
- Roles
- Permissions
- CRM Users

Authentication and authorization are intentionally separated.

---

# Why OpenID Connect?

OpenID Connect (OIDC) is built on OAuth 2.0.

OAuth answers:

> "Can this application access a resource?"

OIDC answers:

> "Who is the user?"

Since a CRM must know **who** is logged in, OIDC is the appropriate protocol.

---

# Why Authorization Code Flow?

We selected:

```
Authorization Code Flow + PKCE
```

instead of:

- Implicit Flow
- Password Grant
- Direct Access Grant

The browser only receives an **authorization code**.

The backend exchanges that code for tokens securely.

Advantages:

- Tokens never exposed unnecessarily
- Client Secret remains on the backend
- Supports refresh tokens
- Industry standard

---

# Why Confidential Client?

The CRM backend is trusted.

It can securely store:

- Client Secret
- Refresh Tokens

Therefore the OIDC client is configured as a **Confidential Client**.

The browser is **not trusted**, so it never receives these credentials.

---

# Infrastructure Components

During this phase we started:

```
Go API

↓

PostgreSQL

↓

Keycloak

↓

Keycloak PostgreSQL
```

Everything runs inside Docker.

---

# PostgreSQL

PostgreSQL serves two different purposes.

## CRM Database

Stores:

- Users
- Organisations
- Contacts
- Deals
- Activities

## Keycloak Database

Stores:

- Users
- Credentials
- Sessions
- Clients
- Realm configuration

Keeping these databases separate avoids coupling authentication data with CRM business data.

---

# CRM Realm

Created:

```
crm
```

A realm is an isolated identity space.

Everything related to this CRM belongs inside this realm.

Examples:

- Users
- Clients
- Roles
- Groups

---

# OIDC Client

Created:

```
crm-backend
```

Configuration:

- Confidential Client
- Standard Flow Enabled
- Direct Access Disabled
- Implicit Disabled

Redirect URI:

```
http://localhost:8080/auth/callback
```

Post Logout Redirect:

```
http://localhost:3000/*
```

---

# Why Not Public Client?

Public clients cannot safely store secrets.

Our backend **can**.

Using a confidential client allows:

- Secure code exchange
- Refresh tokens
- Backend-only authentication

---

# Development User

Created:

```
developer@example.com
```

Purpose:

Provide a predictable user account during development.

This avoids creating a new account every time the environment is recreated.

---

# Docker Networking

Docker containers cannot access services through:

```
localhost
```

because:

```
localhost

↓

Current Container
```

Instead they communicate through Docker DNS.

Example:

```
keycloak:8080
```

This becomes important later when the Go backend communicates with Keycloak.

---

# OIDC Discovery

One of the most important achievements of this phase was enabling discovery.

Endpoint:

```
http://localhost:8081/realms/crm/.well-known/openid-configuration
```

Discovery returns:

- Authorization Endpoint
- Token Endpoint
- Logout Endpoint
- UserInfo Endpoint
- JWKS
- Supported Algorithms

This removes the need to hardcode protocol endpoints.

---

# Issuer

The issuer is:

```
http://localhost:8081/realms/crm
```

Every ID Token must contain:

```
iss = http://localhost:8081/realms/crm
```

The backend will verify this during authentication.

---

# Commands Used

Start containers

```bash
docker compose -f compose.dev.yaml up -d
```

Verify Keycloak

```bash
curl http://localhost:8081
```

Verify discovery

```bash
curl http://localhost:8081/realms/crm/.well-known/openid-configuration
```

---

# Verification

Successful verification returns JSON similar to:

```json
{
  "issuer": "http://localhost:8081/realms/crm",
  "authorization_endpoint": "...",
  "token_endpoint": "...",
  "jwks_uri": "..."
}
```

This confirms:

- Realm exists
- OIDC enabled
- Discovery working

---

# Security Considerations

## Never commit client secrets

Client Secret belongs only in:

```
.env
```

Never:

```
.env.example
```

---

## Never disable issuer validation

The backend must verify:

```
iss
```

matches the configured realm.

---

## Keep Direct Access disabled

The Password Grant is discouraged.

The browser should never submit credentials directly to the backend.

---

## Browser never receives Client Secret

Only the backend knows:

```
OIDC_CLIENT_SECRET
```

---

# Alternatives Considered

## Build authentication ourselves

Rejected.

Too much security responsibility.

---

## Auth0

Rejected.

Cloud-hosted.

Wanted self-hosted.

---

## Supabase Auth

Rejected.

Wanted dedicated identity platform.

---

## Public Client

Rejected.

Backend is trusted.

---

# Lessons Learned

Authentication infrastructure should be established before writing authentication code.

Separating identity from authorization keeps the architecture cleaner.

OIDC Discovery greatly simplifies backend configuration.

Docker networking requires understanding container DNS versus localhost.

---

# Completion Checklist

- [x] PostgreSQL running
- [x] Keycloak running
- [x] CRM Realm
- [x] OIDC Client
- [x] Development User
- [x] Discovery Endpoint
- [x] Docker networking configured
- [x] Backend configuration updated

---

# Git Commit

```
feat(auth): configure Keycloak realm and OIDC infrastructure
```

---

# Next Phase

The next phase builds the Go authentication foundation.

Specifically:

- Configuration package
- OIDC client
- Provider discovery
- Docker-aware transport
- OAuth2 configuration
- ID Token verifier
- Authentication health endpoints

At the end of the next phase, the Go backend will understand how to communicate with Keycloak, although users will still not be able to log in.