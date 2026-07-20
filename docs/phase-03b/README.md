# Phase 03B – Secure OIDC Login Initiation

> Goal: Implement the first half of the OIDC login flow by generating secure login parameters, storing one-time authentication state, and redirecting the browser to Keycloak.

---

# Table of Contents

- Overview
- Objectives
- Authentication Flow
- What Was Implemented
- Login Transaction
- State
- Nonce
- PKCE
- Secure Random Values
- Login Transaction Store
- One-Time Consumption
- Return URL Validation
- Login Handler
- Keycloak Redirect
- Testing
- Security Considerations
- Alternatives Considered
- Common Pitfalls
- Lessons Learned
- Completion Checklist
- Git Commit
- Next Phase

---

# Overview

Phase 3A established communication between the Go backend and Keycloak.

The backend can now:

- Load OIDC configuration
- Discover Keycloak endpoints
- Build an OAuth2 client
- Build an ID Token verifier
- Confirm authentication infrastructure is healthy

Phase 3B builds the first user-facing authentication endpoint:

```text
GET /auth/login
```

This endpoint does not authenticate the user directly.

Instead, it creates a secure login transaction and redirects the browser to Keycloak.

Keycloak then displays the login page and authenticates the user.

---

# Objectives

By the end of this phase, the Go backend should:

- Expose `GET /auth/login`
- Generate a unique OAuth state value
- Generate a unique OpenID Connect nonce
- Generate a PKCE code verifier
- Generate an S256 PKCE challenge
- Store the login transaction temporarily
- Validate post-login return paths
- Redirect the browser to Keycloak
- Prevent login transaction replay
- Expire abandoned login attempts

The callback is not implemented in this phase.

After signing in through Keycloak, the browser may reach:

```text
http://localhost:8080/auth/callback
```

and receive a `404` until Phase 3C is completed.

---

# Authentication Flow

```text
Browser
   │
   │ GET /auth/login
   ▼
Go Backend
   │
   ├── Generate state
   ├── Generate nonce
   ├── Generate PKCE verifier
   ├── Derive S256 challenge
   ├── Validate return path
   └── Store login transaction
   │
   │ HTTP 302
   ▼
Keycloak Authorization Endpoint
   │
   ▼
Keycloak Login Page
```

The complete login flow will eventually become:

```text
Browser
   │
   ▼
GET /auth/login
   │
   ▼
Go generates secure transaction
   │
   ▼
Redirect to Keycloak
   │
   ▼
User authenticates
   │
   ▼
Keycloak redirects to /auth/callback
   │
   ▼
Go validates state, nonce, PKCE and token
```

Phase 3B implements only the first half.

---

# What Was Implemented

The following files were introduced:

```text
apps/api/internal/auth/
├── handler.go
├── login.go
├── random.go
├── transaction.go
└── transaction_store.go
```

Their responsibilities are:

| File | Responsibility |
|---|---|
| `handler.go` | Authentication handler construction and shared configuration |
| `login.go` | Implements `GET /auth/login` |
| `random.go` | Generates cryptographically secure random values |
| `transaction.go` | Defines temporary login transaction data |
| `transaction_store.go` | Saves and consumes one-time login transactions |

---

# Login Transaction

Each authentication attempt creates a temporary transaction.

Example structure:

```go
type LoginTransaction struct {
	State        string
	Nonce        string
	CodeVerifier string
	ReturnTo     string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}
```

The transaction connects the initial login request with the future callback.

Without it, the backend would not know:

- Whether it initiated the login
- Which nonce belongs to the ID Token
- Which PKCE verifier belongs to the authorization code
- Where the user should be redirected afterward
- Whether the callback is expired or replayed

---

# State

The OAuth `state` parameter protects the login flow against login CSRF and callback substitution.

The flow is:

```text
Go generates state
        │
        ▼
State sent to Keycloak
        │
        ▼
Keycloak returns the same state
        │
        ▼
Go validates and consumes state
```

Example authorization parameter:

```text
state=oGEA9vMRYcoPGmNUrJ6FwYdFPyBti3hdmuWLMMDVxy4
```

Every authentication attempt receives a different value.

