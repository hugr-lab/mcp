#!/usr/bin/env bash
#
# Integration Test: Synthea Data Source Registration (T048)
#
# Tests that Synthea data source is registered in Hugr
# Continues from T047 (environment must be running)
#
# Dependencies: T047
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
echo "Integration Test: Synthea Registration (T048)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
LDE_DIR="${PROJECT_ROOT}/lde"

# Track test status
TEST_FAILED=0

# Cleanup function (DO NOT stop environment - T049 will continue)
cleanup() {
    if [ $TEST_FAILED -ne 0 ]; then
        echo ""
        echo "→ Test failed, stopping environment..."
        cd "${LDE_DIR}" && ./scripts/stop.sh > /dev/null 2>&1 || true
        echo -e "${PASS} Environment stopped"
    else
        echo ""
        echo -e "${YELLOW}⚠${NC} Environment left running for T049"
    fi
}

# Set trap to ensure cleanup runs
trap cleanup EXIT

# Step 1: Verify environment is running
echo ""
echo "→ Verifying environment is running..."
if curl -f -s http://localhost:8080/healthz > /dev/null 2>&1; then
    echo -e "${PASS} Hugr is running (continuing from T047)"
else
    echo -e "${FAIL} Hugr is not running. Run test-synthea-generation.sh first!"
    echo -e "${YELLOW}⚠${NC} Skipping cleanup (environment not started by this test)"
    exit 1
fi

# Step 2: Load SECRET_KEY
if [ -f "${LDE_DIR}/.env" ]; then
    source "${LDE_DIR}/.env"
else
    echo -e "${FAIL} .env file not found"
    echo -e "${YELLOW}⚠${NC} Skipping cleanup (environment not started by this test)"
    exit 1
fi

if [ -z "${SECRET_KEY:-}" ]; then
    echo -e "${FAIL} SECRET_KEY not set in .env"
    echo -e "${YELLOW}⚠${NC} Skipping cleanup (environment not started by this test)"
    exit 1
fi

# Step 3: Query Hugr for data sources
echo ""
echo "→ Querying Hugr for data sources..."
QUERY='{"query":"query { core { data_sources { name type as_module } } }"}'

RESPONSE=$(curl -s -X POST http://localhost:8080/graphql \
    -H "Content-Type: application/json" \
    -H "x-hugr-secret: ${SECRET_KEY}" \
    -d "$QUERY" 2>&1)

if [ $? -ne 0 ]; then
    echo -e "${FAIL} Failed to query Hugr"
    echo "Response: $RESPONSE"
    TEST_FAILED=1
    exit 1
fi

# Check for errors in response
ERROR_MSG=$(echo "$RESPONSE" | jq -r '.errors[0].message // empty' 2>/dev/null)
if [ -n "$ERROR_MSG" ]; then
    echo -e "${FAIL} GraphQL query error: $ERROR_MSG"
    echo "Full response: $RESPONSE"
    TEST_FAILED=1
    exit 1
fi

# Step 4: Verify "synthea" data source exists
echo ""
echo "→ Verifying 'synthea' data source exists..."
SYNTHEA_NAME=$(echo "$RESPONSE" | jq -r '.data.core.data_sources[] | select(.name == "synthea") | .name' 2>/dev/null)

if [ "$SYNTHEA_NAME" = "synthea" ]; then
    echo -e "${PASS} Data source 'synthea' found"
else
    echo -e "${FAIL} Data source 'synthea' not found"
    echo "Available data sources:"
    echo "$RESPONSE" | jq '.data.core.data_sources'
    TEST_FAILED=1
    exit 1
fi

# Step 5: Verify type is "duckdb"
echo ""
echo "→ Verifying type is 'duckdb'..."
SYNTHEA_TYPE=$(echo "$RESPONSE" | jq -r '.data.core.data_sources[] | select(.name == "synthea") | .type' 2>/dev/null)

if [ "$SYNTHEA_TYPE" = "duckdb" ]; then
    echo -e "${PASS} Type is 'duckdb'"
else
    echo -e "${FAIL} Type is '$SYNTHEA_TYPE', expected 'duckdb'"
    TEST_FAILED=1
fi

# Step 6: Verify as_module is true
echo ""
echo "→ Verifying as_module is true..."
SYNTHEA_AS_MODULE=$(echo "$RESPONSE" | jq -r '.data.core.data_sources[] | select(.name == "synthea") | .as_module' 2>/dev/null)

if [ "$SYNTHEA_AS_MODULE" = "true" ]; then
    echo -e "${PASS} as_module is true"
else
    echo -e "${FAIL} as_module is '$SYNTHEA_AS_MODULE', expected 'true'"
    TEST_FAILED=1
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $TEST_FAILED -eq 0 ]; then
    echo -e "${PASS} Synthea Registration Test PASSED"
    echo ""
    echo "   Next: Run test-synthea-verification.sh (T049)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${FAIL} Synthea Registration Test FAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
