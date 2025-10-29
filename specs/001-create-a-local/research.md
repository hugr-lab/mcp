# Research: Local Development Environment for Hugr MCP

**Feature**: 001-create-a-local
**Date**: 2025-10-27
**Status**: Complete

## Overview
This document consolidates research findings for implementing a local development environment (LDE) for the Hugr MCP service using Docker Compose.

## Key Research Areas

### 1. Hugr Examples Project Configuration

**Decision**: Use docker-compose.yml and service configurations from `/Users/vgribanov/projects/hugr-lab/examples` as the foundation

**Rationale**:
- Proven working configuration for Hugr ecosystem
- Already includes all required services (Hugr, Keycloak, Postgres, Redis, MinIO)
- Service versions are tested and compatible
- Keycloak realm configuration available with admin/viewer/analyst roles
- Consistent with existing team practices

**Implementation Approach**:
1. Reference `examples/docker-compose.yml` for service definitions
2. Extract Keycloak realm configuration from examples
3. Mirror volume structure (`.local/` directory pattern)
4. Use same service versions and network configuration
5. Adapt for MCP-specific needs (port exposure, environment variables)

**Alternatives Considered**:
- **From scratch configuration**: Rejected due to risk of version mismatches and missing configurations
- **Official Hugr Docker images**: Rejected as examples project provides complete tested setup

---

### 2. Synthea Data Generation in DuckDB Format

**Decision**: Use Docker-based Synthea generation with DuckDB export pipeline

**Rationale**:
- Synthea project at `/Users/vgribanov/projects/synthea` provides Docker support
- DuckDB format efficient for local development (single file, no server required)
- 10,000 patients balances data realism with generation time (~5-10 minutes)
- Docker isolation avoids system Java dependencies

**Implementation Approach**:
1. Create script that uses Synthea Docker image: `docker run --rm -v $(pwd)/data:/output synthetichealth/synthea:latest -p 10000`
2. Convert CSV output to DuckDB using DuckDB CLI or Python duckdb library
3. Store resulting DuckDB file in `lde/data/synthea.duckdb`
4. Register with Hugr as data source via GraphQL mutation

**Data Generation Parameters**:
- Population: 10,000 patients
- State: Default (Massachusetts for consistency with Synthea defaults)
- Seed: Optional fixed seed for reproducibility
- Modules: All default (encounters, medications, conditions, observations, etc.)

**Alternatives Considered**:
- **Pre-generated dataset**: Rejected to ensure fresh, customizable data
- **PostgreSQL output**: Rejected as DuckDB simpler for local development
- **Manual data loading**: Rejected as too time-consuming and error-prone

---

### 3. Open Payments Data Integration

**Decision**: Use Open Payments 2023 data processing approach from hugr examples project

**Rationale**:
- Examples project has proven data loading scripts
- Open Payments 2023 is latest complete year (2024 data may be incomplete)
- Existing transformation/loading logic handles data quality issues
- CSV/Parquet formats supported by Hugr

**Implementation Approach**:
1. Reference examples project data loading scripts at `/Users/vgribanov/projects/hugr-lab/examples`
2. Adapt scripts for LDE directory structure
3. Download Open Payments 2023 data (or use cached copy if available)
4. Transform and load into Hugr following examples pattern
5. Create Hugr data source via GraphQL mutation

**Data Source Details**:
- Dataset: CMS Open Payments General Payments 2023
- Format: CSV (converted to Parquet for efficiency if needed)
- Size: ~10GB raw, varies after filtering
- Fields: Payment information, physician details, manufacturer data

**Alternatives Considered**:
- **Smaller subset**: Rejected as full 2023 data better represents production scale
- **Different year**: Rejected as 2023 is most recent complete dataset
- **Sample data**: Rejected as examples project already provides production approach

---

### 4. Keycloak Realm Configuration

**Decision**: Export and adapt realm configuration from hugr examples project with admin/viewer/analyst roles

**Rationale**:
- Examples project has pre-configured realm matching requirements
- Avoids manual UI configuration steps
- Ensures consistency across developer machines
- Version-controlled realm configuration

**Implementation Approach**:
1. Export realm JSON from examples Keycloak instance
2. Store in `lde/keycloak/realm-config.json`
3. Configure docker-compose to import realm on startup
4. Environment variable for admin password (`.env` file)

**Realm Configuration**:
- Realm name: `hugr` (or as defined in examples)
- Client: Hugr GraphQL API client
- Users: Pre-created test users for each role
  - admin@example.com (admin role)
  - viewer@example.com (viewer role)
  - analyst@example.com (analyst role)
- Credentials: Default passwords in documentation

**Alternatives Considered**:
- **Manual realm setup**: Rejected as not reproducible
- **Simplified auth**: Rejected as doesn't match production auth model
- **Single admin user**: Rejected as role-based testing requires multiple roles

