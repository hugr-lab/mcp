#!/usr/bin/env bash
# T004: Contract test for start.sh interface
# Tests: --help, --reset, --no-data, --verbose flags, exit codes, output format
# Expected: MUST FAIL until start.sh is implemented

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
START_SCRIPT="$PROJECT_ROOT/lde/scripts/start.sh"

# Test result tracking
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
test_case "start.sh exists and is executable"
if [ -f "$START_SCRIPT" ] && [ -x "$START_SCRIPT" ]; then
    pass "Script exists and is executable"
else
    fail "Script does not exist or is not executable at $START_SCRIPT"
fi

# Test 2: Script uses correct shebang
test_case "start.sh has correct shebang"
if [ -f "$START_SCRIPT" ]; then
    SHEBANG=$(head -n1 "$START_SCRIPT")
    if [ "$SHEBANG" = "#!/usr/bin/env bash" ]; then
        pass "Correct shebang: #!/usr/bin/env bash"
    else
        fail "Incorrect shebang: $SHEBANG (expected: #!/usr/bin/env bash)"
    fi
else
    fail "Script does not exist"
fi

# Test 3: Script has strict mode (set -euo pipefail)
test_case "start.sh uses strict mode"
if [ -f "$START_SCRIPT" ]; then
    if grep -q "set -euo pipefail" "$START_SCRIPT"; then
        pass "Strict mode enabled"
    else
        fail "Strict mode not found (expected: set -euo pipefail)"
    fi
else
    fail "Script does not exist"
fi

# Test 4: --help flag works
test_case "start.sh --help displays usage"
if [ -f "$START_SCRIPT" ] && [ -x "$START_SCRIPT" ]; then
    OUTPUT=$("$START_SCRIPT" --help 2>&1 || true)
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

# Test 5: --help mentions all flags
test_case "start.sh --help documents all flags"
if [ -f "$START_SCRIPT" ] && [ -x "$START_SCRIPT" ]; then
    OUTPUT=$("$START_SCRIPT" --help 2>&1 || true)

    MISSING_FLAGS=""
    echo "$OUTPUT" | grep -q -- "--reset" || MISSING_FLAGS="$MISSING_FLAGS --reset"
    echo "$OUTPUT" | grep -q -- "--no-data" || MISSING_FLAGS="$MISSING_FLAGS --no-data"
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

# Test 6: Exit codes are documented (contract requirement)
test_case "start.sh follows exit code contract"
if [ -f "$START_SCRIPT" ]; then
    # Check if exit codes 0-4 are used appropriately
    # This is a static check - we'll verify actual behavior in integration tests
    if grep -q "exit 1\|exit 2\|exit 3\|exit 4" "$START_SCRIPT"; then
        pass "Script uses standard exit codes"
    else
        fail "Script doesn't implement expected exit codes (1=prereq, 2=docker, 3=health, 4=data)"
    fi
else
    fail "Script does not exist"
fi

# Test 7: Script checks prerequisites
test_case "start.sh checks for required prerequisites"
if [ -f "$START_SCRIPT" ]; then
    MISSING=""
    grep -q "docker" "$START_SCRIPT" || MISSING="$MISSING docker"
    grep -q "curl" "$START_SCRIPT" || MISSING="$MISSING curl"
    grep -q "jq" "$START_SCRIPT" || MISSING="$MISSING jq"

    if [ -z "$MISSING" ]; then
        pass "Script checks for required tools"
    else
        fail "Script doesn't check for:$MISSING"
    fi
else
    fail "Script does not exist"
fi

# Test 8: Script uses proper output symbols (✓, ✗, →)
test_case "start.sh uses proper output symbols"
if [ -f "$START_SCRIPT" ]; then
    MISSING=""
    grep -q "✓" "$START_SCRIPT" || MISSING="$MISSING ✓"
    grep -q "→" "$START_SCRIPT" || MISSING="$MISSING →"

    if [ -z "$MISSING" ]; then
        pass "Script uses proper output symbols (✓, →)"
    else
        fail "Script missing symbols:$MISSING"
    fi
else
    fail "Script does not exist"
fi

# Test 9: Script handles --reset flag
test_case "start.sh recognizes --reset flag"
if [ -f "$START_SCRIPT" ]; then
    if grep -q -- "--reset" "$START_SCRIPT"; then
        pass "Script handles --reset flag"
    else
        fail "Script doesn't handle --reset flag"
    fi
else
    fail "Script does not exist"
fi

# Test 10: Script handles --no-data flag
test_case "start.sh recognizes --no-data flag"
if [ -f "$START_SCRIPT" ]; then
    if grep -q -- "--no-data" "$START_SCRIPT"; then
        pass "Script handles --no-data flag"
    else
        fail "Script doesn't handle --no-data flag"
    fi
else
    fail "Script does not exist"
fi

# Test 11: Script handles --verbose flag
test_case "start.sh recognizes --verbose/-v flag"
if [ -f "$START_SCRIPT" ]; then
    if grep -q -- "--verbose\|-v" "$START_SCRIPT"; then
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
echo "Test Results: start.sh Interface Contract"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until start.sh is implemented (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All interface contract tests passed!${NC}"
    exit 0
fi
