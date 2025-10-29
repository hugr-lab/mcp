#!/usr/bin/env bash
set -euo pipefail

# cleanup.sh - Complete cleanup of LDE environment
# This script stops all services, removes containers, and clears all data

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

# Show usage
show_usage() {
    cat << EOF
Usage: cleanup.sh [OPTIONS]

Complete cleanup of the LDE environment. This will:
  1. Stop and remove all Docker containers
  2. Remove Docker volumes and networks
  3. Clear the data/ directory (databases, schemas)
  4. Clear the .local/ directory (service data)

Options:
  --keep-data       Keep data/ directory (preserve databases)
  --keep-local      Keep .local/ directory (preserve service data)
  --keep-images     Keep Docker images (don't prune)
  --force           Skip confirmation prompt
  -h, --help        Show this help message

Examples:
  ./cleanup.sh                  # Full cleanup with confirmation
  ./cleanup.sh --force          # Full cleanup without confirmation
  ./cleanup.sh --keep-data      # Clean everything except databases
  ./cleanup.sh --keep-local     # Clean everything except service data

⚠️  WARNING: This will permanently delete all data unless --keep-* flags are used!

EOF
    exit 0
}

# Parse command line arguments
KEEP_DATA=false
KEEP_LOCAL=false
KEEP_IMAGES=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --keep-data)
            KEEP_DATA=true
            shift
            ;;
        --keep-local)
            KEEP_LOCAL=true
            shift
            ;;
        --keep-images)
            KEEP_IMAGES=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        -h|--help)
            show_usage
            ;;
        *)
            log_error "Unknown option: $1"
            echo "  Use --help for usage information" >&2
            exit 1
            ;;
    esac
done

# Show warning and confirmation
show_warning() {
    echo ""
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}⚠  WARNING: DESTRUCTIVE OPERATION${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "This will permanently delete:"
    echo ""

    if [[ "$KEEP_DATA" == false ]]; then
        echo -e "  ${RED}✗${NC} data/ directory (databases, schemas)"
        echo "     - core.duckdb (~2.5 MB)"
        echo "     - synthea.duckdb (~24 MB)"
        echo "     - openpayments.duckdb (~524 KB)"
        echo "     - All GraphQL schemas"
    else
        echo -e "  ${GREEN}✓${NC} data/ directory (PRESERVED)"
    fi

    echo ""

    if [[ "$KEEP_LOCAL" == false ]]; then
        echo -e "  ${RED}✗${NC} .local/ directory (service data)"
        echo "     - PostgreSQL data (northwind database)"
        echo "     - Redis cache"
        echo "     - MinIO objects"
        echo "     - Keycloak realms and users"
    else
        echo -e "  ${GREEN}✓${NC} .local/ directory (PRESERVED)"
    fi

    echo ""
    echo -e "  ${RED}✗${NC} All Docker containers"
    echo -e "  ${RED}✗${NC} Docker networks (lde-network)"

    if [[ "$KEEP_IMAGES" == false ]]; then
        echo -e "  ${RED}✗${NC} Unused Docker images"
    else
        echo -e "  ${GREEN}✓${NC} Docker images (PRESERVED)"
    fi

    echo ""
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Confirm action
confirm_cleanup() {
    if [[ "$FORCE" == true ]]; then
        return 0
    fi

    show_warning

    read -p "Are you sure you want to proceed? (yes/no): " -r
    echo
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log_info "Cleanup cancelled"
        exit 0
    fi
}

# Stop and remove containers
stop_containers() {
    log_info "Stopping and removing Docker containers..."

    cd "${LDE_DIR}"

    # Stop containers gracefully
    if docker compose -f docker-compose.yml ps -q 2>/dev/null | grep -q .; then
        docker compose -f docker-compose.yml down --volumes 2>/dev/null || true
        log_success "Containers stopped and removed"
    else
        log_success "No running containers found"
    fi

    # Remove any orphaned containers
    local orphaned=$(docker ps -a --filter "name=lde-" -q 2>/dev/null || true)
    if [[ -n "$orphaned" ]]; then
        echo "$orphaned" | xargs docker rm -f >/dev/null 2>&1 || true
        log_success "Removed orphaned containers"
    fi

    # Remove network
    if docker network ls --filter "name=lde-network" -q 2>/dev/null | grep -q .; then
        docker network rm lde-network >/dev/null 2>&1 || true
        log_success "Removed network"
    fi

    echo ""
}

# Clear data directory
clear_data() {
    if [[ "$KEEP_DATA" == true ]]; then
        log_info "Keeping data/ directory (--keep-data flag)"
        echo ""
        return 0
    fi

    log_info "Clearing data/ directory..."

    local data_dir="${LDE_DIR}/data"

    if [[ -d "$data_dir" ]]; then
        # Calculate size before deletion
        local size=$(du -sh "$data_dir" 2>/dev/null | cut -f1 || echo "unknown")

        # Remove all contents except .gitkeep
        find "$data_dir" -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true

        log_success "Cleared data/ directory ($size freed)"
    else
        log_success "data/ directory does not exist"
    fi

    echo ""
}

# Clear .local directory
clear_local() {
    if [[ "$KEEP_LOCAL" == true ]]; then
        log_info "Keeping .local/ directory (--keep-local flag)"
        echo ""
        return 0
    fi

    log_info "Clearing .local/ directory..."

    local local_dir="${LDE_DIR}/.local"

    if [[ -d "$local_dir" ]]; then
        # Calculate size before deletion
        local size=$(du -sh "$local_dir" 2>/dev/null | cut -f1 || echo "unknown")

        # Remove entire directory
        rm -rf "$local_dir"

        # Recreate empty directory
        mkdir -p "$local_dir"

        log_success "Cleared .local/ directory ($size freed)"
    else
        log_success ".local/ directory does not exist"
    fi

    echo ""
}

# Prune Docker images
prune_images() {
    if [[ "$KEEP_IMAGES" == true ]]; then
        log_info "Keeping Docker images (--keep-images flag)"
        echo ""
        return 0
    fi

    log_info "Pruning unused Docker images..."

    # Prune dangling images only (not used by any container)
    docker image prune -f >/dev/null 2>&1 || true

    log_success "Pruned unused images"

    echo ""
}

# Show summary
show_summary() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}Cleanup Complete${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "The LDE environment has been cleaned up."
    echo ""
    echo "Next steps:"
    echo "  • Start fresh: ${LDE_DIR}/scripts/start.sh" or "{LDE_DIR}/scripts/start.sh --no-data"
    echo "  • Load data:   ${LDE_DIR}/scripts/load-data.sh"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Main execution
main() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}LDE Cleanup Script${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    confirm_cleanup

    echo ""
    log_info "Starting cleanup process..."
    echo ""

    stop_containers
    clear_data
    clear_local
    prune_images

    show_summary
}

# Run main function
main "$@"
