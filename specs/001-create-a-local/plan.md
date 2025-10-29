
# Implementation Plan: Local Development Environment for Hugr MCP Service

**Branch**: `001-create-a-local` | **Date**: 2025-10-27 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/Users/vgribanov/projects/hugr-lab/mcp/specs/001-create-a-local/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code, or `AGENTS.md` for all other agents).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Create a reproducible local development environment (LDE) for the Hugr MCP service using Docker Compose to orchestrate all required dependencies (Hugr, Keycloak, Postgres, Redis, MinIO). Provide shell scripts for environment lifecycle management (start, stop, reset) and data loading scripts to populate Hugr with healthcare datasets: 10,000 Synthea-generated patients (DuckDB format) and Open Payments 2023 data. Environment data persists in `.local` directory following the hugr examples project pattern, with optional reset functionality to start fresh.

## Technical Context
**Language/Version**: Shell scripts (Bash 4.0+), Docker Compose v2.0+
**Primary Dependencies**:
- Docker Engine 20.10+
- Docker Compose v2.0+
- curl (for GraphQL mutations)
- jq (for JSON processing)
- Local repositories: `/Users/vgribanov/projects/hugr-lab/examples`, `/Users/vgribanov/projects/synthea`

**Services** (versions from hugr examples project):
- Hugr (latest from hugr examples)
- Keycloak (latest from hugr examples)
- PostgreSQL (latest from hugr examples)
- Redis (latest from hugr examples)
- MinIO (latest from hugr examples)

**Storage**:
- Volume persistence in `.local/` directory
- Service-specific volumes: pg-data, minio, keycloak, redis-data
- Generated data: Synthea DuckDB files, Open Payments CSV/Parquet

**Testing**: Shell script validation, service health checks, data verification queries
**Target Platform**: macOS (Darwin 24.6.0), Linux-compatible
**Project Type**: Infrastructure/DevOps (single project - shell scripts and Docker configs)
**Performance Goals**:
- Environment startup: < 2 minutes for all services
- Data loading: < 10 minutes for 10,000 patients + Open Payments 2023
- Service health checks: < 30 seconds total

**Constraints**:
- Must use configurations from hugr examples project for consistency
- Keycloak realm: admin, viewer, analyst roles (from hugr examples)
- Port exposure: Hugr GraphQL endpoint accessible from host
- Data format: Synthea in DuckDB, Open Payments as in hugr examples
- No hardcoded paths except documented local repo locations

**Scale/Scope**:
- 10,000 Synthea patients
- Open Payments 2023 full dataset
- 2 data sources in Hugr
- 5 Docker services
- 3-4 shell scripts (start, stop, load-data, optional utilities)

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Note**: This feature is infrastructure setup (LDE), not MCP service functionality. Constitutional principles apply as follows:

### I. Schema-First Development
- [x] N/A - Infrastructure feature, no schema introspection
- [x] Services independently testable (health checks, individual service startup)
- [x] Clear separation: Docker services, data loading, scripts

### II. Tool-Based Interface
- [x] N/A - Shell scripts, not MCP tools
- [x] Scripts follow clear input/output patterns (flags, exit codes, stdout/stderr)
- [x] Separation maintained: start vs stop vs load-data

### III. Test-First Development (NON-NEGOTIABLE)
- [x] Validation tests for shell scripts (syntax, required tools, path verification)
- [x] Integration tests for service startup/health
- [x] Data verification tests after loading
- [x] TDD cycle: Write validation/health check tests → Implement scripts → Verify

### IV. Lazy Stepwise Discovery
- [x] N/A - Fixed infrastructure components
- [x] Progressive setup: Services → Data → Verification
- [x] Health checks verify individual service readiness before proceeding

### V. Observability & Error Clarity
- [x] Script output shows status of each operation
- [x] Error messages include actionable context (missing deps, failed services, port conflicts)
- [x] Debug mode via verbose flags (-v, --verbose)
- [x] Service logs accessible via `docker compose logs`

### Additional Infrastructure Principles
- [x] Reproducibility: Same setup on any machine with documented prerequisites
- [x] Idempotency: Scripts can be re-run safely
- [x] Cleanup: Reset flag for clean slate
- [x] Documentation: README/quickstart for first-time setup

**Complexity Justifications**: None required - straightforward infrastructure setup aligns with constitution's observability and testing principles

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
lde/                            # Local Development Environment directory
├── docker-compose.yml          # Service orchestration
├── .env.example               # Environment variable template
├── scripts/
│   ├── start.sh               # Start environment
│   ├── stop.sh                # Stop environment
│   ├── load-data.sh           # Load example datasets
│   └── health-check.sh        # Verify all services healthy
├── keycloak/
│   └── realm-config.json      # Keycloak realm export (from hugr examples)
├── data/
│   └── .gitkeep              # Data sources will be generated here
└── .local/                    # Persisted volumes (git-ignored)
    ├── pg-data/
    ├── minio/
    ├── keycloak/
    └── redis-data/

tests/
└── lde/
    ├── test-scripts.sh        # Script syntax and tool availability
    ├── test-services.sh       # Service health checks
    └── test-data.sh           # Data loading verification
