#!/usr/bin/env bash

set -Eeuo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-compose.dev.yaml}"
API_SERVICE="${API_SERVICE:-api}"
KEYCLOAK_SERVICE="${KEYCLOAK_SERVICE:-keycloak}"

API_URL="${API_URL:-http://localhost:8080}"
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8081}"
REALM="${KEYCLOAK_REALM:-crm}"

EXPECTED_ISSUER="${OIDC_ISSUER_URL:-${KEYCLOAK_URL}/realms/${REALM}}"
EXPECTED_CLIENT_ID="${OIDC_CLIENT_ID:-crm-backend}"
EXPECTED_REDIRECT_URL="${OIDC_REDIRECT_URL:-http://localhost:8080/auth/callback}"

PASS_COUNT=0
FAIL_COUNT=0

info() {
  printf '\n[INFO] %s\n' "$1"
}

pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  printf '[PASS] %s\n' "$1"
}

fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  printf '[FAIL] %s\n' "$1" >&2
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

check_command() {
  local command_name="$1"

  if command_exists "${command_name}"; then
    pass "Required command exists: ${command_name}"
  else
    fail "Required command is missing: ${command_name}"
  fi
}

check_file() {
  local path="$1"

  if [[ -f "${path}" ]]; then
    pass "File exists: ${path}"
  else
    fail "Missing file: ${path}"
  fi
}

check_directory() {
  local path="$1"

  if [[ -d "${path}" ]]; then
    pass "Directory exists: ${path}"
  else
    fail "Missing directory: ${path}"
  fi
}

compose() {
  docker compose -f "${COMPOSE_FILE}" "$@"
}

service_is_running() {
  local service="$1"
  local container_id

  container_id="$(compose ps -q "${service}" 2>/dev/null || true)"

  if [[ -z "${container_id}" ]]; then
    return 1
  fi

  [[ "$(docker inspect -f '{{.State.Running}}' "${container_id}" 2>/dev/null)" == "true" ]]
}

check_http_status() {
  local name="$1"
  local url="$2"
  local expected_status="${3:-200}"
  local status

  status="$(
    curl \
      --silent \
      --show-error \
      --output /dev/null \
      --write-out '%{http_code}' \
      --max-time 10 \
      "${url}" || true
  )"

  if [[ "${status}" == "${expected_status}" ]]; then
    pass "${name} returned HTTP ${expected_status}"
  else
    fail "${name} returned HTTP ${status:-unavailable}; expected ${expected_status}"
  fi
}

json_field() {
  local json="$1"
  local field="$2"

  printf '%s' "${json}" |
    docker run --rm -i python:3.13-alpine \
      python -c "
import json
import sys

document = json.load(sys.stdin)
value = document.get('${field}')

if value is None:
    raise SystemExit(1)

if isinstance(value, bool):
    print(str(value).lower())
else:
    print(value)
"
}

check_json_field() {
  local name="$1"
  local json="$2"
  local field="$3"
  local expected="$4"
  local actual

  actual="$(json_field "${json}" "${field}" 2>/dev/null || true)"

  if [[ "${actual}" == "${expected}" ]]; then
    pass "${name}: ${field}=${expected}"
  else
    fail "${name}: expected ${field}=${expected}, got ${actual:-missing}"
  fi
}

info "Checking required local commands"

check_command docker
check_command curl

if ! docker compose version >/dev/null 2>&1; then
  fail "Docker Compose plugin is unavailable"
else
  pass "Docker Compose plugin is available"
fi

info "Checking Phase 3A project structure"

check_file "${COMPOSE_FILE}"
check_file "apps/api/go.mod"
check_file "apps/api/cmd/api/main.go"
check_file "apps/api/internal/config/config.go"
check_file "apps/api/internal/auth/oidc.go"
check_file "apps/api/internal/auth/transport.go"
check_file "apps/api/internal/server/server.go"
check_file "apps/api/Dockerfile.dev"

check_directory "apps/api/internal/auth"
check_directory "apps/api/internal/config"
check_directory "apps/api/internal/server"

info "Checking Docker services"

if service_is_running "${API_SERVICE}"; then
  pass "API service is running"
else
  fail "API service is not running"
fi

if service_is_running "${KEYCLOAK_SERVICE}"; then
  pass "Keycloak service is running"
else
  fail "Keycloak service is not running"
fi

info "Checking API container environment"

