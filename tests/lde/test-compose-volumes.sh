#!/usr/bin/env bash
# T012: Contract test for docker-compose.yml - Volume persistence pattern
# Tests: All data volumes mount to .local/ subdirectories
# Expected: MUST FAIL until docker-compose.yml has volume configuration

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

# Test 2: Postgres volume uses .local/pg-data
test_case "postgres uses .local/pg-data volume"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 20 "^  postgres:" | grep -q "\.local/pg-data"; then
        pass "postgres volume configured: .local/pg-data"
    else
        fail "postgres missing .local/pg-data volume"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 3: Redis volume uses .local/redis-data
test_case "redis uses .local/redis-data volume"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 20 "^  redis:" | grep -q "\.local/redis-data"; then
        pass "redis volume configured: .local/redis-data"
    else
        fail "redis missing .local/redis-data volume"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 4: MinIO volume uses .local/minio
test_case "minio uses .local/minio volume"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 20 "^  minio:" | grep -q "\.local/minio"; then
        pass "minio volume configured: .local/minio"
    else
        fail "minio missing .local/minio volume"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 5: Keycloak volume uses .local/keycloak
test_case "keycloak uses .local/keycloak volume"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 20 "^  keycloak:" | grep -q "\.local/keycloak"; then
        pass "keycloak data volume configured: .local/keycloak"
    else
        fail "keycloak missing .local/keycloak volume"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 6: Keycloak realm config volume
test_case "keycloak has realm config volume mount"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 30 "^  keycloak:" | grep -q "keycloak/realm-config\.json.*:/opt/keycloak"; then
        pass "keycloak realm config volume configured"
    else
        fail "keycloak missing realm config volume"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 7: At least 4 .local/ volume mounts
test_case "At least 4 .local/ volume mounts defined"
if [ -n "$CONFIG" ]; then
    LOCAL_MOUNT_COUNT=$(echo "$CONFIG" | grep -c "\.local/" || echo "0")

    if [ "$LOCAL_MOUNT_COUNT" -ge 4 ]; then
        pass "Found $LOCAL_MOUNT_COUNT .local/ volume mounts"
    else
        fail "Only found $LOCAL_MOUNT_COUNT .local/ volume mounts (expected at least 4)"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 8: Postgres volume points to correct container path
test_case "postgres volume mounts to /var/lib/postgresql/data"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 20 "^  postgres:" | grep "\.local/pg-data" | grep -q "/var/lib/postgresql/data"; then
        pass "postgres volume mounts to correct path"
    else
        fail "postgres volume doesn't mount to /var/lib/postgresql/data"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 9: Redis volume points to correct container path
test_case "redis volume mounts to /data"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 20 "^  redis:" | grep "\.local/redis-data" | grep -q ":/data"; then
        pass "redis volume mounts to correct path"
    else
        fail "redis volume doesn't mount to /data"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Test 10: MinIO volume points to correct container path
test_case "minio volume mounts to /data"
if [ -n "$CONFIG" ]; then
    if echo "$CONFIG" | grep -A 20 "^  minio:" | grep "\.local/minio" | grep -q ":/data"; then
        pass "minio volume mounts to correct path"
    else
        fail "minio volume doesn't mount to /data"
    fi
else
    fail "Cannot check: compose file invalid"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: docker-compose.yml Volumes"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until docker-compose.yml volumes are configured (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All volume tests passed!${NC}"
    exit 0
fi
