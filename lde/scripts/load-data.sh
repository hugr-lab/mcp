#!/usr/bin/env bash
set -euo pipefail

# load-data.sh - Unified data loading script for all LDE data sources
# Handles data generation, database setup, and registration with Hugr

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
DATA_DIR="${LDE_DIR}/data"
DATA_LOADERS_DIR="${LDE_DIR}/data-loaders"

# Default values
FORCE=false
VERBOSE=false
SKIP_NORTHWIND=false
SKIP_SYNTHEA=false
SKIP_OPENPAYMENTS=false
SKIP_OWM=false

# Load environment variables
if [[ -f "${LDE_DIR}/.env" ]]; then
    set -a
    source "${LDE_DIR}/.env"
    set +a
fi

# Configuration
SECRET_KEY="${SECRET_KEY:-local-dev-secret-key-change-in-production}"
HUGR_URL="http://localhost:19000/query"
SYNTHEA_COUNT="${SYNTHEA_PATIENT_COUNT:-100}"
SYNTHEA_STATE="${SYNTHEA_STATE:-Massachusetts}"

# Track start time
START_TIME=$(date +%s)

# Usage function
show_help() {
    cat << EOF
Usage: load-data.sh [OPTIONS]

Unified data loading script for all LDE data sources.
Generates/prepares data and registers with Hugr.

Options:
  --force               Force regenerate data even if exists
  -v, --verbose         Enable verbose output
  --skip-northwind      Skip Northwind (PostgreSQL)
  --skip-synthea        Skip Synthea generation
  --skip-openpayments   Skip Open Payments
  --skip-owm            Skip OpenWeatherMap
  --help                Show this help message

Note: Embedding model is automatically registered during startup by start.sh

Environment Variables:
  SECRET_KEY                  Hugr authentication secret
  SYNTHEA_PATIENT_COUNT       Number of patients to generate (default: 100)
  SYNTHEA_STATE               US state for generation (default: Massachusetts)
  OPENWEATHERMAP_API_KEY      API key for OpenWeatherMap

Examples:
  ./load-data.sh                    # Load all data sources
  ./load-data.sh --force            # Force regenerate all data
  ./load-data.sh --skip-synthea     # Skip synthea generation
  ./load-data.sh -v                 # Verbose output

EOF
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --skip-northwind)
            SKIP_NORTHWIND=true
            shift
            ;;
        --skip-synthea)
            SKIP_SYNTHEA=true
            shift
            ;;
        --skip-openpayments)
            SKIP_OPENPAYMENTS=true
            shift
            ;;
        --skip-owm)
            SKIP_OWM=true
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

# Function to execute GraphQL mutation
execute_graphql() {
    local query="$1"
    local description="$2"

    log_info "$description"

    local response=$(curl -s -X POST "$HUGR_URL" \
        -H "x-hugr-secret: $SECRET_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"query\":\"$query\"}")

    log_verbose "Response: $response"

    if echo "$response" | jq -e '.errors' > /dev/null 2>&1; then
        local error_msg=$(echo "$response" | jq -r '.errors[0].message')
        # Check if it's a duplicate key error (already exists)
        if echo "$error_msg" | grep -q "Duplicate key\|already exists"; then
            log_success "Already exists (skipped)"
            return 0
        else
            log_error "Failed: $error_msg"
            return 1
        fi
    else
        log_success "Success"
        return 0
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check SECRET_KEY
    if [[ -z "${SECRET_KEY:-}" ]]; then
        log_error "SECRET_KEY not found in .env file"
        exit 1
    fi

    # Check docker
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker is required"
        exit 1
    fi

    # Check DuckDB CLI
    if ! command -v duckdb >/dev/null 2>&1; then
        log_warning "DuckDB CLI not found (required for Synthea/Open Payments)"
    fi

    # Check jq
    if ! command -v jq >/dev/null 2>&1; then
        log_error "jq is required"
        exit 1
    fi

    # Create data directories
    mkdir -p "${DATA_DIR}"
    mkdir -p "${DATA_DIR}/schemas"

    log_success "Prerequisites checked"
    echo ""
}

