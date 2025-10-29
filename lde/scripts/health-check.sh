#!/bin/bash
set -euo pipefail

# health-check.sh - Check health status of all LDE services
# Verifies that PostgreSQL, Redis, MinIO, Keycloak, and Hugr are healthy

# Color codes (only use if stdout is a terminal)
if [[ -t 1 ]]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    GREEN=''
    RED=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LDE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Default values
VERBOSE=false
WAIT_MODE=false
TIMEOUT=120

# Usage function
show_help() {
    cat << EOF
Usage: health-check.sh [OPTIONS]

Check health status of all Local Development Environment services.

Options:
  -v, --verbose     Enable verbose output
  --wait [TIMEOUT]  Wait for services to become healthy (default: 120s)
  --help            Show this help message

Exit Codes:
  0  - All services healthy
  1  - One or more services unhealthy
  2  - Timeout waiting for services

Examples:
  ./health-check.sh              # Check current status
  ./health-check.sh --wait       # Wait up to 120s for all healthy
  ./health-check.sh --wait 60    # Wait up to 60s
  ./health-check.sh --verbose    # Detailed output

Services Checked:
  • PostgreSQL (pg_isready)
  • Redis (redis-cli ping)
  • MinIO (HTTP health endpoint)
  • Keycloak (HTTP health endpoint)
  • Hugr (HTTP healthz endpoint)

EOF
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --wait)
            WAIT_MODE=true
            if [[ $# -gt 1 ]] && [[ $2 =~ ^[0-9]+$ ]]; then
                TIMEOUT=$2
                shift
            fi
            shift
            ;;
        --help)
            show_help
            ;;
        *)
            echo -e "${RED}✗ Unknown option: $1${NC}" >&2
            echo "  Use --help for usage information" >&2
            exit 1
            ;;
    esac
done

# Logging functions
log_info() {
    echo -e "${BLUE}→${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}⚠${NC}  $1"
}

log_verbose() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${YELLOW}  [DEBUG]${NC} $1" >&2
    fi
}

# Service health check results (using simple variables for Bash 3.2 compatibility)
POSTGRES_STATUS=""
POSTGRES_TIME=""
POSTGRES_ERROR=""

REDIS_STATUS=""
REDIS_TIME=""
REDIS_ERROR=""

MINIO_STATUS=""
MINIO_TIME=""
MINIO_ERROR=""

KEYCLOAK_STATUS=""
KEYCLOAK_TIME=""
KEYCLOAK_ERROR=""

HUGR_STATUS=""
HUGR_TIME=""
HUGR_ERROR=""

# Check PostgreSQL
check_postgres() {
    local start_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')

    log_verbose "Checking PostgreSQL..."

    if docker exec lde-postgres pg_isready -U hugr >/dev/null 2>&1; then
        local end_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')
        local response_time=$((end_time - start_time))
        POSTGRES_STATUS="healthy"
        POSTGRES_TIME="${response_time}ms"
        return 0
    else
        POSTGRES_STATUS="unhealthy"
        POSTGRES_ERROR="pg_isready check failed"
        return 1
    fi
}

# Check Redis
check_redis() {
    local start_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')

    log_verbose "Checking Redis..."

    if docker exec lde-redis redis-cli ping 2>/dev/null | grep -q PONG; then
        local end_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')
        local response_time=$((end_time - start_time))
        REDIS_STATUS="healthy"
        REDIS_TIME="${response_time}ms"
        return 0
    else
        REDIS_STATUS="unhealthy"
        REDIS_ERROR="redis-cli ping failed"
        return 1
    fi
}

# Check MinIO
check_minio() {
    local start_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')

    log_verbose "Checking MinIO..."

    if curl -sf http://localhost:19003/minio/health/live >/dev/null 2>&1; then
        local end_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')
        local response_time=$((end_time - start_time))
        MINIO_STATUS="healthy"
        MINIO_TIME="${response_time}ms"
        return 0
    else
        MINIO_STATUS="unhealthy"
        MINIO_ERROR="Health endpoint not responding"
        return 1
    fi
}

# Check Keycloak
check_keycloak() {
    local start_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')

    log_verbose "Checking Keycloak..."

    # Check if Keycloak is responding (any 2xx or 404 response means it's running)
    if curl -sf http://localhost:19005/ >/dev/null 2>&1 || curl -s -o /dev/null -w "%{http_code}" http://localhost:19005/ 2>/dev/null | grep -q "^[24]"; then
        local end_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')
        local response_time=$((end_time - start_time))
        KEYCLOAK_STATUS="healthy"
        KEYCLOAK_TIME="${response_time}ms"
        return 0
    else
        KEYCLOAK_STATUS="unhealthy"
        KEYCLOAK_ERROR="Server not responding"
        return 1
    fi
}

# Check Hugr
check_hugr() {
    local start_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')

    log_verbose "Checking Hugr..."

    if curl -sf http://localhost:19006/health >/dev/null 2>&1; then
        local end_time=$(perl -MTime::HiRes -e 'printf("%.0f",Time::HiRes::time()*1000)')
        local response_time=$((end_time - start_time))
        HUGR_STATUS="healthy"
        HUGR_TIME="${response_time}ms"
        return 0
    else
        HUGR_STATUS="unhealthy"
        HUGR_ERROR="Health endpoint not responding"
        return 1
    fi
}

