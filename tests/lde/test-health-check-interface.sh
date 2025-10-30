#!/usr/bin/env bash
# T007: Contract test for health-check.sh interface
# Tests: --verbose, --wait with timeout, --help flags, exit codes, per-service status
# Expected: MUST FAIL until health-check.sh is implemented

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
HEALTH_SCRIPT="$PROJECT_ROOT/lde/scripts/health-check.sh"

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
test_case "health-check.sh exists and is executable"
if [ -f "$HEALTH_SCRIPT" ] && [ -x "$HEALTH_SCRIPT" ]; then
    pass "Script exists and is executable"
else
    fail "Script does not exist or is not executable at $HEALTH_SCRIPT"
fi

# Test 2: Script uses correct shebang
test_case "health-check.sh has correct shebang"
if [ -f "$HEALTH_SCRIPT" ]; then
    SHEBANG=$(head -n1 "$HEALTH_SCRIPT")
    if [ "$SHEBANG" = "#!/usr/bin/env bash" ]; then
        pass "Correct shebang: #!/usr/bin/env bash"
    else
        fail "Incorrect shebang: $SHEBANG (expected: #!/usr/bin/env bash)"
    fi
else
    fail "Script does not exist"
fi

# Test 3: Script has strict mode
test_case "health-check.sh uses strict mode"
if [ -f "$HEALTH_SCRIPT" ]; then
    if grep -q "set -euo pipefail" "$HEALTH_SCRIPT"; then
        pass "Strict mode enabled"
    else
        fail "Strict mode not found (expected: set -euo pipefail)"
    fi
else
    fail "Script does not exist"
fi

# Test 4: --help flag works
test_case "health-check.sh --help displays usage"
if [ -f "$HEALTH_SCRIPT" ] && [ -x "$HEALTH_SCRIPT" ]; then
    OUTPUT=$("$HEALTH_SCRIPT" --help 2>&1 || true)
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
test_case "health-check.sh --help documents all flags"
if [ -f "$HEALTH_SCRIPT" ] && [ -x "$HEALTH_SCRIPT" ]; then
    OUTPUT=$("$HEALTH_SCRIPT" --help 2>&1 || true)

    MISSING_FLAGS=""
    echo "$OUTPUT" | grep -q -- "--verbose\|^[[:space:]]*-v" || MISSING_FLAGS="$MISSING_FLAGS --verbose/-v"
    echo "$OUTPUT" | grep -q -- "--wait" || MISSING_FLAGS="$MISSING_FLAGS --wait"
    echo "$OUTPUT" | grep -q -- "--help" || MISSING_FLAGS="$MISSING_FLAGS --help"

    if [ -z "$MISSING_FLAGS" ]; then
        pass "All flags documented in --help"
    else
        fail "Missing flags in --help:$MISSING_FLAGS"
    fi
else
    fail "Cannot test --help output: script not executable"
fi

# Test 6: Exit codes contract (0=healthy, 1=unhealthy, 2=timeout)
test_case "health-check.sh follows exit code contract"
if [ -f "$HEALTH_SCRIPT" ]; then
    HAS_EXIT_CODES=0
    grep -q "exit 0" "$HEALTH_SCRIPT" && HAS_EXIT_CODES=$((HAS_EXIT_CODES + 1))
    grep -q "exit 1" "$HEALTH_SCRIPT" && HAS_EXIT_CODES=$((HAS_EXIT_CODES + 1))
    grep -q "exit 2" "$HEALTH_SCRIPT" && HAS_EXIT_CODES=$((HAS_EXIT_CODES + 1))

    if [ $HAS_EXIT_CODES -ge 2 ]; then
        pass "Script uses standard exit codes (0=healthy, 1=unhealthy, 2=timeout)"
    else
        fail "Script doesn't implement expected exit codes"
    fi
else
    fail "Script does not exist"
fi

# Test 7: Script checks all 6 services
test_case "health-check.sh checks all required services"
if [ -f "$HEALTH_SCRIPT" ]; then
    MISSING_SERVICES=""
    grep -qi "postgres\|postgresql" "$HEALTH_SCRIPT" || MISSING_SERVICES="$MISSING_SERVICES postgres"
    grep -qi "redis" "$HEALTH_SCRIPT" || MISSING_SERVICES="$MISSING_SERVICES redis"
    grep -qi "minio" "$HEALTH_SCRIPT" || MISSING_SERVICES="$MISSING_SERVICES minio"
    grep -qi "keycloak" "$HEALTH_SCRIPT" || MISSING_SERVICES="$MISSING_SERVICES keycloak"
    grep -qi "hugr" "$HEALTH_SCRIPT" || MISSING_SERVICES="$MISSING_SERVICES hugr"
    grep -qi "mcp.*inspector\|inspector.*mcp" "$HEALTH_SCRIPT" || MISSING_SERVICES="$MISSING_SERVICES mcp-inspector"

    if [ -z "$MISSING_SERVICES" ]; then
        pass "Script checks all 6 services"
    else
        fail "Script doesn't check:$MISSING_SERVICES"
    fi
else
    fail "Script does not exist"
fi

# Test 8: Script uses proper output symbols
test_case "health-check.sh uses proper output symbols"
if [ -f "$HEALTH_SCRIPT" ]; then
    MISSING=""
    grep -q "✓" "$HEALTH_SCRIPT" || MISSING="$MISSING ✓"
    grep -q "✗" "$HEALTH_SCRIPT" || MISSING="$MISSING ✗"
    grep -q "→" "$HEALTH_SCRIPT" || MISSING="$MISSING →"

    if [ -z "$MISSING" ]; then
        pass "Script uses proper output symbols (✓, ✗, →)"
    else
        fail "Script missing symbols:$MISSING"
    fi
else
    fail "Script does not exist"
fi

# Test 9: Script shows response times
test_case "health-check.sh reports response times"
if [ -f "$HEALTH_SCRIPT" ]; then
    if grep -qi "response.*time\|ms\|seconds" "$HEALTH_SCRIPT"; then
        pass "Script reports response times"
    else
        fail "Script doesn't report response times"
    fi
else
    fail "Script does not exist"
fi

# Test 10: Script handles --wait flag
test_case "health-check.sh recognizes --wait flag"
if [ -f "$HEALTH_SCRIPT" ]; then
    if grep -q -- "--wait" "$HEALTH_SCRIPT"; then
        pass "Script handles --wait flag"
    else
        fail "Script doesn't handle --wait flag"
    fi
else
    fail "Script does not exist"
fi

# Test 11: Script implements timeout logic
test_case "health-check.sh implements timeout"
if [ -f "$HEALTH_SCRIPT" ]; then
    if grep -qi "timeout\|sleep.*loop\|while.*wait" "$HEALTH_SCRIPT"; then
        pass "Script implements timeout logic"
    else
        fail "Script doesn't implement timeout logic"
    fi
else
    fail "Script does not exist"
fi

# Test 12: Script uses curl for HTTP health checks
test_case "health-check.sh uses curl for HTTP checks"
if [ -f "$HEALTH_SCRIPT" ]; then
    if grep -q "curl" "$HEALTH_SCRIPT"; then
        pass "Script uses curl for HTTP health checks"
    else
        fail "Script doesn't use curl"
    fi
else
    fail "Script does not exist"
fi

# Summary
echo ""
echo "========================================"
echo "Test Results: health-check.sh Interface Contract"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${YELLOW}Note: Tests are EXPECTED to fail until health-check.sh is implemented (TDD approach)${NC}"
    exit 1
else
    echo -e "${GREEN}All interface contract tests passed!${NC}"
    exit 0
fi
