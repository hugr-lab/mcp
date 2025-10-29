#!/usr/bin/env bash
#
# Integration Test: MinIO Health and S3 Operations (T044)
#
# Tests MinIO service health, S3 operations, and console accessibility
# Can run in parallel with T042-T043, T045-T046
#
# Dependencies: T023-T029, T041
# Exit codes: 0 = pass, 1 = fail

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
PASS="${GREEN}✓${NC}"
FAIL="${RED}✗${NC}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Integration Test: MinIO (T044)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
LDE_DIR="${PROJECT_ROOT}/lde"

# Track test status
TEST_FAILED=0

# Cleanup function
cleanup() {
    echo ""
    echo "→ Cleaning up..."
    cd "${LDE_DIR}" && ./scripts/stop.sh > /dev/null 2>&1 || true
    echo -e "${PASS} Environment stopped"
}

# Set trap to ensure cleanup runs
trap cleanup EXIT

# Step 1: Start environment with --no-data
echo ""
echo "→ Starting environment (--no-data)..."
if cd "${LDE_DIR}" && ./scripts/start.sh --no-data > /tmp/minio-test-start.log 2>&1; then
    echo -e "${PASS} Environment started"
else
    echo -e "${FAIL} Failed to start environment"
    cat /tmp/minio-test-start.log
    exit 1
fi

# Wait a moment for services to fully initialize
sleep 2

# Step 2: Test health endpoint
echo ""
echo "→ Testing MinIO health endpoint..."
if curl -f -s http://localhost:9000/minio/health/live > /tmp/minio-test-health.txt 2>&1; then
    echo -e "${PASS} MinIO health endpoint accessible"
else
    echo -e "${FAIL} MinIO health endpoint check failed"
    cat /tmp/minio-test-health.txt
    TEST_FAILED=1
fi

# Step 3: Test S3 operations (create bucket, upload, download)
echo ""
echo "→ Testing S3 operations..."

# MinIO credentials from .env
MINIO_USER="minioadmin"
MINIO_PASS="minioadmin"
TEST_BUCKET="test-integration-$(date +%s)"
TEST_FILE="/tmp/minio-test-file.txt"
TEST_CONTENT="MinIO integration test content - $(date)"

# Create test file
echo "$TEST_CONTENT" > "$TEST_FILE"

# Configure AWS CLI to use MinIO
export AWS_ACCESS_KEY_ID="$MINIO_USER"
export AWS_SECRET_ACCESS_KEY="$MINIO_PASS"
export AWS_DEFAULT_REGION="us-east-1"
MINIO_ENDPOINT="http://localhost:9000"

# Create bucket
echo "  → Creating bucket '$TEST_BUCKET'..."
if aws --endpoint-url "$MINIO_ENDPOINT" s3 mb "s3://$TEST_BUCKET" > /tmp/minio-test-bucket.txt 2>&1; then
    echo -e "  ${PASS} Bucket created"
else
    echo -e "  ${FAIL} Failed to create bucket"
    cat /tmp/minio-test-bucket.txt
    TEST_FAILED=1
fi

# Upload file
echo "  → Uploading test file..."
if aws --endpoint-url "$MINIO_ENDPOINT" s3 cp "$TEST_FILE" "s3://$TEST_BUCKET/test-object.txt" > /tmp/minio-test-upload.txt 2>&1; then
    echo -e "  ${PASS} File uploaded"
else
    echo -e "  ${FAIL} Failed to upload file"
    cat /tmp/minio-test-upload.txt
    TEST_FAILED=1
fi

# Download file
echo "  → Downloading test file..."
DOWNLOADED_FILE="/tmp/minio-test-downloaded.txt"
if aws --endpoint-url "$MINIO_ENDPOINT" s3 cp "s3://$TEST_BUCKET/test-object.txt" "$DOWNLOADED_FILE" > /tmp/minio-test-download.txt 2>&1; then
    echo -e "  ${PASS} File downloaded"

    # Verify content matches
    if diff "$TEST_FILE" "$DOWNLOADED_FILE" > /dev/null 2>&1; then
        echo -e "  ${PASS} File content matches"
    else
        echo -e "  ${FAIL} File content mismatch"
        TEST_FAILED=1
    fi
else
    echo -e "  ${FAIL} Failed to download file"
    cat /tmp/minio-test-download.txt
    TEST_FAILED=1
fi

# Cleanup test bucket
aws --endpoint-url "$MINIO_ENDPOINT" s3 rb "s3://$TEST_BUCKET" --force > /dev/null 2>&1 || true
rm -f "$TEST_FILE" "$DOWNLOADED_FILE"

# Step 4: Verify console accessible
echo ""
echo "→ Testing MinIO console accessibility..."
if curl -f -s -o /dev/null http://localhost:9001/login 2>&1; then
    echo -e "${PASS} MinIO console accessible"
else
    echo -e "${FAIL} MinIO console not accessible"
    TEST_FAILED=1
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $TEST_FAILED -eq 0 ]; then
    echo -e "${PASS} MinIO Integration Test PASSED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${FAIL} MinIO Integration Test FAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