---

### 5. Volume Persistence Strategy

**Decision**: Use `.local/` directory with service-specific subdirectories, matching hugr examples pattern

**Rationale**:
- Consistent with existing hugr examples structure
- Git-ignored by default to avoid committing data
- Easy to reset (delete `.local/` directory)
- Clear organization by service

**Volume Structure**:
```
lde/.local/
├── pg-data/       # PostgreSQL data
├── minio/         # MinIO object storage
├── keycloak/      # Keycloak database
└── redis-data/    # Redis persistence
```

**Volume Configuration** (docker-compose.yml):
```yaml
volumes:
  postgres-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: ${PWD}/.local/pg-data
  # Similar for other services
```

**Alternatives Considered**:
- **Docker named volumes**: Rejected as harder to locate and reset
- **System-wide volumes**: Rejected to keep environment self-contained
- **No persistence**: Rejected as data reload after every restart too slow

---

### 6. Service Health Checking

**Decision**: Implement multi-level health checking: Docker health checks + application-level readiness

**Rationale**:
- Docker healthchecks ensure container is running
- Application-level checks verify service is ready for requests
- Progressive health checking provides clear feedback

**Implementation Approach**:
1. **Docker Health Checks** (in docker-compose.yml):
   - Postgres: `pg_isready -U hugr`
   - Redis: `redis-cli ping`
   - MinIO: `curl -f http://localhost:9000/minio/health/live`
   - Keycloak: `curl -f http://localhost:8080/health/ready`
   - Hugr: `curl -f http://localhost:8080/healthz` (adjust to actual endpoint)

2. **Script-based Health Check** (`health-check.sh`):
   - Wait for all Docker health checks to pass
   - Test GraphQL endpoint with introspection query
   - Verify Keycloak authentication
   - Test S3 (MinIO) connectivity
   - Verify database connectivity via Hugr

**Alternatives Considered**:
- **Basic port checking**: Rejected as insufficient (port open != service ready)
- **No health checking**: Rejected as leads to flaky data loading
- **wait-for-it script**: Considered but custom script more appropriate for multi-service validation

---

### 7. GraphQL Mutation Approach for Data Source Registration

**Decision**: Use curl-based GraphQL mutations following the exact pattern from hugr examples project

**Rationale**:
- No additional dependencies beyond curl + jq
- Direct GraphQL interaction, no wrapper libraries needed
- Easy to debug and modify
- Consistent with examples project approach
- Proven mutation structure from open-payments example

**Mutation Pattern** (based on hugr examples):

**Step 1: Create Data Source**
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{
    "query": "mutation addSyntheaDataSource { core { insert_data_sources(data: { name: \"synthea\", description: \"Synthea synthetic patient data\", type: \"duckdb\", prefix: \"syn\", path: \"/data/synthea.duckdb\", read_only: true, as_module: true }) { name type description as_module path self_defined prefix read_only disabled } } }"
  }'
```

**Step 2: Load Data Source**
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{
    "query": "mutation loadSyntheaDataSource { function { core { load_data_sources(name: \"synthea\") { success message } } } }"
  }'
```

**Step 3: Enable Self-Described Schema (Optional)**
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{
    "query": "mutation updateSyntheaDataSource { core { update_data_sources(filter: { name: { eq: \"synthea\" } }, data: { self_defined: true }) { name description type prefix path read_only as_module disabled } } }"
  }'

# Then reload
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{
    "query": "mutation reloadSyntheaDataSource { function { core { load_data_sources(name: \"synthea\") { success message } } } }"
  }'
```

**Open Payments Data Source** (following examples pattern):
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{
    "query": "mutation addOpenPaymentsDataSource { core { insert_data_sources(data: { name: \"op2023\", description: \"Open Payments 2023\", type: \"duckdb\", prefix: \"op2023\", path: \"/data/openpayments.duckdb\", read_only: true, as_module: true }) { name type description as_module path self_defined prefix read_only disabled } } }"
  }'

# Load it
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{
    "query": "mutation loadOpenPaymentsDataSource { function { core { load_data_sources(name: \"op2023\") { success message } } } }"
  }'
```

**Data Source Parameters** (from examples):
- `name`: Unique identifier for the data source
- `description`: Human-readable description
- `type`: "duckdb" for DuckDB databases
- `prefix`: Prefix for GraphQL types (e.g., "syn_", "op2023_")
- `path`: Absolute path to DuckDB file
- `read_only`: true (for query-only access)
- `as_module`: true (data source becomes a top-level GraphQL module)
- `self_defined`: true (use auto-generated schema from DuckDB)

**Authentication**:

