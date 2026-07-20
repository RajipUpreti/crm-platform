# Phase 03A – OIDC Foundation

> Goal: Build the Go authentication foundation and establish secure communication with Keycloak using OpenID Connect.

---

# Table of Contents

- Overview
- Objectives
- Authentication Flow
- Why This Phase Exists
- OIDC Architecture
- Configuration Package
- OIDC Provider Discovery
- Docker Networking
- Docker-aware Transport
- OAuth2 Configuration
- ID Token Verifier
- Health Endpoints
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

With the authentication infrastructure in place (Phase 2), the next step is teaching the Go backend **how to communicate with Keycloak**.

At this point the backend still cannot:

- Log users in
- Exchange authorization codes
- Validate ID Tokens
- Create sessions

Instead, this phase focuses on preparing all the building blocks that will make those operations possible.

Think of this phase as establishing communication between two systems before implementing any business logic.

---

# Objectives

By the end of this phase the backend should:

- Read configuration
- Validate configuration
- Discover the Keycloak OIDC endpoints
- Build an OAuth2 client
- Build an ID Token verifier
- Verify connectivity to Keycloak
- Expose authentication health endpoints

No browser login happens yet.

---

# Authentication Flow

```
Browser

↓

Go API

↓

Read Configuration

↓

Discover Keycloak

↓

Build OAuth2 Client

↓

Build ID Token Verifier

↓

Expose Authentication Services
```

No user interaction exists yet.

---

# Why This Phase Exists

Authentication involves several moving parts.

Before redirecting a browser to Keycloak we must know:

- Where to redirect
- Where to exchange tokens
- Which signing keys validate tokens
- Which algorithms are supported
- Which issuer to trust

All of that comes from OIDC Discovery.

---

# OIDC Architecture

```
                    Browser
                        │
                        ▼
                  Go Backend
                        │
                        ▼
      /.well-known/openid-configuration
                        │
                        ▼
                   Keycloak
```

Once discovery succeeds the backend automatically learns:

- Authorization Endpoint
- Token Endpoint
- UserInfo Endpoint
- Logout Endpoint
- JWKS Endpoint

instead of hardcoding them.

---

# Configuration Package

The first component built was:

```
internal/config
```

Instead of reading:

```go
os.Getenv(...)
```

throughout the application, configuration is loaded once.

```
Environment Variables

↓

Config Package

↓

Validated Config

↓

Application
```

---

# Why Centralize Configuration?

Without a configuration package:

- Environment variables become scattered
- Missing values fail later
- Testing becomes difficult
- Validation is inconsistent

Centralizing configuration provides one trusted source.

The application fails immediately if required configuration is missing.

Failing early is significantly easier to debug than discovering configuration errors during the first authentication attempt.

---

# OIDC Provider Discovery

Instead of hardcoding endpoints like:

```
/token

/auth

/logout
```

we ask Keycloak for its discovery document.

```
http://localhost:8081/realms/crm/.well-known/openid-configuration
```

The response contains every endpoint required by the backend.

Advantages:

- Future proof
- Standards compliant
- Less hardcoded configuration
- Easier upgrades

---

# Docker Networking Problem

Docker introduces an interesting issue.

The browser sees:

```
localhost:8081
```

The API container sees:

```
keycloak:8080
```

Those are different addresses.

However, OIDC requires the Issuer to remain exactly:

```
http://localhost:8081/realms/crm
```

Changing the issuer would invalidate token verification.

---

# Docker-aware HTTP Transport

Instead of changing the issuer, we introduced a custom HTTP transport.

Conceptually:

```
OIDC Request

↓

http://localhost:8081

↓

Transport rewrites network destination

↓

keycloak:8080

↓

Keycloak
```

The backend still believes it is talking to:

```
localhost:8081
```

while Docker routes traffic internally.

This preserves issuer validation.

---

# Why Not Disable Issuer Verification?

One possibility was:

```
SkipIssuerCheck = true
```

