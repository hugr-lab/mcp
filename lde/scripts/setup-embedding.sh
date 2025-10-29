#!/usr/bin/env bash

# setup-embedding.sh - Register embedding model data source with Hugr
#
# This script registers the EmbeddingGemma model as a data source if it doesn't already exist.
# It's called automatically during startup by start.sh

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LDE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Load environment
if [[ -f "${LDE_DIR}/.env" ]]; then
    set -a
    source "${LDE_DIR}/.env"
    set +a
fi

# Configuration
HUGR_URL="${HUGR_URL:-http://localhost:19000/query}"
SECRET_KEY="${SECRET_KEY:-local-dev-secret-key-change-in-production}"
EMBEDDINGS_URL="${EMBEDDINGS_URL:-http://host.docker.internal:19007}"
EMBEDDINGS_MODEL="${EMBEDDINGS_MODEL:-ai/embeddinggemma}"

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

# Parse arguments
VERBOSE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Log verbose output
log_verbose() {
    if [[ "$VERBOSE" == true ]]; then
        echo "  [DEBUG] $1"
    fi
}

# Main setup
main() {
    log_info "Setting up embedding model data source..."

    # Check if Hugr is accessible
    if ! curl -sf "${HUGR_URL}" >/dev/null 2>&1; then
        log_error "Cannot connect to Hugr at ${HUGR_URL}"
        exit 1
    fi

    # Register embedding model data source
    log_verbose "Registering EmbeddingGemma data source"

    local response=$(curl -s -X POST "$HUGR_URL" \
        -H "x-hugr-secret: $SECRET_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"mutation { core { insert_data_sources(data: { name: \\\"emb_gemma\\\", type: \\\"embedding\\\", prefix: \\\"\\\", as_module: false, description: \\\"Text embedding model (EmbeddingGemma 308M)\\\", self_defined: false, read_only: false, path: \\\"[\$EMBEDDINGS_URL]/embeddings?model=[\$EMBEDDINGS_MODEL]\\\" }) { name } } }\"}")

    log_verbose "Response: $response"

    # Check response
    if echo "$response" | jq -e '.data.core.insert_data_sources.name' > /dev/null 2>&1; then
        log_success "Embedding model data source registered"
        return 0
    elif echo "$response" | jq -e '.errors[] | select(.message | contains("Duplicate key"))' > /dev/null 2>&1; then
        log_success "Embedding model data source already exists"
        return 0
    else
        local error=$(echo "$response" | jq -r '.errors[0].message // "Unknown error"')
        log_error "Failed to register embedding model: $error"
        return 1
    fi
}

# Run main function
main "$@"
