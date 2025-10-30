# Implementation Plan: LDE MCP Inspector

**Branch**: `002-lde-mcp-inspector` | **Date**: 2025-10-30 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-lde-mcp-inspector/spec.md`

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
Add the MCP Inspector service to the Local Development Environment (LDE) for testing and validating MCP server functionalities. The MCP Inspector provides a web-based interactive UI for developers to test MCP servers, explore resources/tools/prompts, and view real-time responses. This enhances the LDE by adding a developer tool for MCP debugging alongside the existing services (Hugr, PostgreSQL, Redis, MinIO, Keycloak).

## Technical Context
**Language/Version**: Bash 4.0+ (shell scripts), YAML (Docker Compose), Markdown (documentation)
**Primary Dependencies**: Docker Compose v2.0+, MCP Inspector Docker image from `ghcr.io/modelcontextprotocol/inspector:latest`
**Storage**: Docker volumes for Inspector test history persistence (`.local/mcp-inspector/`)
**Testing**: Shell script contract tests, Docker Compose validation tests
**Target Platform**: Docker-based local development environment (Linux, macOS)
**Project Type**: Infrastructure/DevOps - Docker Compose service orchestration
**Performance Goals**: Service startup <30s, health check response <5s
**Constraints**: Must use port 19007, must integrate with existing LDE lifecycle scripts
**Scale/Scope**: Single service addition to 5-service LDE stack, ~6 files modified (docker-compose.yml, .env.example, 2 READMEs, 3 shell scripts)

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Schema-First Development
- [x] N/A - This feature is infrastructure (Docker service), not MCP service schema development

### II. Tool-Based Interface
- [x] N/A - This feature is infrastructure support for MCP tool testing, not MCP tool implementation

### III. Test-First Development (NON-NEGOTIABLE)
- [ ] Contract tests for Docker Compose service configuration written before implementation
- [ ] Integration tests for LDE scripts with MCP Inspector written before implementation
- [ ] TDD cycle documented in Phase 2

### IV. Lazy Stepwise Discovery
- [x] N/A - This feature is infrastructure, not schema discovery implementation

### V. Observability & Error Clarity
- [ ] Shell script error handling with actionable messages planned
- [ ] Health check logging for MCP Inspector service planned
- [ ] Documentation includes troubleshooting section

**Complexity Justifications** (if any gates failed): None - feature follows constitutional principles where applicable

## Project Structure

### Documentation (this feature)
```
specs/002-lde-mcp-inspector/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
lde/
├── docker-compose.yml          # [MODIFY] Add mcp-inspector service definition
├── .env.example                # [MODIFY] Add MCP Inspector environment variables
├── README.md                   # [MODIFY] Add MCP Inspector service documentation
├── scripts/
│   ├── start.sh                # [MODIFY] Add MCP Inspector to startup sequence
│   ├── stop.sh                 # [MODIFY] Add MCP Inspector to shutdown sequence
│   ├── health-check.sh         # [MODIFY] Add MCP Inspector health check
│   ├── cleanup.sh              # [MODIFY] Add MCP Inspector data cleanup
│   └── load-data.sh            # [NO CHANGE] Data loading not affected
└── .local/
    └── mcp-inspector/          # [CREATE] Persistent volume for test history

README.md                       # [MODIFY] Add MCP Inspector reference in main README

