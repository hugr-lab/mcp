#!/usr/bin/env bash
set -euo pipefail

# stop.sh - Stop the Local Development Environment
# Gracefully shuts down all services while preserving data volumes

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

# Default values
VERBOSE=false

# Usage function
show_help() {
    cat << EOF
Usage: stop.sh [OPTIONS]

Stop the Local Development Environment (preserves data volumes).

Options:
  -v, --verbose  Enable verbose output
  --help         Show this help message

Exit Codes:
  0  - Success (all services stopped)
  1  - Docker compose failure

Examples:
  ./stop.sh           # Normal shutdown
  ./stop.sh --verbose # Detailed output

Note: Data is preserved in .local/ directory. Use './start.sh' to restart.
      Use './start.sh --reset' to wipe data and start fresh.

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

log_verbose() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${YELLOW}  [DEBUG]${NC} $1" >&2
    fi
}

# Stop Docker services
stop_services() {
    log_info "Stopping Docker services..."

    cd "${LDE_DIR}"

    if [[ "$VERBOSE" == true ]]; then
        docker compose -f docker-compose.yml down
    else
        docker compose -f docker-compose.yml down 2>&1 | grep -E '(Stopping|Stopped|Removing|Removed|✔)' || true
    fi

    local exit_code=${PIPESTATUS[0]}

    if [[ $exit_code -ne 0 ]]; then
        log_error "Failed to stop Docker services"
        echo "  → Check running containers: docker ps" >&2
        echo "  → Force stop: docker compose -f ${LDE_DIR}/docker-compose.yml down -v" >&2
        exit 1
    fi

    log_success "All services stopped"
    echo ""
}

# Display status
show_status() {
    log_success "Data persisted in ${LDE_DIR}/.local/"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "  Services have been stopped."
    echo -e "  Your data is preserved and will be available on next start."
    echo ""
    echo -e "  ${BLUE}To restart:${NC} ${LDE_DIR}/scripts/start.sh"
    echo -e "  ${BLUE}To reset:${NC}   ${LDE_DIR}/scripts/start.sh --reset"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Verify services are stopped
verify_stopped() {
    log_verbose "Verifying all containers are stopped..."

    local running_containers=$(docker ps --filter "name=lde-" --format "{{.Names}}" 2>/dev/null || true)

    if [[ -z "$running_containers" ]]; then
        log_verbose "All LDE containers stopped"
        return 0
    else
        echo -e "${YELLOW}⚠${NC}  Some containers still running:" >&2
        echo "$running_containers" | sed 's/^/    /' >&2
        return 0  # Not fatal, just informational
    fi
}

# Trap for cleanup on signals
trap 'echo -e "\n${RED}✗${NC} Interrupted" >&2; exit 130' INT TERM

# Main execution
main() {
    echo ""
    stop_services
    verify_stopped
    show_status
}

main "$@"
