# Research: LDE MCP Inspector

**Feature**: 002-lde-mcp-inspector
**Date**: 2025-10-30
**Status**: Complete

## Overview
Research findings for adding the MCP Inspector service to the Local Development Environment Docker Compose stack.

## 1. MCP Inspector Docker Deployment

### Decision
Use the official MCP Inspector Docker image `ghcr.io/modelcontextprotocol/inspector:latest` with custom port mappings to fit LDE port range (19000-19020).

### Configuration Details
- **Image**: `ghcr.io/modelcontextprotocol/inspector:latest`
- **Ports**:
  - Default: 6274 (client UI), 6277 (proxy server)
  - LDE mapping: 19007 for unified access (both UI and proxy)
- **Volumes**: Requires persistent storage for test history
- **Environment Variables**:
  - `HOST`: Network interface binding (default: localhost)
  - `ALLOWED_ORIGINS`: CORS configuration
  - `MCP_PROXY_AUTH_TOKEN`: Custom authentication token
  - `DANGEROUSLY_OMIT_AUTH`: Disable auth (not recommended for production)

### Rationale
- Official image ensures compatibility and updates
- Port 19007 follows LDE sequential port pattern (after Hugr health endpoint on 19006)
- Authentication enabled by default aligns with security best practices
- Volume mount required to persist test history per FR-009

### Alternatives Considered
- **Build from source**: Rejected - unnecessary complexity, official image sufficient
- **Port range 6274-6277**: Rejected - conflicts with LDE standard port range
- **No persistence**: Rejected - spec requires test history persistence

## 2. LDE Health Check Patterns

### Decision
Check MCP Inspector container running status only, without HTTP endpoint verification.

### Pattern Analysis
From `lde/scripts/health-check.sh`:
- **PostgreSQL**: Uses `docker exec` with `pg_isready` command
- **Redis**: Uses `docker exec` with `redis-cli ping` command
- **MinIO**: HTTP endpoint check via `curl` to `/minio/health/live`
- **Keycloak**: HTTP endpoint check via `curl` to root path
- **Hugr**: HTTP endpoint check via `curl` to `/health`

### Implementation for MCP Inspector
Based on clarification (container status only):
```bash
check_mcp_inspector() {
    local start_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')

    if docker ps --filter "name=lde-mcp-inspector" --filter "status=running" --format "{{.Names}}" | grep -q "lde-mcp-inspector"; then
        local end_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')
        local response_time=$((end_time - start_time))
        MCP_INSPECTOR_STATUS="healthy"
        MCP_INSPECTOR_TIME="${response_time}ms"
        return 0
    else
        MCP_INSPECTOR_STATUS="unhealthy"
        MCP_INSPECTOR_ERROR="Container not running"
        return 1
    fi
}
```

### Rationale
- Follows existing LDE pattern for service status tracking
- Container status check is sufficient per clarification requirements
- Response time tracking maintains consistency with other services
- Simpler than HTTP checks, no endpoint dependencies

### Alternatives Considered
- **HTTP health endpoint**: Rejected per clarification - container status sufficient
- **No health check**: Rejected - inconsistent with LDE service pattern

## 3. LDE Volume Management Patterns

### Decision
Create `.local/mcp-inspector/` directory for persistent test history, clean up in `cleanup.sh` alongside other service data.

### Pattern Analysis
From `lde/docker-compose.yml` and `cleanup.sh`:
- **Volume Structure**: All services use `.local/<service-name>/` pattern
  - `.local/pg-data/` - PostgreSQL
  - `.local/redis-data/` - Redis
  - `.local/minio/` - MinIO
  - `.local/keycloak/` - Keycloak
- **Cleanup Behavior**: `cleanup.sh` removes entire `.local/` directory unless `--keep-local` flag used
- **Data vs Local**:
  - `.local/` = service runtime data (always cleaned)
  - `data/` = application data (optionally preserved with `--keep-data`)

### Implementation
```yaml
volumes:
  - ./.local/mcp-inspector:/data
```

Cleanup in `cleanup.sh`:
- Automatically cleaned when `.local/` directory removed
- No special handling needed - follows existing pattern

### Rationale
- Consistent with LDE volume naming convention
- Test history is service data, not application data - belongs in `.local/`
- Automatic cleanup via existing `cleanup.sh` logic per FR-011 clarification

### Alternatives Considered
- **Store in `data/` directory**: Rejected - test history is service data, not application database
- **Named Docker volume**: Rejected - LDE pattern uses host bind mounts for easier access and backup

## 4. MCP Inspector Authentication Configuration

### Decision
Enable hybrid authentication: accept Keycloak tokens OR allow unauthenticated localhost access.

### Configuration Approach
```yaml
environment:
  - HOST=0.0.0.0  # Allow remote access
  - ALLOWED_ORIGINS=http://localhost:19007,http://localhost:19005  # Include Keycloak origin
  - DANGEROUSLY_OMIT_AUTH=false  # Keep auth enabled by default
```

Additional configuration:
- Set `MCP_PROXY_AUTH_TOKEN` to random value on first start
- Document token retrieval from container logs
- Document localhost bypass behavior (if supported by Inspector)

### Rationale
- Hybrid auth aligns with FR-010 requirements
- Default-secure with auth enabled
- Keycloak integration possible via CORS configuration
- Localhost convenience for developers

### Alternatives Considered
- **Disable auth entirely**: Rejected - security risk, against best practices
- **Keycloak-only auth**: Rejected - too complex for local dev, no existing Inspector integration
- **Secret key only**: Rejected - doesn't match LDE user auth pattern

## 5. Docker Compose Service Dependencies

### Decision
No explicit service dependencies for MCP Inspector.

### Rationale
- MCP Inspector is a standalone testing tool
- Does not depend on other LDE services for functionality
- Can start independently and wait for user to connect to target servers
- Prevents cascade failures if other services are slow to start

### Alternatives Considered
- **Depend on Hugr**: Rejected - Inspector can test any MCP server, not just Hugr
- **Depend on Keycloak**: Rejected - auth integration not required for functionality

## 6. Environment Variable Defaults

### Decision
Add the following to `.env.example`:

```bash
# MCP Inspector
MCP_INSPECTOR_IMAGE=ghcr.io/modelcontextprotocol/inspector:latest
MCP_INSPECTOR_PORT=19007
MCP_INSPECTOR_HOST=0.0.0.0
MCP_INSPECTOR_ALLOWED_ORIGINS=http://localhost:19007
```

### Rationale
- Consistent with LDE pattern of externalizing configuration
- Allows users to customize image version
- Port override capability for conflict resolution
- CORS configuration flexibility

### Alternatives Considered
- **Hardcode in docker-compose.yml**: Rejected - reduces flexibility
- **Minimal env vars**: Rejected - users may need to customize image version or ports

## Summary

| Decision Area | Choice | Key Rationale |
|---------------|--------|---------------|
| Docker Image | `ghcr.io/modelcontextprotocol/inspector:latest` | Official, maintained, compatible |
| Port Mapping | 19007 (single port for both UI and proxy) | Follows LDE sequential pattern |
| Health Check | Container running status only | Per clarification, simpler approach |
| Persistence | `.local/mcp-inspector/` bind mount | Follows LDE volume pattern |
| Cleanup | Auto-removed with `.local/` directory | Follows existing cleanup logic |
| Authentication | Hybrid (tokens + localhost bypass) | Balances security and dev convenience |
| Dependencies | None | Independent service, prevents cascades |
| Env Variables | Image, port, host, CORS in `.env.example` | Configuration flexibility |

## Next Steps
Proceed to Phase 1: Design & Contracts with these research findings as foundation.
