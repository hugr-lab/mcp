# Script Interface Contracts

## Overview
This document defines the expected interface contracts for all shell scripts in the LDE.

## start.sh Contract

### Interface
```bash
start.sh [--reset] [-v|--verbose] [--no-data] [--help]
```

### Parameters
- `--reset`: Optional flag to wipe all volumes before starting
- `-v, --verbose`: Optional flag for detailed output
- `--no-data`: Optional flag to skip data loading
- `--help`: Display usage information

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success - all services healthy |
| 1 | Prerequisites missing |
| 2 | Docker compose failure |
| 3 | Health checks failed |
| 4 | Data loading failed (when not --no-data) |

### Expected Output
1. **Prerequisite Check Phase**
   ```
   ✓ Docker installed (version X.Y.Z)
   ✓ Docker Compose installed (version A.B.C)
   ✓ curl installed
   ✓ jq installed
   ✓ Examples repository found at /path/to/examples
   ✓ Synthea repository found at /path/to/synthea
   ```

2. **Initialization Phase** (if --reset or first run)
   ```
   → Creating .local directory structure...
   ✓ Created .local/pg-data
   ✓ Created .local/redis-data
   ✓ Created .local/minio
   ✓ Created .local/keycloak
   ```

3. **Service Startup Phase**
   ```
   → Starting Docker services...
   [+] Running 5/5
    ✓ Container lde-postgres-1   Healthy
    ✓ Container lde-redis-1      Healthy
    ✓ Container lde-minio-1      Healthy
    ✓ Container lde-keycloak-1   Healthy
    ✓ Container lde-hugr-1       Healthy
   ```

4. **Health Check Phase**
   ```
   → Verifying service health...
   ✓ PostgreSQL ready (23ms)
   ✓ Redis ready (12ms)
   ✓ MinIO ready (45ms)
   ✓ Keycloak ready (1.2s)
   ✓ Hugr GraphQL ready (890ms)
   ```

5. **Data Loading Phase** (unless --no-data)
   ```
   → Loading data sources...
   ✓ Synthea data loaded (10000 patients)
   ✓ Open Payments data loaded (11.8M records)
   ```

6. **Connection Information**
   ```
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Local Development Environment Ready
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   Hugr GraphQL:  http://localhost:8080/graphql
   Keycloak:      http://localhost:8180
   MinIO Console: http://localhost:9001

   Test Credentials:
   - admin@example.com    / admin123    (admin role)
   - analyst@example.com  / analyst123  (analyst role)
   - viewer@example.com   / viewer123   (viewer role)

   Data Sources:
   - synthea_patients (10,000 patients)
   - open_payments_2023 (11.8M records)

   To stop: ./lde/scripts/stop.sh
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   ```

### Test Contract
```bash
# Test: Script exists and is executable
test -x lde/scripts/start.sh

# Test: Help flag works
lde/scripts/start.sh --help
expected_exit_code=0
expected_output_contains="Usage:"

# Test: Prerequisites check detects missing tools
mock_missing_docker && lde/scripts/start.sh
expected_exit_code=1
expected_stderr_contains="Docker required"

# Test: Services start successfully
lde/scripts/start.sh --no-data
expected_exit_code=0
verify_docker_ps_shows_5_healthy_containers

# Test: Reset flag wipes volumes
touch lde/.local/test-marker
lde/scripts/start.sh --reset --no-data
expected_exit_code=0
test ! -f lde/.local/test-marker  # Marker should be gone

# Test: Idempotency - can run multiple times
lde/scripts/start.sh --no-data
lde/scripts/start.sh --no-data
expected_exit_code=0
```

---

## stop.sh Contract

### Interface
```bash
stop.sh [-v|--verbose] [--help]
```

### Parameters
- `-v, --verbose`: Optional flag for detailed output
- `--help`: Display usage information

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success - all services stopped |
| 1 | Docker compose failure |

