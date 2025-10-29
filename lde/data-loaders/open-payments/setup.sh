#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DATA_DIR="../../data/openpayments"
DOWNLOAD_URL="https://download.cms.gov/openpayments/PGYR2023_P01302025_01212025.zip"
DB_FILE="openpayments.duckdb"
SQL_SCRIPT="schema.sql"
FORCE=false

# Help function
show_help() {
    echo "Open Payments DuckDB Database Setup Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -f, --force         Force setup without confirmation"
    echo "  -d, --data-dir      Specify data directory (default: ../../data/openpayments)"
    echo "  -u, --url           Specify download URL"
    echo "  -s, --sql-script    Specify SQL script file (default: schema.sql)"
    echo "  --db-file           Specify DuckDB file name (default: openpayments.duckdb)"
    echo ""
    echo "Examples:"
    echo "  $0                              # Setup with default settings"
    echo "  $0 --force                      # Setup without confirmation"
    echo "  $0 --data-dir ./data            # Use custom data directory"
    echo "  $0 --sql-script custom.sql     # Use custom SQL script"
    echo ""
    echo "Note: This script downloads Open Payments data from CMS and loads it into DuckDB"
    echo "      Make sure you have sufficient disk space and internet connection"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -d|--data-dir)
            DATA_DIR="$2"
            shift 2
            ;;
        -u|--url)
            DOWNLOAD_URL="$2"
            shift 2
            ;;
        -s|--sql-script)
            SQL_SCRIPT="$2"
            shift 2
            ;;
        --db-file)
            DB_FILE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

FILENAME=$(basename "$DOWNLOAD_URL")

echo -e "${BLUE}🏥 Open Payments DuckDB Database Setup${NC}"
echo "====================================="

echo -e "${BLUE}📋 Configuration:${NC}"
echo "   Data Directory: $DATA_DIR"
echo "   Download URL: $DOWNLOAD_URL"
echo "   Archive File: $FILENAME"
echo "   DuckDB File: $DB_FILE"
echo "   SQL Script: $SQL_SCRIPT"

# Check if SQL script exists
if [ ! -f "$SQL_SCRIPT" ]; then
    echo -e "${RED}❌ SQL script '$SQL_SCRIPT' not found!${NC}"
    echo "Please ensure the SQL script exists in the current directory."
    exit 1
fi

# Check if DuckDB is available
if ! command -v duckdb &> /dev/null; then
    echo -e "${RED}❌ DuckDB is not installed or not in PATH!${NC}"
    echo "Please install DuckDB first:"
    echo "  # macOS"
    echo "  brew install duckdb"
    echo "  # Linux"
    echo "  wget https://github.com/duckdb/duckdb/releases/latest/download/duckdb_cli-linux-amd64.zip"
    exit 1
fi

# Check if database already exists
if [ -f "$DB_FILE" ]; then
    echo -e "${YELLOW}⚠️  Database '$DB_FILE' already exists!${NC}"
    
    if [ "$FORCE" = false ]; then
        echo -e "${YELLOW}This will drop the existing database and recreate it.${NC}"
        echo -e "${RED}⚠️  ALL DATA IN '$DB_FILE' DATABASE WILL BE LOST!${NC}"
        echo ""
        echo -n "Are you sure you want to continue? (y/N): "
        read -r response
        
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            echo "Setup cancelled."
            exit 0
        fi
    fi
    
    # Remove existing database
    echo -e "${YELLOW}🔄 Removing existing database '$DB_FILE'...${NC}"
    rm "$DB_FILE"
    echo -e "${GREEN}✅ Existing database removed${NC}"
fi

# Create data directory if it doesn't exist
echo -e "${YELLOW}📁 Creating data directory...${NC}"
mkdir -p "$DATA_DIR"

