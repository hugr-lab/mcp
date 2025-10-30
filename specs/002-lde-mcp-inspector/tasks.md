# Tasks: LDE MCP Inspector

**Feature**: 002-lde-mcp-inspector
**Input**: Design documents from `/specs/002-lde-mcp-inspector/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Tech stack: Bash 4.0+, Docker Compose v2.0+, YAML, Markdown
   → Structure: Infrastructure/DevOps modifications
2. Load design documents:
   → data-model.md: 4 entities (Service, Configuration, Test History, Health Status)
   → contracts/: 3 contract files (docker-compose, env vars, scripts)
   → quickstart.md: 5 integration scenarios
3. Generate tasks by category:
   → Setup: Environment preparation
   → Tests: 7 contract tests (3 files × multiple checks + 5 integration tests)
   → Core: 6 file modifications (docker-compose, env, 4 scripts)
   → Integration: End-to-end validation
   → Polish: Documentation, cleanup
4. Apply task rules:
   → Contract tests in different files = [P]
   → Script modifications = sequential (dependency order)
   → Documentation updates = [P]
5. Number tasks sequentially (T001-T022)
6. Validate: All contracts tested, all scenarios covered
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- All file paths are absolute from repository root

## Phase 3.1: Setup

- [x] **T001** Verify Docker and Docker Compose installed (versions: Docker 20.10+, Compose v2.0+)
  - Command: `docker --version && docker compose version`
  - Exit if requirements not met

- [x] **T002** Create `.local/mcp-inspector/` directory structure
  - Path: `lde/.local/mcp-inspector/`
  - Ensure parent directories exist
  - Set appropriate permissions

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests (Parallel - Different Files)

- [ ] **T003** [P] Contract test: Docker Compose service schema validation in `tests/lde/test-mcp-inspector-compose.sh`
  - Validate `mcp-inspector` service exists in docker-compose.yml
  - Check required fields: image, container_name, ports, environment, volumes, restart
  - Verify image pattern matches `ghcr.io/modelcontextprotocol/inspector:*`
  - Verify container name is `lde-mcp-inspector`
  - Verify ports include 19007 mapping
  - Verify environment includes HOST and ALLOWED_ORIGINS
  - Verify volume mounts `.local/mcp-inspector`
  - Test MUST FAIL initially (service not yet added)
  - Exit code: 1 if validation fails, 0 if passes

- [ ] **T004** [P] Contract test: Environment variables in `tests/lde/test-mcp-inspector-env.sh`
  - Verify `.env.example` contains MCP_INSPECTOR_IMAGE
  - Verify `.env.example` contains MCP_INSPECTOR_PORT with default 19007
  - Verify `.env.example` contains MCP_INSPECTOR_HOST with default 0.0.0.0
  - Verify `.env.example` contains MCP_INSPECTOR_ALLOWED_ORIGINS
  - Verify section header comment "# MCP Inspector" present
  - Test MUST FAIL initially (env vars not yet added)
  - Exit code: 1 if validation fails, 0 if passes

- [ ] **T005** [P] Contract test: Script interfaces in `tests/lde/test-mcp-inspector-scripts.sh`
  - Verify `start.sh` includes MCP Inspector in startup sequence
  - Verify `stop.sh` includes MCP Inspector in shutdown
  - Verify `health-check.sh` checks MCP Inspector status (container running)
  - Verify `health-check.sh` counts 6 services total
  - Verify `cleanup.sh` mentions MCP Inspector in warning text
  - Test MUST FAIL initially (scripts not yet modified)
  - Exit code: 1 if validation fails, 0 if passes

### Integration Tests (Parallel - Different Scenarios)

- [ ] **T006** [P] Integration test: Start LDE with MCP Inspector in `tests/lde/test-integration-start.sh`
  - Run `./lde/scripts/start.sh`
  - Verify exit code 0
  - Verify output contains "MCP Inspector" and port 19007
  - Verify container `lde-mcp-inspector` is running
  - Verify port 19007 responds to HTTP (curl check)
  - Test MUST FAIL initially (service not configured)
  - Cleanup: `./lde/scripts/cleanup.sh --force`

- [ ] **T007** [P] Integration test: Health check includes Inspector in `tests/lde/test-integration-health.sh`
  - Start LDE (prerequisite)
  - Run `./lde/scripts/health-check.sh`
  - Verify exit code 0
  - Verify output contains "MCP Inspector: ✓ Healthy"
  - Verify total shows "(6/6)"
  - Verify response time displayed
  - Test MUST FAIL initially (health check not updated)
  - Cleanup: `./lde/scripts/cleanup.sh --force`

- [ ] **T008** [P] Integration test: Persistence across restarts in `tests/lde/test-integration-persistence.sh`
  - Start LDE
  - Create test file in `.local/mcp-inspector/test-marker.txt`
  - Run `./lde/scripts/stop.sh`
  - Verify test file still exists
  - Run `./lde/scripts/start.sh`
  - Verify test file still exists
  - Verify container remounted volume
  - Test MUST FAIL initially (volume not configured)
  - Cleanup: `./lde/scripts/cleanup.sh --force`

- [ ] **T009** [P] Integration test: Cleanup removes data in `tests/lde/test-integration-cleanup.sh`
  - Start LDE
  - Create test data in `.local/mcp-inspector/`
  - Run `./lde/scripts/cleanup.sh --force`
  - Verify `.local/mcp-inspector/` removed
  - Verify container removed
  - Test with `--keep-local` flag
  - Verify `.local/mcp-inspector/` preserved when flag used
  - Test MUST FAIL initially (cleanup not updated)

- [ ] **T010** [P] Integration test: Stop gracefully shuts down in `tests/lde/test-integration-stop.sh`
  - Start LDE
  - Run `./lde/scripts/stop.sh`
  - Verify exit code 0
  - Verify no `lde-mcp-inspector` container running
  - Verify `.local/mcp-inspector/` data still exists
  - Test MUST FAIL initially (stop script not updated)
  - Cleanup: `rm -rf ./lde/.local/mcp-inspector`

## Phase 3.3: Core Implementation (ONLY after tests T003-T010 are failing)

**IMPORTANT**: Run all tests before starting implementation to verify they fail

### Docker Compose Configuration

- [x] **T011** Add MCP Inspector service to `lde/docker-compose.yml`
  - Add service block for `mcp-inspector` after `hugr` service
  - Set image: `${MCP_INSPECTOR_IMAGE:-ghcr.io/modelcontextprotocol/inspector:latest}`
  - Set container_name: `lde-mcp-inspector`
  - Map ports: `${MCP_INSPECTOR_PORT:-19007}:6274` and `${MCP_INSPECTOR_PORT:-19007}:6277`
  - Add environment:
    - `HOST=${MCP_INSPECTOR_HOST:-0.0.0.0}`
    - `ALLOWED_ORIGINS=${MCP_INSPECTOR_ALLOWED_ORIGINS:-http://localhost:19007}`
  - Add volume: `./.local/mcp-inspector:/data`
  - Set restart: `unless-stopped`
  - Verify: `docker compose -f lde/docker-compose.yml config` validates
  - Verify: T003 contract test now passes

- [x] **T012** Add MCP Inspector environment variables to `lde/.env.example`
  - Add section header comment: `# MCP Inspector`
  - Add: `MCP_INSPECTOR_IMAGE=ghcr.io/modelcontextprotocol/inspector:latest`
  - Add: `MCP_INSPECTOR_PORT=19007`
  - Add: `MCP_INSPECTOR_HOST=0.0.0.0`
  - Add: `MCP_INSPECTOR_ALLOWED_ORIGINS=http://localhost:19007`
  - Place after Keycloak section, before manual additions
  - Verify: T004 contract test now passes

### Script Modifications (Sequential - Dependency Order)

- [x] **T013** Update `lde/scripts/health-check.sh` to check MCP Inspector
  - Add status variables: `MCP_INSPECTOR_STATUS`, `MCP_INSPECTOR_TIME`, `MCP_INSPECTOR_ERROR`
  - Add `check_mcp_inspector()` function:
    - Check container running status via `docker ps`
    - Set status to "healthy" if running, "unhealthy" if not
    - Record response time
  - Call `check_mcp_inspector` in `check_all_services()`
  - Update `display_status()`:
    - Add MCP Inspector to service list
    - Update total from 5 to 6
    - Display status with color coding
  - Update `wait_for_healthy()`: change service count from 5 to 6
  - Update help text to list MCP Inspector
  - Verify: T005 contract test passes (health check portion)
  - Verify: T007 integration test passes

- [x] **T014** Update `lde/scripts/start.sh` to include MCP Inspector
  - Service starts automatically via `docker compose up` (no code change needed)
  - Update service count validation from 5 to 6
  - Update success message to include MCP Inspector:
    - Add line: `MCP Inspector: http://localhost:19007`
  - Verify: Output displays MCP Inspector URL
  - Verify: T006 integration test passes

- [x] **T015** Update `lde/scripts/stop.sh` to handle MCP Inspector
  - Service stops automatically via `docker compose stop` (no code change needed)
  - Verify: MCP Inspector container stops with other services
  - Verify: T010 integration test passes

- [x] **T016** Update `lde/scripts/cleanup.sh` to mention MCP Inspector
  - In `show_warning()` function, add to `.local/` directory list:
    - Line: `echo "     - MCP Inspector test history"`
  - No functional change to cleanup logic (auto-cleaned with `.local/`)
  - Verify: Warning mentions MCP Inspector
  - Verify: T009 integration test passes

## Phase 3.4: Integration Validation

- [ ] **T017** Run full integration test suite
  - Execute all integration tests (T006-T010) in sequence
  - Verify all tests pass
  - Capture any failures and fix related implementation
  - Exit code must be 0 for all tests

- [ ] **T018** Verify service count consistency across all scripts
  - Grep for hardcoded "5" service references
  - Ensure all updated to "6" where appropriate
  - Check `start.sh`, `health-check.sh`, test files
  - Verify: No inconsistencies remain

## Phase 3.5: Documentation

- [x] **T019** [P] Update `lde/README.md` with MCP Inspector section
  - Add MCP Inspector to Services overview (update count from 5 to 6)
  - Add new section after Keycloak:
    ```markdown
    ### MCP Inspector (`http://localhost:19007`)
    - Web-based testing and debugging tool for MCP servers
    - Features: Interactive UI, connection management, test history
    - Volume: `.local/mcp-inspector`
    - Test history persists across restarts, removed by cleanup.sh
    ```
  - Update Port Mappings table:
    - Add row: `| MCP Inspector | 19007 | 6274, 6277 | Web UI + Proxy |`
  - Update Services section in Quick Start expected output
  - Update Troubleshooting section with MCP Inspector examples
  - Reference quickstart.md for validation scenarios

- [x] **T020** [P] Update root `README.md` with MCP Inspector reference
  - Update LDE services list to include MCP Inspector (6 services)
  - Add brief description: "Web-based MCP server testing tool"
  - Update port range reference (19000-19007 instead of 19000-19006)
  - Add link to `lde/README.md` for details

- [x] **T021** [P] Verify existing test scripts account for 6 services
  - Check `tests/lde/test-start-interface.sh`
  - Update expected service count from 5 to 6
  - Update any service name assertions
  - Check `tests/lde/test-compose-validity.sh`
  - Add MCP Inspector to service validation list

## Phase 3.6: Polish & Validation

- [ ] **T022** Execute quickstart.md validation scenarios
  - Follow all 5 scenarios in `specs/002-lde-mcp-inspector/quickstart.md`
  - Scenario 1: Start LDE and verify Inspector accessible
  - Scenario 2: Health check shows Inspector status
  - Scenario 3: Test history persists across restarts
  - Scenario 4: Cleanup removes test history
  - Scenario 5: Stop gracefully shuts down Inspector
  - Check all success criteria boxes
  - Document any issues found
  - All scenarios must pass for feature completion

## Dependencies

```
Setup (T001-T002)
  ↓
Tests Written (T003-T010) - All parallel, must fail initially
  ↓
Docker Compose (T011) ──────┐
  ↓                         │
Environment Variables (T012)│
  ↓                         │
Health Check Script (T013)  │← Depends on T011 (service must exist)
  ↓                         │
Start Script (T014) ────────┘
  ↓
Stop Script (T015)
  ↓
Cleanup Script (T016)
  ↓
Integration Validation (T017-T018)
  ↓
Documentation (T019-T021) - All parallel
  ↓
Final Validation (T022)
```

**Key Constraints**:
- T003-T010 MUST be written and failing before T011
- T013 depends on T011 (service definition must exist)
- T014-T016 sequential (start → stop → cleanup order)
- T019-T021 parallel (different files)

## Parallel Execution Examples

### Contract Tests (T003-T005)
```bash
# Run in parallel - different test files:
./tests/lde/test-mcp-inspector-compose.sh &
./tests/lde/test-mcp-inspector-env.sh &
./tests/lde/test-mcp-inspector-scripts.sh &
wait
```

### Integration Tests (T006-T010)
```bash
# Run in sequence - each test needs clean environment:
./tests/lde/test-integration-start.sh
./tests/lde/test-integration-health.sh
./tests/lde/test-integration-persistence.sh
./tests/lde/test-integration-cleanup.sh
./tests/lde/test-integration-stop.sh
```

### Documentation (T019-T021)
```bash
# Edit in parallel - different files:
# Terminal 1: Edit lde/README.md
# Terminal 2: Edit README.md
# Terminal 3: Edit tests/lde/test-*.sh
```

## Validation Checklist

*GATE: Verify before marking feature complete*

- [x] All contracts (3 files) have corresponding tests (T003-T005)
- [x] All integration scenarios (5 scenarios) have tests (T006-T010)
- [x] All tests come before implementation (T003-T010 before T011-T016)
- [x] Parallel tasks are truly independent (different files, no shared state)
- [x] Each task specifies exact file path
- [x] Script modifications in dependency order (T013→T014→T015→T016)
- [x] Documentation updates are parallel (T019-T021)
- [x] Service count updated consistently (5→6 everywhere)

## Notes

- **[P] tasks**: Can run in parallel (different files, no dependencies)
- **TDD Critical**: All tests (T003-T010) must be written and failing before implementation starts (T011)
- **No Implementation Code**: This is infrastructure configuration (YAML, shell scripts, Markdown)
- **Health Check Method**: Container running status only (per clarification), no HTTP endpoint checks
- **Port Assignment**: 19007 is sequential after Hugr health endpoint (19006)
- **Volume Strategy**: `.local/mcp-inspector/` follows LDE pattern, auto-cleaned with `.local/`
- **Test Cleanup**: Each integration test must clean up after itself

## Task Execution Strategy

1. **Day 1**: Setup + All Tests (T001-T010)
   - Verify environment
   - Write all contract tests (expect failures)
   - Write all integration tests (expect failures)
   - Confirm all 8 tests fail correctly

2. **Day 2**: Core Implementation (T011-T016)
   - Add Docker Compose service
   - Add environment variables
   - Update scripts in order
   - Watch tests turn green

3. **Day 3**: Documentation + Validation (T017-T022)
   - Run integration suite
   - Update documentation
   - Execute quickstart scenarios
   - Final verification

**Total Tasks**: 22
**Parallel Tasks**: 8 (T003-T005, T006-T010, T019-T021)
**Sequential Tasks**: 14 (Setup, Implementation, Validation)
**Estimated Completion**: 2-3 days
