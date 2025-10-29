#!/usr/bin/env bash
# T006: Contract test for load-data.sh interface
# Tests: --skip-*, --force, --verbose, --help flags, exit codes
# Expected: MUST FAIL until load-data.sh is implemented

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOAD_DATA_SCRIPT="$PROJECT_ROOT/lde/scripts/load-data.sh"

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

# Test 1: Script exists and is executable
test_case "load-data.sh exists and is executable"
if [ -f "$LOAD_DATA_SCRIPT" ] && [ -x "$LOAD_DATA_SCRIPT" ]; then
    pass "Script exists and is executable"
else
    fail "Script does not exist or is not executable at $LOAD_DATA_SCRIPT"
fi

# Test 2: Script uses correct shebang
test_case "load-data.sh has correct shebang"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    SHEBANG=$(head -n1 "$LOAD_DATA_SCRIPT")
    if [ "$SHEBANG" = "#!/usr/bin/env bash" ]; then
        pass "Correct shebang: #!/usr/bin/env bash"
    else
        fail "Incorrect shebang: $SHEBANG (expected: #!/usr/bin/env bash)"
    fi
else
    fail "Script does not exist"
fi

# Test 3: Script has strict mode
test_case "load-data.sh uses strict mode"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -q "set -euo pipefail" "$LOAD_DATA_SCRIPT"; then
        pass "Strict mode enabled"
    else
        fail "Strict mode not found (expected: set -euo pipefail)"
    fi
else
    fail "Script does not exist"
fi

# Test 4: --help flag works
test_case "load-data.sh --help displays usage"
if [ -f "$LOAD_DATA_SCRIPT" ] && [ -x "$LOAD_DATA_SCRIPT" ]; then
    OUTPUT=$("$LOAD_DATA_SCRIPT" --help 2>&1 || true)
    EXIT_CODE=$?

    if [ $EXIT_CODE -eq 0 ]; then
        if echo "$OUTPUT" | grep -qi "usage"; then
            pass "--help flag works and displays usage (exit code: 0)"
        else
            fail "--help flag doesn't display 'Usage' in output"
        fi
    else
        fail "--help flag returned exit code $EXIT_CODE (expected: 0)"
    fi
else
    fail "Cannot test --help: script not executable"
fi

# Test 5: --help documents all flags
test_case "load-data.sh --help documents all flags"
if [ -f "$LOAD_DATA_SCRIPT" ] && [ -x "$LOAD_DATA_SCRIPT" ]; then
    OUTPUT=$("$LOAD_DATA_SCRIPT" --help 2>&1 || true)

    MISSING_FLAGS=""
    echo "$OUTPUT" | grep -q -- "--skip-northwind" || MISSING_FLAGS="$MISSING_FLAGS --skip-northwind"
    echo "$OUTPUT" | grep -q -- "--skip-synthea" || MISSING_FLAGS="$MISSING_FLAGS --skip-synthea"
    echo "$OUTPUT" | grep -q -- "--skip-openpayments" || MISSING_FLAGS="$MISSING_FLAGS --skip-openpayments"
    echo "$OUTPUT" | grep -q -- "--skip-owm" || MISSING_FLAGS="$MISSING_FLAGS --skip-owm"
    echo "$OUTPUT" | grep -q -- "--force" || MISSING_FLAGS="$MISSING_FLAGS --force"
    echo "$OUTPUT" | grep -q -- "--verbose\|^[[:space:]]*-v" || MISSING_FLAGS="$MISSING_FLAGS --verbose/-v"
    echo "$OUTPUT" | grep -q -- "--help" || MISSING_FLAGS="$MISSING_FLAGS --help"

    if [ -z "$MISSING_FLAGS" ]; then
        pass "All flags documented in --help"
    else
        fail "Missing flags in --help:$MISSING_FLAGS"
    fi
else
    fail "Cannot test --help output: script not executable"
fi

