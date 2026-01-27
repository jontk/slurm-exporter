#!/bin/bash

# Test all SLURM API versions with local slurm-client
# Collect actual metrics and compare against specification

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://rocky9.ar.jontk.com:6820"
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Njk0NzEzNTcsImlhdCI6MTc2OTQ2OTU1Nywic3VuIjoicm9vdCJ9.zZDnnFuxBR166CC8U2QnF6ZoDwzPg17CuXR0zmCp4to"
OUTPUT_DIR="/tmp/metrics-all-versions"

mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Testing All SLURM API Versions - Local slurm-client${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"

test_version() {
    local version=$1
    echo -e "${YELLOW}Testing ${version}...${NC}"

    local config="/tmp/config-${version}.yaml"
    cat > "$config" << EOF
slurm:
  api_version: "${version}"
  base_url: "${BASE_URL}"
  auth:
    type: "jwt"
    token: "${JWT_TOKEN}"

validation:
  allow_insecure_connections: true

metrics:
  address: "127.0.0.1"
  port: 8080
EOF

    local metrics_file="${OUTPUT_DIR}/metrics-${version}.txt"
    local summary_file="${OUTPUT_DIR}/summary-${version}.txt"

    # Run exporter briefly and collect metrics
    timeout 15 ./slurm-exporter --config "$config" > /dev/null 2>&1 &
    local PID=$!

    sleep 5

    # Collect metrics
    if curl -s "http://127.0.0.1:8080/metrics" > "$metrics_file" 2>/dev/null; then
        # Extract unique SLURM metrics
        local slurm_metrics=$(grep "^slurm_" "$metrics_file" | grep -v "^slurm_exporter" | sed 's/{.*//' | sort -u)
        local total=$(echo "$slurm_metrics" | wc -l)

        echo -e "${GREEN}✓ Collected ${total} metrics${NC}"

        # Generate summary
        {
            echo "=== API Version: ${version} ==="
            echo ""
            echo "Total SLURM metrics: $total"
            echo ""
            echo "Breakdown by collector:"
            echo "  cluster:     $(echo "$slurm_metrics" | grep "^slurm_cluster" | wc -l) metrics"
            echo "  nodes:       $(echo "$slurm_metrics" | grep "^slurm_node" | wc -l) metrics"
            echo "  jobs:        $(echo "$slurm_metrics" | grep "^slurm_job" | wc -l) metrics"
            echo "  accounts:    $(echo "$slurm_metrics" | grep "^slurm_account" | wc -l) metrics"
            echo "  partitions:  $(echo "$slurm_metrics" | grep "^slurm_partition" | wc -l) metrics"
            echo "  users:       $(echo "$slurm_metrics" | grep "^slurm_user" | wc -l) metrics"
            echo "  qos:         $(echo "$slurm_metrics" | grep "^slurm_qos" | wc -l) metrics"
            echo "  system:      $(echo "$slurm_metrics" | grep "^slurm_system" | wc -l) metrics"
            echo "  tres:        $(echo "$slurm_metrics" | grep "^slurm_tres" | wc -l) metrics"
            echo "  licenses:    $(echo "$slurm_metrics" | grep "^slurm_licenses" | wc -l) metrics"
            echo "  shares:      $(echo "$slurm_metrics" | grep "^slurm_shares" | wc -l) metrics"
            echo "  scheduler:   $(echo "$slurm_metrics" | grep "^slurm_scheduler" | wc -l) metrics"
            echo "  other:       $(echo "$slurm_metrics" | grep -v "^slurm_\(cluster\|node\|job\|account\|partition\|user\|qos\|system\|tres\|licenses\|shares\|scheduler\)" | wc -l) metrics"
            echo ""
            echo "Detailed metric list:"
            echo "$slurm_metrics"
        } > "$summary_file"

        cat "$summary_file"
    else
        echo -e "${RED}✗ Failed to collect metrics${NC}"
    fi

    # Kill exporter
    kill $PID 2>/dev/null || true
    wait $PID 2>/dev/null || true
    sleep 2

    echo ""
}

# Build
echo -e "${BLUE}Building exporter...${NC}"
go build -o slurm-exporter ./cmd/slurm-exporter 2>&1 | grep -i error || echo -e "${GREEN}✓ Build successful${NC}"
echo ""

# Test each version
for version in v0.0.40 v0.0.41 v0.0.42 v0.0.43 v0.0.44; do
    test_version "$version"
done

# Create comparison
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Summary Comparison${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"

{
    echo "Version | Total | Cluster | Nodes | Jobs | Accounts | Partitions | Users | QoS | System | TRES | Licenses | Shares | Scheduler"
    echo "--------|-------|---------|-------|------|----------|------------|-------|-----|--------|------|----------|--------|----------"
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
            local scheduler=$(grep "scheduler:" "$OUTPUT_DIR/summary-${version}.txt" | awk '{print $2}')
            printf "%s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s\n" \
                "$version" "$total" "$cluster" "$nodes" "$jobs" "$accounts" "$partitions" "$users" "$qos" "$system" "$tres" "$licenses" "$shares" "$scheduler"
        fi
    done
} | tee "$OUTPUT_DIR/comparison.txt"

echo -e "\n${BLUE}Results saved to:${NC} $OUTPUT_DIR"
echo "  - metrics-*.txt (raw output)"
echo "  - summary-*.txt (detailed analysis)"
echo "  - comparison.txt (version comparison)"