A callback must be rejected when state is:

- Missing
- Unknown
- Expired
- Already consumed
- Different from the original transaction

---

# Why State Is Necessary

Without state validation, an attacker could potentially initiate an authentication request and trick another browser into completing an unrelated callback.

State ensures that the callback belongs to the login flow initiated by the application.

State is not a user ID.

It is an opaque, unpredictable correlation value.

Do not encode sensitive information directly into it.

---

# Nonce

The OpenID Connect `nonce` parameter binds the returned ID Token to the login request.

Example:

```text
nonce=cA4tJclsb5oN_AYrpuAVCg62mSasfU5G0W3ehXv4EFs
```

During Phase 3C, the callback will:

```text
1. Load the transaction using state.
2. Verify the ID Token.
3. Read the token's nonce claim.
4. Compare it to the stored nonce.
5. Reject any mismatch.
```

State protects the authorization callback.

Nonce protects the identity token.

They solve related but different problems and both should be used.

---

# PKCE

PKCE stands for:

```text
Proof Key for Code Exchange
```

It adds protection to the Authorization Code flow.

The backend generates:

```text
code_verifier
```

Then derives:

```text
code_challenge = BASE64URL(SHA256(code_verifier))
```

Only the challenge is sent to Keycloak.

The original verifier remains stored in the login transaction.

During Phase 3C, the backend submits the verifier when exchanging the authorization code.

Keycloak verifies that:

```text
SHA256(received verifier)
```

matches the original challenge.

---

# Why Use PKCE With a Confidential Client?

A confidential backend client already authenticates using a client secret.

PKCE still provides additional protection if an authorization code is intercepted.

Using both gives:

```text
Client authentication
+
Authorization-code binding
```

This is a stronger implementation than relying on the client secret alone.

---

# S256 Challenge Method

The authorization request uses:

```text
code_challenge_method=S256
```

Example:

```text
code_challenge=OP6BGt1C1zQsAugmoMHoMe-S3ElHEmIPbs2MbJcMCPM
```

Do not use:

```text
code_challenge_method=plain
```

The plain method sends a directly reusable verifier value and provides weaker protection.

S256 is the correct default.

---

# Secure Random Values

State and nonce are generated using:

```go
crypto/rand
```

They are encoded with:

```go
base64.RawURLEncoding
```

This produces values that are:

- Cryptographically unpredictable
- Safe for URL query parameters
- Compact
- Free from unnecessary padding

Example helper:

```go
func GenerateRandomValue(byteLength int) (string, error)
```

The implementation requires at least 16 random bytes.

The login flow uses 32 bytes for both state and nonce.

---

# Why Not Use `math/rand`?

`math/rand` is designed for simulations and ordinary pseudo-random behavior.

It is not suitable for security tokens.

Avoid generating authentication state with:

```go
math/rand
time.Now().UnixNano()
incrementing counters
```

These may be predictable.

Authentication values must come from a cryptographically secure source.

---

# Login Transaction Store

The application introduces this interface:

```go
type LoginTransactionStore interface {
	Save(
		ctx context.Context,
		transaction LoginTransaction,
	) error

	Consume(
		ctx context.Context,
		state string,
	) (LoginTransaction, error)

	DeleteExpired(
		ctx context.Context,
		now time.Time,
	) (int, error)
}
```

The current implementation is:

```text
MemoryLoginTransactionStore
```

It stores transactions in a synchronized in-memory map.

---

# Why Use a Store Interface?

The handler should depend on required behavior rather than storage technology.

The same interface can later be implemented by:

```text
Redis
PostgreSQL
Encrypted cookies
Distributed cache
```

The authentication handler does not need to change when storage changes.

This is a useful abstraction because the storage boundary is real and likely to evolve.

---

# One-Time Consumption

The store exposes:

```go
Consume(...)
```

rather than:

```go
Get(...)
```

Consume performs these steps atomically:

```text
Find transaction
        │
        ▼
Delete transaction
        │
        ▼
Validate expiration
        │
        ▼
Return transaction
```

This prevents the same state value from being reused.

---

# Why Not Get and Delete Separately?

This pattern is weaker:

```go
transaction := store.Get(state)
store.Delete(state)
```

