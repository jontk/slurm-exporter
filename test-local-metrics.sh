#!/bin/bash

# Test all API versions against local slurm-client
# Sequential testing with proper cleanup

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://rocky9.ar.jontk.com:6820"
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Njk0NzEzNTcsImlhdCI6MTc2OTQ2OTU1Nywic3VuIjoicm9vdCJ9.zZDnnFuxBR166CC8U2QnF6ZoDwzPg17CuXR0zmCp4to"
METRICS_PORT=9100
TEST_DIR="/tmp/metrics-local-test"
OUTPUT_DIR="${TEST_DIR}/results"

mkdir -p "$OUTPUT_DIR"

API_VERSIONS=("v0.0.40" "v0.0.41" "v0.0.42" "v0.0.43" "v0.0.44")

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Testing Local slurm-client Metrics Collection${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"

create_config() {
    local version=$1
    local config_file="${TEST_DIR}/config-${version}.yaml"

    cat > "$config_file" << EOF
slurm:
  api_version: "${version}"
  base_url: "${BASE_URL}"
  auth:
    type: "jwt"
    token: "${JWT_TOKEN}"

validation:
  allow_insecure_connections: true

collectors:
  cluster:
    enabled: true
    interval: 10s
  nodes:
    enabled: true
    interval: 10s
  jobs:
    enabled: true
    interval: 10s
  accounts:
    enabled: true
    interval: 10s
  partitions:
    enabled: true
    interval: 10s
  users:
    enabled: true
    interval: 10s
  qos:
    enabled: true
    interval: 10s
  tres:
    enabled: true
    interval: 10s
  licenses:
    enabled: true
    interval: 10s
  shares:
    enabled: true
    interval: 10s
  system:
    enabled: true
    interval: 10s

metrics:
  address: "127.0.0.1"
  port: ${METRICS_PORT}
EOF
    echo "$config_file"
}

test_version() {
    local version=$1

    echo -e "${YELLOW}━━━ Testing ${version} ━━━${NC}"

    # Kill any existing exporter
    pkill -9 slurm-exporter 2>/dev/null || true
    sleep 1

    local config_file=$(create_config "$version")
    local metrics_file="${OUTPUT_DIR}/metrics-${version}.txt"

    # Start exporter in background
    ./slurm-exporter --config "$config_file" > "${TEST_DIR}/log-${version}.txt" 2>&1 &
    EXPORTER_PID=$!

    # Wait for startup
    sleep 5

    # Check if still running
    if ! kill -0 $EXPORTER_PID 2>/dev/null; then
        echo -e "${RED}✗ Exporter failed to start${NC}"
        cat "${TEST_DIR}/log-${version}.txt" | tail -10
        return 1
    fi

    # Try to collect metrics
    if curl -s "http://127.0.0.1:${METRICS_PORT}/metrics" > "$metrics_file" 2>/dev/null; then
        local metric_count=$(grep "^slurm_" "$metrics_file" | grep -v "^slurm_exporter" | sed 's/{.*//' | sort -u | wc -l)
        echo -e "${GREEN}✓ Collected ${metric_count} unique SLURM metrics${NC}"
    else
        echo -e "${RED}✗ Failed to collect metrics${NC}"
        cat "${TEST_DIR}/log-${version}.txt" | tail -10
    fi

    # Cleanup
    kill $EXPORTER_PID 2>/dev/null || true
    wait $EXPORTER_PID 2>/dev/null || true
    sleep 2

    # Generate detailed analysis
    analyze_metrics "$version" "$metrics_file"
}

analyze_metrics() {
    local version=$1
    local metrics_file=$2
    local analysis_file="${OUTPUT_DIR}/analysis-${version}.txt"

    {
        echo "=== Metrics Analysis for ${version} ==="
        echo ""
        echo "Raw metric output:"
        grep "^slurm_" "$metrics_file" | grep -v "^slurm_exporter" | sed 's/{.*//' | sort -u
        echo ""
        echo "=== Summary ==="
        local total=$(grep "^slurm_" "$metrics_file" | grep -v "^slurm_exporter" | sed 's/{.*//' | sort -u | wc -l)
        echo "Total SLURM metrics: $total"
        echo ""
        echo "Breakdown by collector:"
        echo "  cluster:     $(grep "^slurm_cluster" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  nodes:       $(grep "^slurm_node" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  jobs:        $(grep "^slurm_job" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  accounts:    $(grep "^slurm_account" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  partitions:  $(grep "^slurm_partition" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  users:       $(grep "^slurm_user" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  qos:         $(grep "^slurm_qos" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  system:      $(grep "^slurm_system" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  tres:        $(grep "^slurm_tres" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  licenses:    $(grep "^slurm_licenses" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  shares:      $(grep "^slurm_shares" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
        echo "  scheduler:   $(grep "^slurm_scheduler" "$metrics_file" | sed 's/{.*//' | sort -u | wc -l) metrics"
    } > "$analysis_file"

    cat "$analysis_file"
}

# Build first
echo -e "${BLUE}Building exporter...${NC}\n"
if go build -o slurm-exporter ./cmd/slurm-exporter 2>&1 | grep -i error; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Build successful\n${NC}"

# Test each version
for version in "${API_VERSIONS[@]}"; do
    test_version "$version"
    echo ""
done

# Cleanup
pkill -9 slurm-exporter 2>/dev/null || true

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Analysis files saved to: $OUTPUT_DIR${NC}\n"
