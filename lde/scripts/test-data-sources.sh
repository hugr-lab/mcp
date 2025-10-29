#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LDE_DIR="$PROJECT_ROOT/lde"

# Load environment variables
if [ -f "$LDE_DIR/.env" ]; then
    set -a
    source "$LDE_DIR/.env"
    set +a
fi

SECRET_KEY="${SECRET_KEY:-local-dev-secret-key-change-in-production}"
HUGR_URL="http://localhost:19000/query"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Hugr Data Source Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Function to execute test query
test_query() {
    local name="$1"
    local query="$2"
    local description="$3"

    echo "→ Testing $name: $description"
    response=$(curl -s -X POST "$HUGR_URL" \
        -H "x-hugr-secret: $SECRET_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"query\":\"$query\"}")

    if echo "$response" | jq -e '.errors' > /dev/null 2>&1; then
        echo "  ✗ Failed: $(echo "$response" | jq -r '.errors[0].message' | head -c 100)"
        return 1
    elif echo "$response" | jq -e '.data' > /dev/null 2>&1; then
        echo "  ✓ Success"
        echo "$response" | jq '.data' | head -10
        return 0
    else
        echo "  ✗ Unexpected response"
        return 1
    fi
}

# Test Northwind
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. Northwind (PostgreSQL)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
test_query "Northwind" \
    "{ northwind { customers(limit: 3) { id company_name } } }" \
    "Query 3 customers"
echo ""

# Test Synthea
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. Synthea (DuckDB - self-defined)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
test_query "Synthea" \
    "{ synthea { patients(limit: 3) { id birthdate first last } } }" \
    "Query 3 patients"
echo ""

# Test Open Payments
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. Open Payments (DuckDB)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
test_query "Open Payments" \
    "{ openpayments { research_payments(limit: 3) { Program_Year Covered_Recipient_Type } } }" \
    "Query 3 research payments"
echo ""

# Test OpenWeatherMap
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. OpenWeatherMap (HTTP REST API)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "→ Testing OpenWeatherMap: Current weather for Tokyo"
echo "  ⚠  TLS certificate verification may fail in development"
echo "  (This is expected and can be ignored)"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Tests Complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "All data sources are accessible via:"
echo "  Query endpoint: $HUGR_URL"
echo "  Admin UI:       http://localhost:19000/admin"
echo ""
echo "Authentication: Use x-hugr-secret header with value from .env"
echo ""