This was rejected.

Issuer validation protects against accepting tokens from an unexpected Identity Provider.

Disabling this check weakens authentication.

It is better to solve networking correctly.

---

# OAuth2 Configuration

Once discovery succeeds we build an OAuth2 configuration.

This configuration contains:

- Client ID
- Client Secret
- Redirect URI
- Authorization Endpoint
- Token Endpoint

This object will later generate login URLs.

---

# ID Token Verifier

An ID Token is a signed JWT.

Before trusting it we must verify:

- Signature
- Issuer
- Audience
- Expiration

The verifier created during this phase will later validate every login.

The verifier is created once during startup and reused.

This avoids unnecessary overhead.

---

# Why Create the Verifier Once?

Every verifier contains:

- Remote JWKS
- Configuration
- Signing algorithms

Creating a verifier on every request would:

- Increase latency
- Increase network traffic
- Increase object allocation

A singleton verifier is significantly more efficient.

---

# Health Endpoints

Two endpoints were added.

## General Health

```
GET /health
```

Purpose:

Verify that the API is running.

---

## Authentication Health

```
GET /health/auth
```

Purpose:

Verify that authentication initialized correctly.

Example:

```json
{
    "status":"ok",
    "issuer":"http://localhost:8081/realms/crm",
    "clientId":"crm-backend"
}
```

This endpoint is useful while debugging OIDC configuration.

---

# Commands Used

Install dependencies

```bash
go get github.com/coreos/go-oidc/v3/oidc

go get golang.org/x/oauth2
```

Clean modules

```bash
go mod tidy
```

Verify formatting

```bash
gofmt -w .
```

Run static analysis

```bash
go vet ./...
```

Run tests

```bash
go test ./...
```

Verify authentication health

```bash
curl http://localhost:8080/health/auth
```

---

# Verification

Authentication health should return:

```json
{
    "status":"ok"
}
```

OIDC Discovery should succeed.

No errors should appear during startup.

The backend should successfully communicate with Keycloak.

---

# Security Considerations

## Fail Fast

The application exits if:

- Client Secret missing
- Issuer missing
- Discovery fails

This prevents partially configured deployments.

---

## Never Disable Issuer Validation

The Issuer uniquely identifies the Identity Provider.

Accepting mismatched issuers could allow tokens from unexpected providers.

---

## No Tokens Yet

During this phase:

- No Access Tokens stored
- No Refresh Tokens stored
- No Sessions created

Only communication with Keycloak exists.

---

## Secrets Remain in Environment Variables

Client Secret is never committed.

Only:

```
.env
```

contains secrets.

---

# Alternatives Considered

## Hardcoded Endpoints

Rejected.

Future Keycloak changes could require code modifications.

Discovery already solves this problem.

---

## Multiple HTTP Clients

Rejected.

One shared OIDC client is easier to manage.

---

## Lazy Initialization

The backend could initialize OIDC during the first login.

Rejected.

Startup validation provides earlier feedback.

Authentication failures should occur during startup, not during user login.

---

# Lessons Learned

Authentication infrastructure should initialize during application startup.

Docker networking requires careful handling when using OIDC.

Provider Discovery removes almost all hardcoded endpoint configuration.

Configuration validation is an important security feature.

---

# Completion Checklist

- [x] Configuration package
- [x] OIDC discovery
- [x] OAuth2 configuration
- [x] ID Token verifier
- [x] Docker-aware transport
- [x] Authentication health endpoint
- [x] Startup validation
- [x] Successful Keycloak communication

---

# Git Commit

```
feat(auth): add OIDC configuration and Keycloak discovery
```

---

# Next Phase

The next phase begins the actual authentication flow.

Specifically:

- Login Handler
- State
- Nonce
- PKCE
- Login Transaction Store
- Redirect to Keycloak

At the end of the next phase, clicking **Login** will redirect the browser to Keycloak with a secure Authorization Code + PKCE request.

The callback and session creation will still be implemented in the following phase.