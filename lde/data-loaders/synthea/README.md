# Synthea Data Loader

This loader generates synthetic healthcare data using [Synthea](https://github.com/synthetichealth/synthea) and loads it into a DuckDB database.

## Overview

Synthea is a synthetic patient generator that creates realistic but not real patient data. This is useful for:
- Testing healthcare applications
- Research and analytics
- Training and demonstrations
- FHIR and HL7 compliance testing

## Files

- `Dockerfile` - Docker image configuration that downloads pre-built Synthea JAR
- `synthea.properties` - Synthea configuration (enables CSV export only)
- `schema.sql` - DuckDB schema with 18 tables and spatial support
- `load.sql` - ETL script to load CSV files into DuckDB
- `generate-and-load.sh` - Main orchestration script

## Database Schema

The schema includes:
- **patients** (with geospatial support via GEOMETRY)
- **organizations**, **providers**, **payers**
- **encounters**, **conditions**, **procedures**, **observations**
- **medications**, **immunizations**, **allergies**, **careplans**
- **devices**, **imaging_studies**, **supplies**
- **claims**, **claims_transactions**, **payer_transitions**

All tables use surrogate BIGINT keys with sequences, maintaining original UUIDs as `source_id`.

## Usage

### Generate Data

```bash
# Generate 100 patients (default)
./generate-and-load.sh

# Generate 1000 patients from California
./generate-and-load.sh -s California -c 1000

# Generate and save to specific database
./generate-and-load.sh -c 5000 -d ../../data/synthea.duckdb

# Clean and regenerate
./generate-and-load.sh --clean -c 500
```

### Options

```
-s, --state STATE        US state for data generation (default: Massachusetts)
-c, --count COUNT        Number of patients (default: 100)
-d, --database DB_NAME   DuckDB database name (default: synthea.duckdb)
-o, --output OUTPUT_DIR  Output directory for CSV files (default: output)
--skip-build            Skip Docker image build
--skip-generate         Skip data generation (use existing CSV)
--skip-db               Skip database creation
--clean                 Clean output directory first
-h, --help              Show help
```

### Prerequisites

- Docker Desktop or Docker Engine
- DuckDB CLI (`brew install duckdb` on macOS)

### Example Output

```
🏥 Synthea Data Generation and DuckDB Setup
===========================================

[INFO] Checking prerequisites...
[SUCCESS] Prerequisites check passed
[INFO] Building Docker image 'synthea'...
[SUCCESS] Docker image built successfully
[INFO] Generating synthetic healthcare data...
[INFO] State: Massachusetts, Patient count: 10
...
[SUCCESS] Data generation completed
[INFO] Generated CSV files:
...
[SUCCESS] Database setup completed: synthea.duckdb

Database statistics:
┌──────────────┬───────────┐
│  table_name  │ row_count │
├──────────────┼───────────┤
│ conditions   │       411 │
│ encounters   │       473 │
│ medications  │       417 │
│ observations │      8847 │
│ patients     │        10 │
│ procedures   │      1466 │
└──────────────┴───────────┘
```

## Loading into Hugr

Once the database is generated, register it as a Hugr data source:

```graphql
mutation {
  core {
    insert_data_sources(data: {
      name: "synthea"
      description: "Synthea synthetic healthcare data"
      type: "duckdb"
      prefix: "synthea"
      path: "/hugr-data/synthea.duckdb"
      self_defined: true
      read_only: true
      as_module: true
    }) {
      name
    }
  }
}
```

Then load it:

```graphql
mutation {
  function {
    core {
      load_data_source(name: "synthea") {
        success
        message
      }
    }
  }
}
```

## Implementation Details

### Dockerfile

Uses `openjdk:17-slim` and downloads the pre-built Synthea JAR from GitHub releases. This is much faster than building from source (which would require cloning and running gradlew).

### Data Generation

Synthea generates CSV files for all patient data. The generation process:
1. Downloads Synthea JAR (~150MB, cached in Docker layer)
2. Runs `java -jar synthea.jar -p $COUNT $STATE`
3. Generates CSV files in `output/csv/` directory
4. Takes ~1-2 minutes for 100 patients

### Database Loading

The ETL process:
1. Creates DuckDB schema with sequences and foreign keys
2. Installs spatial extension for geospatial support
3. Reads CSV files with `read_csv_auto()`
4. Transforms data (timestamps, geometry from lat/lon)
5. Loads into normalized tables with surrogate keys
6. Maintains original UUIDs as `source_id` for traceability

### Performance

- 10 patients: ~24MB database, ~5 seconds
- 100 patients: ~150MB database, ~30 seconds
- 1,000 patients: ~1.5GB database, ~5 minutes
- 10,000 patients: ~15GB database, ~50 minutes

## Troubleshooting

### Docker Build Fails

```bash
# Pull base image manually
docker pull openjdk:17-slim

# Try building again
./generate-and-load.sh
```

### DuckDB Not Found

```bash
# macOS
brew install duckdb

# Linux
wget https://github.com/duckdb/duckdb/releases/latest/download/duckdb_cli-linux-amd64.zip
unzip duckdb_cli-linux-amd64.zip
sudo mv duckdb /usr/local/bin/
```

### Generation Takes Too Long

```bash
# Use fewer patients for testing
./generate-and-load.sh -c 10

# Or skip generation if CSV files exist
./generate-and-load.sh --skip-generate
```

## References

- [Synthea GitHub](https://github.com/synthetichealth/synthea)
- [Synthea Wiki](https://github.com/synthetichealth/synthea/wiki)
- [DuckDB Spatial Extension](https://duckdb.org/docs/extensions/spatial)
