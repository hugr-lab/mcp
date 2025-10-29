#!/usr/bin/env bash

# lde.sh - Local Development Environment Manager
#
# A convenience wrapper for managing the Hugr LDE from the project root.
# Provides commands to start, stop, cleanup, and check health of the environment.

set -euo pipefail

# Color codes (only if stdout is a terminal)
if [[ -t 1 ]]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    GREEN=''
    RED=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Script directory and LDE directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LDE_DIR="${SCRIPT_DIR}/lde"

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
Usage: ./lde.sh <command> [options]

Commands:
  start [options]     Start the LDE environment
  stop [options]      Stop the LDE environment
  cleanup [options]   Clean up the LDE environment
  health [options]    Check health of LDE services
  load [options]      Load data sources
  logs [service]      Show logs for all or specific service
  help                Show this help message

Options are passed through to the underlying scripts.

Examples:
  ./lde.sh start              # Start LDE (auto-detects if data exists)
  ./lde.sh start --verbose    # Start with verbose output
  ./lde.sh stop               # Stop all services
  ./lde.sh cleanup --force    # Clean up without confirmation
  ./lde.sh health --wait      # Wait for services to become healthy
  ./lde.sh load --force       # Force reload all data
  ./lde.sh logs hugr          # Show logs for hugr service

For detailed options, run:
  ./lde.sh start --help
  ./lde.sh cleanup --help
  ./lde.sh health --help
  ./lde.sh load --help

See lde.md for comprehensive documentation.

EOF
}

# Check if data exists
data_exists() {
    local data_dir="${LDE_DIR}/data"

    # Check for key database files
    if [[ -f "${data_dir}/core.duckdb" ]] && \
       [[ -f "${data_dir}/synthea.duckdb" ]] && \
       [[ -f "${data_dir}/openpayments.duckdb" ]]; then
        return 0
    fi

    return 1
}

# Start command
cmd_start() {
    local args=("$@")

    # Check if data already exists and --no-data flag is not set
    if data_exists && [[ ! " ${args[*]} " =~ " --no-data " ]]; then
        log_info "Data already exists, starting without data loading..."
        args+=("--no-data")
    fi

    # Execute start script from LDE directory
    cd "${LDE_DIR}"
    bash "${LDE_DIR}/scripts/start.sh" "${args[@]+"${args[@]}"}"
}

# Stop command
cmd_stop() {
    cd "${LDE_DIR}"
    bash "${LDE_DIR}/scripts/stop.sh" "$@"
}

# Cleanup command
cmd_cleanup() {
    cd "${LDE_DIR}"
    bash "${LDE_DIR}/scripts/cleanup.sh" "$@"
}

# Health check command
cmd_health() {
    cd "${LDE_DIR}"
    bash "${LDE_DIR}/scripts/health-check.sh" "$@"
}

# Load data command
cmd_load() {
    cd "${LDE_DIR}"
    bash "${LDE_DIR}/scripts/load-data.sh" "$@"
}

# Logs command
cmd_logs() {
    local service="${1:-}"

    cd "${LDE_DIR}"

    if [[ -n "$service" ]]; then
        docker compose logs -f "$service"
    else
        docker compose logs -f
    fi
}

# Main execution
main() {
    if [[ $# -eq 0 ]]; then
        show_usage
        exit 1
    fi

    local command="$1"
    shift

    case "$command" in
        start)
            cmd_start "$@"
            ;;
        stop)
            cmd_stop "$@"
            ;;
        cleanup)
            cmd_cleanup "$@"
            ;;
        health)
            cmd_health "$@"
            ;;
        load)
            cmd_load "$@"
            ;;
        logs)
            cmd_logs "$@"
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            log_error "Unknown command: $command"
            echo ""
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
