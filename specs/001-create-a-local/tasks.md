# Tasks: Local Development Environment for Hugr MCP Service

**Input**: Design documents from `/Users/vgribanov/projects/hugr-lab/mcp/specs/001-create-a-local/`
**Prerequisites**: plan.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓, quickstart.md ✓

## Execution Flow (main)
```
1. ✓ Loaded plan.md - Shell/Docker stack, infrastructure project
2. ✓ Loaded research.md - 9 technical decisions documented
3. ✓ Loaded data-model.md - 5 services, 2 data sources, 4 scripts
4. ✓ Loaded contracts/ - docker-compose-schema.yml, script-interfaces.md
5. ✓ Loaded quickstart.md - 5-step verification workflow
6. Generated 47 tasks across 9 phases
7. Applied TDD ordering (tests before implementation)
8. Marked parallel tasks [P] (different files, no dependencies)
9. Validated completeness: All contracts tested ✓, All scripts implemented ✓
10. SUCCESS - Tasks ready for execution
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
Infrastructure project: `lde/` directory at repository root
- Docker configs: `lde/docker-compose.yml`, `lde/.env.example`
- Scripts: `lde/scripts/*.sh`
- Tests: `tests/lde/*.sh`
- Documentation: `lde/README.md`

---

## Phase 3.1: Setup & Prerequisites

- [ ] **T001** Create directory structure
  - Create `lde/` directory
  - Create `lde/scripts/` directory
  - Create `lde/keycloak/` directory
  - Create `lde/data/` directory with `.gitkeep`
  - Create `tests/lde/` directory
  - **Files**: Directory structure only
  - **Dependencies**: None (first task)

- [ ] **T002** Configure .gitignore
  - Add `lde/.local/` to `.gitignore`
  - Add `lde/data/*.duckdb` to `.gitignore`
  - Add `lde/.env` to `.gitignore` (keep `.env.example`)
  - **Files**: `.gitignore`
  - **Dependencies**: T001

- [ ] **T003** Create .env.example template
  - Create `lde/.env.example` with all environment variables from data-model.md
  - Include: Service images, database config, Redis, MinIO, Keycloak, SECRET_KEY, data paths
  - Document each variable with inline comments
  - **Files**: `lde/.env.example`
  - **Dependencies**: T001

---

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests - Scripts (All Parallel)

- [ ] **T004 [P]** Contract test: start.sh interface
  - Test file: `tests/lde/test-start-interface.sh`
  - Verify: --help flag, --reset flag, --no-data flag, --verbose flag
  - Verify: Exit codes (0=success, 1=prereq missing, 2=docker fail, 3=health fail, 4=data fail)
  - Verify: Output format (✓, ✗, → symbols, color when tty)
  - Mock: Docker commands to test without actual services
  - **Files**: `tests/lde/test-start-interface.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T004-T007

- [ ] **T005 [P]** Contract test: stop.sh interface
  - Test file: `tests/lde/test-stop-interface.sh`
  - Verify: --help flag, --verbose flag
  - Verify: Exit codes (0=success, 1=docker fail)
  - Verify: Data persistence confirmation message
  - Mock: Docker compose down command
  - **Files**: `tests/lde/test-stop-interface.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T004-T007

- [ ] **T006 [P]** Contract test: load-data.sh interface
  - Test file: `tests/lde/test-load-data-interface.sh`
  - Verify: --synthea-only, --openpayments-only, --force, --verbose, --help flags
  - Verify: Exit codes (0-5 per contract)
  - Verify: Progress indicators for long operations
  - Verify: Idempotency (skips if data exists unless --force)
  - Mock: curl, docker, duckdb commands
  - **Files**: `tests/lde/test-load-data-interface.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T004-T007

- [ ] **T007 [P]** Contract test: health-check.sh interface
  - Test file: `tests/lde/test-health-check-interface.sh`
  - Verify: --verbose flag, --wait flag with timeout, --help flag
  - Verify: Exit codes (0=healthy, 1=unhealthy, 2=timeout)
  - Verify: Per-service status output with response times
  - Verify: Failure mode outputs (connection details, suggestions)
  - Mock: Service health endpoints
  - **Files**: `tests/lde/test-health-check-interface.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T004-T007

### Contract Tests - Docker Compose (All Parallel)

- [ ] **T008 [P]** Contract test: docker-compose.yml YAML validity
  - Test file: `tests/lde/test-compose-validity.sh`
  - Verify: `docker compose config` succeeds (syntax valid)
  - Verify: YAML structure matches schema
  - Test with missing docker-compose.yml (should fail gracefully)
  - **Files**: `tests/lde/test-compose-validity.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T008-T015

- [ ] **T009 [P]** Contract test: All required services defined
  - Test file: `tests/lde/test-compose-services.sh`
  - Verify: `docker compose config --services` includes: hugr, postgres, redis, minio, keycloak
  - Verify: Exactly 5 services, no extras
  - **Files**: `tests/lde/test-compose-services.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T008-T015

- [ ] **T010 [P]** Contract test: Service dependencies configured
  - Test file: `tests/lde/test-compose-dependencies.sh`
  - Verify: Hugr depends_on includes postgres, redis, minio, keycloak
  - Verify: All dependencies use `condition: service_healthy`
  - Parse YAML to check dependency graph
  - **Files**: `tests/lde/test-compose-dependencies.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T008-T015

- [ ] **T011 [P]** Contract test: Healthchecks configured
  - Test file: `tests/lde/test-compose-healthchecks.sh`
  - Verify: All 5 services have healthcheck.test defined
  - Verify: Interval, timeout, retries set for each service
  - Verify: Postgres: `pg_isready`, Redis: `redis-cli ping`, MinIO: curl, Keycloak: curl, Hugr: curl
  - **Files**: `tests/lde/test-compose-healthchecks.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T008-T015

- [ ] **T012 [P]** Contract test: Volume persistence pattern
  - Test file: `tests/lde/test-compose-volumes.sh`
  - Verify: All data volumes mount to `.local/` subdirectories
  - Verify: postgres → `.local/pg-data`, redis → `.local/redis-data`, minio → `.local/minio`, keycloak → `.local/keycloak`
  - Verify: Keycloak realm config volume: `./keycloak/realm-config.json:/opt/keycloak/data/import/realm.json`
  - **Files**: `tests/lde/test-compose-volumes.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T008-T015

- [ ] **T013 [P]** Contract test: Port exposures
  - Test file: `tests/lde/test-compose-ports.sh`
  - Verify: Hugr 8080:8080, Keycloak 8180:8080, MinIO 9000:9000 and 9001:9001
  - Verify: Postgres 5432:5432, Redis 6379:6379 (optional but expected)
  - **Files**: `tests/lde/test-compose-ports.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T008-T015

- [ ] **T014 [P]** Contract test: Environment variables
  - Test file: `tests/lde/test-compose-env.sh`
  - Verify: Hugr has DATABASE_URL, REDIS_URL, S3_ENDPOINT, KEYCLOAK_URL, KEYCLOAK_REALM
  - Verify: Postgres has POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB
  - Verify: MinIO has MINIO_ROOT_USER, MINIO_ROOT_PASSWORD
  - Verify: Keycloak has KEYCLOAK_ADMIN, KEYCLOAK_ADMIN_PASSWORD, KC_HTTP_PORT
  - **Files**: `tests/lde/test-compose-env.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T008-T015

- [ ] **T015 [P]** Contract test: Keycloak import configuration
  - Test file: `tests/lde/test-compose-keycloak.sh`
  - Verify: Keycloak command includes `--import-realm`
  - Verify: Realm file volume mounted correctly
  - **Files**: `tests/lde/test-compose-keycloak.sh`
  - **Dependencies**: T001
  - **Parallel Group**: T008-T015

---

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Docker Configuration

- [ ] **T016** Create docker-compose.yml with all 5 services
  - File: `lde/docker-compose.yml`
  - Define services: hugr, postgres, redis, minio, keycloak
  - Configure networks for inter-service communication
  - Use environment variables from .env file
  - **Files**: `lde/docker-compose.yml`
  - **Dependencies**: T003, T008-T015 (tests must fail first)
  - **References**: data-model.md lines 30-135 for service definitions

- [ ] **T017** Configure PostgreSQL service
  - Add to `lde/docker-compose.yml`
  - Image: `postgres:15-alpine` (or from .env)
  - Environment: POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB
  - Volume: `.local/pg-data:/var/lib/postgresql/data`
  - Healthcheck: `pg_isready -U hugr`
  - **Files**: `lde/docker-compose.yml`
  - **Dependencies**: T016

- [ ] **T018** Configure Redis service
  - Add to `lde/docker-compose.yml`
  - Image: `redis:7-alpine` (or from .env)
  - Command: `redis-server --appendonly yes`
  - Volume: `.local/redis-data:/data`
  - Healthcheck: `redis-cli ping`
  - **Files**: `lde/docker-compose.yml`
  - **Dependencies**: T016

- [ ] **T019** Configure MinIO service
  - Add to `lde/docker-compose.yml`
  - Image: `minio/minio:latest` (or from .env)
  - Command: `server /data --console-address ":9001"`
  - Environment: MINIO_ROOT_USER, MINIO_ROOT_PASSWORD
  - Ports: 9000:9000, 9001:9001
  - Volume: `.local/minio:/data`
  - Healthcheck: `curl -f http://localhost:9000/minio/health/live`
  - **Files**: `lde/docker-compose.yml`
  - **Dependencies**: T016

- [ ] **T020** Configure Keycloak service
  - Add to `lde/docker-compose.yml`
  - Image: `quay.io/keycloak/keycloak:latest` (or from .env)
  - Command: `start-dev --import-realm`
  - Environment: KEYCLOAK_ADMIN, KEYCLOAK_ADMIN_PASSWORD, KC_HTTP_PORT=8080
  - Ports: 8180:8080
  - Volumes: realm-config.json, `.local/keycloak:/opt/keycloak/data/h2`
  - Healthcheck: `curl -f http://localhost:8080/health/ready`
  - **Files**: `lde/docker-compose.yml`
  - **Dependencies**: T016

- [ ] **T021** Configure Hugr service
  - Add to `lde/docker-compose.yml`
  - Image: `hugr/hugr:latest` (or from .env)
  - Environment: DATABASE_URL, REDIS_URL, S3_ENDPOINT, KEYCLOAK_URL, KEYCLOAK_REALM, SECRET_KEY
  - Ports: 8080:8080
  - depends_on: postgres, redis, minio, keycloak (all with condition: service_healthy)
  - Healthcheck: `curl -f http://localhost:8080/healthz`
  - **Files**: `lde/docker-compose.yml`
  - **Dependencies**: T016, T017, T018, T019, T020

- [ ] **T022** Extract Keycloak realm configuration
  - Source: `/Users/vgribanov/projects/hugr-lab/examples` Keycloak configuration
  - Create: `lde/keycloak/realm-config.json`
  - Include: Realm "hugr", 3 roles (admin, viewer, analyst), 3 test users, hugr-graphql client
  - **Files**: `lde/keycloak/realm-config.json`
  - **Dependencies**: T001
  - **References**: data-model.md lines 311-370, research.md section 4

### Script Implementation

- [ ] **T023** Implement start.sh - Prerequisites check
  - File: `lde/scripts/start.sh`
  - Shebang: `#!/usr/bin/env bash`
  - Strict mode: `set -euo pipefail`
  - Parse flags: --reset, --no-data, --verbose, --help
  - Check: docker, docker-compose, curl, jq installed
  - Check: Examples repo exists at path from .env
  - Check: Synthea repo exists at path from .env
  - Display ✓ for each check
  - Exit code 1 if any prerequisite missing
  - **Files**: `lde/scripts/start.sh`
  - **Dependencies**: T004 (test must fail first), T016-T022 (compose file exists)

- [ ] **T024** Implement start.sh - Directory initialization
  - Add to `lde/scripts/start.sh`
  - If --reset flag: remove `.local/` directory entirely
  - Create `.local/` subdirectories: pg-data, redis-data, minio, keycloak
  - Display "→ Creating .local directory structure..."
  - Display "✓ Created .local/{subdir}" for each
  - **Files**: `lde/scripts/start.sh`
  - **Dependencies**: T023

- [ ] **T025** Implement start.sh - Service startup
  - Add to `lde/scripts/start.sh`
  - Run `docker compose -f lde/docker-compose.yml up -d`
  - Display "→ Starting Docker services..."
  - Capture output and show service status
  - Exit code 2 if docker compose fails
  - **Files**: `lde/scripts/start.sh`
  - **Dependencies**: T024

- [ ] **T026** Implement start.sh - Health check wait
  - Add to `lde/scripts/start.sh`
  - Call `lde/scripts/health-check.sh --wait 120`
  - Display "→ Verifying service health..."
  - Exit code 3 if health checks fail or timeout
  - Display per-service status with response times
  - **Files**: `lde/scripts/start.sh`
  - **Dependencies**: T025, T030 (health-check.sh exists)

- [ ] **T027** Implement start.sh - Data loading
  - Add to `lde/scripts/start.sh`
  - Skip if --no-data flag set
  - Call `lde/scripts/load-data.sh`
  - Display "→ Loading data sources..."
  - Exit code 4 if data loading fails
  - Display data loading summary
  - **Files**: `lde/scripts/start.sh`
  - **Dependencies**: T026, T031 (load-data.sh exists)

- [ ] **T028** Implement start.sh - Connection information
  - Add to `lde/scripts/start.sh`
  - Display formatted output with:
    - Hugr GraphQL URL (http://localhost:8080/graphql)
    - Keycloak URL (http://localhost:8180)
    - MinIO Console URL (http://localhost:9001)
    - Test credentials for 3 users
    - Data sources loaded (if data loading succeeded)
    - Command to stop: `./lde/scripts/stop.sh`
  - Use box drawing characters for visual appeal
  - **Files**: `lde/scripts/start.sh`
  - **Dependencies**: T027

- [ ] **T029** Implement stop.sh
  - File: `lde/scripts/stop.sh`
  - Shebang, strict mode, parse flags (--verbose, --help)
  - Run `docker compose -f lde/docker-compose.yml down` (preserves volumes)
  - Display "→ Stopping Docker services..."
  - Display service stop status
  - Display "✓ Data persisted in lde/.local/"
  - Display restart command
  - Exit code 0 on success, 1 on docker failure
  - **Files**: `lde/scripts/stop.sh`
  - **Dependencies**: T005 (test must fail first), T016-T022

- [ ] **T030** Implement health-check.sh
  - File: `lde/scripts/health-check.sh`
  - Shebang, strict mode, parse flags (--verbose, --wait TIMEOUT, --help)
  - Check each service health endpoint:
    - Postgres: `pg_isready -h localhost -U hugr`
    - Redis: `redis-cli -h localhost ping`
    - MinIO: `curl -f http://localhost:9000/minio/health/live`
    - Keycloak: `curl -f http://localhost:8180/health/ready`
    - Hugr: `curl -f http://localhost:8080/healthz`
  - Display per-service status (✓ Healthy or ✗ Unhealthy) with response time
  - If --wait: loop with timeout, checking every 5 seconds
  - Exit code 0 (all healthy), 1 (some unhealthy), 2 (timeout)
  - Display actionable error messages with docker logs command suggestions
  - **Files**: `lde/scripts/health-check.sh`
  - **Dependencies**: T007 (test must fail first), T016-T022

- [ ] **T031** Implement load-data.sh - Authentication setup
  - File: `lde/scripts/load-data.sh`
  - Shebang, strict mode, parse flags (--synthea-only, --openpayments-only, --force, --verbose, --help)
  - Load SECRET_KEY from .env file
  - Exit code 1 if SECRET_KEY not found
  - Display "→ Checking prerequisites..."
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T006 (test must fail first), T003 (.env.example exists)

- [ ] **T032** Implement load-data.sh - Hugr health verification
  - Add to `lde/scripts/load-data.sh`
  - Call `lde/scripts/health-check.sh` to verify Hugr is healthy
  - Exit code 2 if Hugr not healthy
  - Display "✓ Hugr is healthy"
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T031, T030

- [ ] **T033** Implement load-data.sh - Synthea generation
  - Add to `lde/scripts/load-data.sh`
  - Skip if --openpayments-only flag
  - Check if `lde/data/synthea.duckdb` exists and skip unless --force
  - Use Docker to run Synthea: `docker run --rm -v $(pwd)/lde/data:/output synthetichealth/synthea:latest -p 10000`
  - Display progress: "→ Generating Synthea data (10,000 patients)..."
  - Display progress bar or periodic updates
  - Exit code 3 if generation fails
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T032

- [ ] **T034** Implement load-data.sh - DuckDB conversion
  - Add to `lde/scripts/load-data.sh`
  - Convert Synthea CSV files to DuckDB database
  - Create `lde/data/synthea.duckdb`
  - Import all 9 Synthea tables: patients, encounters, conditions, medications, observations, procedures, immunizations, allergies, careplans
  - Display "→ Converting to DuckDB format..."
  - Display "✓ Created synthea.duckdb (size in MB)"
  - Exit code 3 if conversion fails
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T033

- [ ] **T035** Implement load-data.sh - Synthea data source registration
  - Add to `lde/scripts/load-data.sh`
  - GraphQL mutation: `core { insert_data_sources(data: {...}) }`
  - Use curl with `x-hugr-secret: ${SECRET_KEY}` header
  - Parameters: name="synthea", type="duckdb", prefix="syn", path="/data/synthea.duckdb", read_only=true, as_module=true
  - Display "→ Registering data sources with Hugr..."
  - Display "✓ Created synthea data source"
  - Exit code 4 if registration fails
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T034
  - **References**: research.md lines 209-227 for exact mutation

- [ ] **T036** Implement load-data.sh - Synthea data source loading
  - Add to `lde/scripts/load-data.sh`
  - GraphQL mutation: `function { core { load_data_sources(name: "synthea") { success message } } }`
  - Use curl with `x-hugr-secret: ${SECRET_KEY}` header
  - Enable self-defined schema (optional): update data source with self_defined=true, then reload
  - Display "✓ Synthea data loaded"
  - Exit code 4 if loading fails
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T035
  - **References**: research.md lines 219-244 for exact mutations

- [ ] **T037** Implement load-data.sh - Open Payments processing
  - Add to `lde/scripts/load-data.sh`
  - Skip if --synthea-only flag
  - Check for existing `lde/data/openpayments.duckdb` and skip unless --force
  - Reference examples project setup script at `/Users/vgribanov/projects/hugr-lab/examples/examples/open-payments/setup.sh`
  - Download Open Payments 2023 data (or use cached)
  - Convert to DuckDB format: create tables for general_payments, research_payments, ownership_information
  - Display "→ Loading Open Payments 2023 data..."
  - Display progress bar
  - Display "✓ Converted to DuckDB (size in GB)"
  - Exit code 3 if processing fails
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T032

- [ ] **T038** Implement load-data.sh - Open Payments data source registration
  - Add to `lde/scripts/load-data.sh`
  - GraphQL mutation: `core { insert_data_sources(data: {...}) }`
  - Parameters: name="op2023", type="duckdb", prefix="op2023", path="/data/openpayments.duckdb", read_only=true, as_module=true
  - GraphQL mutation: `function { core { load_data_sources(name: "op2023") { success message } } }`
  - Display "✓ Open Payments data loaded"
  - Exit code 4 if registration/loading fails
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T037
  - **References**: research.md lines 247-263 for exact mutations

- [ ] **T039** Implement load-data.sh - Verification queries
  - Add to `lde/scripts/load-data.sh`
  - Query Synthea: verify patient count = 10000
    - GraphQL: `{ synthea { patients_aggregation { count: _rows_count } } }`
  - Query Open Payments: verify records exist, check year range
    - GraphQL: `{ op2023 { general_payments_aggregation { count: _rows_count } } }`
  - Display "→ Verifying data sources..."
  - Display "✓ Patient count: 10,000" and "✓ Open Payments records: ~11.8M"
  - Exit code 5 if verification fails
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T036, T038

- [ ] **T040** Implement load-data.sh - Summary output
  - Add to `lde/scripts/load-data.sh`
  - Display formatted summary box:
    - Data Sources: 2
    - Total Records: ~11.9M
    - Disk Usage: calculate from .duckdb file sizes
    - Load Time: track from start of script
  - Use box drawing characters
  - **Files**: `lde/scripts/load-data.sh`
  - **Dependencies**: T039

- [ ] **T041** Add executable permissions to all scripts
  - Run: `chmod +x lde/scripts/*.sh`
  - **Files**: All scripts in lde/scripts/
  - **Dependencies**: T023-T040

---

## Phase 3.4: Integration Tests

### Service Health Tests (All Parallel)

- [ ] **T042 [P]** Integration test: PostgreSQL health and connectivity
  - Test file: `tests/lde/test-postgres-integration.sh`
  - Start environment with `lde/scripts/start.sh --no-data`
  - Test: `pg_isready -h localhost -U hugr`
  - Test: Connect with psql and execute simple query
  - Verify: Database "hugr" exists
  - Stop environment with `lde/scripts/stop.sh`
  - **Files**: `tests/lde/test-postgres-integration.sh`
  - **Dependencies**: T023-T029, T041
  - **Parallel Group**: T042-T046

- [ ] **T043 [P]** Integration test: Redis health and connectivity
  - Test file: `tests/lde/test-redis-integration.sh`
  - Start environment with `lde/scripts/start.sh --no-data`
  - Test: `redis-cli -h localhost ping`
  - Test: SET/GET key-value pair
  - Verify: Appendonly mode enabled
  - Stop environment
  - **Files**: `tests/lde/test-redis-integration.sh`
  - **Dependencies**: T023-T029, T041
  - **Parallel Group**: T042-T046

- [ ] **T044 [P]** Integration test: MinIO health and S3 operations
  - Test file: `tests/lde/test-minio-integration.sh`
  - Start environment with `lde/scripts/start.sh --no-data`
  - Test: Health endpoint `curl http://localhost:9000/minio/health/live`
  - Test: Create bucket, upload file, download file using mc or curl
  - Verify: Console accessible at http://localhost:9001
  - Stop environment
  - **Files**: `tests/lde/test-minio-integration.sh`
  - **Dependencies**: T023-T029, T041
  - **Parallel Group**: T042-T046

- [ ] **T045 [P]** Integration test: Keycloak health and realm loading
  - Test file: `tests/lde/test-keycloak-integration.sh`
  - Start environment with `lde/scripts/start.sh --no-data`
  - Test: Health endpoint `curl http://localhost:8180/health/ready`
  - Test: Realm "hugr" exists: `curl http://localhost:8180/realms/hugr`
  - Test: Obtain token for admin@example.com
  - Verify: Token contains expected roles
  - Stop environment
  - **Files**: `tests/lde/test-keycloak-integration.sh`
  - **Dependencies**: T023-T029, T041
  - **Parallel Group**: T042-T046

- [ ] **T046 [P]** Integration test: Hugr GraphQL endpoint and introspection
  - Test file: `tests/lde/test-hugr-integration.sh`
  - Start environment with `lde/scripts/start.sh --no-data`
  - Test: Health endpoint `curl http://localhost:8080/healthz`
  - Test: GraphQL introspection query `{ __schema { queryType { name } } }`
  - Verify: Response contains valid GraphQL schema
  - Verify: Can query with SECRET_KEY header
  - Stop environment
  - **Files**: `tests/lde/test-hugr-integration.sh`
  - **Dependencies**: T023-T029, T041
  - **Parallel Group**: T042-T046

### Data Loading Tests (Sequential - environment state dependent)

- [ ] **T047** Integration test: Synthea data generation
  - Test file: `tests/lde/test-synthea-generation.sh`
  - Start environment with `lde/scripts/start.sh --no-data`
  - Run: `lde/scripts/load-data.sh --synthea-only`
  - Verify: `lde/data/synthea.duckdb` file created
  - Verify: File size reasonable (100-500MB)
  - Verify: All 9 tables present in DuckDB
  - **Files**: `tests/lde/test-synthea-generation.sh`
  - **Dependencies**: T031-T036, T041

- [ ] **T048** Integration test: Synthea data source registration
  - Test file: `tests/lde/test-synthea-registration.sh`
  - Continue from T047 (environment still running)
  - Query Hugr: List data sources
  - Verify: "synthea" data source exists
  - Verify: Type is "duckdb", as_module is true
  - **Files**: `tests/lde/test-synthea-registration.sh`
  - **Dependencies**: T047

- [ ] **T049** Integration test: Synthea data verification
  - Test file: `tests/lde/test-synthea-verification.sh`
  - Continue from T048
  - Query: `{ synthea { patients_aggregation { count: _rows_count } } }`
  - Verify: Count equals 10000
  - Query sample patient data
  - Query: List tables in synthea module
  - Verify: 9 tables accessible
  - Stop environment
  - **Files**: `tests/lde/test-synthea-verification.sh`
  - **Dependencies**: T048

- [ ] **T050** Integration test: Open Payments data processing
  - Test file: `tests/lde/test-openpayments-processing.sh`
  - Start environment with `lde/scripts/start.sh --no-data`
  - Run: `lde/scripts/load-data.sh --openpayments-only`
  - Verify: `lde/data/openpayments.duckdb` file created
  - Verify: File size reasonable (1-3GB)
  - Verify: 3 tables present: general_payments, research_payments, ownership_information
  - **Files**: `tests/lde/test-openpayments-processing.sh`
  - **Dependencies**: T037-T038, T041

- [ ] **T051** Integration test: Open Payments data verification
  - Test file: `tests/lde/test-openpayments-verification.sh`
  - Continue from T050
  - Query: `{ op2023 { general_payments_aggregation { count: _rows_count } } }`
  - Verify: Count > 10 million records
  - Query sample payment data
  - Query: Aggregate by state
  - Stop environment
  - **Files**: `tests/lde/test-openpayments-verification.sh`
  - **Dependencies**: T050

### Full Workflow Tests (Sequential)

- [ ] **T052** Integration test: Full start sequence (ABSENT → HEALTHY)
  - Test file: `tests/lde/test-full-startup.sh`
  - Ensure `.local/` does not exist (clean state)
  - Run: `lde/scripts/start.sh` (with data loading)
  - Verify: All services healthy
  - Verify: Both data sources loaded
  - Verify: Can query both Synthea and Open Payments
  - **Files**: `tests/lde/test-full-startup.sh`
  - **Dependencies**: T023-T040, T041

- [ ] **T053** Integration test: Stop and restart with data persistence
  - Test file: `tests/lde/test-restart-persistence.sh`
  - Continue from T052 (environment running with data)
  - Query and save sample data (patient ID, payment ID)
  - Run: `lde/scripts/stop.sh`
  - Verify: `.local/` directory still exists
  - Run: `lde/scripts/start.sh` (should skip data loading)
  - Verify: Startup faster (< 1 minute)
  - Query same sample data - verify it's identical
  - **Files**: `tests/lde/test-restart-persistence.sh`
  - **Dependencies**: T052

- [ ] **T054** Integration test: Reset flag (wipe and fresh start)
  - Test file: `tests/lde/test-reset-flag.sh`
  - Continue from T053 (environment running)
  - Create marker file: `touch lde/.local/test-marker`
  - Run: `lde/scripts/stop.sh`
  - Run: `lde/scripts/start.sh --reset --no-data`
  - Verify: Marker file does not exist (`.local/` was wiped)
  - Verify: All services healthy
  - Verify: No data sources (data loading skipped)
  - Stop environment
  - **Files**: `tests/lde/test-reset-flag.sh`
  - **Dependencies**: T053

- [ ] **T055** Integration test: Authentication flow (SECRET_KEY)
  - Test file: `tests/lde/test-auth-secret-key.sh`
  - Start environment with `lde/scripts/start.sh`
  - Load SECRET_KEY from .env
  - Test: Query with `x-hugr-secret` header - should succeed
  - Test: Query without header - should fail (401)
  - Test: Query with wrong secret - should fail (403)
  - Test: Mutation (data source query) with SECRET_KEY - should succeed
  - **Files**: `tests/lde/test-auth-secret-key.sh`
  - **Dependencies**: T052

- [ ] **T056** Integration test: Authentication flow (OIDC tokens)
  - Test file: `tests/lde/test-auth-oidc.sh`
  - Continue from T055 (environment running)
  - Obtain token for admin@example.com from Keycloak
  - Test: Query with Bearer token - should succeed
  - Obtain token for viewer@example.com
  - Test: Query with viewer token - verify role-based filtering
  - Obtain token for analyst@example.com
  - Test: Query with analyst token - verify role-based access
  - Stop environment
  - **Files**: `tests/lde/test-auth-oidc.sh`
  - **Dependencies**: T055

---

## Phase 3.5: Documentation & Polish

- [ ] **T057 [P]** Create README.md in lde/ directory
  - File: `lde/README.md`
  - Sections:
    - Overview of local development environment
    - Prerequisites (Docker, Docker Compose, curl, jq)
    - Quick Start (reference quickstart.md)
    - Scripts documentation (start.sh, stop.sh, load-data.sh, health-check.sh)
    - Port mappings table
    - Troubleshooting common issues
    - Links to quickstart.md and data-model.md
  - **Files**: `lde/README.md`
  - **Dependencies**: T001-T056 completed
  - **Parallel Group**: T057-T058

- [ ] **T058 [P]** Document environment variables and ports
  - Add to `lde/README.md`
  - Table: All environment variables from .env.example with descriptions
  - Table: Port mappings (host:container) for all services
  - Table: Default credentials for Keycloak users
  - Security note: Change SECRET_KEY and passwords for non-local use
  - **Files**: `lde/README.md`
  - **Dependencies**: T057
  - **Parallel Group**: T057-T058

---

## Dependencies Summary

```
Setup (T001-T003)
  ↓
Contract Tests (T004-T015) [All parallel - different test files]
  ↓
Docker Implementation (T016-T022)
  ├─ T016 (base compose file)
  ├─ T017-T021 (services, depend on T016)
  └─ T022 (Keycloak config, parallel with compose)
  ↓
Script Implementation (T023-T041)
  ├─ T023-T028 (start.sh, sequential dependencies)
  ├─ T029 (stop.sh, parallel with start after T022)
  ├─ T030 (health-check.sh, needed by start.sh)
  ├─ T031-T040 (load-data.sh, sequential dependencies)
  └─ T041 (permissions, after all scripts)
  ↓
Integration Tests (T042-T056)
  ├─ Service Health (T042-T046) [Parallel - different test files]
  ├─ Data Loading (T047-T051) [Sequential - shared environment state]
  └─ Full Workflow (T052-T056) [Sequential - shared environment state]
  ↓
Documentation (T057-T058) [Parallel - different sections]
```

## Parallel Execution Examples

### Contract Tests (Phase 3.2)
All 12 contract tests can run in parallel:
```bash
# Launch all script interface tests together:
./tests/lde/test-start-interface.sh &
./tests/lde/test-stop-interface.sh &
./tests/lde/test-load-data-interface.sh &
./tests/lde/test-health-check-interface.sh &

# Launch all docker compose tests together:
./tests/lde/test-compose-validity.sh &
./tests/lde/test-compose-services.sh &
./tests/lde/test-compose-dependencies.sh &
./tests/lde/test-compose-healthchecks.sh &
./tests/lde/test-compose-volumes.sh &
./tests/lde/test-compose-ports.sh &
./tests/lde/test-compose-env.sh &
./tests/lde/test-compose-keycloak.sh &
wait
```

### Service Health Tests (Phase 3.4)
5 service health tests can run in parallel:
```bash
./tests/lde/test-postgres-integration.sh &
./tests/lde/test-redis-integration.sh &
./tests/lde/test-minio-integration.sh &
./tests/lde/test-keycloak-integration.sh &
./tests/lde/test-hugr-integration.sh &
wait
```

## Validation Checklist

- [x] All contracts have corresponding tests (T004-T015 for docker-compose + scripts)
- [x] All entities have implementation tasks (5 services: T017-T021, 4 scripts: T023-T040)
- [x] All tests come before implementation (Phase 3.2 before Phase 3.3)
- [x] Parallel tasks truly independent (marked [P], different files)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] All quickstart scenarios have integration tests (T042-T056)
- [x] TDD workflow maintained throughout

## Notes

- **[P] tasks**: Different files, no dependencies, can run simultaneously
- **Verify tests fail**: Before implementing, run contract tests - they must fail
- **Commit after each task**: Git commit after completing each task for clean history
- **Environment state**: Integration tests T047-T056 share environment state, must run sequentially
- **References**:
  - GraphQL mutations: research.md section 7
  - Service configs: data-model.md lines 30-370
  - Keycloak realm: data-model.md lines 311-370
  - Port mappings: research.md section 9
  - Authentication: research.md section 7 (SECRET_KEY approach)

---

**Total Tasks**: 58
**Parallel Tasks**: 19 (marked [P])
**Sequential Tasks**: 39
**Estimated Completion Time**:
- Phase 3.1 (Setup): 30 minutes
- Phase 3.2 (Contract Tests): 2-3 hours (parallel execution)
- Phase 3.3 (Implementation): 6-8 hours
- Phase 3.4 (Integration Tests): 4-6 hours
- Phase 3.5 (Documentation): 1-2 hours
**Total**: 14-20 hours of focused development

**Ready for execution**: Each task is specific, includes file paths, and can be completed independently by an LLM or developer.