Two concurrent callback requests could potentially read the same transaction before either deletes it.

Using one synchronized `Consume` operation reduces replay risk.

---

# Transaction Expiration

Each login transaction receives:

```text
CreatedAt
ExpiresAt
```

The default lifetime is:

```text
10 minutes
```

This is long enough for a user to complete authentication but short enough to limit the lifetime of abandoned login attempts.

Expired transactions are rejected.

---

# Expired Transaction Cleanup

Abandoned login attempts may never reach the callback.

Therefore, the in-memory store includes:

```go
DeleteExpired(...)
```

A background cleanup process periodically removes expired entries.

This prevents unbounded memory growth during development.

In production, Redis TTLs would usually handle expiration automatically.

---

# Limitations of the In-Memory Store

The current implementation is intentionally temporary.

It is suitable for:

- Local development
- One API process
- Initial testing
- Learning the flow

It is not suitable for:

- Multiple API replicas
- Rolling deployments
- High availability
- Long-running production systems

Problems include:

```text
API restart deletes transactions
Callback may reach another replica
State is not shared across processes
```

---

# Recommended Production Enhancement

Use Redis:

```text
Key:
oidc-login:{state}

Value:
nonce
code verifier
return path
timestamps

TTL:
10 minutes
```

The callback should use an atomic operation such as:

```text
GETDEL
```

to retrieve and remove the transaction.

PostgreSQL is also possible, but Redis is a better fit for short-lived authentication state.

---

# Return URL Validation

The login endpoint accepts an optional destination:

```text
/auth/login?return_to=/contacts
```

Valid examples:

```text
/dashboard
/contacts
/app/acme/deals
/onboarding
```

Invalid examples:

```text
https://evil.example
//evil.example
javascript:alert(1)
dashboard
```

Invalid destinations fall back to:

```text
/dashboard
```

---

# Why Validate Return URLs?

Without validation, the application could become an open redirector.

An attacker might send:

```text
http://localhost:8080/auth/login?return_to=https://fake.example
```

The user would authenticate through the real CRM and then be redirected to an attacker-controlled site.

Restricting redirects to local application paths prevents this.

---

# Local Path Rules

A safe return path must:

- Start with `/`
- Not start with `//`
- Not contain an absolute scheme
- Not contain a host
- Remain within the frontend application

The validation is implemented in:

```go
isSafeLocalPath(...)
```

---

# Login Handler

The login endpoint is:

```text
GET /auth/login
```

Its responsibilities are:

```text
1. Generate state.
2. Generate nonce.
3. Generate PKCE verifier.
4. Derive S256 challenge.
5. Resolve a safe return path.
6. Store the login transaction.
7. Build the Keycloak authorization URL.
8. Return a redirect.
```

The handler does not process passwords.

It never renders a login form.

Keycloak owns the user authentication interface.

---

# Authorization URL Generation

The backend uses:

```go
OAuth2Config.AuthCodeURL(...)
```

with:

```go
oidc.Nonce(nonce)
oauth2.S256ChallengeOption(codeVerifier)
```

This produces a redirect URL containing:

```text
client_id=crm-backend
response_type=code
redirect_uri=http://localhost:8080/auth/callback
scope=openid profile email
state=<random>
nonce=<random>
code_challenge=<derived>
code_challenge_method=S256
```

---

# Why Return HTTP 302?

The login endpoint returns:

```text
HTTP/1.1 302 Found
```

with a `Location` header pointing to Keycloak.

This causes a normal browser navigation.

No JavaScript is required.

Example:

```text
Location: http://localhost:8081/realms/crm/protocol/openid-connect/auth?...
```

A standard link can start authentication:

```html
<a href="http://localhost:8080/auth/login">
  Sign in
</a>
```

---

# Method-Aware Routing

The route is registered as:

```go
mux.HandleFunc(
	"GET /auth/login",
	authHandler.Login,
)
```

Modern Go `ServeMux` supports HTTP-method-aware route patterns.

This means only `GET` requests match the route.

---

# HEAD Request Behavior

This command:

```bash
curl -I http://localhost:8080/auth/login
```

sends a `HEAD` request.

