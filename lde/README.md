# Local Development Environment (LDE)

Reproducible local development environment for the Hugr MCP service using Docker Compose.

## Overview

This LDE provides:
- **5 Docker services**: Hugr, PostgreSQL (TimescaleDB with pgvector), Redis (20GB L2 cache), MinIO, Keycloak
- **AI Models**: EmbeddingGemma (308M) for embeddings, GPT-OSS-20B for summarization via Docker models
- **5 data sources**: Northwind (PostgreSQL), Synthea (DuckDB), Open Payments (DuckDB), OpenWeatherMap (HTTP), EmbeddingGemma (embedding)
- **Shell scripts**: Environment lifecycle management (start, stop, health-check, load-data)
- **Data persistence**: All data stored in `.local/` and `data/` directories
- **Self-contained**: No external repository dependencies

## Quick Start

### Prerequisites

- Docker Desktop 4.36+ or Docker Engine 20.10+ with AI extension
- Docker Compose v2.0+
- curl, jq
- 50GB RAM available (20GB Hugr + 20GB Redis + 10GB for models and other services)
- 30GB free disk space (includes model downloads)
- Ports available: 19000-19008
- GPU recommended for AI models (NVIDIA, AMD, or Apple Silicon)

### First Time Setup

```bash
# 1. Navigate to repository root
cd /path/to/hugr-lab/mcp

# 2. Copy environment template
cp lde/.env.example lde/.env

# 3. (Optional) Edit .env to customize configuration
#    Update SECRET_KEY, patient count, etc.

# 4. Start the environment
./lde/scripts/start.sh
```

**Duration**: 5-8 minutes (includes data generation)

**Expected Output**:
```
✓ Docker installed
✓ Prerequisites checked
→ Creating .local directory structure...
→ Starting Docker services...
✓ All services healthy (5/5)
→ Loading data sources...
✓ Synthea data loaded (10,000 patients)
✓ Open Payments data loaded (11.8M records)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Local Development Environment Ready
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Hugr GraphQL:  http://localhost:19000/query
Keycloak:      http://localhost:19005
MinIO Console: http://localhost:19004

Test Credentials:
- admin@example.com    / admin123    (admin role)
- analyst@example.com  / analyst123  (analyst role)
- viewer@example.com   / viewer123   (viewer role)

Data Sources:
- northwind (PostgreSQL sample database)
- synthea (10,000 patients in DuckDB)
- openpayments (~11.8M records in DuckDB)
- owm (OpenWeatherMap REST API)
- emb_gemma (EmbeddingGemma 308M model)

To stop: ./lde/scripts/stop.sh
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Services

### Hugr GraphQL (`http://localhost:19000/query`)
- GraphQL API server with L2 Redis cache
- Memory limit: 20GB
- AI Models: EmbeddingGemma, GPT-OSS-20B (via Docker models)
- Depends on: PostgreSQL, Redis, MinIO, Keycloak
- Health: `http://localhost:19006/health`

### PostgreSQL (`localhost:19001`)
- Database: TimescaleDB with pgvector extension
- Image: `timescale/timescaledb-ha:pg16`
- Credentials: `hugr:hugr@localhost:19001/hugr`
- Volume: `.local/pg-data`

### Redis (`localhost:19002`)
- L2 cache for Hugr with AOF persistence
- Memory limit: 20GB (maxmemory-policy: allkeys-lru)
- Image: `redis:8-alpine`
- Volume: `.local/redis-data`

### AI Models (Docker Models)
See [DOCKER_MODELS.md](DOCKER_MODELS.md) for detailed configuration.
- **EmbeddingGemma** (`ai/embeddinggemma`): 308M parameter embedding model
- **GPT-OSS-20B** (`ai/gpt-oss-20b`): 21B parameter LLM for summarization

### MinIO (`http://localhost:19003`, Console: `http://localhost:19004`)
- S3-compatible object storage
- Credentials: `minioadmin:minioadmin`
- Volume: `.local/minio`

### Keycloak (`http://localhost:19005`)
- Authentication server
- Admin: `admin:admin`
- Realm: `hugr` (auto-imported)
- Roles: admin, analyst, viewer
- Volume: `.local/keycloak`

## Scripts

### `start.sh` - Start Environment

```bash
./lde/scripts/start.sh [OPTIONS]
```

**Options**:
- `--no-data`: Skip data loading step
- `--verbose`: Enable detailed output
- `--help`: Show usage

**Exit Codes**:
- `0`: Success
- `1`: Prerequisites missing
- `2`: Docker compose failure
- `3`: Health checks failed
- `4`: Data loading failed

