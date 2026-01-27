#!/bin/bash
set -euo pipefail

# SLURM Server Configuration
SLURM_HOST="rocky9.ar.jontk.com"
SLURM_PORT="6820"
SLURM_BASE_URL="http://${SLURM_HOST}:${SLURM_PORT}"

# Test Configuration
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXPORTER_PORT="9341"
EXPORTER_LOG="/tmp/slurm-exporter-adapter-test.log"
RESULTS_FILE="/tmp/adapter-test-results.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================================"
echo "SLURM Exporter Adapter Integration Test"
echo "================================================"
echo "Target SLURM Server: $SLURM_BASE_URL"
echo "Exporter Port: $EXPORTER_PORT"
echo ""

# Step 1: Get JWT Token
echo "[1/5] Acquiring JWT token from SLURM server..."
JWT_TOKEN=$(ssh "root@${SLURM_HOST}" "scontrol token" | grep -oP 'token=\K.*' || true)

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}✗ Failed to acquire JWT token${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Token acquired (length: ${#JWT_TOKEN})${NC}"
TOKEN_PREVIEW="${JWT_TOKEN:0:20}...${JWT_TOKEN: -20}"
echo "  Token: $TOKEN_PREVIEW"
echo ""

# Step 2: Create test configuration with auto-detection
echo "[2/5] Creating test configuration..."

cat > "/tmp/slurm-exporter-adapter-test.yaml" <<EOF
server:
  address: ":${EXPORTER_PORT}"
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"
  timeout: 30s
  read_timeout: 15s
  write_timeout: 10s
  idle_timeout: 60s

slurm:
  base_url: "${SLURM_BASE_URL}"
  api_version: ""  # Empty for auto-detection
  use_adapters: true
  auth:
    type: jwt
    token: "${JWT_TOKEN}"
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s
  rate_limit:
    requests_per_second: 10.0
    burst_size: 20

collectors:
  global:
    default_interval: 30s
    default_timeout: 10s
    max_concurrency: 5
  cluster:
    enabled: true
    interval: 30s
    timeout: 10s
  nodes:
    enabled: true
    interval: 30s
    timeout: 10s
  jobs:
    enabled: true
    interval: 15s
    timeout: 10s

logging:
  level: debug
  format: json
  output: stdout

metrics:
  namespace: slurm
  subsystem: exporter
EOF

if [ ! -f "/tmp/slurm-exporter-adapter-test.yaml" ]; then
    echo -e "${RED}✗ Failed to create configuration file${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Configuration created${NC}"
echo "  Config: /tmp/slurm-exporter-adapter-test.yaml"
echo ""

# Step 3: Build and start exporter
echo "[3/5] Building and starting exporter..."

# Kill any existing exporter
pkill -f "slurm-exporter" || true
sleep 1

# Build the exporter if needed
if [ ! -f "./cmd/slurm-exporter/slurm-exporter" ]; then
    echo "  Building slurm-exporter..."
    go build -o cmd/slurm-exporter/slurm-exporter ./cmd/slurm-exporter
fi

# Start the exporter in background
./cmd/slurm-exporter/slurm-exporter \
    -config /tmp/slurm-exporter-adapter-test.yaml \
    > "$EXPORTER_LOG" 2>&1 &

EXPORTER_PID=$!
echo -e "${GREEN}✓ Exporter started (PID: $EXPORTER_PID)${NC}"

# Wait for exporter to be ready
echo "  Waiting for exporter to become ready..."
MAX_WAIT=30
WAIT_TIME=0

while [ $WAIT_TIME -lt $MAX_WAIT ]; do
    if curl -sf "http://localhost:${EXPORTER_PORT}/ready" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Exporter is ready${NC}"
        break
    fi
    echo -n "."
    sleep 1
    WAIT_TIME=$((WAIT_TIME + 1))
done

if [ $WAIT_TIME -ge $MAX_WAIT ]; then
    echo -e "${RED}✗ Exporter failed to become ready${NC}"
    echo "Exporter logs:"
    tail -30 "$EXPORTER_LOG"
    kill $EXPORTER_PID || true
    exit 1
fi

echo ""

# Step 4: Test adapter functionality
echo "[4/5] Testing adapter functionality..."

# Check API version detection
echo "  Checking API version detection..."
METRICS=$(curl -sf "http://localhost:${EXPORTER_PORT}/metrics" 2>/dev/null || echo "")

if echo "$METRICS" | grep -q "slurm_"; then
    echo -e "${GREEN}✓ Successfully retrieved SLURM metrics${NC}"

    # Count metrics
    METRIC_COUNT=$(echo "$METRICS" | grep -c "^slurm_" || true)
    echo "    Metric samples: ~$METRIC_COUNT"
else
    echo -e "${YELLOW}⚠ No SLURM metrics found${NC}"
fi

# Check health status
echo "  Checking health status..."
HEALTH=$(curl -sf "http://localhost:${EXPORTER_PORT}/health" 2>/dev/null || echo "{}")

if echo "$HEALTH" | grep -q "healthy"; then
    echo -e "${GREEN}✓ Exporter is healthy${NC}"
else
    echo -e "${YELLOW}⚠ Exporter health status unclear${NC}"
fi

# Check exporter logs for API version detection
echo "  Checking logs for API version auto-detection..."
if grep -q "Auto-detected SLURM API version" "$EXPORTER_LOG"; then
    DETECTED_VERSION=$(grep "Auto-detected SLURM API version" "$EXPORTER_LOG" | head -1 | grep -oP 'version=\K\S+' || true)
    echo -e "${GREEN}✓ API version auto-detection occurred${NC}"
    echo "    Detected version: $DETECTED_VERSION"
else
    echo -e "${YELLOW}⚠ API version auto-detection message not found${NC}"
fi

echo ""

# Step 5: Collect test results
echo "[5/5] Collecting test results..."

# Create results JSON
cat > "$RESULTS_FILE" <<EOF
{
  "test_run": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "slurm_server": "${SLURM_BASE_URL}",
    "exporter_port": ${EXPORTER_PORT},
    "configuration": {
      "use_adapters": true,
      "api_version_auto_detection": true,
      "auth_type": "jwt"
    }
  },
  "results": {
    "exporter_running": $([ -d "/proc/$EXPORTER_PID" ] && echo "true" || echo "false"),
    "metrics_available": $([ -n "$METRICS" ] && echo "true" || echo "false"),
    "health_check_passed": $(echo "$HEALTH" | grep -q "healthy" && echo "true" || echo "false"),
    "api_version_detected": $(grep -q "Auto-detected SLURM API version" "$EXPORTER_LOG" && echo "true" || echo "false")
  },
  "logs": {
    "exporter_log": "${EXPORTER_LOG}"
  }
}
EOF

echo -e "${GREEN}✓ Results saved${NC}"
echo "  Results file: $RESULTS_FILE"
echo ""

# Summary
echo "================================================"
echo "Test Summary:"
echo "================================================"

if [ -d "/proc/$EXPORTER_PID" ]; then
    echo -e "${GREEN}✓ Exporter running${NC}"
else
    echo -e "${RED}✗ Exporter not running${NC}"
fi

if [ -n "$METRICS" ]; then
    echo -e "${GREEN}✓ Metrics available${NC}"
else
    echo -e "${RED}✗ No metrics available${NC}"
fi

if echo "$HEALTH" | grep -q "healthy"; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${YELLOW}⚠ Health check unclear${NC}"
fi

if grep -q "Auto-detected SLURM API version" "$EXPORTER_LOG"; then
    echo -e "${GREEN}✓ API version auto-detection working${NC}"
else
    echo -e "${YELLOW}⚠ API version auto-detection not detected${NC}"
fi

echo ""
echo "To view detailed logs:"
echo "  tail -f $EXPORTER_LOG"
echo ""
echo "To keep exporter running, press Ctrl+C"
echo "Exporter PID: $EXPORTER_PID"
echo ""

# Keep the test running and show logs
trap "kill $EXPORTER_PID || true; exit 0" EXIT INT TERM
tail -f "$EXPORTER_LOG"
