#!/bin/bash

# Comprehensive test of all SLURM API versions
# Tests v0.0.40 through v0.0.44 and compares metrics against specification

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://rocky9.ar.jontk.com:6820"
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Njk0NzEzNTcsImlhdCI6MTc2OTQ2OTU1Nywic3VuIjoicm9vdCJ9.zZDnnFuxBR166CC8U2QnF6ZoDwzPg17CuXR0zmCp4to"
OUTPUT_DIR="/tmp/metrics-comprehensive-test"
METRICS_PORT=19100

mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Comprehensive SLURM API Version Testing${NC}"
echo -e "${BLUE}Testing v0.0.40 through v0.0.44 with Local slurm-client${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"

cleanup_exporter() {
    pkill -9 slurm-exporter 2>/dev/null || true
    sleep 2
}

create_config() {
    local version=$1
    local output_file=$2

    cat > "$output_file" << EOF
slurm:
  api_version: "$version"
  base_url: "$BASE_URL"
  auth:
    type: "jwt"
    token: "$JWT_TOKEN"
  timeout: 30s

validation:
  allow_insecure_connections: true

metrics:
  address: "127.0.0.1"
  port: $METRICS_PORT

server:
  address: "127.0.0.1:$METRICS_PORT"

collectors:
  global:
    default_interval: 10s
    default_timeout: 5s

  cluster:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  nodes:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  jobs:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  accounts:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  associations:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  partitions:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  users:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  qos:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  reservations:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  licenses:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  shares:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  diagnostics:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  tres:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  wckeys:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  clusters:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  performance:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true
  system:
    enabled: true
    interval: 10s
    filters:
      metrics:
        enable_all: true

logging:
  level: "info"
EOF
}

test_version() {
    local version=$1
    echo -e "${YELLOW}Testing ${version}...${NC}"

    cleanup_exporter

    local config="/tmp/config-${version}.yaml"
    create_config "$version" "$config"

    local metrics_file="${OUTPUT_DIR}/metrics-${version}.txt"
    local summary_file="${OUTPUT_DIR}/summary-${version}.txt"

    # Start exporter
    timeout 20 ./slurm-exporter --config "$config" > /dev/null 2>&1 &
    local PID=$!

    sleep 8

    # Collect metrics
    if curl -s "http://127.0.0.1:$METRICS_PORT/metrics" > "$metrics_file" 2>/dev/null; then
        # Extract unique SLURM metrics (exclude internal exporter metrics)
        local slurm_metrics=$(grep "^slurm_" "$metrics_file" | grep -v "^slurm_exporter" | sed 's/{.*//' | sort -u)
        local total=$(echo "$slurm_metrics" | wc -l)

        echo -e "${GREEN}✓ Collected ${total} unique SLURM metrics${NC}"

        # Generate detailed summary
        {
            echo "=== API Version: ${version} ==="
            echo ""
            echo "Total SLURM metrics: $total"
            echo ""
            echo "Breakdown by collector:"
            echo "  cluster:      $(echo "$slurm_metrics" | grep -c "^slurm_cluster" || true) metrics"
            echo "  nodes:        $(echo "$slurm_metrics" | grep -c "^slurm_node" || true) metrics"
            echo "  jobs:         $(echo "$slurm_metrics" | grep -c "^slurm_job" || true) metrics"
            echo "  accounts:     $(echo "$slurm_metrics" | grep -c "^slurm_account" || true) metrics"
            echo "  partitions:   $(echo "$slurm_metrics" | grep -c "^slurm_partition" || true) metrics"
            echo "  users:        $(echo "$slurm_metrics" | grep -c "^slurm_user" || true) metrics"
            echo "  qos:          $(echo "$slurm_metrics" | grep -c "^slurm_qos" || true) metrics"
            echo "  system:       $(echo "$slurm_metrics" | grep -c "^slurm_system" || true) metrics"
            echo "  tres:         $(echo "$slurm_metrics" | grep -c "^slurm_tres" || true) metrics"
            echo "  licenses:     $(echo "$slurm_metrics" | grep -c "^slurm_licenses" || true) metrics"
            echo "  shares:       $(echo "$slurm_metrics" | grep -c "^slurm_shares" || true) metrics"
            echo "  associations: $(echo "$slurm_metrics" | grep -c "^slurm_association" || true) metrics"
            echo "  diagnostics:  $(echo "$slurm_metrics" | grep -c "^slurm_diagnostics" || true) metrics"
            echo "  wckeys:       $(echo "$slurm_metrics" | grep -c "^slurm_wckeys" || true) metrics"
            echo "  performance:  $(echo "$slurm_metrics" | grep -c "^slurm_performance" || true) metrics"
            echo "  clusters:     $(echo "$slurm_metrics" | grep -c "^slurm_clusters" || true) metrics"
            echo ""
            echo "All metrics collected:"
            echo "$slurm_metrics"
        } > "$summary_file"

        cat "$summary_file" | head -30
        echo -e "${GREEN}✓ Full summary saved to: $summary_file${NC}"
    else
        echo -e "${RED}✗ Failed to collect metrics for ${version}${NC}"
    fi

    kill $PID 2>/dev/null || true
    wait $PID 2>/dev/null || true
    echo ""
}

# Test each version
for version in v0.0.40 v0.0.41 v0.0.42 v0.0.43 v0.0.44; do
    test_version "$version"
done

# Create comparison
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Version Comparison Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"

{
    echo "Version | Total | Cluster | Nodes | Jobs | Accounts | Partitions | Users | QoS | System | TRES | Licenses | Shares | Associations | Diagnostics | WCKeys | Performance | Clusters"
    echo "--------|-------|---------|-------|------|----------|------------|-------|-----|--------|------|----------|--------|--------------|-------------|--------|-------------|----------"

    for version in v0.0.40 v0.0.41 v0.0.42 v0.0.43 v0.0.44; do
        if [ -f "$OUTPUT_DIR/summary-${version}.txt" ]; then
            local total=$(grep "^Total SLURM" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $NF}')
            local cluster=$(grep "cluster:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local nodes=$(grep "nodes:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local jobs=$(grep "jobs:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local accounts=$(grep "accounts:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local partitions=$(grep "partitions:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local users=$(grep "users:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local qos=$(grep "qos:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local system=$(grep "system:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local tres=$(grep "tres:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local licenses=$(grep "licenses:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local shares=$(grep "shares:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local associations=$(grep "associations:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local diagnostics=$(grep "diagnostics:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local wckeys=$(grep "wckeys:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local performance=$(grep "performance:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            local clusters=$(grep "clusters:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')

            printf "%s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s\n" \
                "$version" "$total" "$cluster" "$nodes" "$jobs" "$accounts" "$partitions" "$users" "$qos" "$system" "$tres" "$licenses" "$shares" "$associations" "$diagnostics" "$wckeys" "$performance" "$clusters"
        fi
    done
} | tee "$OUTPUT_DIR/comparison.txt"

cleanup_exporter

echo -e "\n${BLUE}Results saved to: ${OUTPUT_DIR}${NC}"
echo "  - metrics-*.txt (raw Prometheus output)"
echo "  - summary-*.txt (detailed analysis per version)"
echo "  - comparison.txt (version comparison table)"