if service_is_running "${API_SERVICE}"; then
  API_ENV="$(
    compose exec -T "${API_SERVICE}" sh -c '
      printf "OIDC_ISSUER_URL=%s\n" "${OIDC_ISSUER_URL:-}"
      printf "OIDC_CLIENT_ID=%s\n" "${OIDC_CLIENT_ID:-}"
      printf "OIDC_REDIRECT_URL=%s\n" "${OIDC_REDIRECT_URL:-}"
      printf "OIDC_DOCKER_KEYCLOAK_ADDRESS=%s\n" "${OIDC_DOCKER_KEYCLOAK_ADDRESS:-}"
      if [ -n "${OIDC_CLIENT_SECRET:-}" ]; then
        printf "OIDC_CLIENT_SECRET_SET=true\n"
      else
        printf "OIDC_CLIENT_SECRET_SET=false\n"
      fi
    ' 2>/dev/null || true
  )"

  if grep -q "OIDC_ISSUER_URL=${EXPECTED_ISSUER}" <<<"${API_ENV}"; then
    pass "OIDC_ISSUER_URL is correct"
  else
    fail "OIDC_ISSUER_URL is missing or incorrect"
  fi

  if grep -q "OIDC_CLIENT_ID=${EXPECTED_CLIENT_ID}" <<<"${API_ENV}"; then
    pass "OIDC_CLIENT_ID is correct"
  else
    fail "OIDC_CLIENT_ID is missing or incorrect"
  fi

  if grep -q "OIDC_REDIRECT_URL=${EXPECTED_REDIRECT_URL}" <<<"${API_ENV}"; then
    pass "OIDC_REDIRECT_URL is correct"
  else
    fail "OIDC_REDIRECT_URL is missing or incorrect"
  fi

  if grep -q "OIDC_DOCKER_KEYCLOAK_ADDRESS=keycloak:8080" <<<"${API_ENV}"; then
    pass "Docker Keycloak address is correct"
  else
    fail "OIDC_DOCKER_KEYCLOAK_ADDRESS should be keycloak:8080"
  fi

  if grep -q "OIDC_CLIENT_SECRET_SET=true" <<<"${API_ENV}"; then
    pass "OIDC client secret is configured"
  else
    fail "OIDC client secret is missing"
  fi
fi

info "Checking Keycloak public OIDC discovery"

check_http_status \
  "Keycloak discovery endpoint" \
  "${KEYCLOAK_URL}/realms/${REALM}/.well-known/openid-configuration"

DISCOVERY_JSON="$(
  curl \
    --silent \
    --show-error \
    --max-time 10 \
    "${KEYCLOAK_URL}/realms/${REALM}/.well-known/openid-configuration" \
    2>/dev/null || true
)"

if [[ -n "${DISCOVERY_JSON}" ]]; then
  check_json_field \
    "Keycloak discovery" \
    "${DISCOVERY_JSON}" \
    "issuer" \
    "${EXPECTED_ISSUER}"
else
  fail "Could not read Keycloak discovery JSON"
fi

info "Checking Keycloak connectivity from the API container"

if service_is_running "${API_SERVICE}"; then
  if compose exec -T "${API_SERVICE}" \
    wget -qO- \
    "http://keycloak:8080/realms/${REALM}/.well-known/openid-configuration" \
    >/dev/null 2>&1; then
    pass "API container can reach Keycloak through Docker DNS"
  else
    fail "API container cannot reach http://keycloak:8080"
  fi
fi

info "Checking API endpoints"

check_http_status "API health endpoint" "${API_URL}/health"
check_http_status "OIDC health endpoint" "${API_URL}/health/auth"

HEALTH_JSON="$(
  curl \
    --silent \
    --show-error \
    --max-time 10 \
    "${API_URL}/health" \
    2>/dev/null || true
)"

if [[ -n "${HEALTH_JSON}" ]]; then
  check_json_field "API health" "${HEALTH_JSON}" "status" "ok"
fi

AUTH_HEALTH_JSON="$(
  curl \
    --silent \
    --show-error \
    --max-time 10 \
    "${API_URL}/health/auth" \
    2>/dev/null || true
)"

if [[ -n "${AUTH_HEALTH_JSON}" ]]; then
  check_json_field \
    "OIDC health" \
    "${AUTH_HEALTH_JSON}" \
    "status" \
    "ok"

  check_json_field \
    "OIDC health" \
    "${AUTH_HEALTH_JSON}" \
    "issuer" \
    "${EXPECTED_ISSUER}"

  check_json_field \
    "OIDC health" \
    "${AUTH_HEALTH_JSON}" \
    "clientId" \
    "${EXPECTED_CLIENT_ID}"

  check_json_field \
    "OIDC health" \
    "${AUTH_HEALTH_JSON}" \
    "redirectUrl" \
    "${EXPECTED_REDIRECT_URL}"
else
  fail "Could not read OIDC health JSON"
fi

info "Running Go checks inside Docker"

if service_is_running "${API_SERVICE}"; then
  if compose exec -T "${API_SERVICE}" \
    sh -c 'test -z "$(gofmt -l .)"'; then
    pass "Go source files are formatted"
  else
    fail "Some Go files need formatting; run gofmt -w ."
  fi

  if compose exec -T "${API_SERVICE}" go vet ./...; then
    pass "go vet completed successfully"
  else
    fail "go vet failed"
  fi

  if compose exec -T "${API_SERVICE}" go test ./...; then
    pass "Go tests completed successfully"
  else
    fail "Go tests failed"
  fi

  if compose exec -T "${API_SERVICE}" go mod tidy; then
    pass "go mod tidy completed successfully"
  else
    fail "go mod tidy failed"
  fi

  if compose exec -T "${API_SERVICE}" \
    git diff --exit-code -- go.mod go.sum >/dev/null 2>&1; then
    pass "go.mod and go.sum were already tidy"
  else
    fail "go mod tidy changed go.mod or go.sum; review and commit the changes"
  fi
fi

printf '\n========================================\n'
printf 'Phase 3A check complete\n'
printf 'Passed: %d\n' "${PASS_COUNT}"
printf 'Failed: %d\n' "${FAIL_COUNT}"
printf '========================================\n'

if ((FAIL_COUNT > 0)); then
  exit 1
fi

printf 'Phase 3A is complete.\n'