# Test 6: Exit codes contract (0-5)
test_case "load-data.sh follows exit code contract"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    # Check if exit codes 1-5 are used
    HAS_EXIT_CODES=0
    grep -q "exit 1" "$LOAD_DATA_SCRIPT" && HAS_EXIT_CODES=$((HAS_EXIT_CODES + 1))
    grep -q "exit 2" "$LOAD_DATA_SCRIPT" && HAS_EXIT_CODES=$((HAS_EXIT_CODES + 1))
    grep -q "exit 3" "$LOAD_DATA_SCRIPT" && HAS_EXIT_CODES=$((HAS_EXIT_CODES + 1))
    grep -q "exit 4" "$LOAD_DATA_SCRIPT" && HAS_EXIT_CODES=$((HAS_EXIT_CODES + 1))
    grep -q "exit 5" "$LOAD_DATA_SCRIPT" && HAS_EXIT_CODES=$((HAS_EXIT_CODES + 1))

    if [ $HAS_EXIT_CODES -ge 3 ]; then
        pass "Script uses multiple exit codes (1-5 per contract)"
    else
        fail "Script doesn't implement expected exit codes (1=prereq, 2=health, 3=generation, 4=registration, 5=verification)"
    fi
else
    fail "Script does not exist"
fi

# Test 7: Script checks for Hugr health
test_case "load-data.sh checks Hugr health"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -q "health-check\|hugr.*health\|healthz" "$LOAD_DATA_SCRIPT"; then
        pass "Script checks Hugr health"
    else
        fail "Script doesn't check Hugr health"
    fi
else
    fail "Script does not exist"
fi

# Test 8: Script uses proper output symbols
test_case "load-data.sh uses proper output symbols"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    MISSING=""
    grep -q "✓" "$LOAD_DATA_SCRIPT" || MISSING="$MISSING ✓"
    grep -q "→" "$LOAD_DATA_SCRIPT" || MISSING="$MISSING →"

    if [ -z "$MISSING" ]; then
        pass "Script uses proper output symbols (✓, →)"
    else
        fail "Script missing symbols:$MISSING"
    fi
else
    fail "Script does not exist"
fi

# Test 9: Script handles --skip-northwind flag
test_case "load-data.sh recognizes --skip-northwind flag"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -q -- "--skip-northwind" "$LOAD_DATA_SCRIPT"; then
        pass "Script handles --skip-northwind flag"
    else
        fail "Script doesn't handle --skip-northwind flag"
    fi
else
    fail "Script does not exist"
fi

# Test 10: Script handles --skip-synthea flag
test_case "load-data.sh recognizes --skip-synthea flag"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -q -- "--skip-synthea" "$LOAD_DATA_SCRIPT"; then
        pass "Script handles --skip-synthea flag"
    else
        fail "Script doesn't handle --skip-synthea flag"
    fi
else
    fail "Script does not exist"
fi

# Test 11: Script handles --skip-openpayments flag
test_case "load-data.sh recognizes --skip-openpayments flag"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -q -- "--skip-openpayments" "$LOAD_DATA_SCRIPT"; then
        pass "Script handles --skip-openpayments flag"
    else
        fail "Script doesn't handle --skip-openpayments flag"
    fi
else
    fail "Script does not exist"
fi

# Test 12: Script handles --skip-owm flag
test_case "load-data.sh recognizes --skip-owm flag"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -q -- "--skip-owm" "$LOAD_DATA_SCRIPT"; then
        pass "Script handles --skip-owm flag"
    else
        fail "Script doesn't handle --skip-owm flag"
    fi
else
    fail "Script does not exist"
fi

# Test 13: Script handles --force flag
test_case "load-data.sh recognizes --force flag"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -q -- "--force" "$LOAD_DATA_SCRIPT"; then
        pass "Script handles --force flag"
    else
        fail "Script doesn't handle --force flag"
    fi
else
    fail "Script does not exist"
fi

# Test 14: Script mentions data verification
test_case "load-data.sh includes data verification"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -qi "verif\|patient.*count\|record.*count" "$LOAD_DATA_SCRIPT"; then
        pass "Script includes verification logic"
    else
        fail "Script doesn't include verification logic"
    fi
else
    fail "Script does not exist"
fi

# Test 15: Script uses GraphQL mutations
test_case "load-data.sh uses GraphQL API"
if [ -f "$LOAD_DATA_SCRIPT" ]; then
    if grep -qi "graphql\|curl.*8080\|x-hugr-secret" "$LOAD_DATA_SCRIPT"; then
        pass "Script uses GraphQL API"
    else
        fail "Script doesn't use GraphQL API"
    fi
else
    fail "Script does not exist"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: load-data.sh Interface Contract"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until load-data.sh is implemented (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All interface contract tests passed!${NC}"
    exit 0
fi
