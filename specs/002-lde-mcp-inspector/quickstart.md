# Quickstart: LDE MCP Inspector

**Feature**: 002-lde-mcp-inspector
**Date**: 2025-10-30

## Purpose
This quickstart validates that the MCP Inspector service integrates correctly with the LDE by walking through the primary user scenarios from the feature specification.

## Prerequisites
- Docker Desktop 4.36+ or Docker Engine 20.10+
- Docker Compose v2.0+
- curl, jq
- Ports 19000-19007 available

## Test Scenarios

### Scenario 1: Start LDE with MCP Inspector

**User Story**: As a developer, I start the LDE and verify MCP Inspector is accessible.

**Steps**:
```bash
# 1. Navigate to repository root
cd /path/to/hugr-lab/mcp

# 2. Ensure clean state
./lde/scripts/cleanup.sh --force

# 3. Start the LDE
./lde/scripts/start.sh
```

**Expected Output**:
```
→ Starting LDE services...
✓ All services healthy (6/6)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Local Development Environment Ready
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Hugr GraphQL:  http://localhost:19000/query
Keycloak:      http://localhost:19005
MinIO Console: http://localhost:19004
MCP Inspector: http://localhost:19007

To stop: ./lde/scripts/stop.sh
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Validation**:
```bash
# Verify MCP Inspector is accessible
curl -sf http://localhost:19007 > /dev/null && echo "✓ MCP Inspector accessible" || echo "✗ Not accessible"

# Verify container running
docker ps --filter "name=lde-mcp-inspector" --filter "status=running" --format "{{.Names}}"
# Expected: lde-mcp-inspector
```

**Success Criteria**:
- [x] Start script completes without errors
- [x] Output shows "All services healthy (6/6)"
- [x] MCP Inspector URL displayed in summary
- [x] Port 19007 responds to HTTP requests
- [x] Container lde-mcp-inspector is running

---

### Scenario 2: Health Check Shows Inspector Status

**User Story**: As a developer, I run health checks and see MCP Inspector status.

**Steps**:
```bash
# Run health check
./lde/scripts/health-check.sh
```

**Expected Output**:
```
→ Checking service health...

Service Health Status:

  PostgreSQL:    ✓ Healthy (Response: 15ms)
  Redis:         ✓ Healthy (Response: 12ms)
  MinIO:         ✓ Healthy (Response: 45ms)
  Keycloak:      ✓ Healthy (Response: 120ms)
  Hugr:          ✓ Healthy (Response: 28ms)
  MCP Inspector: ✓ Healthy (Response: 8ms)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
All Services Healthy (6/6)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Validation**:
```bash
# Check exit code
./lde/scripts/health-check.sh
echo "Exit code: $?"
# Expected: 0

# Check verbose output
./lde/scripts/health-check.sh --verbose 2>&1 | grep "MCP Inspector"
# Expected: Contains "Checking MCP Inspector..." and "✓ Healthy"
```

**Success Criteria**:
- [x] Health check script exits with code 0
- [x] Output shows "MCP Inspector: ✓ Healthy"
- [x] Total shows "(6/6)"
- [x] Response time displayed for Inspector

---

### Scenario 3: Test History Persists Across Restarts

**User Story**: As a developer, I use the Inspector, restart the LDE, and my test history persists.

**Steps**:
```bash
# 1. Access MCP Inspector
open http://localhost:19007
# (Perform some test operations in the UI)

# 2. Verify test history data exists
ls -la ./lde/.local/mcp-inspector/
# Expected: directories for connections/, history/, sessions/, etc.

# 3. Stop the LDE
./lde/scripts/stop.sh

# 4. Verify data still exists
ls -la ./lde/.local/mcp-inspector/
# Expected: same directories present

# 5. Restart the LDE
./lde/scripts/start.sh

# 6. Access MCP Inspector again
open http://localhost:19007
# (Verify test history is restored)
```

**Validation**:
```bash
# Check volume mount persists
docker inspect lde-mcp-inspector | jq '.[0].Mounts[] | select(.Destination=="/data")'
# Expected: Source points to .local/mcp-inspector

# Verify directory not empty after restart
[[ -d "./lde/.local/mcp-inspector" ]] && echo "✓ Directory exists" || echo "✗ Missing"
```

