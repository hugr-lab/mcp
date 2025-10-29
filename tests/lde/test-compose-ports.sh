#!/usr/bin/env bash
# T013: Contract test for docker-compose.yml - Port exposures
# Tests: Required ports exposed for all services
# Expected: MUST FAIL until docker-compose.yml has port configuration

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

# Test 2: Hugr port 8080 exposed
test_case "hugr exposes port 8080:8080"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  hugr:" | grep -q "8080:8080"; then
        pass "hugr port 8080:8080 configured"
    else
        fail "hugr port 8080:8080 not found"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 3: Keycloak port 8180 exposed
test_case "keycloak exposes port 8180:8080"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  keycloak:" | grep -q "8180:8080"; then
        pass "keycloak port 8180:8080 configured"
    else
        fail "keycloak port 8180:8080 not found"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 4: MinIO port 9000 exposed
test_case "minio exposes port 9000:9000"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  minio:" | grep -q "9000:9000"; then
        pass "minio port 9000:9000 configured"
    else
        fail "minio port 9000:9000 not found"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 5: MinIO console port 9001 exposed
test_case "minio exposes port 9001:9001"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  minio:" | grep -q "9001:9001"; then
        pass "minio console port 9001:9001 configured"
    else
        fail "minio console port 9001:9001 not found"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 6: Postgres port (optional but recommended)
test_case "postgres port exposure (optional)"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  postgres:" | grep -q "5432:5432"; then
        pass "postgres port 5432:5432 configured (optional)"
    else
        pass "postgres port not exposed (acceptable - internal only)"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 7: Redis port (optional but recommended)
test_case "redis port exposure (optional)"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  redis:" | grep -q "6379:6379"; then
        pass "redis port 6379:6379 configured (optional)"
    else
        pass "redis port not exposed (acceptable - internal only)"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 8: Essential ports all present
test_case "All essential ports present (8080, 8180, 9000, 9001)"
if [ -n "$CONFIG" ]; then
    MISSING_PORTS=""
    echo "$CONFIG" | grep -q "8080:8080" || MISSING_PORTS="$MISSING_PORTS 8080"
    echo "$CONFIG" | grep -q "8180:8080" || MISSING_PORTS="$MISSING_PORTS 8180"
    echo "$CONFIG" | grep -q "9000:9000" || MISSING_PORTS="$MISSING_PORTS 9000"
    echo "$CONFIG" | grep -q "9001:9001" || MISSING_PORTS="$MISSING_PORTS 9001"

    if [ -z "$MISSING_PORTS" ]; then
        pass "All essential ports configured"
    else
        fail "Missing essential ports:$MISSING_PORTS"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 9: No port conflicts
test_case "No duplicate host ports"
if [ -n "$CONFIG" ]; then
    # Extract all host ports and check for duplicates
    HOST_PORTS=$(echo "$CONFIG" | grep -oP '^\s+- "\K\d+(?=:)' | sort)
    DUPLICATE_PORTS=$(echo "$HOST_PORTS" | uniq -d)

    if [ -z "$DUPLICATE_PORTS" ]; then
        pass "No duplicate host ports"
    else
        fail "Duplicate host ports found: $DUPLICATE_PORTS"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: docker-compose.yml Ports"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until docker-compose.yml ports are configured (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All port tests passed!${NC}"
    exit 0
fi