# Verify Hugr is healthy
verify_hugr_health() {
    log_info "Verifying Hugr is healthy..."

    # Check health endpoint (port 19006 maps to internal port 14000)
    if curl -sf http://localhost:19006/health >/dev/null 2>&1; then
        log_success "Hugr is healthy"
    else
        log_error "Hugr is not healthy or not running"
        echo "  → Start services: ${LDE_DIR}/scripts/start.sh" >&2
        exit 2
    fi

    echo ""
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 1. NORTHWIND (PostgreSQL)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

load_northwind() {
    if [[ "$SKIP_NORTHWIND" == true ]]; then
        log_verbose "Skipping Northwind (--skip-northwind flag)"
        return 0
    fi

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}1. Northwind (PostgreSQL)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Create database
    log_info "Creating northwind database..."
    docker exec lde-postgres psql -U hugr -d hugr -c "CREATE DATABASE northwind;" 2>/dev/null || log_verbose "Database may already exist"

    # Load dump
    log_info "Loading northwind dump..."
    docker exec -i lde-postgres psql -U hugr -d northwind < "${DATA_LOADERS_DIR}/northwind/northwind_dump.sql" > /dev/null 2>&1
    log_success "Northwind data loaded"

    # Copy schema
    log_info "Copying schema files..."
    mkdir -p "${DATA_DIR}/schemas/northwind"
    cp "${DATA_LOADERS_DIR}/northwind/schema.graphql" "${DATA_DIR}/schemas/northwind/"
    log_success "Schema copied"

    # Register data source
    execute_graphql \
        "mutation { core { insert_data_sources(data: { name: \\\"northwind\\\", type: \\\"postgres\\\", prefix: \\\"nw\\\", as_module: true, path: \\\"postgres://hugr:hugr@postgres:5432/northwind\\\", read_only: false, description: \\\"Northwind database example\\\", catalogs: [{ name: \\\"northwind\\\", type: \\\"uri\\\", description: \\\"Northwind schema\\\", path: \\\"/data/schemas/northwind\\\" }] }) { name } } }" \
        "Registering Northwind data source"

    # Load data source
    execute_graphql \
        "mutation { function { core { load_data_source(name: \\\"northwind\\\") { success message } } } }" \
        "Loading Northwind data source"

    echo ""
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 2. SYNTHEA (DuckDB with self-defined schema)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

load_synthea() {
    if [[ "$SKIP_SYNTHEA" == true ]]; then
        log_verbose "Skipping Synthea (--skip-synthea flag)"
        return 0
    fi

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}2. Synthea (DuckDB - self-defined schema)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    local synthea_db="${DATA_DIR}/synthea.duckdb"
    local synthea_script="${DATA_LOADERS_DIR}/synthea/generate-and-load.sh"

    # Check if data already exists
    if [[ -f "$synthea_db" ]] && [[ "$FORCE" != true ]]; then
        log_success "Synthea data already exists at ${synthea_db}"
        log_info "Use --force to regenerate"
    else
        # Generate and load using the proper script
        log_info "Generating Synthea data (${SYNTHEA_COUNT} patients from ${SYNTHEA_STATE})..."

        if [[ ! -x "$synthea_script" ]]; then
            chmod +x "$synthea_script"
        fi

        # Run the generation script
        cd "${DATA_LOADERS_DIR}/synthea"

        # Use absolute paths since we're changing directories
        local abs_synthea_db="$(cd "${DATA_DIR}" && pwd)/synthea.duckdb"
        local temp_output="$(cd "${DATA_DIR}" && pwd)/.tmp/synthea"
        local gen_args="-c ${SYNTHEA_COUNT} -s ${SYNTHEA_STATE} -d ${abs_synthea_db} -o ${temp_output}"

        if [[ "$VERBOSE" == true ]]; then
            bash "$synthea_script" $gen_args
        else
            bash "$synthea_script" $gen_args > /dev/null 2>&1
        fi

        cd "${LDE_DIR}"

        log_success "Synthea data generated"
    fi

    # Register data source
    execute_graphql \
        "mutation { core { insert_data_sources(data: { name: \\\"synthea\\\", type: \\\"duckdb\\\", prefix: \\\"synthea\\\", as_module: true, path: \\\"/data/synthea.duckdb\\\", read_only: true, self_defined: true, description: \\\"Synthea synthetic patient data\\\" }) { name } } }" \
        "Registering Synthea data source"

    # Load data source
    execute_graphql \
        "mutation { function { core { load_data_source(name: \\\"synthea\\\") { success message } } } }" \
        "Loading Synthea data source"

    echo ""
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 3. OPEN PAYMENTS (DuckDB with GraphQL schema)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

load_openpayments() {
    if [[ "$SKIP_OPENPAYMENTS" == true ]]; then
        log_verbose "Skipping Open Payments (--skip-openpayments flag)"
        return 0
    fi

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}3. Open Payments (DuckDB - with GraphQL schema)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    local op_db="${DATA_DIR}/openpayments.duckdb"
    local op_setup="${DATA_LOADERS_DIR}/open-payments/setup.sh"

    # Check if data already exists
    if [[ -f "$op_db" ]] && [[ "$FORCE" != true ]]; then
        log_success "Open Payments data already exists at ${op_db}"
        log_info "Use --force to regenerate"
    else
        # Run setup script to download and load real data from CMS
        log_info "Downloading and loading Open Payments data from CMS..."
        log_info "This will download ~1GB and may take several minutes..."

        cd "${DATA_LOADERS_DIR}/open-payments"

        # Run setup with force flag and absolute paths
        local abs_op_db="$(cd "${DATA_DIR}" && pwd)/openpayments.duckdb"
        local temp_data_dir="$(cd "${DATA_DIR}" && pwd)/.tmp/openpayments"
        local setup_args="--force --db-file ${abs_op_db} --data-dir ${temp_data_dir}"

        if bash "$op_setup" $setup_args 2>&1 | tee /tmp/openpayments-setup.log; then
            log_success "Open Payments data downloaded and loaded"
        else
            log_error "Failed to setup Open Payments data"
            log_error "Check /tmp/openpayments-setup.log for details"
            cd "${LDE_DIR}"
            return 1
        fi

        cd "${LDE_DIR}"
    fi

    # Copy schemas
    log_info "Copying schema files..."
    mkdir -p "${DATA_DIR}/schemas/openpayments"
    cp "${DATA_LOADERS_DIR}/open-payments/schemas/schema.graphql" "${DATA_DIR}/schemas/openpayments/"
    cp "${DATA_LOADERS_DIR}/open-payments/schemas/extra.graphql" "${DATA_DIR}/schemas/openpayments/"
    log_success "Schemas copied"

    # Register data source
    execute_graphql \
        "mutation { core { insert_data_sources(data: { name: \\\"openpayments\\\", type: \\\"duckdb\\\", prefix: \\\"op\\\", as_module: true, path: \\\"/data/openpayments.duckdb\\\", read_only: true, description: \\\"Open Payments 2023 data\\\", catalogs: [{ name: \\\"openpayments\\\", type: \\\"uri\\\", description: \\\"Open Payments schema\\\", path: \\\"/data/schemas/openpayments\\\" }] }) { name } } }" \
        "Registering Open Payments data source"

    # Load data source
    execute_graphql \
        "mutation { function { core { load_data_source(name: \\\"openpayments\\\") { success message } } } }" \
        "Loading Open Payments data source"

    echo ""
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 4. OPENWEATHERMAP (HTTP REST API)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

load_openweathermap() {
    if [[ "$SKIP_OWM" == true ]]; then
        log_verbose "Skipping OpenWeatherMap (--skip-owm flag)"
        return 0
    fi

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}4. OpenWeatherMap (HTTP REST API)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [[ -z "${OPENWEATHERMAP_API_KEY:-}" ]]; then
        log_warning "OPENWEATHERMAP_API_KEY not set in .env (skipping)"
        echo ""
        return 0
    fi

    # Copy spec files
    log_info "Copying spec files..."
    mkdir -p "${DATA_DIR}/schemas/openweathermap"
    cp "${DATA_LOADERS_DIR}/openweathermap/spec.yaml" "${DATA_DIR}/schemas/openweathermap/"
    cp "${DATA_LOADERS_DIR}/openweathermap/schema.graphql" "${DATA_DIR}/schemas/openweathermap/"
    log_success "Spec files copied"

    # Register data source with complex path
    log_info "Registering OpenWeatherMap data source"

    local response=$(curl -s -X POST "$HUGR_URL" \
        -H "x-hugr-secret: $SECRET_KEY" \
        -H "Content-Type: application/json" \
        -d "{
            \"query\": \"mutation { core { insert_data_sources(data: { name: \\\"owm\\\", type: \\\"http\\\", prefix: \\\"owm\\\", as_module: true, read_only: true, self_defined: true, description: \\\"OpenWeatherMap REST API\\\", path: \\\"https://api.openweathermap.org?x-hugr-spec-path=/data/schemas/openweathermap/spec.yaml&x-hugr-security={\\\\\\\"schema_name\\\\\\\":\\\\\\\"owm\\\\\\\",\\\\\\\"api_key\\\\\\\":\\\\\\\"${OPENWEATHERMAP_API_KEY}\\\\\\\"}\\\" }) { name } } }\"
        }")

    log_verbose "Response: $response"

    if echo "$response" | jq -e '.data.core.insert_data_sources' > /dev/null 2>&1; then
        log_success "Success"
    else
        local error=$(echo "$response" | jq -r '.errors[0].message // "Unknown error"')
        if echo "$error" | grep -q "Duplicate key\|already exists"; then
            log_success "Already exists (skipped)"
        else
            log_error "Failed: $error"
        fi
    fi

    # Load data source
    execute_graphql \
        "mutation { function { core { load_data_source(name: \\\"owm\\\") { success message } } } }" \
        "Loading OpenWeatherMap data source"

    echo ""
}