Depending on route and handler checks, it may return:

```text
405 Method Not Allowed
Allow: GET
```

That does not mean the login route is broken.

Use a real GET request:

```bash
curl -sS -D - -o /dev/null \
  http://localhost:8080/auth/login
```

Expected:

```text
HTTP/1.1 302 Found
Location: http://localhost:8081/realms/crm/protocol/openid-connect/auth?...
```

---

# Verification Commands

## Verify the redirect

```bash
curl -sS -D - -o /dev/null \
  http://localhost:8080/auth/login
```

Expected:

```text
HTTP/1.1 302 Found
Location: http://localhost:8081/realms/crm/protocol/openid-connect/auth?...
```

---

## Display only the redirect URL

```bash
curl -sS -D - -o /dev/null \
  http://localhost:8080/auth/login |
  grep -i '^location:'
```

---

## Run the request twice

First request:

```bash
curl -sS -D - -o /dev/null \
  http://localhost:8080/auth/login |
  grep -i '^location:'
```

Second request:

```bash
curl -sS -D - -o /dev/null \
  http://localhost:8080/auth/login |
  grep -i '^location:'
```

The following values must differ:

```text
state
nonce
code_challenge
```

---

# Example Verification Output

First request:

```text
state=oGEA9vMRYcoPGmNUrJ6FwYdFPyBti3hdmuWLMMDVxy4
nonce=cA4tJclsb5oN_AYrpuAVCg62mSasfU5G0W3ehXv4EFs
code_challenge=OP6BGt1C1zQsAugmoMHoMe-S3ElHEmIPbs2MbJcMCPM
```

Second request:

```text
state=d2t024LZ4hdWqzVJWv2BjxQPqznUWX9WX9mCs0sQcfM
nonce=zQX2k1JKggn6edDzduBmcC6T5axd6eez9WDG6iwo3Xs
code_challenge=l1JmEzHCx3M3fuwINKNrFthqoKDlleK1StSj4CHx1EY
```

This confirms that each login attempt receives new security values.

---

# Browser Test

Open:

```text
http://localhost:8080/auth/login
```

Expected behavior:

```text
Go login endpoint
        │
        ▼
Keycloak authorization endpoint
        │
        ▼
Keycloak login screen
```

After entering the development user credentials, Keycloak should redirect to:

```text
http://localhost:8080/auth/callback
```

Because the callback is not yet implemented, a `404` is expected at the end of Phase 3B.

That proves that:

- Login initiation works
- Keycloak accepts the client configuration
- The redirect URI is valid
- The user can authenticate
- Keycloak returns an authorization code

---

# Tests Added

Tests should cover:

```text
Safe local paths
External URL rejection
Protocol-relative URL rejection
Transaction save and consume
One-time state use
Expired transaction rejection
Expired transaction cleanup
```

Example commands:

```bash
docker compose -f compose.dev.yaml exec api \
  gofmt -w .
```

```bash
docker compose -f compose.dev.yaml exec api \
  go vet ./...
```

```bash
docker compose -f compose.dev.yaml exec api \
  go test ./...
```

---

# Security Considerations

## State Must Be Unpredictable

State must come from `crypto/rand`.

Predictable state weakens callback correlation.

---

## Nonce Must Be Verified Later

Generating nonce is not sufficient.

Phase 3C must compare the ID Token nonce against the stored transaction nonce.

---

## PKCE Verifier Must Stay Private

The code verifier must not be:

- Included in the browser URL
- Logged
- Sent to Next.js
- Exposed through health endpoints

Only the derived challenge is sent to Keycloak.

---

## Login Transactions Must Be Single Use

A state value must be consumed and deleted during callback handling.

A second callback using the same state must fail.

---

## Avoid Logging Sensitive Values

Do not log:

```text
state
nonce
code verifier
authorization code
access token
refresh token
ID Token
client secret
```

Even though state and nonce are short-lived, they are part of an active authentication transaction.

---

## Use Generic Browser Errors

If transaction storage fails, return:

```text
could not start authentication
```

Log the internal error server-side.

Do not expose internal implementation details to the browser.

---

# Alternatives Considered

## Signed Cookie Login State

