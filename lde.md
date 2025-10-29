# Hugr Local Development Environment (LDE)

Complete local development environment for Hugr MCP service with Docker Compose.

## Quick Start

```bash
# Start the environment (automatically detects if data exists)
./lde.sh start

# Or start from the lde directory
./lde/scripts/start.sh
```

**First run**: 5-8 minutes (includes data generation)
**Subsequent runs**: ~30 seconds (data already exists)

## What You Get

### Services (5 Docker containers)
- **Hugr** - GraphQL API server (http://localhost:19000/query)
- **PostgreSQL** - TimescaleDB with pgvector (localhost:19001)
- **Redis** - 20GB L2 cache (localhost:19002)
- **MinIO** - S3-compatible storage (http://localhost:19004)
- **Keycloak** - Authentication server (http://localhost:19005)

### Data Sources (5 sources)
- **Northwind** - PostgreSQL sample database (~5MB)
- **Synthea** - Synthetic patient data, 10K patients (~500MB)
- **Open Payments** - CMS payment data, 11.8M records (~3GB)
- **OpenWeatherMap** - Weather REST API
- **EmbeddingGemma** - Text embedding model (308M)

### AI Models
- **EmbeddingGemma** (308M) - Text embeddings
- **GPT-OSS-20B** (21B) - Text summarization

## Management Script

Use `./lde.sh` from the project root for convenient management:

```bash
# Start environment
./lde.sh start              # Auto-detects if data exists
./lde.sh start --verbose    # Start with detailed output
./lde.sh start --no-data    # Start without loading data

# Stop environment
./lde.sh stop

# Full cleanup
./lde.sh cleanup --force    # Remove all data and containers

# Check health
./lde.sh health             # Quick health check
./lde.sh health --wait      # Wait for services to become healthy

# Load/reload data
./lde.sh load               # Load all data sources
./lde.sh load --force       # Force reload

# View logs
./lde.sh logs               # All services
./lde.sh logs hugr          # Specific service

# Help
./lde.sh help
```

### Smart Data Detection

The `lde.sh start` command automatically checks if data exists:
- **Data exists**: Starts with `--no-data` flag (faster)
- **No data**: Loads all data sources (first run)

## Common Workflows

### First Time Setup

```bash
# 1. Copy environment template
cp lde/.env.example lde/.env

# 2. (Optional) Edit configuration
nano lde/.env

# 3. Start everything
./lde.sh start
```

### Daily Development

```bash
# Start (quick if data exists)
./lde.sh start

# Work with Hugr...

# Stop when done
./lde.sh stop
```

### Refresh Data

```bash
# Reload all data sources
./lde.sh load --force

# Or reload specific source
./lde/scripts/load-data.sh --skip-northwind --skip-synthea --force
```

### Complete Reset

```bash
# Full cleanup and fresh start
./lde.sh cleanup --force
./lde.sh start
```

## Configuration

### Environment Variables (`.env`)

Key settings:

```bash
# Service Images
HUGR_IMAGE=hugr/hugr:latest
POSTGRES_IMAGE=timescale/timescaledb-ha:pg16
REDIS_IMAGE=redis:8-alpine

# Authentication
SECRET_KEY=local-dev-secret-key-change-in-production

# Data Generation
SYNTHEA_COUNT=10000
SYNTHEA_STATE=Massachusetts

# API Keys
OPENWEATHERMAP_API_KEY=your_api_key_here

# AI Models
EMBEDDINGS_URL=http://host.docker.internal:19007
EMBEDDINGS_MODEL=ai/embeddinggemma
```

See `lde/.env.example` for complete list.

### Ports

| Service | Port | Purpose |
|---------|------|---------|
| Hugr GraphQL | 19000 | API endpoint |
| Hugr Health | 19006 | Health check |
| PostgreSQL | 19001 | Database |
| Redis | 19002 | Cache |
| MinIO API | 19003 | S3 API |
| MinIO Console | 19004 | Web UI |
| Keycloak | 19005 | Auth server |

## Authentication

### SECRET_KEY (Scripts)

For data loading and automation:

```bash
curl -X POST http://localhost:19000/query \
  -H "x-hugr-secret: local-dev-secret-key-change-in-production" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ core { data_sources { name } } }"}'
```

### OIDC Tokens (Users)

For role-based access:

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:19005/realms/hugr/protocol/openid-connect/token \
  -d "client_id=hugr-graphql" \
  -d "username=admin@example.com" \
  -d "password=admin123" \
  -d "grant_type=password" \
  | jq -r '.access_token')

# Use token
curl -X POST http://localhost:19000/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ synthea { patients_aggregation { count: _rows_count } } }"}'
```

**Test Users:**
- `admin@example.com` / `admin123` (admin)
- `analyst@example.com` / `analyst123` (analyst)
- `viewer@example.com` / `viewer123` (viewer)

## Data Sources

### Northwind (PostgreSQL)

Sample database with orders, products, customers.

**Query:**
```graphql
{
  nw {
    orders(limit: 10) {
      order_id
      customer_id
      order_date
    }
  }
}
```

### Synthea (DuckDB)

Synthetic healthcare data with realistic patient records.

**Query:**
```graphql
{
  synthea {
    patients_aggregation {
      count: _rows_count
    }
    patients(limit: 5) {
      id
      first
      last
      gender
      race
    }
  }
}
```

**Configuration:**
- Patient count: `SYNTHEA_COUNT=10000` in `.env`
- State: `SYNTHEA_STATE=Massachusetts`

### Open Payments (DuckDB)

CMS pharmaceutical/device company payments to healthcare providers.

**Query:**
```graphql
{
  op {
    general_payments_aggregation {
      count: _rows_count
    }
    general_payments(limit: 5) {
      physician_name
      manufacturer_name
      payment_amount
      payment_date
    }
  }
}
```

### OpenWeatherMap (HTTP)

Real-time weather data via REST API.

**Setup:**
1. Get API key: https://openweathermap.org/api
2. Set in `.env`: `OPENWEATHERMAP_API_KEY=your_key`

**Query:**
```graphql
{
  owm {
    weather(q: "Boston,MA,US") {
      main {
        temp
        humidity
      }
      weather {
        description
      }
    }
  }
}
```

### EmbeddingGemma (Embedding)

Text embedding model for semantic search.

**Query:**
```graphql
mutation {
  embedding {
    embed(text: "Hello world") {
      vector
      dimension
    }
  }
}
```

## Troubleshooting

### Services Won't Start

```bash
# Check Docker
docker ps

# Check port conflicts
lsof -i :19000

# View logs
./lde.sh logs

# Full cleanup and retry
./lde.sh cleanup --force
./lde.sh start
```

### Health Checks Failing

```bash
# Detailed health check
./lde.sh health --verbose

# Check specific service
docker compose -f lde/docker-compose.yml logs hugr

# Restart unhealthy service
docker compose -f lde/docker-compose.yml restart hugr
```

### Data Loading Fails

```bash
# Verify services are healthy first
./lde.sh health

# Reload with verbose output
./lde/scripts/load-data.sh --verbose --force

# Check specific data loader logs
tail -f /tmp/openpayments-setup.log
```

### Out of Disk Space

```bash
# Check usage
du -sh lde/data/*
du -sh lde/.local/*

# Clean temporary files
rm -rf lde/data/.tmp/*

# Full cleanup
./lde.sh cleanup --force
```

## Directory Structure

```
mcp/
├── lde.sh                  # Management script (project root)
├── lde.md                  # This documentation
└── lde/
    ├── docker-compose.yml  # Service definitions
    ├── .env                # Configuration
    ├── scripts/
    │   ├── start.sh        # Start services
    │   ├── stop.sh         # Stop services
    │   ├── cleanup.sh      # Clean up environment
    │   ├── health-check.sh # Health verification
    │   └── load-data.sh    # Load all data sources
    ├── data-loaders/       # Data generation scripts
    │   ├── northwind/
    │   ├── synthea/
    │   ├── open-payments/
    │   └── openweathermap/
    ├── data/               # Generated databases
    │   ├── .tmp/           # Temporary CSV files (auto-cleaned)
    │   ├── core.duckdb
    │   ├── synthea.duckdb
    │   └── openpayments.duckdb
    └── .local/             # Persistent volumes
        ├── pg-data/
        ├── redis-data/
        ├── minio/
        └── keycloak/
```

## Performance

### Resource Requirements

**Minimum:**
- RAM: 8GB (4GB for services + 4GB for OS)
- Disk: 15GB free
- CPU: 4 cores

**Recommended:**
- RAM: 16GB (allows 20GB Redis cache)
- Disk: 30GB free
- CPU: 8 cores
- GPU: Optional for AI models

### Startup Times

| Scenario | Time |
|----------|------|
| First run (with data) | 5-8 minutes |
| Subsequent runs (data exists) | ~30 seconds |
| Start without data | ~1 minute |
| Data reload only | ~10-15 minutes |

### Data Sizes

| Source | Size | Records |
|--------|------|---------|
| Northwind | ~5MB | Sample DB |
| Synthea | ~500MB | 10K patients |
| Open Payments | ~3GB | 11.8M payments |
| **Total** | **~3.5GB** | |

## Advanced Usage

### Custom Data Configuration

```bash
# Generate more patients
echo "SYNTHEA_COUNT=100000" >> lde/.env
./lde.sh load --force

# Use different state
echo "SYNTHEA_STATE=California" >> lde/.env
./lde.sh load --force
```

### Selective Data Loading

```bash
# Load only Northwind and Synthea
./lde/scripts/load-data.sh --skip-openpayments --skip-owm

# Reload only Open Payments
./lde/scripts/load-data.sh --skip-northwind --skip-synthea --skip-owm --force
```

### Development Workflow

```bash
# Start without data for faster iteration
./lde.sh start --no-data

# Test schema changes...

# Load specific data when needed
./lde/scripts/load-data.sh --skip-openpayments
```

### Accessing Databases Directly

```bash
# PostgreSQL
psql postgresql://hugr:hugr@localhost:19001/hugr

# Redis
redis-cli -h localhost -p 19002

# DuckDB (Synthea)
duckdb lde/data/synthea.duckdb

# DuckDB (Open Payments)
duckdb lde/data/openpayments.duckdb
```

## CI/CD Integration

```bash
# Start in CI environment
./lde.sh start --no-data

# Run tests against Hugr
curl -f http://localhost:19000/query

# Cleanup
./lde.sh cleanup --force
```

## Documentation

### LDE Documentation
- **Full README**: `lde/README.md` - Comprehensive LDE guide
- **Data Loaders**: `lde/data-loaders/README.md` - Data generation details
- **Docker Models**: `lde/DOCKER_MODELS.md` - AI models setup

### Project Documentation
- **Quickstart**: `specs/001-create-a-local/quickstart.md`
- **Tasks**: `specs/001-create-a-local/tasks.md`
- **Data Model**: `specs/001-create-a-local/data-model.md`

## Support

### Getting Help

```bash
# Command help
./lde.sh help
./lde.sh start --help
./lde.sh cleanup --help

# Check logs
./lde.sh logs
./lde.sh logs hugr
```

### Common Issues

1. **Port conflicts**: Change ports in `.env`
2. **Out of memory**: Increase Docker memory limit
3. **Disk space**: Run `./lde.sh cleanup --force`
4. **Health checks fail**: Check `./lde.sh logs`

## License

See project LICENSE file.
