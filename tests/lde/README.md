# LDE Contract Tests

This directory contains contract tests for the Local Development Environment (LDE) following Test-Driven Development (TDD) principles.

## Test Philosophy

These tests were written **BEFORE** the implementation, following the TDD "Red-Green-Refactor" cycle:

1. **RED**: Write tests that fail (we are here)
2. **GREEN**: Write minimal code to make tests pass
3. **REFACTOR**: Improve code while keeping tests green

## Test Structure

### Phase 3.2: Contract Tests (T004-T015)

All 12 contract test files validate that implementations meet their specifications defined in:
- `/specs/001-create-a-local/contracts/script-interfaces.md`
- `/specs/001-create-a-local/contracts/docker-compose-schema.yml`

### Script Interface Tests (T004-T007)

| Test ID | File | Tests |
|---------|------|-------|
| T004 | `test-start-interface.sh` | start.sh interface (flags, exit codes, output format) |
| T005 | `test-stop-interface.sh` | stop.sh interface (flags, exit codes, data persistence) |
| T006 | `test-load-data-interface.sh` | load-data.sh interface (flags, exit codes, GraphQL usage) |
| T007 | `test-health-check-interface.sh` | health-check.sh interface (flags, exit codes, service checks) |

### Docker Compose Contract Tests (T008-T015)

| Test ID | File | Tests |
|---------|------|-------|
| T008 | `test-compose-validity.sh` | YAML syntax and docker compose config validation |
| T009 | `test-compose-services.sh` | All 5 required services defined (hugr, postgres, redis, minio, keycloak) |
| T010 | `test-compose-dependencies.sh` | Service dependencies with service_healthy conditions |
| T011 | `test-compose-healthchecks.sh` | Healthchecks for all services |
| T012 | `test-compose-volumes.sh` | Volume persistence using .local/ pattern |
| T013 | `test-compose-ports.sh` | Required port exposures (8080, 8180, 9000, 9001) |
| T014 | `test-compose-env.sh` | Environment variables for all services |
| T015 | `test-compose-keycloak.sh` | Keycloak --import-realm configuration |

## Running Tests

### Run All Contract Tests

```bash
./tests/lde/run-all-contract-tests.sh
```

This runs all 12 contract tests and provides a summary.

### Run Individual Tests

```bash
./tests/lde/test-start-interface.sh
./tests/lde/test-compose-validity.sh
# ... etc
```

### Run Tests in Parallel (Different Groups)

Script interface tests can run in parallel:
```bash
./tests/lde/test-start-interface.sh &
./tests/lde/test-stop-interface.sh &
./tests/lde/test-load-data-interface.sh &
./tests/lde/test-health-check-interface.sh &
wait
```

Docker compose tests can run in parallel:
```bash
./tests/lde/test-compose-*.sh &
wait
```

## Expected Behavior

### Before Implementation (Current State)
- ✗ All tests should FAIL
- This indicates we're in the TDD "Red" phase
- Ready to proceed with implementation

### After Implementation
- ✓ All tests should PASS
- This indicates implementation meets contracts
- Ready for integration testing (Phase 3.4)

## Test Output

Each test provides:
- Clear test case descriptions
- Pass/fail status with ✓ and ✗ symbols
- Color-coded output (green=pass, red=fail, yellow=info)
- Summary statistics
- Helpful reminder that failures are expected in TDD

Example output:
```
Test 1: start.sh exists and is executable
✗ Script does not exist or is not executable

========================================
Test Results: start.sh Interface Contract
========================================
Tests run:    11
Tests passed: 0
Tests failed: 11
========================================
Note: Tests are EXPECTED to fail until start.sh is implemented (TDD approach)
```

## Contract Specifications

Each test validates specific contract requirements:

### Script Contracts
- Shebang: `#!/usr/bin/env bash`
- Strict mode: `set -euo pipefail`
- Flags: `--help`, `--verbose/-v`, and specific flags per script
- Exit codes: Specific codes for different failure modes
- Output: Use of ✓, ✗, → symbols for consistent UX
- Idempotency: Safe to run multiple times

### Docker Compose Contracts
- YAML validity
- Service definitions
- Dependency graph (hugr depends on all others)
- Healthchecks (test, interval, timeout, retries)
- Volume persistence (.local/ pattern)
- Port exposures (no conflicts)
- Environment variables (all required vars)
- Keycloak realm import

## Next Steps

1. **Verify all tests fail**: Run `./tests/lde/run-all-contract-tests.sh`
2. **Implement Phase 3.3**: Create docker-compose.yml and scripts
3. **Watch tests turn green**: Re-run tests as implementation progresses
4. **Integration tests**: Move to Phase 3.4 when all contract tests pass

## References

- Task breakdown: `/specs/001-create-a-local/tasks.md`
- Script contracts: `/specs/001-create-a-local/contracts/script-interfaces.md`
- Compose schema: `/specs/001-create-a-local/contracts/docker-compose-schema.yml`
