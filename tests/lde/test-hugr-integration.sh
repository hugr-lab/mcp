#!/usr/bin/env bash
#
# Integration Test: Hugr GraphQL Endpoint and Introspection (T046)
#
# Tests Hugr GraphQL service health, introspection, and SECRET_KEY authentication
# Can run in parallel with T042-T045
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
echo "Integration Test: Hugr GraphQL (T046)"
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
if cd "${LDE_DIR}" && ./scripts/start.sh --no-data > /tmp/hugr-test-start.log 2>&1; then
    echo -e "${PASS} Environment started"
else
    echo -e "${FAIL} Failed to start environment"
    cat /tmp/hugr-test-start.log
    exit 1
fi

# Wait a moment for services to fully initialize
sleep 3

# Step 2: Test health endpoint
echo ""
echo "→ Testing Hugr health endpoint..."
if curl -f -s http://localhost:8080/healthz > /tmp/hugr-test-health.txt 2>&1; then
    echo -e "${PASS} Hugr health endpoint accessible"
else
    echo -e "${FAIL} Hugr health endpoint check failed"
    cat /tmp/hugr-test-health.txt
    TEST_FAILED=1
fi

# Step 3: Test GraphQL introspection query
echo ""
echo "→ Testing GraphQL introspection query..."
INTROSPECTION_QUERY='{"query":"{ __schema { queryType { name } } }"}'

INTROSPECTION_RESPONSE=$(curl -s -X POST http://localhost:8080/graphql \
    -H "Content-Type: application/json" \
    -d "$INTROSPECTION_QUERY" 2>&1)

if [ $? -eq 0 ]; then
    QUERY_TYPE=$(echo "$INTROSPECTION_RESPONSE" | jq -r '.data.__schema.queryType.name' 2>/dev/null || echo "")
    if [ "$QUERY_TYPE" = "Query" ]; then
        echo -e "${PASS} GraphQL introspection successful"
    else
        echo -e "${FAIL} GraphQL introspection returned unexpected result"
        echo "Response: $INTROSPECTION_RESPONSE"
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} GraphQL introspection query failed"
    echo "Response: $INTROSPECTION_RESPONSE"
    TEST_FAILED=1
fi

# Step 4: Verify can query with SECRET_KEY header
echo ""
echo "→ Testing query with SECRET_KEY authentication..."

# Load SECRET_KEY from .env file
if [ -f "${LDE_DIR}/.env" ]; then
    source "${LDE_DIR}/.env"
else
    echo -e "${FAIL} .env file not found"
    TEST_FAILED=1
fi

if [ -n "${SECRET_KEY:-}" ]; then
    # Try a simple query with SECRET_KEY
    SECRET_QUERY='{"query":"{ __typename }"}'

    SECRET_RESPONSE=$(curl -s -X POST http://localhost:8080/graphql \
        -H "Content-Type: application/json" \
        -H "x-hugr-secret: ${SECRET_KEY}" \
        -d "$SECRET_QUERY" 2>&1)

    if [ $? -eq 0 ]; then
        TYPENAME=$(echo "$SECRET_RESPONSE" | jq -r '.data.__typename' 2>/dev/null || echo "")
        if [ "$TYPENAME" = "Query" ]; then
            echo -e "${PASS} Query with SECRET_KEY successful"
        else
            # Check if it's an error response
            ERROR_MSG=$(echo "$SECRET_RESPONSE" | jq -r '.errors[0].message' 2>/dev/null || echo "")
            if [ -n "$ERROR_MSG" ]; then
                echo -e "${FAIL} Query with SECRET_KEY failed: $ERROR_MSG"
                TEST_FAILED=1
            else
                echo -e "${FAIL} Query with SECRET_KEY returned unexpected result"
                echo "Response: $SECRET_RESPONSE"
                TEST_FAILED=1
            fi
        fi
    else
        echo -e "${FAIL} Query with SECRET_KEY failed"
        echo "Response: $SECRET_RESPONSE"
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} SECRET_KEY not set in .env file"
    TEST_FAILED=1
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $TEST_FAILED -eq 0 ]; then
    echo -e "${PASS} Hugr GraphQL Integration Test PASSED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${FAIL} Hugr GraphQL Integration Test FAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