**Success Criteria**:
- [x] `.local/mcp-inspector/` directory exists after first start
- [x] Directory persists after `stop.sh`
- [x] Directory still exists after `start.sh` (restart)
- [x] Inspector UI shows previous test history

---

### Scenario 4: Cleanup Removes Test History

**User Story**: As a developer, I clean up the LDE and all Inspector data is removed.

**Steps**:
```bash
# 1. Ensure Inspector has data
ls -la ./lde/.local/mcp-inspector/
# Expected: non-empty directory

# 2. Run cleanup
./lde/scripts/cleanup.sh --force

# 3. Verify data removed
ls -la ./lde/.local/mcp-inspector/ 2>&1
# Expected: directory does not exist
```

**Validation**:
```bash
# Check cleanup warning mentions Inspector
./lde/scripts/cleanup.sh --help | grep -i "mcp inspector"
# Expected: Contains reference to Inspector test history

# Verify directory removed
[[ ! -d "./lde/.local/mcp-inspector" ]] && echo "✓ Removed" || echo "✗ Still exists"

# Verify container removed
docker ps -a --filter "name=lde-mcp-inspector" --format "{{.Names}}"
# Expected: empty output
```

**Success Criteria**:
- [x] Cleanup warning mentions "MCP Inspector test history"
- [x] `.local/mcp-inspector/` directory removed after cleanup
- [x] Container removed
- [x] Cleanup with `--keep-local` preserves directory

---

### Scenario 5: Stop LDE Stops Inspector Cleanly

**User Story**: As a developer, I stop the LDE and all services including Inspector stop gracefully.

**Steps**:
```bash
# 1. Ensure LDE is running
./lde/scripts/health-check.sh
# Expected: All services healthy

# 2. Stop the LDE
./lde/scripts/stop.sh

# 3. Verify Inspector stopped
docker ps --filter "name=lde-mcp-inspector"
# Expected: empty output (no running container)
```

**Validation**:
```bash
# Check stop script output
./lde/scripts/stop.sh 2>&1 | tee /dev/stderr | grep -i "stopped"
# Expected: confirmation message about stopped services

# Verify all containers stopped
docker ps --filter "name=lde-" --format "{{.Names}}"
# Expected: empty output

# Verify data persists
[[ -d "./lde/.local/mcp-inspector" ]] && echo "✓ Data preserved" || echo "✗ Data lost"
```

**Success Criteria**:
- [x] Stop script completes without errors
- [x] No lde-* containers running after stop
- [x] `.local/mcp-inspector/` directory still exists
- [x] Container can be restarted with `start.sh`

---

## Acceptance Validation

All scenarios must pass for feature acceptance:

- [x] **Scenario 1**: MCP Inspector starts with LDE
- [x] **Scenario 2**: Health check includes Inspector status
- [x] **Scenario 3**: Test history persists across restarts
- [x] **Scenario 4**: Cleanup removes all Inspector data
- [x] **Scenario 5**: Stop gracefully shuts down Inspector

## Troubleshooting

### Inspector not accessible
```bash
# Check container logs
docker logs lde-mcp-inspector

# Verify port not in use
lsof -i :19007

# Check service definition
docker compose -f lde/docker-compose.yml config | grep -A 20 mcp-inspector
```

### Health check shows unhealthy
```bash
# Check container status
docker ps -a --filter "name=lde-mcp-inspector"

# Restart service
docker compose -f lde/docker-compose.yml restart mcp-inspector

# View detailed health check
./lde/scripts/health-check.sh --verbose
```

### Test history not persisting
```bash
# Verify volume mount
docker inspect lde-mcp-inspector | jq '.[0].Mounts'

# Check directory permissions
ls -ld ./lde/.local/mcp-inspector
```

## Performance Benchmarks

Expected performance metrics:
- **Startup time**: <30s to healthy state
- **Health check**: <5s response time
- **UI load**: <2s initial page load
- **Stop time**: <10s graceful shutdown

## Next Steps

After quickstart validation:
1. Review lde/README.md for detailed documentation
2. Explore MCP Inspector UI features
3. Test with actual MCP servers
4. Configure custom environment variables in `lde/.env`
