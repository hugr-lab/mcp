#!/usr/bin/env bash
# T005: Contract test for stop.sh interface
# Tests: --help, --verbose flags, exit codes, data persistence message
# Expected: MUST FAIL until stop.sh is implemented

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
STOP_SCRIPT="$PROJECT_ROOT/lde/scripts/stop.sh"

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
test_case "stop.sh exists and is executable"
if [ -f "$STOP_SCRIPT" ] && [ -x "$STOP_SCRIPT" ]; then
    pass "Script exists and is executable"
else
    fail "Script does not exist or is not executable at $STOP_SCRIPT"
fi

# Test 2: Script uses correct shebang
test_case "stop.sh has correct shebang"
if [ -f "$STOP_SCRIPT" ]; then
    SHEBANG=$(head -n1 "$STOP_SCRIPT")
    if [ "$SHEBANG" = "#!/usr/bin/env bash" ]; then
        pass "Correct shebang: #!/usr/bin/env bash"
    else
        fail "Incorrect shebang: $SHEBANG (expected: #!/usr/bin/env bash)"
    fi
else
    fail "Script does not exist"
fi

# Test 3: Script has strict mode
test_case "stop.sh uses strict mode"
if [ -f "$STOP_SCRIPT" ]; then
    if grep -q "set -euo pipefail" "$STOP_SCRIPT"; then
        pass "Strict mode enabled"
    else
        fail "Strict mode not found (expected: set -euo pipefail)"
    fi
else
    fail "Script does not exist"
fi

# Test 4: --help flag works
test_case "stop.sh --help displays usage"
if [ -f "$STOP_SCRIPT" ] && [ -x "$STOP_SCRIPT" ]; then
    OUTPUT=$("$STOP_SCRIPT" --help 2>&1 || true)
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

# Test 5: --help mentions required flags
test_case "stop.sh --help documents flags"
if [ -f "$STOP_SCRIPT" ] && [ -x "$STOP_SCRIPT" ]; then
    OUTPUT=$("$STOP_SCRIPT" --help 2>&1 || true)

    MISSING_FLAGS=""
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

# Test 6: Exit codes contract (0=success, 1=docker fail)
test_case "stop.sh follows exit code contract"
if [ -f "$STOP_SCRIPT" ]; then
    # Check if exit codes 0 and 1 are used appropriately
    if grep -q "exit 0\|exit 1" "$STOP_SCRIPT"; then
        pass "Script uses standard exit codes"
    else
        fail "Script doesn't implement expected exit codes (0=success, 1=failure)"
    fi
else
    fail "Script does not exist"
fi

# Test 7: Script uses docker compose down
test_case "stop.sh uses docker compose down"
if [ -f "$STOP_SCRIPT" ]; then
    if grep -q "docker compose down\|docker-compose down" "$STOP_SCRIPT"; then
        pass "Script uses docker compose down"
    else
        fail "Script doesn't use docker compose down command"
    fi
else
    fail "Script does not exist"
fi

# Test 8: Script uses proper output symbols
test_case "stop.sh uses proper output symbols"
if [ -f "$STOP_SCRIPT" ]; then
    MISSING=""
    grep -q "✓" "$STOP_SCRIPT" || MISSING="$MISSING ✓"
    grep -q "→" "$STOP_SCRIPT" || MISSING="$MISSING →"

    if [ -z "$MISSING" ]; then
        pass "Script uses proper output symbols (✓, →)"
    else
        fail "Script missing symbols:$MISSING"
    fi
else
    fail "Script does not exist"
fi

# Test 9: Script mentions data persistence
test_case "stop.sh mentions data persistence"
if [ -f "$STOP_SCRIPT" ]; then
    if grep -qi "persist\|data.*\.local" "$STOP_SCRIPT"; then
        pass "Script mentions data persistence"
    else
        fail "Script doesn't mention data persistence in output"
    fi
else
    fail "Script does not exist"
fi

# Test 10: Script provides restart instructions
test_case "stop.sh provides restart command"
if [ -f "$STOP_SCRIPT" ]; then
    if grep -q "start\.sh\|restart" "$STOP_SCRIPT"; then
        pass "Script provides restart instructions"
    else
        fail "Script doesn't provide restart command"
    fi
else
    fail "Script does not exist"
fi

# Test 11: Script handles --verbose flag
test_case "stop.sh recognizes --verbose/-v flag"
if [ -f "$STOP_SCRIPT" ]; then
    if grep -q -- "--verbose\|-v" "$STOP_SCRIPT"; then
        pass "Script handles --verbose/-v flag"
    else
        fail "Script doesn't handle --verbose/-v flag"
    fi
else
    fail "Script does not exist"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: stop.sh Interface Contract"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until stop.sh is implemented (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All interface contract tests passed!${NC}"
    exit 0
fi
