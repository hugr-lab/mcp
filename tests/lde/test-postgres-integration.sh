#!/usr/bin/env bash
#
# Integration Test: PostgreSQL Health and Connectivity (T042)
#
# Tests PostgreSQL service health, connectivity, and database verification
# Can run in parallel with T043-T046
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
echo "Integration Test: PostgreSQL (T042)"
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
if cd "${LDE_DIR}" && ./scripts/start.sh --no-data > /tmp/postgres-test-start.log 2>&1; then
    echo -e "${PASS} Environment started"
else
    echo -e "${FAIL} Failed to start environment"
    cat /tmp/postgres-test-start.log
    exit 1
fi

# Wait a moment for services to fully initialize
sleep 2

# Step 2: Test pg_isready
echo ""
echo "→ Testing pg_isready..."
if pg_isready -h localhost -U hugr -d hugr > /dev/null 2>&1; then
    echo -e "${PASS} pg_isready check passed"
else
    echo -e "${FAIL} pg_isready check failed"
    TEST_FAILED=1
fi

# Step 3: Test connection with psql
echo ""
echo "→ Testing psql connection..."
if echo "SELECT 1 as test;" | PGPASSWORD=hugr psql -h localhost -U hugr -d hugr -t -A > /tmp/postgres-test-query.txt 2>&1; then
    RESULT=$(cat /tmp/postgres-test-query.txt | grep -v "^$" | head -1)
    if [ "$RESULT" = "1" ]; then
        echo -e "${PASS} psql connection successful"
    else
        echo -e "${FAIL} psql query returned unexpected result: $RESULT"
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} psql connection failed"
    cat /tmp/postgres-test-query.txt
    TEST_FAILED=1
fi

# Step 4: Verify database "hugr" exists
echo ""
echo "→ Verifying database 'hugr' exists..."
if echo "SELECT datname FROM pg_database WHERE datname = 'hugr';" | PGPASSWORD=hugr psql -h localhost -U hugr -d hugr -t -A > /tmp/postgres-test-db.txt 2>&1; then
    DB_NAME=$(cat /tmp/postgres-test-db.txt | grep -v "^$" | head -1)
    if [ "$DB_NAME" = "hugr" ]; then
        echo -e "${PASS} Database 'hugr' exists"
    else
        echo -e "${FAIL} Database 'hugr' not found"
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} Failed to query database list"
    cat /tmp/postgres-test-db.txt
    TEST_FAILED=1
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $TEST_FAILED -eq 0 ]; then
    echo -e "${PASS} PostgreSQL Integration Test PASSED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${FAIL} PostgreSQL Integration Test FAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
