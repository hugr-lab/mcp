#!/usr/bin/env bash
set -euo pipefail

# start.sh - Start the Local Development Environment
# Handles prerequisites, initialization, service startup, health checks, and data loading

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

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LDE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJECT_ROOT="$(cd "${LDE_DIR}/.." && pwd)"

# Default values
RESET=false
NO_DATA=false
VERBOSE=false

# Usage function
show_help() {
    cat << EOF
Usage: start.sh [OPTIONS]

Start the Local Development Environment with all required services.

Options:
  --reset        Wipe all volumes and start fresh
  --no-data      Skip data loading step
  -v, --verbose  Enable verbose output
  --help         Show this help message

Exit Codes:
  0  - Success (all services healthy)
  1  - Prerequisites missing (docker, docker-compose, etc.)
  2  - Docker compose failed
  3  - Health checks failed
  4  - Data loading failed (if --no-data not set)

Examples:
  ./start.sh                    # Normal startup with data loading
  ./start.sh --no-data          # Start without loading data
  ./start.sh --reset            # Clean slate (wipes volumes)
  ./start.sh --reset --no-data  # Clean start, no data loading

EOF
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --reset)
            RESET=true
            shift
            ;;
        --no-data)
            NO_DATA=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
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
    echo -e "${YELLOW}⚠${NC}  $1" >&2
}