### Expected Output
```
→ Stopping Docker services...
[+] Stopping 5/5
 ✓ Container lde-hugr-1       Stopped
 ✓ Container lde-keycloak-1   Stopped
 ✓ Container lde-minio-1      Stopped
 ✓ Container lde-redis-1      Stopped
 ✓ Container lde-postgres-1   Stopped

✓ All services stopped
✓ Data persisted in lde/.local/

To restart: ./lde/scripts/start.sh
```

### Test Contract
```bash
# Test: Stop works when services running
lde/scripts/start.sh --no-data
lde/scripts/stop.sh
expected_exit_code=0
verify_docker_ps_shows_0_lde_containers

# Test: Stop is idempotent
lde/scripts/stop.sh
lde/scripts/stop.sh
expected_exit_code=0

# Test: Data persists after stop
lde/scripts/start.sh --no-data
touch lde/.local/test-marker
lde/scripts/stop.sh
test -f lde/.local/test-marker  # Marker should persist
```

---

## load-data.sh Contract

### Interface
```bash
load-data.sh [--synthea-only] [--openpayments-only] [--force] [-v|--verbose] [--help]
```

### Parameters
- `--synthea-only`: Load only Synthea data
- `--openpayments-only`: Load only Open Payments data
- `--force`: Force reload even if data exists
- `-v, --verbose`: Detailed output
- `--help`: Usage information

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success - data loaded and verified |
| 1 | Prerequisites missing |
| 2 | Hugr not healthy |
| 3 | Data generation failed |
| 4 | Data source registration failed |
| 5 | Data verification failed |

### Expected Output
1. **Environment Check**
   ```
   → Checking prerequisites...
   ✓ Hugr is healthy
   ✓ Synthea repository found
   ✓ Examples repository found
   ```

2. **Synthea Generation**
   ```
   → Generating Synthea data (10,000 patients)...
   Running Synthea Docker container...
   [====================] 100% (10000/10000 patients)
   ✓ Generated 10,000 patients (3m 45s)

   → Converting to DuckDB format...
   Processing CSV files...
   [====================] 100% (9 tables)
   ✓ Created synthea.duckdb (523 MB)
   ```

3. **Open Payments Loading**
   ```
   → Loading Open Payments 2023 data...
   Downloading dataset (if needed)...
   ✓ Dataset cached locally

   Converting to Parquet...
   [====================] 100% (11.8M records)
   ✓ Converted to Parquet (2.1 GB)
   ```

4. **Hugr Registration**
   ```
   → Registering data sources with Hugr...
   Creating Synthea data source...
   ✓ Created synthea_patients data source

   Creating Open Payments data source...
   ✓ Created open_payments_2023 data source
   ```

5. **Verification**
   ```
   → Verifying data sources...
   Testing Synthea queries...
   ✓ Patient count: 10,000
   ✓ Encounter count: 45,823
   ✓ Sample query successful

   Testing Open Payments queries...
   ✓ Record count: 11,872,943
   ✓ Year range: 2023-2023
   ✓ Sample query successful

   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Data Loading Complete
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Data Sources: 2
   Total Records: ~11.9M
   Disk Usage: ~2.6 GB
   Load Time: 8m 12s
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   ```

### Test Contract
```bash
# Test: Requires Hugr to be running
lde/scripts/stop.sh
lde/scripts/load-data.sh
expected_exit_code=2
expected_stderr_contains="Hugr not healthy"

# Test: Generates Synthea data
lde/scripts/start.sh --no-data
lde/scripts/load-data.sh --synthea-only
expected_exit_code=0
test -f lde/data/synthea.duckdb

# Test: Skips if data exists (unless --force)
lde/scripts/load-data.sh --synthea-only
expected_output_contains="already exists, skipping"
lde/scripts/load-data.sh --synthea-only --force
expected_output_contains="Generating Synthea data"

# Test: Data sources registered in Hugr
lde/scripts/load-data.sh
verify_graphql_query_returns_2_data_sources

# Test: Verification queries succeed
lde/scripts/load-data.sh
verify_synthea_patient_count_equals_10000
verify_openpayments_year_equals_2023
```

