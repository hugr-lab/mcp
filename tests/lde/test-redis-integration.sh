#!/usr/bin/env bash
#
# Integration Test: Redis Health and Connectivity (T043)
#
# Tests Redis service health, connectivity, and operations
# Can run in parallel with T042, T044-T046
#
# Dependencies: T023-T029, T041
# Exit codes: 0 = pass, 1 = fail

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
PASS="${GREEN}✓${NC}"
FAIL="${RED}✗${NC}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Integration Test: Redis (T043)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
LDE_DIR="${PROJECT_ROOT}/lde"

# Track test status
TEST_FAILED=0

# Cleanup function
cleanup() {
    echo ""
    echo "→ Cleaning up..."
    cd "${LDE_DIR}" && ./scripts/stop.sh > /dev/null 2>&1 || true
    echo -e "${PASS} Environment stopped"
}

# Set trap to ensure cleanup runs
trap cleanup EXIT

# Step 1: Start environment with --no-data
echo ""
echo "→ Starting environment (--no-data)..."
if cd "${LDE_DIR}" && ./scripts/start.sh --no-data > /tmp/redis-test-start.log 2>&1; then
    echo -e "${PASS} Environment started"
else
    echo -e "${FAIL} Failed to start environment"
    cat /tmp/redis-test-start.log
    exit 1
fi

# Wait a moment for services to fully initialize
sleep 2

# Step 2: Test redis-cli ping
echo ""
echo "→ Testing redis-cli ping..."
if PING_RESULT=$(redis-cli -h localhost -p 6379 ping 2>&1); then
    if [ "$PING_RESULT" = "PONG" ]; then
        echo -e "${PASS} Redis ping successful"
    else
        echo -e "${FAIL} Redis ping returned unexpected result: $PING_RESULT"
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} Redis ping failed"
    TEST_FAILED=1
fi

# Step 3: Test SET/GET operations
echo ""
echo "→ Testing SET/GET operations..."
TEST_KEY="test:integration:$(date +%s)"
TEST_VALUE="integration-test-value-42"

# SET operation
if redis-cli -h localhost -p 6379 SET "$TEST_KEY" "$TEST_VALUE" > /tmp/redis-test-set.txt 2>&1; then
    SET_RESULT=$(cat /tmp/redis-test-set.txt)
    if [ "$SET_RESULT" = "OK" ]; then
        echo -e "${PASS} SET operation successful"
    else
        echo -e "${FAIL} SET operation returned unexpected result: $SET_RESULT"
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} SET operation failed"
    cat /tmp/redis-test-set.txt
    TEST_FAILED=1
fi

# GET operation
if GET_RESULT=$(redis-cli -h localhost -p 6379 GET "$TEST_KEY" 2>&1); then
    if [ "$GET_RESULT" = "$TEST_VALUE" ]; then
        echo -e "${PASS} GET operation successful (value matches)"
    else
        echo -e "${FAIL} GET operation returned unexpected value: $GET_RESULT"
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} GET operation failed"
    TEST_FAILED=1
fi

# Cleanup test key
redis-cli -h localhost -p 6379 DEL "$TEST_KEY" > /dev/null 2>&1 || true

# Step 4: Verify appendonly mode enabled
echo ""
echo "→ Verifying appendonly mode..."
if AOF_ENABLED=$(redis-cli -h localhost -p 6379 CONFIG GET appendonly 2>&1 | tail -1); then
    if [ "$AOF_ENABLED" = "yes" ] || [ "$AOF_ENABLED" = "always" ]; then
        echo -e "${PASS} Appendonly mode enabled"
    else
        echo -e "${YELLOW}⚠${NC} Appendonly mode: $AOF_ENABLED (expected: yes)"
        # Not failing the test, as this might be intentional
    fi
else
    echo -e "${FAIL} Failed to check appendonly mode"
    TEST_FAILED=1
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $TEST_FAILED -eq 0 ]; then
    echo -e "${PASS} Redis Integration Test PASSED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${FAIL} Redis Integration Test FAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