**Examples**:
```bash
# Normal start (with data loading)
./lde/scripts/start.sh

# Start without data (faster for testing)
./lde/scripts/start.sh --no-data

# Verbose mode
./lde/scripts/start.sh --verbose
```

### `cleanup.sh` - Clean Up Environment

```bash
./lde/scripts/cleanup.sh [OPTIONS]
```

Complete cleanup of the LDE environment. This will:
1. Stop and remove all Docker containers
2. Remove Docker volumes and networks
3. Clear the `data/` directory (databases, schemas)
4. Clear the `.local/` directory (service data)

**Options**:
- `--keep-data`: Keep data/ directory (preserve databases)
- `--keep-local`: Keep .local/ directory (preserve service data)
- `--keep-images`: Keep Docker images (don't prune)
- `--force`: Skip confirmation prompt
- `--help`: Show usage

**Examples**:
```bash
# Full cleanup with confirmation
./lde/scripts/cleanup.sh

# Full cleanup without confirmation
./lde/scripts/cleanup.sh --force

# Clean everything except databases
./lde/scripts/cleanup.sh --keep-data

# Clean everything except service data
./lde/scripts/cleanup.sh --keep-local
```

⚠️ **WARNING**: This will permanently delete all data unless `--keep-*` flags are used!

### `stop.sh` - Stop Environment

```bash
./lde/scripts/stop.sh [OPTIONS]
```

Stops all containers while preserving data in `.local/`.

**Options**:
- `--verbose`: Detailed output
- `--help`: Show usage

### `health-check.sh` - Check Service Health

```bash
./lde/scripts/health-check.sh [OPTIONS]
```

Checks all 5 services with response times.

**Options**:
- `--verbose`: Show detailed health status
- `--wait [TIMEOUT]`: Wait for services to become healthy (default: 120s)
- `--help`: Show usage

**Exit Codes**:
- `0`: All healthy
- `1`: Some unhealthy
- `2`: Timeout

**Examples**:
```bash
# Quick health check
./lde/scripts/health-check.sh

# Wait for services (useful in scripts)
./lde/scripts/health-check.sh --wait 180

# Detailed output
./lde/scripts/health-check.sh --verbose
```

### `load-data.sh` - Load All Data Sources

```bash
./lde/scripts/load-data.sh [OPTIONS]
```

Unified script that loads all 5 data sources:
1. **Northwind** (PostgreSQL) - Sample database
2. **Synthea** (DuckDB) - Synthetic patient data
3. **Open Payments** (DuckDB) - CMS payment data
4. **OpenWeatherMap** (HTTP) - REST API
5. **EmbeddingGemma** (Embedding) - AI model

**Options**:
- `--skip-northwind`: Skip Northwind loading
- `--skip-synthea`: Skip Synthea generation
- `--skip-openpayments`: Skip Open Payments download
- `--skip-owm`: Skip OpenWeatherMap registration
- `--skip-embedding`: Skip embedding model registration
- `--force`: Force reload even if data exists
- `--verbose`: Detailed output
- `--help`: Show usage

**Exit Codes**:
- `0`: Success
- `1`: Prerequisites missing
- `2`: Hugr not healthy
- `3`: Data generation failed
- `4`: Data source registration failed
- `5`: Verification failed

**Examples**:
```bash
# Load all data sources
./lde/scripts/load-data.sh

# Load only Synthea (skip others)
./lde/scripts/load-data.sh --skip-northwind --skip-openpayments --skip-owm --skip-embedding

# Force reload all data
./lde/scripts/load-data.sh --force

# Verbose output
./lde/scripts/load-data.sh --verbose
```

**Data Generation**:
- Synthea and Open Payments CSV files are temporarily stored in `data/.tmp/`
- After DuckDB database creation, CSV files are automatically cleaned up
- Only final DuckDB files remain in `data/` directory

## Data Sources

### Synthea - Synthetic Patient Data

- **Records**: 10,000 patients (configurable)
- **Format**: DuckDB
- **Size**: ~200-500MB
- **Tables**: 20 (patients, encounters, conditions, medications, observations, procedures, etc.)
- **Schema**: Exact copy from `/projects/synthea`
- **Generation**: Docker-based Synthea with reproducible seed

**Configuration** (`.env`):
```bash
SYNTHEA_PATIENT_COUNT=10000
SYNTHEA_STATE=Massachusetts
SYNTHEA_SEED=12345
```

**Hugr Data Source**:
```graphql
{
  synthea {
    patients_aggregation {
      count: _rows_count
    }
  }
}
```

### Open Payments - CMS Payment Data

- **Records**: ~11.8M general payments (2023)
- **Format**: DuckDB
- **Size**: ~2-3GB
- **Tables**: 3 (general_payments, research_payments, ownership_information)
- **Source**: CMS Open Payments downloads (~1GB zip file)
- **Loading**: Automated download, extract, and load via `setup.sh`

**Configuration** (`.env`):
```bash
OPENPAYMENTS_DOWNLOAD_URL=https://download.cms.gov/openpayments/PGYR2023_P01302025_01212025.zip
```

**Hugr Data Source**:
```graphql
{
  op {
    general_payments_aggregation {
      count: _rows_count
    }
  }
}
```

### OpenWeatherMap - Weather REST API

- **Type**: HTTP REST API
- **Format**: OpenAPI 3.0 specification
- **Authentication**: API key (configured in `.env`)
- **Path**: Query parameters with security configuration

**Configuration** (`.env`):
```bash
OPENWEATHERMAP_API_KEY=your_api_key_here
```

**Hugr Data Source**:
```graphql
{
  owm {
    weather(q: "Boston,MA,US", appid: "your_key") {
      main {
        temp
        humidity
      }
    }
  }
}
```

### EmbeddingGemma - Text Embedding Model

- **Type**: Embedding model
- **Model**: Google EmbeddingGemma 308M
- **Deployment**: Docker AI models extension
- **URL**: Configured via environment variables

**Configuration** (`.env`):
```bash
EMBEDDINGS_URL=http://host.docker.internal:19007
EMBEDDINGS_MODEL=ai/embeddinggemma
```

## Authentication

### SECRET_KEY (Data Loading)

For automated scripts and data loading:

```bash
curl -X POST http://localhost:19000/query \
  -H "Content-Type: application/json" \
  -H "x-hugr-secret: local-dev-secret-key-change-in-production" \
  -d '{"query":"{ core { data_sources { name } } }"}'
```

### OIDC Tokens (User-Level)

For role-based testing:

```bash
# Get admin token
TOKEN=$(curl -s -X POST http://localhost:19005/realms/hugr/protocol/openid-connect/token \
  -d "client_id=hugr-graphql" \
  -d "username=admin@example.com" \
  -d "password=admin123" \
  -d "grant_type=password" \
  | jq -r '.access_token')

# Query with token
curl -X POST http://localhost:19000/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"{ synthea { patients_aggregation { count: _rows_count } } }"}'
```

**Test Users**:
- `admin@example.com` / `admin123` (admin role)
- `analyst@example.com` / `analyst123` (analyst role)
- `viewer@example.com` / `viewer123` (viewer role)

## Port Mappings

| Service | Host Port | Container Port | Purpose |
|---------|-----------|----------------|---------|
| Hugr | 19000 | 8080 | GraphQL API |
| PostgreSQL | 19001 | 5432 | Database |
| Redis | 19002 | 6379 | Cache |
| MinIO API | 19003 | 9000 | S3 API |
| MinIO Console | 19004 | 9001 | Web UI |
| Keycloak | 19005 | 8080 | Auth server |

## Environment Variables

See `.env.example` for full list. Key variables:

```bash
# Service Images
POSTGRES_IMAGE=timescale/timescaledb-ha:pg16
REDIS_IMAGE=redis:8-alpine
HUGR_IMAGE=hugr/hugr:latest

# Database
DATABASE_URL=postgresql://hugr:hugr@postgres:5432/hugr

# Authentication
SECRET_KEY=local-dev-secret-key-change-in-production

# Data Generation
SYNTHEA_PATIENT_COUNT=10000
```

## Common Operations

### Restart Environment
```bash
./lde/scripts/stop.sh
./lde/scripts/start.sh
```

### Reset to Clean State
```bash
# Full cleanup
./lde/scripts/cleanup.sh --force

# Start fresh
./lde/scripts/start.sh
```

### Check Logs
```bash
docker compose -f lde/docker-compose.yml logs
docker compose -f lde/docker-compose.yml logs hugr
docker compose -f lde/docker-compose.yml logs postgres
```

### Restart Single Service
```bash
docker compose -f lde/docker-compose.yml restart hugr
```

### Access Databases

**PostgreSQL**:
```bash
psql postgresql://hugr:hugr@localhost:19001/hugr
```

**Redis**:
```bash
redis-cli -h localhost -p 19002
```

**DuckDB (Synthea)**:
```bash
duckdb lde/data/synthea.duckdb
```

## Directory Structure

```
lde/
├── docker-compose.yml          # Service definitions
├── .env.example               # Environment template
├── .env                       # Your configuration (git-ignored)
├── scripts/
│   ├── start.sh               # Start environment
│   ├── stop.sh                # Stop environment
│   ├── cleanup.sh             # Clean up environment
│   ├── health-check.sh        # Health verification
│   └── load-data.sh           # Unified data loading
├── data-loaders/              # See data-loaders/README.md
│   ├── northwind/             # PostgreSQL sample database
│   │   ├── northwind_dump.sql
│   │   └── schema.graphql
│   ├── synthea/               # Synthetic patient data generator
│   │   ├── Dockerfile
│   │   ├── generate-and-load.sh
│   │   ├── schema.sql
│   │   └── load.sql
│   ├── open-payments/         # CMS payment data downloader
│   │   ├── setup.sh
│   │   ├── schema.sql
│   │   └── schemas/*.graphql
│   └── openweathermap/        # Weather API configuration
│       ├── spec.yaml          # OpenAPI specification
│       └── schema.graphql
├── keycloak/
│   └── realm-config.json      # Keycloak realm import
├── data/
│   ├── .gitkeep
│   ├── .tmp/                  # Temporary CSV files (auto-cleaned)
│   ├── core.duckdb            # (generated) Hugr metadata
│   ├── synthea.duckdb         # (generated) Patient data
│   ├── openpayments.duckdb    # (generated) Payment data
│   └── schemas/               # GraphQL schemas
│       ├── northwind/
│       ├── openpayments/
│       └── openweathermap/
└── .local/                    # Persistent volumes (git-ignored)
    ├── pg-data/
    ├── redis-data/
    ├── minio/
    └── keycloak/
```

## Troubleshooting

### Services Won't Start

**Port conflicts**:
```bash
# Find conflicting process
lsof -i :19000

# Or modify .env to use different ports
# Then restart
./lde/scripts/stop.sh
./lde/scripts/start.sh
```

**Docker issues**:
```bash
# Check Docker is running
docker ps

# Check logs
docker compose -f lde/docker-compose.yml logs

# Nuclear option: rebuild
./lde/scripts/stop.sh
docker compose -f lde/docker-compose.yml down -v
./lde/scripts/start.sh --reset
```

### Health Checks Failing

```bash
# Detailed health status
./lde/scripts/health-check.sh --verbose

# Check individual service logs
docker compose -f lde/docker-compose.yml logs postgres
docker compose -f lde/docker-compose.yml logs keycloak

# Restart unhealthy service
docker compose -f lde/docker-compose.yml restart <service>
```

### Data Loading Fails

```bash
# Verify Hugr is healthy first
./lde/scripts/health-check.sh

# Try loading with verbose output
./lde/scripts/load-data.sh --verbose

# Force reload
./lde/scripts/load-data.sh --force
```

### Out of Disk Space

```bash
# Check usage
du -sh lde/.local/*
du -sh lde/data/*

# Clean up
./lde/scripts/stop.sh
rm -rf lde/.local/*
rm -rf lde/data/*.duckdb
./lde/scripts/start.sh --reset
```

## Testing

Contract tests verify interface compliance:
```bash
# Run all contract tests
./tests/lde/run-all-contract-tests.sh

# Run individual tests
./tests/lde/test-start-interface.sh
./tests/lde/test-compose-validity.sh
```

Integration tests verify actual functionality:
```bash
# Run all integration tests
./tests/lde/run-all-integration-tests.sh

# Run specific test
./tests/lde/test-postgres-integration.sh
```

## Performance

### Startup Times
- **First run**: 5-8 minutes (includes data generation)
- **Subsequent runs**: ~30 seconds (data already loaded)
- **With `--no-data`**: ~1 minute

### Resource Usage
- **RAM**: ~4-6GB (all services)
- **Disk**: ~3-4GB (data + volumes)
- **CPU**: Moderate during data loading, low at runtime

## Data Loaders

See `data-loaders/README.md` for detailed documentation on:
- Synthea data generation with Docker
- Open Payments downloading and loading
- Schema sources and SQL scripts
- Standalone usage outside of LDE

## Related Documentation

- **Quickstart**: `/specs/001-create-a-local/quickstart.md`
- **Tasks**: `/specs/001-create-a-local/tasks.md`
- **Data Model**: `/specs/001-create-a-local/data-model.md`
- **Research**: `/specs/001-create-a-local/research.md`
- **Contracts**: `/specs/001-create-a-local/contracts/`

## Support

If issues persist:
1. Check `docker compose logs` for detailed errors
2. Verify all prerequisites are installed
3. Ensure ports are not in use
4. Check available disk space and memory
5. Try `--reset` flag to start from clean state