---

## health-check.sh Contract

### Interface
```bash
health-check.sh [-v|--verbose] [--wait [TIMEOUT]] [--help]
```

### Parameters
- `-v, --verbose`: Detailed output
- `--wait [TIMEOUT]`: Wait for services (default 120s)
- `--help`: Usage information

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | All services healthy |
| 1 | One or more services unhealthy |
| 2 | Timeout waiting for services |

### Expected Output (All Healthy)
```
→ Checking service health...

PostgreSQL:  ✓ Healthy (Response: 18ms)
Redis:       ✓ Healthy (Response: 9ms)
MinIO:       ✓ Healthy (Response: 52ms)
Keycloak:    ✓ Healthy (Response: 1.1s)
Hugr:        ✓ Healthy (Response: 234ms)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
All Services Healthy (5/5)
Total Check Time: 1.4s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Expected Output (Failure)
```
→ Checking service health...

PostgreSQL:  ✓ Healthy (Response: 18ms)
Redis:       ✓ Healthy (Response: 9ms)
MinIO:       ✗ Unhealthy (Connection refused)
Keycloak:    ✗ Unhealthy (HTTP 503)
Hugr:        ⚠ Waiting  (Dependencies not ready)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Health Check Failed (2/5 healthy)

Failed Services:
- MinIO: Connection refused on port 9000
  → Check: docker compose logs minio
- Keycloak: HTTP 503 Service Unavailable
  → Check: docker compose logs keycloak

Warnings:
- Hugr: Waiting for dependencies

Suggestion: Run 'docker compose restart minio keycloak'
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Test Contract
```bash
# Test: Detects healthy services
lde/scripts/start.sh --no-data
lde/scripts/health-check.sh
expected_exit_code=0
expected_output_contains="All Services Healthy (5/5)"

# Test: Detects unhealthy services
lde/scripts/start.sh --no-data
docker compose stop minio
lde/scripts/health-check.sh
expected_exit_code=1
expected_output_contains="MinIO"
expected_output_contains="Unhealthy"

# Test: Wait mode with timeout
lde/scripts/stop.sh
lde/scripts/health-check.sh --wait 10 &
sleep 2
lde/scripts/start.sh --no-data
wait  # Should succeed once services up

# Test: Timeout triggers exit code 2
lde/scripts/stop.sh
lde/scripts/health-check.sh --wait 5
expected_exit_code=2
expected_output_contains="Timeout"
```

---

## Common Contract Elements

### All Scripts Must
1. **Be executable**: `chmod +x scripts/*.sh`
2. **Use bash shebang**: `#!/usr/bin/env bash`
3. **Set strict mode**: `set -euo pipefail`
4. **Support --help flag**: Display usage and exit 0
5. **Use color output**: Green (✓), Red (✗), Yellow (⚠), when stdout is tty
6. **Log to stderr for errors**: Keep stdout for data/structured output
7. **Provide progress indicators**: For long-running operations
8. **Exit with appropriate codes**: Follow contract exit code definitions
9. **Be idempotent**: Safe to run multiple times
10. **Clean up on signal**: Trap SIGINT, SIGTERM for graceful shutdown

### Error Message Format
```bash
echo "✗ Error: <problem description>" >&2
echo "  → Suggestion: <how to fix>" >&2
exit <appropriate_code>
```

### Success Message Format
```bash
echo "✓ <operation> succeeded"
```

### Progress Indicator Format
```bash
echo "→ <operation in progress>..."
# ... do work ...
echo "✓ <operation> complete"
```

## Test Execution

All contract tests implemented in:
- `tests/lde/test-scripts.sh` - Script interface validation
- `tests/lde/test-services.sh` - Service health integration tests
- `tests/lde/test-data.sh` - Data loading verification tests

Run all tests:
```bash
./tests/lde/test-scripts.sh
./tests/lde/test-services.sh
./tests/lde/test-data.sh
```

Expected: All tests pass before implementation is considered complete.
