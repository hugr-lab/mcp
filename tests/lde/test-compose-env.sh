#!/usr/bin/env bash
# T014: Contract test for docker-compose.yml - Environment variables
# Tests: All required environment variables for each service
# Expected: MUST FAIL until docker-compose.yml has environment configuration

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/lde/docker-compose.yml"

# Test result tracking
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo -e "${RED}✗${NC} $1"
}

test_case() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "\n${YELLOW}Test $TESTS_RUN:${NC} $1"
}

# Test 1: File exists
test_case "docker-compose.yml exists"
if [ -f "$COMPOSE_FILE" ]; then
    pass "File exists"
else
    fail "File does not exist at $COMPOSE_FILE"
fi

# Get full config
if [ -f "$COMPOSE_FILE" ]; then
    cd "$PROJECT_ROOT/lde" || exit 1
    CONFIG=$(docker compose config 2>&1 || echo "")
else
    CONFIG=""
fi

# Test 2: Hugr has DATABASE_URL
test_case "hugr has DATABASE_URL environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 50 "^  hugr:" | grep -q "DATABASE_URL"; then
        pass "DATABASE_URL configured for hugr"
    else
        fail "DATABASE_URL missing for hugr"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 3: Hugr has REDIS_URL
test_case "hugr has REDIS_URL environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 50 "^  hugr:" | grep -q "REDIS_URL"; then
        pass "REDIS_URL configured for hugr"
    else
        fail "REDIS_URL missing for hugr"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 4: Hugr has S3_ENDPOINT
test_case "hugr has S3_ENDPOINT environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 50 "^  hugr:" | grep -q "S3_ENDPOINT"; then
        pass "S3_ENDPOINT configured for hugr"
    else
        fail "S3_ENDPOINT missing for hugr"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 5: Hugr has KEYCLOAK_URL
test_case "hugr has KEYCLOAK_URL environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 50 "^  hugr:" | grep -q "KEYCLOAK_URL"; then
        pass "KEYCLOAK_URL configured for hugr"
    else
        fail "KEYCLOAK_URL missing for hugr"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 6: Hugr has KEYCLOAK_REALM
test_case "hugr has KEYCLOAK_REALM environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 50 "^  hugr:" | grep -q "KEYCLOAK_REALM"; then
        pass "KEYCLOAK_REALM configured for hugr"
    else
        fail "KEYCLOAK_REALM missing for hugr"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 7: Postgres has POSTGRES_USER
test_case "postgres has POSTGRES_USER environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  postgres:" | grep -q "POSTGRES_USER"; then
        pass "POSTGRES_USER configured for postgres"
    else
        fail "POSTGRES_USER missing for postgres"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 8: Postgres has POSTGRES_PASSWORD
test_case "postgres has POSTGRES_PASSWORD environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  postgres:" | grep -q "POSTGRES_PASSWORD"; then
        pass "POSTGRES_PASSWORD configured for postgres"
    else
        fail "POSTGRES_PASSWORD missing for postgres"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 9: Postgres has POSTGRES_DB
test_case "postgres has POSTGRES_DB environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  postgres:" | grep -q "POSTGRES_DB"; then
        pass "POSTGRES_DB configured for postgres"
    else
        fail "POSTGRES_DB missing for postgres"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 10: MinIO has MINIO_ROOT_USER
test_case "minio has MINIO_ROOT_USER environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  minio:" | grep -q "MINIO_ROOT_USER"; then
        pass "MINIO_ROOT_USER configured for minio"
    else
        fail "MINIO_ROOT_USER missing for minio"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 11: MinIO has MINIO_ROOT_PASSWORD
test_case "minio has MINIO_ROOT_PASSWORD environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  minio:" | grep -q "MINIO_ROOT_PASSWORD"; then
        pass "MINIO_ROOT_PASSWORD configured for minio"
    else
        fail "MINIO_ROOT_PASSWORD missing for minio"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 12: Keycloak has KEYCLOAK_ADMIN
test_case "keycloak has KEYCLOAK_ADMIN environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  keycloak:" | grep -q "KEYCLOAK_ADMIN"; then
        pass "KEYCLOAK_ADMIN configured for keycloak"
    else
        fail "KEYCLOAK_ADMIN missing for keycloak"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 13: Keycloak has KEYCLOAK_ADMIN_PASSWORD
test_case "keycloak has KEYCLOAK_ADMIN_PASSWORD environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  keycloak:" | grep -q "KEYCLOAK_ADMIN_PASSWORD"; then
        pass "KEYCLOAK_ADMIN_PASSWORD configured for keycloak"
    else
        fail "KEYCLOAK_ADMIN_PASSWORD missing for keycloak"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 14: Keycloak has KC_HTTP_PORT
test_case "keycloak has KC_HTTP_PORT environment variable"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  keycloak:" | grep -q "KC_HTTP_PORT"; then
        pass "KC_HTTP_PORT configured for keycloak"
    else
        fail "KC_HTTP_PORT missing for keycloak"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: docker-compose.yml Environment Variables"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until docker-compose.yml environment variables are configured (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All environment variable tests passed!${NC}"
    exit 0
fi
