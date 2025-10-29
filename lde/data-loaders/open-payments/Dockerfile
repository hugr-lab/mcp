# Dockerfile for Open Payments Data Loading
FROM debian:bullseye-slim

# Install required tools
RUN apt-get update && apt-get install -y \
    wget \
    unzip \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install DuckDB CLI
RUN wget https://github.com/duckdb/duckdb/releases/latest/download/duckdb_cli-linux-amd64.zip \
    && unzip duckdb_cli-linux-amd64.zip \
    && mv duckdb /usr/local/bin/ \
    && rm duckdb_cli-linux-amd64.zip

# Create working directory
WORKDIR /open-payments

# Copy scripts
COPY setup.sh /open-payments/setup.sh
COPY schema.sql /open-payments/schema.sql
RUN chmod +x /open-payments/setup.sh

# Set default data directory
ENV DATA_DIR=/data/openpayments
ENV DB_FILE=/data/openpayments.duckdb

# Default command
CMD ["./setup.sh", "--force", "--data-dir", "/data/openpayments", "--db-file", "/data/openpayments.duckdb"]
