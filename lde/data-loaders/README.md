# Data Loaders

Self-contained data generation and loading scripts for the Hugr LDE. Each loader is standalone and can be used independently or via the unified `load-data.sh` script.

## Overview

| Source | Type | Size | Records | Script |
|--------|------|------|---------|--------|
| Northwind | PostgreSQL | ~5MB | Sample DB | SQL dump |
| Synthea | DuckDB | ~200-500MB | 10,000 patients | `generate-and-load.sh` |
| Open Payments | DuckDB | ~2-3GB | ~11.8M payments | `setup.sh` |
| OpenWeatherMap | HTTP API | N/A | REST API | OpenAPI spec |

## Directory Structure

```
data-loaders/
├── northwind/
│   ├── northwind_dump.sql      # PostgreSQL database dump
│   └── schema.graphql          # Hugr GraphQL schema
├── synthea/
│   ├── Dockerfile              # Docker image with Synthea JAR
│   ├── synthea.properties      # Synthea configuration
│   ├── generate-and-load.sh    # Main generation script
│   ├── schema.sql              # DuckDB schema
│   └── load.sql                # DuckDB data loading
├── open-payments/
│   ├── setup.sh                # Download and load script
│   ├── schema.sql              # DuckDB schema
│   └── schemas/
│       ├── schema.graphql      # Hugr GraphQL schema (base)
│       └── extra.graphql       # Additional types/queries
└── openweathermap/
    ├── spec.yaml               # OpenAPI 3.0 specification
    └── schema.graphql          # Hugr GraphQL schema
```

## Synthea Data Loader

### Overview
Generates synthetic patient data using Synthea and loads it into a DuckDB database following the exact schema from `/projects/synthea`.

### Files
- **schema.sql**: Database schema with 20 tables (patients, encounters, conditions, etc.)
  - Uses surrogate BIGINT ids with sequences
  - Includes spatial extension for geom fields
  - Maintains foreign key relationships
- **load.sql**: SQL script to load CSV data into the database
  - Reads from `output/` directory (Synthea CSV output)
  - Joins related tables using source_id references
- **generate-and-load.sh**: Orchestration script
  - Runs Synthea with configured parameters
  - Creates DuckDB database with schema
  - Loads generated CSV data

### Usage

#### Standalone
```bash
cd synthea

# Generate 1000 patients from California
./generate-and-load.sh -c 1000 -s California -d ~/data/synthea.duckdb

# Use custom output directory (supports absolute paths)
./generate-and-load.sh -c 500 -o /tmp/synthea-output

# Skip Docker build (use existing image)
./generate-and-load.sh --skip-build -c 100
```

#### Options
```
-s, --state STATE        US state for generation (default: Massachusetts)
-c, --count COUNT        Number of patients (default: 100)
-d, --database DB        DuckDB file path (default: synthea.duckdb)
-o, --output DIR         Output directory for CSV files (default: output)
--skip-build            Skip Docker image build
--skip-generate         Skip data generation
--skip-db               Skip database creation
--clean                 Clean output directory first
-h, --help              Show help
```

### Environment Variables
- `SYNTHEA_PATIENT_COUNT`: Number of patients to generate (default: 10000)
- `SYNTHEA_STATE`: State for patient generation (default: Massachusetts)
- `SYNTHEA_SEED`: Random seed for reproducibility (default: 12345)
- `SYNTHEA_OUTPUT_DIR`: Output directory (default: /data)
- `SYNTHEA_DB_FILE`: DuckDB database file path (default: /data/synthea.duckdb)

### Output
- DuckDB database: `synthea.duckdb`
- Size: ~200-500MB (depends on patient count)
- Tables: 20 (patients, encounters, conditions, medications, observations, procedures, immunizations, allergies, careplans, devices, imaging_studies, supplies, claims, claims_transactions, payer_transitions, organizations, payers, providers)

### CSV Cleanup
CSV files are temporarily stored in the output directory. After successful DuckDB creation, the entire output directory is automatically removed to save disk space.

When using with LDE's `load-data.sh`, CSV files are stored in `/path/to/lde/data/.tmp/synthea/` and automatically cleaned up.

## Open Payments Data Loader

### Overview
Downloads and loads CMS Open Payments data into DuckDB.

### Files
- **schema.sql**: Database schema for Open Payments tables
  - `general_payments`: General payment records
  - `research_payments`: Research payment records
  - `ownership_information`: Ownership information
- **setup.sh**: Download and loading script
  - Downloads ZIP archive from CMS
  - Extracts CSV files
  - Creates DuckDB database
  - Loads data into tables

### Usage