# Check if data already exists
CSV_COUNT=$(find "$DATA_DIR" -name "*.csv" 2>/dev/null | wc -l)
if [ "$CSV_COUNT" -gt 0 ] && [ "$FORCE" = false ]; then
    echo -e "${YELLOW}⚠️  Found $CSV_COUNT CSV files in data directory${NC}"
    echo -e "${YELLOW}Use --force to re-download data${NC}"
    echo ""
    echo -e "${BLUE}📊 Existing files:${NC}"
    ls -lh "$DATA_DIR"/*.csv 2>/dev/null | head -5
    if [ "$CSV_COUNT" -gt 5 ]; then
        echo "   ... and $((CSV_COUNT - 5)) more files"
    fi
else
    # Download and extract data
    echo -e "${YELLOW}⬇️  Downloading Open Payments data...${NC}"
    
    cd "$DATA_DIR"
    
    # Clean up existing files if force mode
    if [ "$FORCE" = true ]; then
        echo -e "${YELLOW}🧹 Cleaning up existing files...${NC}"
        rm -f *.csv *.zip 2>/dev/null || true
    fi
    
    # Download the file
    if [ -f "$FILENAME" ]; then
        echo -e "${YELLOW}⚠️  Archive '$FILENAME' already exists. Removing...${NC}"
        rm "$FILENAME"
    fi
    
    echo "URL: $DOWNLOAD_URL"
    echo "Destination: $(pwd)/$FILENAME"
    
    if wget -q --show-progress "$DOWNLOAD_URL"; then
        echo -e "${GREEN}✅ Download completed successfully${NC}"
    else
        echo -e "${RED}❌ Download failed!${NC}"
        exit 1
    fi
    
    # Check file size
    FILE_SIZE=$(du -h "$FILENAME" | cut -f1)
    echo -e "${BLUE}📁 Downloaded file size: $FILE_SIZE${NC}"
    
    # Test zip file integrity
    echo -e "${YELLOW}🔍 Checking archive integrity...${NC}"
    if unzip -t "$FILENAME" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Archive is valid${NC}"
    else
        echo -e "${RED}❌ Archive is corrupted!${NC}"
        rm "$FILENAME"
        exit 1
    fi
    
    # Extract the archive
    echo -e "${YELLOW}📦 Extracting archive...${NC}"
    if unzip -q "$FILENAME"; then
        echo -e "${GREEN}✅ Extraction completed${NC}"
    else
        echo -e "${RED}❌ Extraction failed!${NC}"
        exit 1
    fi
    
    # Remove the archive
    echo -e "${YELLOW}🗑️  Removing archive...${NC}"
    rm "$FILENAME"
    echo -e "${GREEN}✅ Archive removed${NC}"
    
    # Show extracted files
    echo ""
    echo -e "${BLUE}📊 Extracted CSV files:${NC}"
    CSV_FILES=(*.csv)
    if [ ${#CSV_FILES[@]} -gt 0 ] && [ -f "${CSV_FILES[0]}" ]; then
        for file in "${CSV_FILES[@]}"; do
            if [ -f "$file" ]; then
                SIZE=$(du -h "$file" | cut -f1)
                LINES=$(wc -l < "$file" 2>/dev/null || echo "?")
                echo "   📄 $file ($SIZE, $LINES lines)"
            fi
        done
        CSV_COUNT=${#CSV_FILES[@]}
    else
        echo "   No CSV files found"
        CSV_COUNT=0
    fi
    
    cd - >/dev/null
fi

if [ "$CSV_COUNT" -eq 0 ]; then
    echo -e "${RED}❌ No CSV files found in data directory!${NC}"
    exit 1
fi

# Execute SQL script to create database and load data
echo -e "${YELLOW}🔄 Creating DuckDB database and executing SQL script...${NC}"
echo "Database: $DB_FILE"
echo "SQL Script: $SQL_SCRIPT"
echo "Data Directory: $DATA_DIR"

# Set environment variable for SQL script to use
export OPENPAYMENTS_DATA_DIR="$DATA_DIR"

if duckdb "$DB_FILE" < "$SQL_SCRIPT" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Database creation and data loading completed successfully${NC}"
else
    echo -e "${RED}❌ Failed to execute SQL script!${NC}"
    echo "Please check the SQL script and data files."
    exit 1
fi

# Show table list
echo -e "${BLUE}📊 Tables in database:${NC}"
duckdb "$DB_FILE" "SHOW TABLES;" 2>/dev/null | sed 's/^/ - /' || echo "   Could not retrieve table list"

# Show database statistics
echo ""
echo -e "${BLUE}📈 Database Statistics:${NC}"

# Get database file size
DB_SIZE=$(du -h "$DB_FILE" | cut -f1)
echo "   - Database file size: $DB_SIZE"

# Remove open payments data directory if it was created
if [ -d "$DATA_DIR" ]; then
    echo -e "${YELLOW}🗑️  Removing data directory '$DATA_DIR'...${NC}"
    rm -rf "$DATA_DIR"
    echo -e "${GREEN}✅ Data directory removed${NC}"
else
    echo -e "${YELLOW}📁 Data directory '$DATA_DIR' does not exist, skipping removal${NC}"
fi

# Count records in tables (if standard Open Payments tables exist)
GENERAL_COUNT=$(duckdb "$DB_FILE" "SELECT COUNT(*) FROM general_payments;" 2>/dev/null | tail -n 1 | tr -d '\r' || echo "N/A")
RESEARCH_COUNT=$(duckdb "$DB_FILE" "SELECT COUNT(*) FROM research_payments;" 2>/dev/null | tail -n 1 | tr -d '\r' || echo "N/A")
OWNERSHIP_COUNT=$(duckdb "$DB_FILE" "SELECT COUNT(*) FROM ownership_information;" 2>/dev/null | tail -n 1 | tr -d '\r' || echo "N/A")

if [ "$GENERAL_COUNT" != "N/A" ]; then
    echo "   - General Payments: $GENERAL_COUNT"
fi
if [ "$RESEARCH_COUNT" != "N/A" ]; then
    echo "   - Research Payments: $RESEARCH_COUNT"
fi
if [ "$OWNERSHIP_COUNT" != "N/A" ]; then
    echo "   - Ownership Information: $OWNERSHIP_COUNT"
fi

# Show total data directory size
DATA_SIZE=$(du -sh "$DATA_DIR" | cut -f1)

echo ""
echo -e "${GREEN}🎉 Open Payments database setup completed successfully!${NC}"
echo ""
echo -e "${BLUE}📝 Database details:${NC}"
echo "   Database: $DB_FILE"
echo "   Data Directory: $DATA_DIR ($DATA_SIZE)"
echo "   CSV Files: $CSV_COUNT"
echo ""
echo -e "${BLUE}🔗 Next steps:${NC}"
echo "   • Test connection: duckdb $DB_FILE"
echo "   • Query data: duckdb $DB_FILE \"SELECT COUNT(*) FROM duckdb_tables();\""
echo "   • Create GraphQL schema for hugr"
echo "   • Explore data structure and relationships"
echo ""
echo -e "${BLUE}💡 Example queries:${NC}"
echo "   duckdb $DB_FILE \"DESCRIBE general_payments;\""
echo "   duckdb $DB_FILE \"SELECT * FROM general_payments LIMIT 5;\""
echo ""