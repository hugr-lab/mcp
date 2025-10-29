# Quickstart: Local Development Environment

**Feature**: 001-create-a-local
**Audience**: Developers setting up local Hugr MCP development environment
**Time**: ~15 minutes (first run), ~2 minutes (subsequent runs)

## Prerequisites

Before starting, ensure you have:

1. **Docker & Docker Compose**
   ```bash
   docker --version  # Should be 20.10+
   docker compose version  # Should be v2.0+
   ```

2. **Required Tools**
   ```bash
   curl --version
   jq --version
   ```

3. **Local Repositories**
   ```bash
   # Verify paths exist
   ls /Users/vgribanov/projects/hugr-lab/examples
   ls /Users/vgribanov/projects/synthea
   ```

4. **System Resources**
   - At least 8GB RAM available
   - At least 10GB free disk space
   - Ports available: 8080, 8180, 5432, 6379, 9000, 9001

## Quick Start (First Time Setup)

### Step 1: Start the Environment

From the repository root:

```bash
cd /Users/vgribanov/projects/hugr-lab/mcp
./lde/scripts/start.sh
```

**Expected Duration**: 5-8 minutes
- Docker image pulls: 2-3 minutes
- Service startup: 1-2 minutes
- Data generation & loading: 5-10 minutes (Synthea + Open Payments)

**Expected Output**:
```
✓ Docker installed (version 24.0.6)
✓ Docker Compose installed (version v2.23.0)
✓ curl installed
✓ jq installed
✓ Examples repository found
✓ Synthea repository found

→ Creating .local directory structure...
✓ Created .local/pg-data
✓ Created .local/redis-data
✓ Created .local/minio
✓ Created .local/keycloak

→ Starting Docker services...
[+] Running 5/5
 ✓ Container lde-postgres-1   Healthy
 ✓ Container lde-redis-1      Healthy
 ✓ Container lde-minio-1      Healthy
 ✓ Container lde-keycloak-1   Healthy
 ✓ Container lde-hugr-1       Healthy

→ Verifying service health...
✓ PostgreSQL ready (23ms)
✓ Redis ready (12ms)
✓ MinIO ready (45ms)
✓ Keycloak ready (1.2s)
✓ Hugr GraphQL ready (890ms)

→ Loading data sources...
→ Generating Synthea data (10,000 patients)...
[====================] 100% (10000/10000 patients)
✓ Generated 10,000 patients (3m 45s)
✓ Synthea data loaded (10000 patients)

→ Loading Open Payments 2023 data...
[====================] 100% (11.8M records)
✓ Open Payments data loaded (11.8M records)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Local Development Environment Ready
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Hugr GraphQL:  http://localhost:19000/query
Keycloak:      http://localhost:19005
MinIO Console: http://localhost:19004

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

### Step 2: Verify the Installation

Run a quick health check:

```bash
./lde/scripts/health-check.sh
```

**Expected Output**:
```
→ Checking service health...

PostgreSQL:  ✓ Healthy (Response: 18ms)
Redis:       ✓ Healthy (Response: 9ms)
MinIO:       ✓ Healthy (Response: 52ms)
Keycloak:    ✓ Healthy (Response: 1.1s)
Hugr:        ✓ Healthy (Response: 234ms)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
All Services Healthy (5/5)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Step 3: Test GraphQL Access

Open your browser or use curl to test the Hugr GraphQL endpoint:

```bash
curl -X POST http://localhost:19000/query \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __schema { queryType { name } } }"}'
```

**Expected Response**:
```json
{
  "data": {
    "__schema": {
      "queryType": {
        "name": "Query"
      }
    }
  }
}
```

### Step 4: Test Authentication

**Option A: Using SECRET_KEY (simpler, for data operations)**

```bash
# SECRET_KEY is set in .env file, load it
source lde/.env

# Query Synthea data using x-hugr-secret header
curl -X POST http://localhost:19000/query \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{
    "query": "query { synthea { patients_aggregation { count: _rows_count } } }"
  }' | jq
```

