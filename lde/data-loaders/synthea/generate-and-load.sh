#!/bin/bash

# Synthea Data Generation and DuckDB Setup Script
# This script builds the Docker image, generates synthetic healthcare data,
# and loads it into a DuckDB database.

set -e  # Exit on any error

# Configuration
DEFAULT_STATE="Massachusetts"
DEFAULT_COUNT=100
DEFAULT_DB_NAME="synthea.duckdb"
OUTPUT_DIR="output"
DOCKER_IMAGE="synthea"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -s, --state STATE        US state for data generation (default: $DEFAULT_STATE)"
    echo "  -c, --count COUNT        Number of patients to generate (default: $DEFAULT_COUNT)"
    echo "  -d, --database DB_NAME   DuckDB database name (default: $DEFAULT_DB_NAME)"
    echo "  -o, --output OUTPUT_DIR  Output directory for CSV files (default: $OUTPUT_DIR)"
    echo "  --skip-build            Skip Docker image build"
    echo "  --skip-generate         Skip data generation (use existing CSV files)"
    echo "  --skip-db               Skip database creation and loading"
    echo "  --clean                 Clean output directory before generation"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                          # Use defaults"
    echo "  $0 -s California -c 5000                   # Generate 5000 patients from California"
    echo "  $0 --skip-generate --database my_data.db   # Only load existing CSV data"
    echo "  $0 --clean -c 1000                         # Clean and generate 1000 patients"
}

# Parse command line arguments
parse_args() {
    SYNTHEA_STATE="$DEFAULT_STATE"
    SYNTHEA_COUNT="$DEFAULT_COUNT"
    DB_NAME="$DEFAULT_DB_NAME"
    SKIP_BUILD=false
    SKIP_GENERATE=false
    SKIP_DB=false
    CLEAN_OUTPUT=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--state)
                SYNTHEA_STATE="$2"
                shift 2
                ;;
            -c|--count)
                SYNTHEA_COUNT="$2"
                shift 2
                ;;
            -d|--database)
                DB_NAME="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-generate)
                SKIP_GENERATE=true
                shift
                ;;
            --skip-db)
                SKIP_DB=true
                shift
                ;;
            --clean)
                CLEAN_OUTPUT=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if Docker is installed and running
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker is not running. Please start Docker daemon."
        exit 1
    fi

    # Check if DuckDB is installed
    if ! command -v duckdb &> /dev/null; then
        log_error "DuckDB is not installed. Please install DuckDB first."
        echo "You can install it with:"
        echo "  # On macOS:"
        echo "  brew install duckdb"
        echo "  # On Ubuntu/Debian:"
        echo "  wget https://github.com/duckdb/duckdb/releases/latest/download/duckdb_cli-linux-amd64.zip"
        echo "  unzip duckdb_cli-linux-amd64.zip && sudo mv duckdb /usr/local/bin/"
        exit 1
    fi

    # Check if required files exist
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    if [ ! -f "$SCRIPT_DIR/Dockerfile" ]; then
        log_error "Dockerfile not found in $SCRIPT_DIR"
        exit 1
    fi

    if [ ! -f "$SCRIPT_DIR/schema.sql" ]; then
        log_error "schema.sql not found in $SCRIPT_DIR"
        exit 1
    fi

    if [ ! -f "$SCRIPT_DIR/load.sql" ]; then
        log_error "load.sql not found in $SCRIPT_DIR"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Clean output directory
clean_output() {
    if [ "$CLEAN_OUTPUT" = true ]; then
        log_info "Cleaning output directory..."
        if [ -d "$OUTPUT_DIR" ]; then
            rm -rf "$OUTPUT_DIR"
            log_success "Output directory cleaned"
        fi
    fi
}

# Build Docker image
build_docker_image() {
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    if [ "$SKIP_BUILD" = false ]; then
        log_info "Building Docker image '$DOCKER_IMAGE'..."
        if docker build -t "$DOCKER_IMAGE" "$SCRIPT_DIR"; then
            log_success "Docker image built successfully"
        else
            log_error "Failed to build Docker image"
            exit 1
        fi
    else
        log_info "Skipping Docker image build"
        # Check if image exists
        if ! docker image inspect "$DOCKER_IMAGE" &> /dev/null; then
            log_error "Docker image '$DOCKER_IMAGE' not found. Cannot skip build."
            exit 1
        fi
    fi
}

