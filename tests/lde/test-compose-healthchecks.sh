#!/usr/bin/env bash
# T011: Contract test for docker-compose.yml - Healthchecks configured
# Tests: All 5 services have healthcheck with test, interval, timeout, retries
# Expected: MUST FAIL until docker-compose.yml has healthchecks

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

# Function to check healthcheck for a service
check_service_healthcheck() {
    local service=$1
    local expected_command=$2

    test_case "$service has healthcheck configured"

    if [ -z "$CONFIG" ]; then
        fail "Cannot check: compose file invalid"
        return
    fi

    # Extract service section
    SERVICE_SECTION=$(echo "$CONFIG" | grep -A 50 "^  $service:")

    # Check for healthcheck section
    if ! echo "$SERVICE_SECTION" | grep -q "healthcheck:"; then
        fail "$service missing healthcheck section"
        return
    fi

    # Check for test command
    if ! echo "$SERVICE_SECTION" | grep -A 10 "healthcheck:" | grep -q "test:"; then
        fail "$service healthcheck missing test command"
        return
    fi

    # Check for interval
    if ! echo "$SERVICE_SECTION" | grep -A 10 "healthcheck:" | grep -q "interval:"; then
        fail "$service healthcheck missing interval"
        return
    fi

    # Check for timeout
    if ! echo "$SERVICE_SECTION" | grep -A 10 "healthcheck:" | grep -q "timeout:"; then
        fail "$service healthcheck missing timeout"
        return
    fi

    # Check for retries
    if ! echo "$SERVICE_SECTION" | grep -A 10 "healthcheck:" | grep -q "retries:"; then
        fail "$service healthcheck missing retries"
        return
    fi

    # Check for expected command pattern
    if [ -n "$expected_command" ]; then
        if echo "$SERVICE_SECTION" | grep -A 10 "healthcheck:" | grep -q "$expected_command"; then
            pass "$service healthcheck fully configured with correct test command"
        else
            fail "$service healthcheck missing expected command pattern: $expected_command"
        fi
    else
        pass "$service healthcheck fully configured"
    fi
}

# Test 2-6: Check each service
check_service_healthcheck "postgres" "pg_isready"
check_service_healthcheck "redis" "redis-cli"
check_service_healthcheck "minio" "curl.*9000"
check_service_healthcheck "keycloak" "curl.*8080\|health"
check_service_healthcheck "hugr" "curl.*8080\|healthz"

# Test 7: All services have healthcheck
test_case "All 5 services have healthcheck defined"
if [ -n "$CONFIG" ]; then
    HEALTHCHECK_COUNT=$(echo "$CONFIG" | grep -c "healthcheck:" || echo "0")

    if [ "$HEALTHCHECK_COUNT" -ge 5 ]; then
        pass "All services have healthcheck (found $HEALTHCHECK_COUNT healthchecks)"
    else
        fail "Not all services have healthcheck (found $HEALTHCHECK_COUNT, expected 5)"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: docker-compose.yml Healthchecks"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until docker-compose.yml healthchecks are configured (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All healthcheck tests passed!${NC}"
    exit 0
fi
