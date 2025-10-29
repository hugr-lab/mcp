#!/usr/bin/env bash
# Run all Phase 3.2 contract tests (T004-T015)
# This script runs all contract tests and provides a summary

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================"
echo "Phase 3.2 Contract Tests (T004-T015)"
echo "TDD Approach: Tests MUST FAIL until implementation"
echo -e "========================================${NC}\n"

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test and track results
run_test() {
    local test_file=$1
    local test_name=$2
    local test_id=$3

    echo -e "\n${BLUE}Running $test_id: $test_name${NC}"
    echo "----------------------------------------"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if "$SCRIPT_DIR/$test_file" 2>&1; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ $test_id PASSED${NC}"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo -e "${RED}✗ $test_id FAILED (expected in TDD)${NC}"
    fi
}

# Script Interface Tests (T004-T007)
echo -e "\n${YELLOW}=== Script Interface Tests ===${NC}"
run_test "test-start-interface.sh" "start.sh interface" "T004"
run_test "test-stop-interface.sh" "stop.sh interface" "T005"
run_test "test-load-data-interface.sh" "load-data.sh interface" "T006"
run_test "test-health-check-interface.sh" "health-check.sh interface" "T007"

# Docker Compose Tests (T008-T015)
echo -e "\n${YELLOW}=== Docker Compose Contract Tests ===${NC}"
run_test "test-compose-validity.sh" "docker-compose.yml validity" "T008"
run_test "test-compose-services.sh" "Required services defined" "T009"
run_test "test-compose-dependencies.sh" "Service dependencies" "T010"
run_test "test-compose-healthchecks.sh" "Healthchecks configured" "T011"
run_test "test-compose-volumes.sh" "Volume persistence" "T012"
run_test "test-compose-ports.sh" "Port exposures" "T013"
run_test "test-compose-env.sh" "Environment variables" "T014"
run_test "test-compose-keycloak.sh" "Keycloak import config" "T015"

# Summary
echo -e "\n${BLUE}========================================"
echo "Contract Test Summary"
echo -e "========================================${NC}"
echo "Total test suites:  $TOTAL_TESTS"
echo -e "Passed suites:      ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed suites:      ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 12 ] && [ $PASSED_TESTS -eq 0 ]; then
    echo -e "\n${YELLOW}✓ All tests failing as expected (TDD red phase)${NC}"
    echo -e "${YELLOW}  Ready to proceed with implementation (Phase 3.3)${NC}"
    exit 0
elif [ $PASSED_TESTS -eq 12 ] && [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passing (TDD green phase)${NC}"
    echo -e "${GREEN}  Implementation complete!${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠ Partial implementation detected${NC}"
    echo -e "${YELLOW}  $PASSED_TESTS tests passing, $FAILED_TESTS tests failing${NC}"
    exit 1
fi