# Verify loaded data sources
verify_data_sources() {
    log_info "Verifying data sources..."

    local response=$(curl -s -X POST "$HUGR_URL" \
        -H "x-hugr-secret: $SECRET_KEY" \
        -H "Content-Type: application/json" \
        -d '{"query": "{ core { data_sources { name type } } }"}')

    local count=$(echo "$response" | jq '.data.core.data_sources | length')

    log_success "Loaded ${count} data sources"

    if [[ "$VERBOSE" == true ]]; then
        echo "$response" | jq -r '.data.core.data_sources[] | "  • \(.name) (\(.type))"'
    fi

    echo ""
}

# Display summary
show_summary() {
    local end_time=$(date +%s)
    local elapsed=$((end_time - START_TIME))
    local elapsed_min=$((elapsed / 60))
    local elapsed_sec=$((elapsed % 60))

    # Calculate disk usage
    local total_size=0
    [[ -f "${DATA_DIR}/synthea.duckdb" ]] && total_size=$((total_size + $(stat -f%z "${DATA_DIR}/synthea.duckdb" 2>/dev/null || echo 0)))
    [[ -f "${DATA_DIR}/openpayments.duckdb" ]] && total_size=$((total_size + $(stat -f%z "${DATA_DIR}/openpayments.duckdb" 2>/dev/null || echo 0)))
    [[ -f "${DATA_DIR}/core.duckdb" ]] && total_size=$((total_size + $(stat -f%z "${DATA_DIR}/core.duckdb" 2>/dev/null || echo 0)))

    local total_size_mb=$((total_size / 1024 / 1024))

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}Data Loading Complete${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  ${BLUE}Disk Usage:${NC}      ${total_size_mb} MB"
    echo -e "  ${BLUE}Load Time:${NC}       ${elapsed_min}m ${elapsed_sec}s"
    echo ""
    echo "Loaded data sources:"
    echo "  • northwind:     PostgreSQL sample database"
    echo "  • synthea:       Synthetic patient data"
    echo "  • openpayments:  Healthcare payments data"
    echo "  • owm:           OpenWeatherMap API (if API key provided)"
    echo ""
    echo "Note: emb_gemma (embedding model) is registered during startup"
    echo ""
    echo "Query endpoint: http://localhost:19000/query"
    echo "Admin UI:       http://localhost:19000/admin"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Main execution
main() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Hugr Data Loading Script${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    check_prerequisites
    verify_hugr_health

    load_northwind
    load_synthea
    load_openpayments
    load_openweathermap

    verify_data_sources
    show_summary
}

# Run main function with all arguments
main "$@"
