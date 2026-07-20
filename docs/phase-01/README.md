# Phase 01 – Project Foundation

> Goal: Build a production-ready project foundation that is Docker-first, modular, microservice-ready, and easy to scale.

---

# Table of Contents

- Overview
- Objectives
- Architecture Vision
- Why This Architecture?
- Repository Structure
- Why Docker First?
- Project Structure
- Development Workflow
- Commands Used
- Verification
- Lessons Learned
- Alternatives Considered
- Future Improvements
- Completion Checklist
- Git Commit

---

# Overview

Every successful software project starts with a solid foundation.

Instead of immediately writing authentication code or creating database tables, the first phase focuses on creating a maintainable development environment that will continue to support the project as it grows.

The CRM we're building is expected to evolve over time.

Initially it will be a single Go application, but eventually it may become several independent services.

Because of that, the project structure should not assume that only one backend service will ever exist.

This phase establishes the repository layout, Docker-based development environment, documentation structure, and development workflow.

---

# Objectives

By the end of this phase we should have:

- A Git repository
- Docker-first local development
- Monorepo structure
- Go backend directory
- Next.js frontend directory
- Infrastructure directory
- Documentation directory
- Makefile
- Docker Compose
- Development containers

No business logic exists yet.

No authentication exists yet.

No database schema exists yet.

The focus is purely on building the project's foundation.

---

# Architecture Vision

```
                    Browser
                        │
                        ▼
                 Next.js Frontend
                        │
                        ▼
                  Go Backend API
                        │
            ┌───────────┴───────────┐
            ▼                       ▼
        PostgreSQL              Keycloak
```

Everything runs inside Docker.

No application runtime is installed directly on the host operating system.

---

# Why Docker First?

Many projects require developers to install:

- Go
- Node
- npm
- PostgreSQL
- Redis
- Keycloak
- Java
- Migration tools

Every developer ends up with slightly different versions.

This causes:

- "Works on my machine"
- Version conflicts
- Broken onboarding
- Difficult CI parity

Instead, Docker provides one reproducible environment.

Advantages:

- Same Go version everywhere
- Same Node version everywhere
- Same PostgreSQL version
- Same Keycloak version
- Same commands
- Easier onboarding
- Easier CI/CD

The only required software is Docker.

---

# Why a Monorepo?

The repository is organized as a monorepo.

```
crm-platform/

apps/

packages/

infrastructure/

docs/
```

Authentication changes often require changes in:

- Frontend
- Backend
- Infrastructure

Keeping everything together makes those changes atomic.

---

# Why Not Start With Microservices?

Microservices introduce operational complexity.

Examples include:

- Service discovery
- Distributed tracing
- Independent deployments
- Authentication between services
- Event delivery
- Retry logic

These concerns slow development during the early stages.

Instead we use a modular monolith.

```
apps/api

internal/

auth/

contact/

organisation/

deal/
```

Each module is isolated.

Later it can be extracted into its own service.

This approach provides most of the organizational benefits of microservices without the operational cost.

---

# Repository Structure

```
crm-platform/

apps/

api/

web/

packages/

contracts/

go-common/

infrastructure/

keycloak/

postgres/

proxy/

docker/

docs/

compose.dev.yaml

Makefile

README.md
```

---

# Folder Explanation

## apps/

Contains deployable applications.

Initially:

```
api

web
```

Eventually:

```
api-gateway

crm-service

notification-service

worker-service
```

---

## packages/

Contains reusable code.

Examples:

- API contracts
- Shared Go utilities

Business logic should not be shared here.

---

## infrastructure/

Contains everything required to run the platform.

Examples:

- Docker
- Keycloak
- PostgreSQL
- Reverse Proxy

---

## docs/

Contains project documentation.

This includes:

- Development phases
- Architecture
- ADRs
- Roadmap

---

# Why Feature-Based Packages?

Instead of:

```
handlers/

services/

repositories/
```

we prefer:

```
auth/

contact/

organisation/
```

Everything related to a feature stays together.

Advantages:

- Easier navigation
- Easier testing
- Easier extraction into microservices
- Better ownership

---

# Development Workflow

Every developer follows the same workflow.

Clone repository

↓

Run Docker

↓

Start containers

↓

Develop

↓

Commit

↓

Push

No manual software installation.

---

# Commands Used

Clone repository

```bash
git clone https://github.com/rajipupreti/crm-platform.git
cd crm-platform
```

Create structure

```bash
mkdir -p apps/api
mkdir -p apps/web
mkdir -p infrastructure
mkdir -p docs
```

Start development

```bash
docker compose -f compose.dev.yaml up -d
```

---

# Verification

Phase 1 is considered complete when:

- Repository exists
- Docker Compose starts
- Containers are healthy
- Go module exists
- Project structure matches design
- Documentation exists

No application functionality is expected yet.

---

# Lessons Learned

A good project foundation prevents expensive refactoring later.

Starting with Docker avoids environment drift.

Organizing by feature rather than technical layers improves maintainability.

Planning for future services does not mean building microservices immediately.

---

# Alternatives Considered

## Multiple repositories

Rejected because:

- Harder coordination
- More CI pipelines
- Harder onboarding

---

## Installing runtimes locally

Rejected because:

- Version conflicts
- Difficult upgrades
- Environment inconsistency

---

## Immediate microservices

Rejected because:

- Too much operational complexity
- Slower feature development
- Distributed debugging

---

# Future Improvements

Future phases will add:

- Keycloak
- PostgreSQL
- Authentication
- Authorization
- CRM modules
- Monitoring
- CI/CD

---

# Completion Checklist

- [x] Repository initialized
- [x] Docker-first environment
- [x] Monorepo structure
- [x] Go backend directory
- [x] Next.js directory
- [x] Infrastructure directory
- [x] Documentation directory
- [x] Docker Compose
- [x] Makefile

---

# Git Commit

```
chore: initialize docker-first CRM project structure
```

---

# Next Phase

The next phase introduces the authentication infrastructure.

Specifically:

- PostgreSQL
- Keycloak
- CRM Realm
- OIDC Client
- Development User
- OIDC Discovery

No Go authentication code is written until those components are functioning correctly.