#### Standalone
```bash
cd open-payments

# Download and load (default: ../../data/openpayments)
./setup.sh

# Custom paths
./setup.sh \
  --db-file ~/data/openpayments.duckdb \
  --data-dir ~/tmp/openpayments-csv

# Force re-download
./setup.sh --force

# Custom download URL
./setup.sh --url https://download.cms.gov/openpayments/PGYR2022_...zip
```

### Environment Variables
- `OPENPAYMENTS_DATA_DIR`: Temporary directory for CSV files (default: /data/openpayments)
- `OPENPAYMENTS_DB_FILE`: DuckDB database file path (default: /data/openpayments.duckdb)
- `OPENPAYMENTS_DOWNLOAD_URL`: URL to download ZIP archive

### Options
- `--help`: Show help message
- `--force`: Force download and reload without confirmation
- `--data-dir DIR`: Specify data directory
- `--db-file FILE`: Specify database file name
- `--url URL`: Specify download URL
- `--sql-script FILE`: Specify SQL script file

### Output
- DuckDB database: `openpayments.duckdb`
- Size: ~2-3GB
- Tables: 3 (general_payments, research_payments, ownership_information) + providers (aggregated)
- Records: ~11.8M general payments (2023 data)

### CSV Cleanup
Downloaded CSV files are extracted to the data directory specified by `--data-dir`. After successful DuckDB creation, the entire data directory is automatically removed.

When using with LDE's `load-data.sh`, CSV files are stored in `/path/to/lde/data/.tmp/openpayments/` and automatically cleaned up.

### Download Size
- **ZIP file**: ~1GB
- **Extracted CSV**: ~8-10GB
- **Final DuckDB**: ~2-3GB

## Integration with LDE

All data loaders are integrated into the unified `load-data.sh` script:

```bash
# Load all data sources
./lde/scripts/load-data.sh

# Skip specific sources
./lde/scripts/load-data.sh --skip-synthea --skip-openpayments

# Force reload
./lde/scripts/load-data.sh --force
```

### Automatic Cleanup

When using `load-data.sh`, temporary CSV files are automatically cleaned up:

- **Synthea**: CSVs stored in `lde/data/.tmp/synthea/`, removed after DB creation
- **Open Payments**: CSVs stored in `lde/data/.tmp/openpayments/`, removed after DB creation

Only final DuckDB files remain in `lde/data/` directory.

## Schema Source

### Synthea
The SQL scripts are exact copies from `/Users/vgribanov/projects/synthea`:
- Source: `/Users/vgribanov/projects/synthea/schema.sql`
- Source: `/Users/vgribanov/projects/synthea/load.sql`
- **No modifications** - uses the same structure as the original Synthea repository

### Open Payments
The scripts are copied from the hugr examples repository:
- Source: `/Users/vgribanov/projects/hugr-lab/examples/examples/open-payments/schema.sql`
- Source: `/Users/vgribanov/projects/hugr-lab/examples/examples/open-payments/setup.sh`

## Benefits

1. **No External Dependencies**: All required scripts are included in this repository
2. **Dockerized**: Can run in isolated containers
3. **Reproducible**: Fixed schemas and deterministic data generation
4. **Self-Contained**: No need to clone external repositories
5. **Version Controlled**: Schema and loading logic tracked in git

## Troubleshooting

### Synthea Generation Fails
- Check Docker daemon is running
- Ensure sufficient disk space (>2GB)
- Verify Java is available in Docker image

### Open Payments Download Fails
- Check internet connection
- Verify download URL is current
- Check CMS website for updated URLs

### DuckDB Errors
- Install DuckDB CLI: `brew install duckdb` (macOS) or download from [duckdb.org](https://duckdb.org)
- Verify CSV files are properly formatted
- Check database file permissions

### Schema Mismatches
- If Synthea schema changes, update `schema.sql` and `load.sql` from source
- If Open Payments schema changes, update `schema.sql` from examples repo
- Regenerate databases after schema updates

## Maintenance

### Updating Synthea Schema
```bash
# If synthea repository schema changes
cp /Users/vgribanov/projects/synthea/schema.sql lde/data-loaders/synthea/
cp /Users/vgribanov/projects/synthea/load.sql lde/data-loaders/synthea/
```

### Updating Open Payments
```bash
# If examples repository changes
cp /Users/vgribanov/projects/hugr-lab/examples/examples/open-payments/schema.sql lde/data-loaders/open-payments/
cp /Users/vgribanov/projects/hugr-lab/examples/examples/open-payments/setup.sh lde/data-loaders/open-payments/
```

### Testing Data Loaders
```bash
# Test Synthea loader
cd lde/data-loaders/synthea
./generate-and-load.sh

# Test Open Payments loader
cd lde/data-loaders/open-payments
./setup.sh --force

# Verify databases
duckdb /path/to/synthea.duckdb "SHOW TABLES;"
duckdb /path/to/openpayments.duckdb "SHOW TABLES;"
```
