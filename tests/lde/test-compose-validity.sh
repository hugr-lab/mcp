#!/usr/bin/env bash
# T008: Contract test for docker-compose.yml YAML validity
# Tests: YAML syntax, docker compose config succeeds
# Expected: MUST FAIL until docker-compose.yml is created

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
    pass "File exists at $COMPOSE_FILE"
else
    fail "File does not exist at $COMPOSE_FILE"
fi

# Test 2: File is not empty
test_case "docker-compose.yml is not empty"
if [ -f "$COMPOSE_FILE" ]; then
    if [ -s "$COMPOSE_FILE" ]; then
        pass "File is not empty"
    else
        fail "File is empty"
    fi
else
    fail "File does not exist"
fi

# Test 3: File has valid YAML syntax (basic check)
test_case "docker-compose.yml has valid YAML structure"
if [ -f "$COMPOSE_FILE" ]; then
    # Try to parse with grep (basic structure check)
    if grep -q "^services:" "$COMPOSE_FILE"; then
        pass "File contains 'services:' section"
    else
        fail "File missing 'services:' section"
    fi
else
    fail "File does not exist"
fi

# Test 4: Docker compose config validates successfully
test_case "docker compose config validates the file"
if [ -f "$COMPOSE_FILE" ]; then
    cd "$PROJECT_ROOT/lde" || exit 1
    if docker compose config > /dev/null 2>&1; then
        pass "docker compose config succeeded"
    else
        ERROR_OUTPUT=$(docker compose config 2>&1 || true)
        fail "docker compose config failed: ${ERROR_OUTPUT:0:100}"
    fi
else
    fail "File does not exist, cannot validate"
fi

# Test 5: File specifies version (optional but recommended)
test_case "docker-compose.yml specifies version"
if [ -f "$COMPOSE_FILE" ]; then
    if grep -q "^version:" "$COMPOSE_FILE"; then
        VERSION=$(grep "^version:" "$COMPOSE_FILE" | head -1)
        pass "Version specified: $VERSION"
    else
        # Version is optional in newer Docker Compose
        pass "No version specified (optional in Docker Compose v2+)"
    fi
else
    fail "File does not exist"
fi

# Test 6: File uses proper indentation
test_case "docker-compose.yml uses consistent indentation"
if [ -f "$COMPOSE_FILE" ]; then
    # Check for tabs (should use spaces)
    if grep -P '\t' "$COMPOSE_FILE" > /dev/null 2>&1; then
        fail "File contains tabs (should use spaces)"
    else
        pass "File uses spaces for indentation"
    fi
else
    fail "File does not exist"
fi

# Test 7: File has reasonable size
test_case "docker-compose.yml has reasonable size"
if [ -f "$COMPOSE_FILE" ]; then
    SIZE=$(wc -c < "$COMPOSE_FILE")
    if [ "$SIZE" -gt 100 ] && [ "$SIZE" -lt 50000 ]; then
        pass "File size is reasonable ($SIZE bytes)"
    else
        fail "File size is unexpected ($SIZE bytes)"
    fi
else
    fail "File does not exist"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: docker-compose.yml Validity"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until docker-compose.yml is created (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All validity tests passed!${NC}"
    exit 0
fi