**Option 1: SECRET_KEY (Recommended for data loading scripts)**
- Use `x-hugr-secret` header with SECRET_KEY from environment
- Simpler approach, no token management needed
- Ideal for automated scripts and data loading
```bash
# Set in .env file
SECRET_KEY=your-secret-key-here

# Use in curl requests
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{ "query": "..." }'
```

**Option 2: OIDC Token (For user-level operations)**
- Obtain token from Keycloak using admin credentials
- Pass as Bearer token in Authorization header
- Needed for role-based access testing
```bash
TOKEN=$(curl -X POST http://localhost:8180/realms/hugr/protocol/openid-connect/token \
  -d "client_id=hugr-graphql" \
  -d "username=admin@example.com" \
  -d "password=admin123" \
  -d "grant_type=password" \
  | jq -r '.access_token')

# Use in curl requests
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{ "query": "..." }'
```

**Decision for LDE**: Use SECRET_KEY approach in data loading scripts for simplicity. OIDC tokens used only in verification/testing steps.

**Verification Queries**:
```bash
# List data sources
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"query":"{ dataSources { name type description } }"}' | jq

# Count Synthea patients
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"query":"{ synthea { patients { count } } }"}' | jq
```

**Alternatives Considered**:
- **Python script**: Rejected to minimize dependencies
- **Go CLI tool**: Rejected as overkill for simple mutations
- **Manual mutations**: Rejected as not scriptable

---

### 8. Script Idempotency and Error Handling

**Decision**: Implement idempotent scripts with explicit error checking and rollback capability

**Rationale**:
- Scripts may be run multiple times during development
- Clear error messages aid debugging
- Ability to recover from partial failures

**Implementation Patterns**:
```bash
# Check prerequisites
check_prerequisites() {
  command -v docker >/dev/null 2>&1 || { echo "Docker required"; exit 1; }
  command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose required"; exit 1; }
  # ... more checks
}

# Idempotent data loading
load_synthea_data() {
  if data_source_exists "synthea_patients"; then
    echo "Synthea data source already exists, skipping..."
    return 0
  fi
  # ... load data
}

# Error handling
set -euo pipefail  # Exit on error, undefined var, pipe failure
trap 'echo "Error on line $LINENO"' ERR
```

**Features**:
- Verbose mode flag (-v) for debugging
- Dry-run mode (--dry-run) to preview actions
- Reset flag (--reset) for clean slate
- Color-coded output (green=success, red=error, yellow=warning)

**Alternatives Considered**:
- **Simple linear scripts**: Rejected as too fragile
- **Makefile**: Considered but shell scripts more flexible
- **Ansible playbook**: Rejected as overkill for local development

---

### 9. Port Exposure Strategy

**Decision**: Expose only essential ports to host, document port mappings clearly

**Port Mappings**:
- Hugr GraphQL: `8080:8080` (primary interface)
- Keycloak: `8180:8080` (avoid conflict with Hugr)
- PostgreSQL: `5432:5432` (for direct DB access if needed)
- MinIO Console: `9001:9001` (web UI)
- MinIO API: `9000:9000` (S3 API)
- Redis: `6379:6379` (for debugging)

**Rationale**:
- Hugr GraphQL must be accessible for MCP service development
- Keycloak UI needed for user management
- Database ports useful for debugging but optional
- MinIO console for data inspection

**Documentation**:
- Port table in README
- `.env.example` with port configuration
- Instructions for changing ports if conflicts occur

**Alternatives Considered**:
- **All ports exposed**: Rejected as unnecessary and clutters host
- **Minimal ports (Hugr only)**: Rejected as limits debugging capability
- **Random ports**: Rejected as not reproducible

---

## Implementation Dependencies

### Required Tools
- Docker Engine 20.10+
- Docker Compose v2.0+
- curl (http requests)
- jq (JSON processing)
- bash 4.0+ (script compatibility)

### Optional Tools
- DuckDB CLI (for data inspection)
- PostgreSQL client (psql) (for database debugging)

### External Resources
- Hugr examples repository (local clone)
- Synthea repository (local clone)
- Open Payments 2023 dataset (download or cached)

---

## Summary

All research complete. No NEEDS CLARIFICATION items remaining. Ready to proceed to Phase 1 (Design & Contracts).

**Key Decisions**:
1. Base configuration on hugr examples project
2. Docker-based Synthea generation → DuckDB format
3. Open Payments 2023 following examples pattern
4. Keycloak realm export with admin/viewer/analyst roles
5. `.local/` directory for volume persistence
6. Multi-level health checking (Docker + application)
7. curl + jq for GraphQL mutations
8. Idempotent error-handling scripts
9. Essential port exposure with clear documentation

**Next Phase**: Design & Contracts (data model, script interfaces, test contracts)