# Generate synthetic data
generate_data() {
    if [ "$SKIP_GENERATE" = false ]; then
        log_info "Generating synthetic healthcare data..."
        log_info "State: $SYNTHEA_STATE, Patient count: $SYNTHEA_COUNT"

        # Create output directory
        mkdir -p "$OUTPUT_DIR"

        # Run Synthea container
        # Handle both absolute and relative paths for OUTPUT_DIR
        if [[ "$OUTPUT_DIR" = /* ]]; then
            # Absolute path - use as is
            local volume_mount="$OUTPUT_DIR:/output"
        else
            # Relative path - prepend current directory
            local volume_mount="$(pwd)/$OUTPUT_DIR:/output"
        fi

        if docker run --rm \
            -e SYNTHEA_STATE="$SYNTHEA_STATE" \
            -e SYNTHEA_COUNT="$SYNTHEA_COUNT" \
            -v "$volume_mount" \
            "$DOCKER_IMAGE"; then
            log_success "Data generation completed"
        else
            log_error "Data generation failed"
            exit 1
        fi

        # Check if CSV files were generated
        if [ ! -d "$OUTPUT_DIR/csv" ]; then
            log_error "CSV files not found in $OUTPUT_DIR/csv"
            exit 1
        fi

        # Show generated files
        log_info "Generated CSV files:"
        ls -la "$OUTPUT_DIR/csv/"
    else
        log_info "Skipping data generation"
        # Check if CSV files exist
        if [ ! -d "$OUTPUT_DIR/csv" ]; then
            log_error "CSV files not found in $OUTPUT_DIR/csv. Cannot skip generation."
            exit 1
        fi
    fi
}

# Create and populate database
setup_database() {
    if [ "$SKIP_DB" = false ]; then
        log_info "Setting up DuckDB database: $DB_NAME"

        SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

        # Remove existing database if it exists
        if [ -f "$DB_NAME" ]; then
            log_warning "Database $DB_NAME already exists. Removing..."
            rm "$DB_NAME"
        fi

        # Update load.sql to use correct output path
        TEMP_LOAD_SQL="load_temp.sql"
        sed "s|'output/|'$OUTPUT_DIR/csv/|g" "$SCRIPT_DIR/load.sql" > "$TEMP_LOAD_SQL"

        # Create database schema
        log_info "Creating database schema..."
        if duckdb "$DB_NAME" < "$SCRIPT_DIR/schema.sql"; then
            log_success "Database schema created"
        else
            log_error "Failed to create database schema"
            rm -f "$TEMP_LOAD_SQL"
            exit 1
        fi

        # Load data into database
        log_info "Loading data into database..."
        if duckdb "$DB_NAME" < "$TEMP_LOAD_SQL"; then
            log_success "Data loaded successfully"
        else
            log_error "Failed to load data into database"
            rm -f "$TEMP_LOAD_SQL"
            exit 1
        fi

        # Clean up temporary file
        rm -f "$TEMP_LOAD_SQL"

        # Show database statistics
        log_info "Database statistics:"
        duckdb "$DB_NAME" -c "
            SELECT
                'patients' as table_name, COUNT(*) as row_count
            FROM patients
            UNION ALL
            SELECT 'encounters', COUNT(*) FROM encounters
            UNION ALL
            SELECT 'conditions', COUNT(*) FROM conditions
            UNION ALL
            SELECT 'procedures', COUNT(*) FROM procedures
            UNION ALL
            SELECT 'observations', COUNT(*) FROM observations
            UNION ALL
            SELECT 'medications', COUNT(*) FROM medications
            ORDER BY table_name;
        "

        # Clean up CSV files after successful database creation
        log_info "Cleaning up CSV files..."
        if [ -d "$OUTPUT_DIR" ]; then
            rm -rf "$OUTPUT_DIR"
            log_success "CSV files cleaned up"
        fi

        log_success "Database setup completed: $DB_NAME"
    else
        log_info "Skipping database setup"
    fi
}

# Show completion summary
show_summary() {
    echo ""
    log_success "=== SETUP COMPLETED ==="
    echo ""
    echo "Configuration used:"
    echo "  State: $SYNTHEA_STATE"
    echo "  Patient count: $SYNTHEA_COUNT"
    echo "  Database: $DB_NAME"
    echo "  Output directory: $OUTPUT_DIR"
    echo ""
    echo "Next steps:"
    echo "  1. Connect to database: duckdb $DB_NAME"
    echo "  2. Run queries: SELECT COUNT(*) FROM patients;"
    echo "  3. Explore spatial data: SELECT city, COUNT(*) FROM patients GROUP BY city;"
    echo ""
    echo "Example queries:"
    echo "  # Patient demographics"
    echo "  SELECT race, gender, COUNT(*) FROM patients GROUP BY race, gender;"
    echo ""
    echo "  # Most common conditions"
    echo "  SELECT description, COUNT(*) as count"
    echo "  FROM conditions"
    echo "  GROUP BY description"
    echo "  ORDER BY count DESC"
    echo "  LIMIT 10;"
    echo ""
}

# Main execution
main() {
    echo "🏥 Synthea Data Generation and DuckDB Setup"
    echo "==========================================="
    echo ""

    parse_args "$@"
    check_prerequisites
    clean_output
    build_docker_image
    generate_data
    setup_database
    show_summary
}

# Run main function with all arguments
main "$@"
