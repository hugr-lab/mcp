#!/usr/bin/env bash
# T009: Contract test for docker-compose.yml - All required services defined
# Tests: hugr, postgres, redis, minio, keycloak services present
# Expected: MUST FAIL until docker-compose.yml is created with all services

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

# Test 2: Can list services with docker compose
test_case "docker compose config --services works"
if [ -f "$COMPOSE_FILE" ]; then
    cd "$PROJECT_ROOT/lde" || exit 1
    if SERVICES=$(docker compose config --services 2>&1); then
        pass "docker compose config --services succeeded"
        echo "   Services found: $(echo "$SERVICES" | tr '\n' ' ')"
    else
        fail "docker compose config --services failed"
        SERVICES=""
    fi
else
    fail "File does not exist"
    SERVICES=""
fi

# Test 3: Hugr service defined
test_case "hugr service is defined"
if [ -n "$SERVICES" ]; then
    if echo "$SERVICES" | grep -q "^hugr$"; then
        pass "hugr service found"
    else
        fail "hugr service not found"
    fi
else
    fail "Cannot check services: compose file invalid or missing"
fi

# Test 4: PostgreSQL service defined
test_case "postgres service is defined"
if [ -n "$SERVICES" ]; then
    if echo "$SERVICES" | grep -q "^postgres$"; then
        pass "postgres service found"
    else
        fail "postgres service not found"
    fi
else
    fail "Cannot check services: compose file invalid or missing"
fi

# Test 5: Redis service defined
test_case "redis service is defined"
if [ -n "$SERVICES" ]; then
    if echo "$SERVICES" | grep -q "^redis$"; then
        pass "redis service found"
    else
        fail "redis service not found"
    fi
else
    fail "Cannot check services: compose file invalid or missing"
fi

# Test 6: MinIO service defined
test_case "minio service is defined"
if [ -n "$SERVICES" ]; then
    if echo "$SERVICES" | grep -q "^minio$"; then
        pass "minio service found"
    else
        fail "minio service not found"
    fi
else
    fail "Cannot check services: compose file invalid or missing"
fi

# Test 7: Keycloak service defined
test_case "keycloak service is defined"
if [ -n "$SERVICES" ]; then
    if echo "$SERVICES" | grep -q "^keycloak$"; then
        pass "keycloak service found"
    else
        fail "keycloak service not found"
    fi
else
    fail "Cannot check services: compose file invalid or missing"
fi

# Test 8: Exactly 5 services (no extras)
test_case "Exactly 5 services defined (no extras)"
if [ -n "$SERVICES" ]; then
    SERVICE_COUNT=$(echo "$SERVICES" | wc -l)
    if [ "$SERVICE_COUNT" -eq 5 ]; then
        pass "Exactly 5 services defined"
    else
        fail "Expected 5 services, found $SERVICE_COUNT"
    fi
else
    fail "Cannot count services: compose file invalid or missing"
fi

# Test 9: All services have image defined
test_case "All services have image specification"
if [ -f "$COMPOSE_FILE" ]; then
    cd "$PROJECT_ROOT/lde" || exit 1
    CONFIG=$(docker compose config 2>&1 || true)

    MISSING_IMAGES=""
    for service in hugr postgres redis minio keycloak; do
        if ! echo "$CONFIG" | grep -A 10 "^  $service:" | grep -q "image:"; then
            MISSING_IMAGES="$MISSING_IMAGES $service"
        fi
    done

    if [ -z "$MISSING_IMAGES" ]; then
        pass "All services have image specified"
    else
        fail "Services missing image:$MISSING_IMAGES"
    fi
else
    fail "File does not exist"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: docker-compose.yml Services"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until docker-compose.yml is created (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All service definition tests passed!${NC}"
    exit 0
fi
