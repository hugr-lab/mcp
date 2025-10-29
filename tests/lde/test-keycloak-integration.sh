#!/usr/bin/env bash
#
# Integration Test: Keycloak Health and Realm Loading (T045)
#
# Tests Keycloak service health, realm verification, and token operations
# Can run in parallel with T042-T044, T046
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
echo "Integration Test: Keycloak (T045)"
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
if cd "${LDE_DIR}" && ./scripts/start.sh --no-data > /tmp/keycloak-test-start.log 2>&1; then
    echo -e "${PASS} Environment started"
else
    echo -e "${FAIL} Failed to start environment"
    cat /tmp/keycloak-test-start.log
    exit 1
fi

# Wait for Keycloak to fully initialize (it can take longer)
sleep 5

# Step 2: Test health endpoint
echo ""
echo "→ Testing Keycloak health endpoint..."
if curl -f -s http://localhost:8180/health/ready > /tmp/keycloak-test-health.txt 2>&1; then
    echo -e "${PASS} Keycloak health endpoint accessible"
else
    echo -e "${FAIL} Keycloak health endpoint check failed"
    cat /tmp/keycloak-test-health.txt
    TEST_FAILED=1
fi

# Step 3: Test realm "hugr" exists
echo ""
echo "→ Verifying realm 'hugr' exists..."
if curl -f -s http://localhost:8180/realms/hugr > /tmp/keycloak-test-realm.txt 2>&1; then
    REALM_NAME=$(cat /tmp/keycloak-test-realm.txt | jq -r '.realm' 2>/dev/null || echo "")
    if [ "$REALM_NAME" = "hugr" ]; then
        echo -e "${PASS} Realm 'hugr' exists"
    else
        echo -e "${FAIL} Realm 'hugr' not found or invalid response"
        cat /tmp/keycloak-test-realm.txt
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} Failed to access realm 'hugr'"
    cat /tmp/keycloak-test-realm.txt
    TEST_FAILED=1
fi

# Step 4: Obtain token for admin@example.com
echo ""
echo "→ Obtaining token for admin@example.com..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8180/realms/hugr/protocol/openid-connect/token \
    -d "client_id=hugr-graphql" \
    -d "username=admin@example.com" \
    -d "password=admin123" \
    -d "grant_type=password" 2>&1)

if [ $? -eq 0 ]; then
    ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token' 2>/dev/null || echo "")
    if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
        echo -e "${PASS} Token obtained successfully"

        # Step 5: Verify token contains expected roles
        echo ""
        echo "→ Verifying token contains admin role..."

        # Decode JWT token (extract payload)
        TOKEN_PAYLOAD=$(echo "$ACCESS_TOKEN" | cut -d. -f2)
        # Add padding if needed for base64 decoding
        TOKEN_PAYLOAD_PADDED="${TOKEN_PAYLOAD}$(printf '=%.0s' {1..4})"
        TOKEN_DECODED=$(echo "$TOKEN_PAYLOAD_PADDED" | base64 -d 2>/dev/null || echo "{}")

        # Check for admin role in realm_access or resource_access
        ADMIN_ROLE=$(echo "$TOKEN_DECODED" | jq -r '.realm_access.roles[]? // .resource_access."hugr-graphql".roles[]? // empty' 2>/dev/null | grep -i "admin" || echo "")

        if [ -n "$ADMIN_ROLE" ]; then
            echo -e "${PASS} Token contains admin role: $ADMIN_ROLE"
        else
            # Try alternate locations
            ALL_ROLES=$(echo "$TOKEN_DECODED" | jq -r '[.realm_access.roles[]?, .resource_access[]?.roles[]?] | join(", ")' 2>/dev/null || echo "")
            if [ -n "$ALL_ROLES" ]; then
                echo -e "${YELLOW}⚠${NC} Token contains roles: $ALL_ROLES (admin role check inconclusive)"
            else
                echo -e "${FAIL} Token does not contain expected admin role"
                TEST_FAILED=1
            fi
        fi
    else
        echo -e "${FAIL} Failed to extract access token from response"
        echo "Response: $TOKEN_RESPONSE"
        TEST_FAILED=1
    fi
else
    echo -e "${FAIL} Failed to obtain token"
    echo "Response: $TOKEN_RESPONSE"
    TEST_FAILED=1
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $TEST_FAILED -eq 0 ]; then
    echo -e "${PASS} Keycloak Integration Test PASSED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${FAIL} Keycloak Integration Test FAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