The login transaction could be stored in a signed and encrypted cookie.

Advantages:

- No server-side state
- Works across API replicas
- No Redis requirement

Disadvantages:

- More difficult key rotation
- Cookie size limitations
- Requires authenticated encryption
- More sensitive data reaches the browser
- Replay prevention still requires careful handling

For this project, server-side storage is easier to reason about.

---

## PostgreSQL Transaction Store

PostgreSQL could store login transactions.

Advantages:

- Durable
- Shared across instances
- Existing infrastructure

Disadvantages:

- More database writes
- Cleanup required
- Less natural for short-lived data
- Adds database dependency to every login

It remains a valid alternative when Redis is unavailable.

---

## No PKCE

A confidential client can perform Authorization Code flow without PKCE.

Rejected because PKCE adds useful defense against code interception at relatively low implementation cost.

---

## Frontend-Managed OIDC

Next.js could perform authentication directly as a public client.

Rejected for this architecture because:

- Tokens would be managed closer to the browser
- Session handling becomes more complex
- The Go backend is already the application security boundary
- A backend-managed session reduces token exposure

---

## Passing `return_to` Through OAuth State

Some systems encode the return path directly inside the OAuth state value.

Rejected because state should remain opaque and unpredictable.

The return path is stored server-side with the transaction instead.

---

# Common Pitfalls

## Using `curl -I`

`curl -I` sends `HEAD`, not `GET`.

Use:

```bash
curl -sS -D - -o /dev/null \
  http://localhost:8080/auth/login
```

---

## Reusing a PKCE Verifier

Generate a new verifier for every login attempt.

Never keep one global verifier.

---

## Reusing State

Every login attempt must have a unique state value.

---

## Allowing Arbitrary Return URLs

Do not trust:

```text
return_to=https://external.example
```

Restrict return paths to local frontend paths.

---

## Forgetting Transaction Expiration

Abandoned login attempts must not stay valid indefinitely.

---

## Using In-Memory State in Production

One API instance may start the login while another receives the callback.

Use shared storage before horizontal scaling.

---

## Logging the Authorization URL

The URL includes active state, nonce, and PKCE challenge values.

Avoid logging full authorization URLs in production.

---

# Lessons Learned

The login endpoint is not merely a redirect.

It creates a security transaction that must later be validated carefully.

State, nonce, and PKCE solve different problems:

| Mechanism | Primary Protection |
|---|---|
| State | Callback correlation and login CSRF |
| Nonce | ID Token replay and transaction binding |
| PKCE | Authorization-code interception |

A secure OIDC implementation should use all three.

Temporary authentication data should be:

- Random
- Short-lived
- Single use
- Stored securely
- Removed after use

---

# Completion Checklist

- [x] `GET /auth/login` exists
- [x] Login returns HTTP 302
- [x] Redirect points to the CRM Keycloak realm
- [x] Client ID is `crm-backend`
- [x] Response type is `code`
- [x] State is generated securely
- [x] State changes for every request
- [x] Nonce is generated securely
- [x] Nonce changes for every request
- [x] PKCE verifier is generated
- [x] PKCE challenge uses S256
- [x] Challenge changes for every request
- [x] Login transaction is stored
- [x] Transactions expire
- [x] Transactions are consumed only once
- [x] Unsafe return URLs are rejected
- [x] Keycloak login page opens
- [x] Keycloak redirects to `/auth/callback`
- [x] Go formatting passes
- [x] Go vet passes
- [x] Go tests pass

---

# Git Commit

```text
feat(auth): add secure OIDC login initiation
```

---

# Next Phase

Phase 3C will complete the callback side of authentication.

It will implement:

```text
GET /auth/callback
```

The callback will:

```text
1. Validate provider errors.
2. Require authorization code and state.
3. Atomically consume the login transaction.
4. Exchange the code using the PKCE verifier.
5. Extract the ID Token.
6. Verify signature, issuer, audience and expiration.
7. Validate nonce.
8. Read identity claims.
9. Create an application session.
10. Set an HttpOnly cookie.
11. Redirect to the stored frontend destination.
```

After Phase 3C, the first complete login flow will work end to end.