log_verbose() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${YELLOW}  [DEBUG]${NC} $1" >&2
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing=false

    # Check docker
    if command -v docker >/dev/null 2>&1; then
        local docker_version=$(docker --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        log_success "Docker installed (version ${docker_version})"
    else
        log_error "Docker is required but not installed"
        echo "  → Install from: https://docs.docker.com/get-docker/" >&2
        missing=true
    fi

    # Check docker compose
    if docker compose version >/dev/null 2>&1; then
        local compose_version=$(docker compose version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        log_success "Docker Compose installed (version ${compose_version})"
    else
        log_error "Docker Compose v2 is required but not installed"
        echo "  → Install from: https://docs.docker.com/compose/install/" >&2
        missing=true
    fi

    # Check curl
    if command -v curl >/dev/null 2>&1; then
        log_success "curl installed"
    else
        log_error "curl is required but not installed"
        missing=true
    fi

    # Check jq
    if command -v jq >/dev/null 2>&1; then
        log_success "jq installed"
    else
        log_error "jq is required but not installed"
        echo "  → Install with: brew install jq (macOS) or apt install jq (Linux)" >&2
        missing=true
    fi

    # Load .env file
    if [[ -f "${LDE_DIR}/.env" ]]; then
        source "${LDE_DIR}/.env"
        log_success ".env file loaded"
    else
        log_error ".env file not found at ${LDE_DIR}/.env"
        echo "  → Copy .env.example to .env and configure it" >&2
        missing=true
    fi

    if [[ "$missing" == true ]]; then
        exit 1
    fi

    echo ""
}

# Initialize directories
initialize_directories() {
    if [[ "$RESET" == true ]]; then
        log_info "Resetting environment (--reset flag)..."

        # Stop Hugr first if running
        if docker ps -q -f name=lde-hugr >/dev/null 2>&1; then
            log_info "Stopping Hugr container..."
            docker stop lde-hugr >/dev/null 2>&1 || true
        fi

        # Remove .local directory (PostgreSQL, Redis, MinIO, Keycloak data)
        if [[ -d "${LDE_DIR}/.local" ]]; then
            rm -rf "${LDE_DIR}/.local"
            log_success "Removed .local directory"
        fi

        # Remove all data files (DuckDB databases, schemas, etc.)
        if [[ -d "${LDE_DIR}/data" ]]; then
            log_info "Removing all data files and schemas..."
            find "${LDE_DIR}/data" -type f -name "*.duckdb*" -delete 2>/dev/null || true
            find "${LDE_DIR}/data/schemas" -mindepth 1 -delete 2>/dev/null || true
            log_success "Removed data files"
        fi
    fi

    log_info "Creating .local directory structure..."

    mkdir -p "${LDE_DIR}/.local/pg-data"
    log_success "Created .local/pg-data"

    mkdir -p "${LDE_DIR}/.local/redis-data"
    log_success "Created .local/redis-data"

    mkdir -p "${LDE_DIR}/.local/minio"
    log_success "Created .local/minio"

    mkdir -p "${LDE_DIR}/.local/keycloak"
    log_success "Created .local/keycloak"

    # Ensure data directories exist
    mkdir -p "${LDE_DIR}/data/schemas"
    log_success "Created data directories"

    echo ""
}

# Start Docker services
start_services() {
    log_info "Starting Docker services..."

    cd "${LDE_DIR}"

    if [[ "$VERBOSE" == true ]]; then
        docker compose -f docker-compose.yml up -d
    else
        docker compose -f docker-compose.yml up -d 2>&1 | grep -E '(Creating|Starting|Created|Started|Running|✔)' || true
    fi

    if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
        log_error "Failed to start Docker services"
        echo "  → Check logs with: docker compose -f ${LDE_DIR}/docker-compose.yml logs" >&2
        exit 2
    fi

    log_success "Docker services started"
    echo ""
}

# Wait for health checks
wait_for_health() {
    log_info "Verifying service health..."

    if [[ -x "${SCRIPT_DIR}/health-check.sh" ]]; then
        # Increase timeout to 300s for first startup (docker image pull) and data loading
        if "${SCRIPT_DIR}/health-check.sh" --wait 300; then
            log_success "All services healthy"
        else
            log_error "Health checks failed or timed out"
            echo "  → Check service status: docker compose -f ${LDE_DIR}/docker-compose.yml ps" >&2
            echo "  → Check logs: docker compose -f ${LDE_DIR}/docker-compose.yml logs" >&2
            exit 3
        fi
    else
        log_error "health-check.sh not found or not executable"
        exit 3
    fi

    echo ""
}

# Setup roles
setup_roles() {
    log_info "Setting up Hugr roles..."

    if [[ -x "${SCRIPT_DIR}/setup-roles.sh" ]]; then
        if "${SCRIPT_DIR}/setup-roles.sh" ${VERBOSE:+-v}; then
            log_success "Role setup complete"
        else
            log_warning "Role setup failed (may already exist)"
            echo "  → Try running manually: ${SCRIPT_DIR}/setup-roles.sh" >&2
        fi
    else
        log_warning "setup-roles.sh not found or not executable"
    fi

    echo ""
}

# Setup embedding model
setup_embedding() {
    log_info "Setting up embedding model..."

    if [[ -x "${SCRIPT_DIR}/setup-embedding.sh" ]]; then
        if "${SCRIPT_DIR}/setup-embedding.sh" ${VERBOSE:+-v}; then
            log_success "Embedding model setup complete"
        else
            log_warning "Embedding model setup failed (may already exist)"
            echo "  → Try running manually: ${SCRIPT_DIR}/setup-embedding.sh" >&2
        fi
    else
        log_warning "setup-embedding.sh not found or not executable"
    fi

    echo ""
}

# Load data
load_data() {
    if [[ "$NO_DATA" == true ]]; then
        log_info "Skipping data loading (--no-data flag)"
        echo ""
        return 0
    fi

    log_info "Loading data sources..."

    if [[ -x "${SCRIPT_DIR}/load-data.sh" ]]; then
        if "${SCRIPT_DIR}/load-data.sh" ${VERBOSE:+-v}; then
            log_success "Data loading complete"
        else
            log_error "Data loading failed"
            echo "  → Try running manually: ${SCRIPT_DIR}/load-data.sh" >&2
            echo "  → Or skip with: ./start.sh --no-data" >&2
            exit 4
        fi
    else
        log_error "load-data.sh not found or not executable"
        exit 4
    fi

    echo ""
}

# Display connection information
show_connection_info() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}Local Development Environment Ready${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  ${BLUE}Hugr GraphQL:${NC}  http://localhost:19000/query"
    echo -e "  ${BLUE}Keycloak:${NC}      http://localhost:19005"
    echo -e "  ${BLUE}MinIO Console:${NC} http://localhost:19004"
    echo ""
    echo -e "  ${BLUE}Test Credentials:${NC}"
    echo -e "    • admin@example.com    / admin123    (admin role)"
    echo -e "    • analyst@example.com  / analyst123  (analyst role)"
    echo -e "    • viewer@example.com   / viewer123   (viewer role)"
    echo ""

    if [[ "$NO_DATA" == false ]]; then
        echo -e "  ${BLUE}Data Sources:${NC}"
        echo -e "    • Synthea synthetic patient data"
        echo -e "    • Open Payments 2023 dataset"
        echo ""
    fi

    echo -e "  ${BLUE}To stop:${NC} ${LDE_DIR}/scripts/stop.sh"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Trap for cleanup on signals
trap 'echo -e "\n${RED}✗${NC} Interrupted. Run ./stop.sh to clean up if needed." >&2; exit 130' INT TERM

# Main execution
main() {
    echo ""
    check_prerequisites
    initialize_directories
    start_services
    wait_for_health
    setup_roles
    setup_embedding
    load_data
    show_connection_info
}

main "$@"