```

**Structure Decision**: Single project (infrastructure/DevOps). Using `lde/` directory at repository root to contain all environment-related files. The `.local/` directory follows hugr examples pattern for volume persistence. Scripts are executable shell scripts with clear naming. Tests verify each aspect independently (scripts, services, data).

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh claude`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
The `/tasks` command will:
1. Load `.specify/templates/tasks-template.md` as base
2. Generate tasks from Phase 1 artifacts:
   - **From contracts/docker-compose-schema.yml**: Docker Compose validation tests
   - **From contracts/script-interfaces.md**: Script interface contract tests for start.sh, stop.sh, load-data.sh, health-check.sh
   - **From data-model.md**: Service configuration tasks, environment setup tasks
   - **From quickstart.md**: Integration test scenarios matching quickstart verification steps

**Task Categories** (in TDD order):
1. **Setup & Prerequisites** (tasks 1-3):
   - Create directory structure (`lde/`, `.local/`, `scripts/`, `keycloak/`, `data/`)
   - Create `.env.example` template
   - Add `.local/` to `.gitignore`

2. **Contract Tests - Scripts** (tasks 4-11) [P - can run in parallel]:
   - Test `start.sh` interface (help, prerequisites check, exit codes)
   - Test `stop.sh` interface
   - Test `load-data.sh` interface
   - Test `health-check.sh` interface

3. **Contract Tests - Docker Compose** (tasks 12-15) [P]:
   - Test docker-compose.yml YAML validity
   - Test all required services defined
   - Test service dependencies configured
   - Test healthchecks configured
   - Test volume persistence pattern

4. **Implementation - Docker Configuration** (tasks 16-20):
   - Write docker-compose.yml with 5 services
   - Extract Keycloak realm config from examples
   - Create `.env.example` with all variables
   - Configure volume mappings to `.local/`
   - Configure health checks for all services

5. **Implementation - Scripts** (tasks 21-28):
   - Implement `start.sh` (prerequisite checks, directory init, docker compose up, health wait, data load)
   - Implement `stop.sh` (docker compose down, preserve volumes)
   - Implement `health-check.sh` (per-service checks, wait mode)
   - Implement `load-data.sh` (Synthea generation, Open Payments loading, Hugr registration)

6. **Integration Tests - Service Health** (tasks 29-33) [P]:
   - Test PostgreSQL health and connectivity
   - Test Redis health and connectivity
   - Test MinIO health and S3 operations
   - Test Keycloak health and realm loading
   - Test Hugr GraphQL endpoint and introspection

7. **Integration Tests - Data Loading** (tasks 34-38):
   - Test Synthea data generation (10,000 patients)
   - Test DuckDB conversion
   - Test Open Payments data processing
   - Test Hugr data source registration
   - Test data verification queries

8. **Integration Tests - Full Workflow** (tasks 39-43):
   - Test full start sequence (ABSENT → HEALTHY)
   - Test stop/restart with data persistence
   - Test reset flag (wipe and fresh start)
   - Test authentication flow (Keycloak token → Hugr query)
   - Test role-based queries (admin, analyst, viewer)

9. **Documentation** (tasks 44-45):
   - Create README.md in `lde/` directory
   - Document port mappings and access URLs

**Ordering Strategy**:
- **TDD order**: All test tasks before corresponding implementation tasks
- **Dependency order**:
  - Setup → Contract tests → Implementation → Integration tests
  - Docker config before scripts (scripts depend on compose file)
  - Scripts before data loading (data loading uses scripts)
- **Parallel execution markers [P]**:
  - All contract tests can run in parallel (independent)
  - All integration service tests can run in parallel
  - Data loading tests sequential (depend on environment state)

**Estimated Output**: ~45 numbered, ordered tasks in tasks.md

**Task Dependencies**:
```
Setup (1-3)
  ↓
Contract Tests (4-15) [All parallel]
  ↓
Docker Implementation (16-20)
  ↓
Script Implementation (21-28)
  ↓
Integration Tests (29-43)
  ├─ Service Health (29-33) [Parallel]
  ├─ Data Loading (34-38) [Sequential]
  └─ Full Workflow (39-43) [Sequential]
  ↓
Documentation (44-45)
```

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command) - research.md created with all decisions documented
- [x] Phase 1: Design complete (/plan command) - data-model.md, contracts/, quickstart.md, CLAUDE.md created
- [x] Phase 2: Task planning complete (/plan command - describe approach only) - ~45 tasks planned
- [x] Phase 3: Tasks generated (/tasks command) - tasks.md created with 58 numbered tasks
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS (infrastructure feature, principles adapted appropriately)
- [x] Post-Design Constitution Check: PASS (no violations, TDD approach maintained)
- [x] All NEEDS CLARIFICATION resolved (via /clarify command)
- [x] Complexity deviations documented (None required)

**Artifacts Generated**:
- [x] `/specs/001-create-a-local/research.md` (Phase 0)
- [x] `/specs/001-create-a-local/data-model.md` (Phase 1)
- [x] `/specs/001-create-a-local/contracts/docker-compose-schema.yml` (Phase 1)
- [x] `/specs/001-create-a-local/contracts/script-interfaces.md` (Phase 1)
- [x] `/specs/001-create-a-local/quickstart.md` (Phase 1)
- [x] `/CLAUDE.md` (Phase 1 - updated)
- [x] `/specs/001-create-a-local/tasks.md` (Phase 3)

**Next Step**: Begin implementation following tasks.md (58 tasks across 5 phases)

---
*Based on Constitution v2.0.0 - See `.specify/memory/constitution.md`*
