#!/bin/bash
set -euo pipefail

# setup-roles.sh - Setup Hugr roles with appropriate permissions
# Creates admin, analyst, and viewer roles in Hugr

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

# Load environment variables
if [[ -f "${LDE_DIR}/.env" ]]; then
    source "${LDE_DIR}/.env"
fi

# Default values
VERBOSE=false
HUGR_URL="${HUGR_URL:-http://localhost:19000/query}"
SECRET_KEY="${SECRET_KEY:-local-dev-secret-key-change-in-production}"

# Usage function
show_help() {
    cat << EOF
Usage: setup-roles.sh [OPTIONS]

Setup Hugr roles with appropriate permissions.

Options:
  -v, --verbose     Enable verbose output
  --help            Show this help message

Roles Created:
  • admin    - Full access (system role)
  • analyst  - Read-only access, cannot mutate data
  • viewer   - Read-only access, no core module access

Environment Variables:
  HUGR_URL     - Hugr GraphQL endpoint (default: http://localhost:19000/query)
  SECRET_KEY   - Hugr secret key for authentication

Examples:
  ./setup-roles.sh              # Setup all roles
  ./setup-roles.sh --verbose    # Detailed output

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

# Check if Hugr is available
check_hugr() {
    log_verbose "Checking Hugr availability at ${HUGR_URL}..."

    if ! curl -sf "${HUGR_URL}" -X POST \
        -H "Content-Type: application/json" \
        -H "x-hugr-secret: ${SECRET_KEY}" \
        -d '{"query":"{ __schema { queryType { name } } }"}' >/dev/null 2>&1; then
        log_error "Hugr is not available at ${HUGR_URL}"
        log_error "Make sure Hugr is running and SECRET_KEY is correct"
        return 1
    fi

    log_verbose "Hugr is available"
    return 0
}

# Setup analyst role
setup_analyst_role() {
    log_info "Setting up analyst role..."

    local query='mutation { core { insert_roles(data: { name: "analyst", description: "Data analyst with read-only access", permissions: [{ type_name: "Mutation", field_name: "*", disabled: true, hidden: true }] }) { name description } } }'

    log_verbose "Sending mutation..."

    local response=$(curl -s -X POST "${HUGR_URL}" \
        -H "Content-Type: application/json" \
        -H "x-hugr-secret: ${SECRET_KEY}" \
        --data-binary @- <<EOF
{"query": $(echo "$query" | jq -Rs .)}
EOF
)

    log_verbose "Response: ${response}"

    if echo "$response" | jq -e '.data.core.insert_roles.name' >/dev/null 2>&1; then
        log_success "Analyst role created"
        return 0
    elif echo "$response" | jq -e '.errors[] | select(.message | contains("already exists"))' >/dev/null 2>&1; then
        log_success "Analyst role already exists"
        return 0
    else
        log_error "Failed to create analyst role"
        log_error "Response: ${response}"
        return 1
    fi
}

# Setup viewer role
setup_viewer_role() {
    log_info "Setting up viewer role..."

    local query='mutation { core { insert_roles(data: { name: "viewer", description: "Read-only viewer without core module access", permissions: [{ type_name: "Mutation", field_name: "*", disabled: true, hidden: true }, { type_name: "Query", field_name: "core", disabled: true, hidden: true }] }) { name description } } }'

    log_verbose "Sending mutation..."

    local response=$(curl -s -X POST "${HUGR_URL}" \
        -H "Content-Type: application/json" \
        -H "x-hugr-secret: ${SECRET_KEY}" \
        --data-binary @- <<EOF
{"query": $(echo "$query" | jq -Rs .)}
EOF
)

    log_verbose "Response: ${response}"

    if echo "$response" | jq -e '.data.core.insert_roles.name' >/dev/null 2>&1; then
        log_success "Viewer role created"
        return 0
    elif echo "$response" | jq -e '.errors[] | select(.message | contains("already exists"))' >/dev/null 2>&1; then
        log_success "Viewer role already exists"
        return 0
    else
        log_error "Failed to create viewer role"
        log_error "Response: ${response}"
        return 1
    fi
}

# Verify roles
verify_roles() {
    log_info "Verifying roles..."

    local query='{ core { roles { name description permissions { type_name field_name disabled hidden } } } }'

    local response=$(curl -s -X POST "${HUGR_URL}" \
        -H "Content-Type: application/json" \
        -H "x-hugr-secret: ${SECRET_KEY}" \
        -d "{\"query\": \"$(echo "$query" | tr -d '\n')\"}")

    log_verbose "Response: ${response}"

    local role_count=$(echo "$response" | jq -r '.data.core.roles | length' 2>/dev/null || echo "0")

    if [[ "$role_count" -ge 2 ]]; then
        log_success "Roles verified (${role_count} roles configured)"

        if [[ "$VERBOSE" == true ]]; then
            echo ""
            echo -e "${BLUE}Configured Roles:${NC}"
            echo "$response" | jq -r '.data.core.roles[] | "  • \(.name): \(.description)"' 2>/dev/null || true
            echo ""
        fi
        return 0
    else
        log_error "Role verification failed (expected at least 2 roles, found ${role_count})"
        return 1
    fi
}

# Main execution
main() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Hugr Role Setup${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if ! check_hugr; then
        exit 1
    fi

    echo ""

    setup_analyst_role || true
    setup_viewer_role || true

    echo ""

    if verify_roles; then
        echo ""
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}Role Setup Complete${NC}"
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        exit 0
    else
        echo ""
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}Role Setup Failed${NC}"
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        exit 1
    fi
}

main "$@"