**Option B: Using Keycloak Token (for role-based testing)**

```bash
# Get admin token from Keycloak
TOKEN=$(curl -X POST http://localhost:19005/realms/hugr/protocol/openid-connect/token \
  -d "client_id=hugr-graphql" \
  -d "username=admin@example.com" \
  -d "password=admin123" \
  -d "grant_type=password" \
  | jq -r '.access_token')

# Query Synthea data using Bearer token
curl -X POST http://localhost:19000/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "query": "query { synthea { patients_aggregation { count: _rows_count } } }"
  }' | jq
```

**Expected Response**:
```json
{
  "data": {
    "synthea": {
      "patients_aggregation": {
        "count": 10000
      }
    }
  }
}
```

### Step 5: Explore Data Sources

List available data sources:

```bash
# Using SECRET_KEY (recommended for scripts)
curl -X POST http://localhost:19000/query \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: ${SECRET_KEY}" \
  -d '{"query":"query { core { data_sources { name type description } } }"}' | jq
```

**Expected Response**:
```json
{
  "data": {
    "core": {
      "data_sources": [
        {
          "name": "synthea",
          "type": "duckdb",
          "description": "Synthea synthetic patient data"
        },
        {
          "name": "op2023",
          "type": "duckdb",
          "description": "Open Payments 2023"
        }
      ]
    }
  }
}
```

## Common Operations

### Stop the Environment

Preserves all data in `.local/` directory:

```bash
./lde/scripts/stop.sh
```

### Restart the Environment

Quick restart using existing data:

```bash
./lde/scripts/start.sh
```

**Expected Duration**: ~30 seconds (no data reload needed)

### Reset Environment (Clean Slate)

Wipe all data and start fresh:

```bash
./lde/scripts/start.sh --reset
```

**Warning**: This deletes all data in `.local/` directory!

### Start Without Data Loading

For faster testing when data not needed:

```bash
./lde/scripts/start.sh --no-data
```

Then load data separately:

```bash
./lde/scripts/load-data.sh
```

### Load Specific Data Source

Load only Synthea:

```bash
./lde/scripts/load-data.sh --synthea-only
```

Load only Open Payments:

```bash
./lde/scripts/load-data.sh --openpayments-only
```

### Force Reload Data

Reload even if data exists:

```bash
./lde/scripts/load-data.sh --force
```

## Troubleshooting

### Services Won't Start

**Problem**: Docker containers fail to start

**Solution**:
```bash
# Check Docker is running
docker ps

# View service logs
docker compose -f lde/docker-compose.yml logs

# Try resetting
./lde/scripts/stop.sh
./lde/scripts/start.sh --reset
```

### Port Already in Use

**Problem**: Error: "port is already allocated"

**Solution**:
1. Identify conflicting process:
   ```bash
   lsof -i :8080  # Or whichever port is conflicting
   ```

2. Either stop the conflicting process or modify `.env` file:
   ```bash
   # Edit lde/.env
   HUGR_PORT=8081  # Change to available port
   ```

3. Restart:
   ```bash
   ./lde/scripts/stop.sh
   ./lde/scripts/start.sh
   ```

### Health Checks Failing

**Problem**: Services show as unhealthy

**Solution**:
```bash
# Detailed health check
./lde/scripts/health-check.sh --verbose

# Check individual service logs
docker compose -f lde/docker-compose.yml logs postgres
docker compose -f lde/docker-compose.yml logs keycloak
docker compose -f lde/docker-compose.yml logs hugr

# Restart unhealthy service
docker compose -f lde/docker-compose.yml restart <service-name>
```

### Data Loading Fails

**Problem**: Synthea or Open Payments data fails to load

