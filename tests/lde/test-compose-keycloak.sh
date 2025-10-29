#!/usr/bin/env bash
# T015: Contract test for docker-compose.yml - Keycloak import configuration
# Tests: Keycloak command includes --import-realm, realm file volume mounted
# Expected: MUST FAIL until docker-compose.yml has Keycloak import config

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

# Test 2: Keycloak service has command section
test_case "keycloak has command section"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  keycloak:" | grep -q "command:"; then
        pass "keycloak has command section"
    else
        fail "keycloak missing command section"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 3: Keycloak command includes start-dev
test_case "keycloak command includes start-dev"
if [ -n "$CONFIG" ]; then
    KEYCLOAK_SECTION=$(echo "$CONFIG" | grep -A 40 "^  keycloak:")
    if echo "$KEYCLOAK_SECTION" | grep -A 5 "command:" | grep -q "start-dev\|start"; then
        pass "keycloak command includes start-dev"
    else
        fail "keycloak command doesn't include start-dev"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 4: Keycloak command includes --import-realm
test_case "keycloak command includes --import-realm flag"
if [ -n "$CONFIG" ]; then
    KEYCLOAK_SECTION=$(echo "$CONFIG" | grep -A 40 "^  keycloak:")
    if echo "$KEYCLOAK_SECTION" | grep -A 5 "command:" | grep -q -- "--import-realm"; then
        pass "keycloak command includes --import-realm"
    else
        fail "keycloak command missing --import-realm flag"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 5: Keycloak has realm config volume mount
test_case "keycloak has realm-config.json volume mount"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 40 "^  keycloak:" | grep -q "realm-config\.json"; then
        pass "keycloak has realm-config.json volume"
    else
        fail "keycloak missing realm-config.json volume"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 6: Realm config volume points to /opt/keycloak/data/import
test_case "realm config mounts to /opt/keycloak/data/import"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 40 "^  keycloak:" | grep "realm-config\.json" | grep -q "/opt/keycloak/data/import"; then
        pass "realm config mounts to correct path"
    else
        fail "realm config doesn't mount to /opt/keycloak/data/import"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 7: Realm config source path is relative to compose file
test_case "realm config source path is ./keycloak/ or keycloak/"
if [ -n "$CONFIG" ]; then
    VOLUME_LINE=$(echo "$CONFIG" | grep -A 40 "^  keycloak:" | grep "realm-config\.json" || echo "")
    if echo "$VOLUME_LINE" | grep -q "keycloak/realm-config\.json\|./keycloak/realm-config\.json"; then
        pass "realm config source path is correct"
    else
        fail "realm config source path incorrect (expected: keycloak/realm-config.json)"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 8: Keycloak has data persistence volume
test_case "keycloak has data persistence volume (.local/keycloak)"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 40 "^  keycloak:" | grep -q "\.local/keycloak"; then
        pass "keycloak has data persistence volume"
    else
        fail "keycloak missing .local/keycloak volume"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 9: Both volumes present (realm config + data)
test_case "keycloak has both realm config and data volumes"
if [ -n "$CONFIG" ]; then
    KEYCLOAK_VOLUMES=$(echo "$CONFIG" | grep -A 40 "^  keycloak:" | grep -c "volumes:\|realm-config\|\.local/keycloak" || echo "0")

    if [ "$KEYCLOAK_VOLUMES" -ge 3 ]; then
        pass "keycloak has multiple volumes configured"
    else
        fail "keycloak missing required volumes"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: docker-compose.yml Keycloak Import"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until docker-compose.yml Keycloak import is configured (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All Keycloak import tests passed!${NC}"
    exit 0
fi
