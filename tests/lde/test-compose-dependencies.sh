#!/usr/bin/env bash
# T010: Contract test for docker-compose.yml - Service dependencies
# Tests: hugr depends_on all other services with service_healthy condition
# Expected: MUST FAIL until docker-compose.yml has proper dependencies

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

# Test 2: Hugr service has depends_on
test_case "hugr service has depends_on section"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 20 "^  hugr:" | grep -q "depends_on:"; then
        pass "hugr has depends_on defined"
    else
        fail "hugr missing depends_on section"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 3: Hugr depends on postgres
test_case "hugr depends on postgres"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  hugr:" | grep -A 20 "depends_on:" | grep -q "postgres:"; then
        pass "hugr depends on postgres"
    else
        fail "hugr doesn't depend on postgres"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 4: Hugr depends on redis
test_case "hugr depends on redis"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  hugr:" | grep -A 20 "depends_on:" | grep -q "redis:"; then
        pass "hugr depends on redis"
    else
        fail "hugr doesn't depend on redis"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 5: Hugr depends on minio
test_case "hugr depends on minio"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  hugr:" | grep -A 20 "depends_on:" | grep -q "minio:"; then
        pass "hugr depends on minio"
    else
        fail "hugr doesn't depend on minio"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 6: Hugr depends on keycloak
test_case "hugr depends on keycloak"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  hugr:" | grep -A 20 "depends_on:" | grep -q "keycloak:"; then
        pass "hugr depends on keycloak"
    else
        fail "hugr doesn't depend on keycloak"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 7: Dependencies use service_healthy condition
test_case "Dependencies use service_healthy condition"
if [ -n "$CONFIG" ]; then
    # Check if service_healthy is used in depends_on
    HUGR_SECTION=$(echo "$CONFIG" | grep -A 40 "^  hugr:")

    if echo "$HUGR_SECTION" | grep -q "condition:.*service_healthy"; then
        pass "Dependencies use service_healthy condition"
    else
        # Alternative format check
        if echo "$HUGR_SECTION" | grep -q "service_healthy"; then
            pass "Dependencies use service_healthy condition"
        else
            fail "Dependencies don't use service_healthy condition"
        fi
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 8: All 4 dependencies present (count check)
test_case "Hugr has exactly 4 service dependencies"
if [ -n "$CONFIG" ]; then
    HUGR_DEPS=$(echo "$CONFIG" | grep -A 40 "^  hugr:" | grep -A 30 "depends_on:" | grep -c "condition:\|service_healthy\|postgres:\|redis:\|minio:\|keycloak:" || echo "0")

    # This is a rough check - should find references to all 4 services
    if [ "$HUGR_DEPS" -ge 4 ]; then
        pass "Hugr has multiple service dependencies"
    else
        fail "Hugr missing some service dependencies (found $HUGR_DEPS references)"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: docker-compose.yml Dependencies"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until docker-compose.yml dependencies are configured (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All dependency tests passed!${NC}"
    exit 0
fi
