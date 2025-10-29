#!/usr/bin/env bash
#
# Integration Test: Synthea Data Generation (T047)
#
# Tests Synthea data generation and DuckDB file creation
# First test in sequential data loading series (T047-T049)
#
# Dependencies: T031-T036, T041
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
echo "Integration Test: Synthea Generation (T047)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
LDE_DIR="${PROJECT_ROOT}/lde"
DATA_DIR="${LDE_DIR}/data"
SYNTHEA_DB="${DATA_DIR}/synthea.duckdb"

# Track test status
TEST_FAILED=0

# Cleanup function (DO NOT stop environment - T048 will continue)
cleanup() {
    if [ $TEST_FAILED -ne 0 ]; then
        echo ""
        echo "→ Test failed, stopping environment..."
        cd "${LDE_DIR}" && ./scripts/stop.sh > /dev/null 2>&1 || true
        echo -e "${PASS} Environment stopped"
    else
        echo ""
        echo -e "${YELLOW}⚠${NC} Environment left running for T048"
    fi
}

# Set trap to ensure cleanup runs
trap cleanup EXIT

# Step 1: Start environment with --no-data
echo ""
echo "→ Starting environment (--no-data)..."
if cd "${LDE_DIR}" && ./scripts/start.sh --no-data > /tmp/synthea-gen-start.log 2>&1; then
    echo -e "${PASS} Environment started"
else
    echo -e "${FAIL} Failed to start environment"
    cat /tmp/synthea-gen-start.log
    exit 1
fi

# Step 2: Run load-data.sh --synthea-only
echo ""
echo "→ Running load-data.sh --synthea-only..."
echo "   (This may take 3-5 minutes)"

START_TIME=$(date +%s)
if cd "${LDE_DIR}" && ./scripts/load-data.sh --synthea-only > /tmp/synthea-gen-load.log 2>&1; then
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    echo -e "${PASS} Synthea data generation completed (${DURATION}s)"
else
    echo -e "${FAIL} Synthea data generation failed"
    cat /tmp/synthea-gen-load.log
    TEST_FAILED=1
    exit 1
fi

# Step 3: Verify synthea.duckdb file created
echo ""
echo "→ Verifying synthea.duckdb file exists..."
if [ -f "$SYNTHEA_DB" ]; then
    echo -e "${PASS} File exists: $SYNTHEA_DB"
else
    echo -e "${FAIL} File not found: $SYNTHEA_DB"
    TEST_FAILED=1
    exit 1
fi

# Step 4: Verify file size reasonable (100-500MB)
echo ""
echo "→ Verifying file size..."
FILE_SIZE=$(stat -f%z "$SYNTHEA_DB" 2>/dev/null || stat -c%s "$SYNTHEA_DB" 2>/dev/null || echo "0")
FILE_SIZE_MB=$((FILE_SIZE / 1024 / 1024))

if [ $FILE_SIZE_MB -ge 100 ] && [ $FILE_SIZE_MB -le 500 ]; then
    echo -e "${PASS} File size reasonable: ${FILE_SIZE_MB}MB"
elif [ $FILE_SIZE_MB -gt 0 ]; then
    echo -e "${YELLOW}⚠${NC} File size: ${FILE_SIZE_MB}MB (expected 100-500MB, but may vary)"
else
    echo -e "${FAIL} Invalid file size: ${FILE_SIZE_MB}MB"
    TEST_FAILED=1
fi

# Step 5: Verify all 9 tables present in DuckDB
echo ""
echo "→ Verifying 9 tables present in DuckDB..."

# List tables using DuckDB CLI
TABLES=$(duckdb "$SYNTHEA_DB" "SHOW TABLES;" 2>/dev/null | tail -n +4 | head -n -1 | wc -l | tr -d ' ')

if [ "$TABLES" -eq 9 ]; then
    echo -e "${PASS} All 9 tables present"
    # Show table names
    echo "   Tables:"
    duckdb "$SYNTHEA_DB" "SHOW TABLES;" 2>/dev/null | tail -n +4 | head -n -1 | sed 's/^/   - /'
else
    echo -e "${FAIL} Expected 9 tables, found $TABLES"
    duckdb "$SYNTHEA_DB" "SHOW TABLES;" 2>/dev/null
    TEST_FAILED=1
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $TEST_FAILED -eq 0 ]; then
    echo -e "${PASS} Synthea Generation Test PASSED"
    echo ""
    echo "   Next: Run test-synthea-registration.sh (T048)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${FAIL} Synthea Generation Test FAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