# Check all services
check_all_services() {
    check_postgres || true
    check_redis || true
    check_minio || true
    check_keycloak || true
    check_hugr || true
}

# Display service status
display_status() {
    local total=5
    local healthy=0

    echo ""
    echo -e "${BLUE}Service Health Status:${NC}"
    echo ""

    # PostgreSQL
    if [[ "${POSTGRES_STATUS}" == "healthy" ]]; then
        echo -e "  PostgreSQL:  ${GREEN}✓ Healthy${NC} (Response: ${POSTGRES_TIME})"
        ((healthy++))
    else
        echo -e "  PostgreSQL:  ${RED}✗ Unhealthy${NC} (${POSTGRES_ERROR})"
    fi

    # Redis
    if [[ "${REDIS_STATUS}" == "healthy" ]]; then
        echo -e "  Redis:       ${GREEN}✓ Healthy${NC} (Response: ${REDIS_TIME})"
        ((healthy++))
    else
        echo -e "  Redis:       ${RED}✗ Unhealthy${NC} (${REDIS_ERROR})"
    fi

    # MinIO
    if [[ "${MINIO_STATUS}" == "healthy" ]]; then
        echo -e "  MinIO:       ${GREEN}✓ Healthy${NC} (Response: ${MINIO_TIME})"
        ((healthy++))
    else
        echo -e "  MinIO:       ${RED}✗ Unhealthy${NC} (${MINIO_ERROR})"
    fi

    # Keycloak
    if [[ "${KEYCLOAK_STATUS}" == "healthy" ]]; then
        echo -e "  Keycloak:    ${GREEN}✓ Healthy${NC} (Response: ${KEYCLOAK_TIME})"
        ((healthy++))
    else
        echo -e "  Keycloak:    ${RED}✗ Unhealthy${NC} (${KEYCLOAK_ERROR})"
    fi

    # Hugr
    if [[ "${HUGR_STATUS}" == "healthy" ]]; then
        echo -e "  Hugr:        ${GREEN}✓ Healthy${NC} (Response: ${HUGR_TIME})"
        ((healthy++))
    else
        echo -e "  Hugr:        ${RED}✗ Unhealthy${NC} (${HUGR_ERROR})"
    fi

    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    if [[ $healthy -eq $total ]]; then
        echo -e "${GREEN}All Services Healthy (${healthy}/${total})${NC}"
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}Health Check Failed (${healthy}/${total} healthy)${NC}"
        echo ""
        echo -e "${YELLOW}Failed Services:${NC}"

        [[ "${POSTGRES_STATUS}" != "healthy" ]] && echo -e "  • PostgreSQL: ${POSTGRES_ERROR}"
        [[ "${REDIS_STATUS}" != "healthy" ]] && echo -e "  • Redis: ${REDIS_ERROR}"
        [[ "${MINIO_STATUS}" != "healthy" ]] && echo -e "  • MinIO: ${MINIO_ERROR}"
        [[ "${KEYCLOAK_STATUS}" != "healthy" ]] && echo -e "  • Keycloak: ${KEYCLOAK_ERROR}"
        [[ "${HUGR_STATUS}" != "healthy" ]] && echo -e "  • Hugr: ${HUGR_ERROR}"

        echo ""
        echo -e "${YELLOW}Troubleshooting:${NC}"
        echo -e "  → Check container status: docker compose -f ${LDE_DIR}/docker-compose.yml ps"
        echo -e "  → View logs: docker compose -f ${LDE_DIR}/docker-compose.yml logs <service>"
        echo -e "  → Restart services: docker compose -f ${LDE_DIR}/docker-compose.yml restart"
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        return 1
    fi
}

# Wait mode - check repeatedly until all healthy or timeout
wait_for_healthy() {
    log_info "Waiting for services to become healthy (timeout: ${TIMEOUT}s)..."
    echo ""

    local start_time=$(date +%s)
    local elapsed=0
    local attempt=1

    while [[ $elapsed -lt $TIMEOUT ]]; do
        log_verbose "Health check attempt #${attempt}"

        check_all_services

        # Count healthy services
        local healthy=0
        [[ "${POSTGRES_STATUS}" == "healthy" ]] && ((healthy++))
        [[ "${REDIS_STATUS}" == "healthy" ]] && ((healthy++))
        [[ "${MINIO_STATUS}" == "healthy" ]] && ((healthy++))
        [[ "${KEYCLOAK_STATUS}" == "healthy" ]] && ((healthy++))
        [[ "${HUGR_STATUS}" == "healthy" ]] && ((healthy++))

        log_verbose "Healthy services: ${healthy}/5"

        if [[ $healthy -eq 5 ]]; then
            display_status
            return 0
        fi

        # Sleep before next attempt
        sleep 5

        local current_time=$(date +%s)
        elapsed=$((current_time - start_time))
        ((attempt++))
    done

    # Timeout reached
    log_error "Timeout waiting for services to become healthy (${TIMEOUT}s elapsed)"
    echo ""
    display_status
    return 2
}

# Main execution
main() {
    if [[ "$WAIT_MODE" == true ]]; then
        if wait_for_healthy; then
            exit 0
        else
            exit 2
        fi
    else
        log_info "Checking service health..."
        check_all_services
        if display_status; then
            exit 0
        else
            exit 1
        fi
    fi
}

main "$@"