tests/lde/
├── test-compose-validity.sh    # [MODIFY] Add MCP Inspector service validation
└── test-start-interface.sh     # [MODIFY] Add MCP Inspector in service count
```

**Structure Decision**: This is an infrastructure feature modifying the existing LDE Docker Compose orchestration. Changes are primarily configuration files (YAML, shell scripts, Markdown) with no application code. The structure follows the established LDE pattern of services defined in `docker-compose.yml` managed by lifecycle scripts in `lde/scripts/`.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - No NEEDS CLARIFICATION items remain after clarification session
   - Need to research MCP Inspector Docker image configuration
   - Need to research MCP Inspector volume/persistence requirements
   - Need to research LDE health check patterns for consistency

2. **Generate and dispatch research agents**:
   - Research MCP Inspector Docker deployment (image, ports, volumes, env vars)
   - Research existing LDE health check implementation patterns
   - Research existing LDE cleanup script patterns for volume management
   - Research MCP Inspector authentication/security configuration options

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all technical decisions documented

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - MCP Inspector Service entity (Docker Compose service definition)
   - MCP Inspector Configuration entity (environment variables)
   - Test History Data entity (persistent volume structure)
   - Service Health Status entity (health check response format)

2. **Generate API contracts** from functional requirements:
   - Docker Compose service schema (mcp-inspector service block)
   - Environment variable contract (.env.example entries)
   - Script interface contracts (start.sh, stop.sh, health-check.sh, cleanup.sh parameter/behavior contracts)
   - Documentation structure contract (README sections for MCP Inspector)
   - Output contracts to `/contracts/`

3. **Generate contract tests** from contracts:
   - `test-docker-compose-schema.sh`: Validate service definition structure
   - `test-env-variables.sh`: Verify all required env vars present
   - `test-script-interfaces.sh`: Verify scripts handle MCP Inspector correctly
   - `test-health-check-format.sh`: Validate health check response
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Story 1: Start LDE → MCP Inspector accessible on localhost:19007
   - Story 2: Health check → MCP Inspector shows as healthy/running
   - Story 3: Restart LDE → Test history persists
   - Story 4: Cleanup LDE → Test history removed
   - Story 5: Stop LDE → MCP Inspector stops cleanly

5. **Update CLAUDE.md incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh claude`
   - Add Docker Compose v2.0+, MCP Inspector deployment knowledge
   - Keep recent changes (this feature + previous LDE feature)
   - Keep under 150 lines for token efficiency

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, CLAUDE.md

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each contract → contract test task [P]
- Each modified file → implementation task with TDD cycle
- Documentation tasks for README files
- Integration test tasks for end-to-end scenarios

**Task Breakdown**:
1. Contract Tests (parallel [P]):
   - Write Docker Compose schema validation test
   - Write environment variable contract test
   - Write script interface contract tests
   - Write health check format test

2. Implementation (sequential, TDD):
   - Add mcp-inspector service to docker-compose.yml (make tests pass)
   - Add MCP Inspector env vars to .env.example (make tests pass)
   - Update start.sh to include MCP Inspector (make tests pass)
   - Update stop.sh to include MCP Inspector (make tests pass)
   - Update health-check.sh to check MCP Inspector (make tests pass)
   - Update cleanup.sh to remove MCP Inspector data (make tests pass)

3. Documentation (parallel [P]):
   - Update lde/README.md with MCP Inspector section
   - Update root README.md with MCP Inspector reference
   - Verify all documentation cross-references

4. Integration Tests:
   - Full LDE startup with MCP Inspector
   - Health check verification
   - Persistence verification (restart test)
   - Cleanup verification

**Ordering Strategy**:
- TDD order: Contract tests → Implementation → Integration tests
- Dependency order: docker-compose.yml first (defines service), then scripts (manage service), then docs (describe service)
- Mark [P] for parallel execution where files don't depend on each other

**Estimated Output**: 18-22 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following TDD principles)
**Phase 5**: Validation (run all tests, execute quickstart.md, verify integration)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

No violations - this feature is infrastructure support that follows constitutional principles where applicable (TDD, observability).

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [x] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented

**Artifacts Generated**:
- [x] research.md - Technical decisions and alternatives
- [x] data-model.md - Entity definitions and relationships
- [x] contracts/ - Docker Compose, environment, script contracts
  - [x] docker-compose-schema.yaml
  - [x] environment-variables.yaml
  - [x] script-interfaces.yaml
- [x] quickstart.md - User scenario validation steps
- [x] CLAUDE.md - Updated agent context (via update script)
- [x] tasks.md - 22 numbered tasks (8 parallel, 14 sequential)

---
*Based on Constitution v2.0.0 - See `.specify/memory/constitution.md`*