**Solution**:
```bash
# Check Hugr is healthy first
./lde/scripts/health-check.sh

# Try loading with verbose output
./lde/scripts/load-data.sh --verbose

# Load data sources separately
./lde/scripts/load-data.sh --synthea-only --verbose
./lde/scripts/load-data.sh --openpayments-only --verbose

# Force reload
./lde/scripts/load-data.sh --force
```

### Keycloak Authentication Fails

**Problem**: Cannot get authentication token

**Solution**:
```bash
# Verify Keycloak is healthy
curl -f http://localhost:19005/health/ready

# Check realm is loaded
curl http://localhost:19005/realms/hugr | jq

# Verify user exists in realm
# Login to Keycloak admin console: http://localhost:19005
# Username: admin, Password: admin
# Navigate to Realm: hugr → Users
```

### Out of Disk Space

**Problem**: No space left on device

**Solution**:
```bash
# Check disk usage
du -sh lde/.local/*

# Clean up old data
./lde/scripts/stop.sh
rm -rf lde/.local/*
./lde/scripts/start.sh
```

### Containers Keep Restarting

**Problem**: Docker containers in restart loop

**Solution**:
```bash
# Check container status
docker compose -f lde/docker-compose.yml ps

# View recent logs
docker compose -f lde/docker-compose.yml logs --tail=50 <service-name>

# Check resource usage
docker stats

# If out of memory, increase Docker memory limit in Docker Desktop settings
```

## Manual Cleanup

If scripts fail or you need to manually clean up:

```bash
# Stop all containers
docker compose -f lde/docker-compose.yml down

# Remove volumes
rm -rf lde/.local/*

# Remove generated data
rm -rf lde/data/*

# Start fresh
./lde/scripts/start.sh --reset
```

## Accessing Service UIs

### Hugr GraphQL Playground
- URL: http://localhost:19000/query
- Use Keycloak tokens in Authorization header

### Keycloak Admin Console
- URL: http://localhost:19005
- Username: `admin`
- Password: `admin`
- Realm: `hugr`

### MinIO Console
- URL: http://localhost:19004
- Username: `minioadmin`
- Password: `minioadmin`

### PostgreSQL (via psql)
```bash
psql postgresql://hugr:hugr@localhost:19001/hugr
```

### Redis (via redis-cli)
```bash
redis-cli -h localhost -p 19002
```

## Next Steps

Now that your local development environment is running:

1. **Explore the GraphQL Schema**
   - Open http://localhost:19000/query
   - Use introspection to explore available queries

2. **Test Different User Roles**
   - Get tokens for admin, analyst, and viewer users
   - Observe role-based field filtering

3. **Query Healthcare Data**
   - Explore Synthea patient records
   - Query Open Payments transactions
   - Join data across sources

4. **Develop MCP Tools**
   - Use this environment to test MCP service features
   - All Hugr APIs available locally

5. **Run Integration Tests**
   - Test against realistic healthcare datasets
   - Verify role-based access control

## Verification Checklist

- [ ] All 5 services show as healthy
- [ ] Can access Hugr GraphQL endpoint
- [ ] Can authenticate with Keycloak
- [ ] Synthea data source has 10,000 patients
- [ ] Open Payments data source accessible
- [ ] Can query data with admin token
- [ ] Different user roles work (admin, analyst, viewer)
- [ ] Services survive restart (data persists)
- [ ] Can reset environment to clean state

## Success Criteria

Your environment is correctly set up when:

1. **✓ Health check passes**: All 5 services healthy
2. **✓ Authentication works**: Can get token from Keycloak
3. **✓ Data accessible**: Can query both data sources
4. **✓ Persistence works**: Data survives stop/start cycle
5. **✓ Reset works**: `--reset` flag gives clean slate

---

**Estimated Total Time**: 15 minutes (first run), 2 minutes (subsequent runs)

**Support**: If issues persist, check `docker compose logs` for detailed error messages or refer to the troubleshooting section above.
