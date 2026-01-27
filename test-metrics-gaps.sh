#!/bin/bash

# Comprehensive metrics gap analysis script
# Tests all API versions against the local slurm-client
# Compares actual metrics collected vs. expected specification

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://rocky9.ar.jontk.com:6820"
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Njk0NzEzNTcsImlhdCI6MTc2OTQ2OTU1Nywic3VuIjoicm9vdCJ9.zZDnnFuxBR166CC8U2QnF6ZoDwzPg17CuXR0zmCp4to"
METRICS_PORT=9100
EXPORTER_TIMEOUT=10
TEST_DIR="/tmp/metrics-gap-test"
OUTPUT_DIR="${TEST_DIR}/results"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# API versions to test
API_VERSIONS=("v0.0.40" "v0.0.41" "v0.0.42" "v0.0.43" "v0.0.44")

# Expected metrics per collector (from METRICS-SPECIFICATION.md)
declare -A EXPECTED_METRICS=(
    ["cluster"]=6
    ["nodes"]=6
    ["jobs"]=7
    ["accounts"]=6
    ["partitions"]=11
    ["users"]=6
    ["qos"]=13
    ["tres"]=5
    ["licenses"]=5
    ["shares"]=7
    ["system"]=14
    ["scheduler"]=3
)

# Mapping of metric prefixes to collectors
declare -A METRIC_PREFIXES=(
    ["slurm_cluster"]=cluster
    ["slurm_node"]=nodes
    ["slurm_job"]=jobs
    ["slurm_account"]=accounts
    ["slurm_partition"]=partitions
    ["slurm_user"]=users
    ["slurm_qos"]=qos
    ["slurm_tres"]=tres
    ["slurm_licenses"]=licenses
    ["slurm_shares"]=shares
    ["slurm_system"]=system
    ["slurm_scheduler"]=scheduler
)

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}SLURM Exporter Metrics Gap Analysis${NC}"
echo -e "${BLUE}Local slurm-client Testing${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Create test configuration files
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

collector:
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

# Test a specific API version
test_version() {
    local version=$1
    local port=$((9100 + $(echo "$version" | tr -cd '0-9' | awk '{print $1 % 100}')))

    # Kill any existing process on this port
    lsof -ti :$port | xargs kill -9 2>/dev/null || true
    sleep 1

    local config_file=$(create_config "$version")
    local metrics_file="${OUTPUT_DIR}/metrics-${version}.txt"
    local analysis_file="${OUTPUT_DIR}/analysis-${version}.txt"

    echo -e "${YELLOW}Testing API Version: ${version}${NC}"

    # Update port in config
    sed -i "s/port: ${METRICS_PORT}/port: ${port}/" "$config_file"

    # Start exporter
    timeout ${EXPORTER_TIMEOUT} ./slurm-exporter --config "$config_file" 2>&1 | grep -i "listening\|error\|warn" &
    EXPORTER_PID=$!

    # Wait for exporter to start
    sleep 3

    # Collect metrics
    if curl -s "http://127.0.0.1:${port}/metrics" > "$metrics_file" 2>/dev/null; then
        echo -e "${GREEN}✓ Metrics collected successfully${NC}"
    else
        echo -e "${RED}✗ Failed to collect metrics${NC}"
        kill $EXPORTER_PID 2>/dev/null || true
        return 1
    fi

    # Kill exporter
    kill $EXPORTER_PID 2>/dev/null || true
    wait $EXPORTER_PID 2>/dev/null || true
    sleep 1

    # Analyze metrics
    analyze_metrics "$version" "$metrics_file" "$analysis_file"
}

# Analyze collected metrics against specification
analyze_metrics() {
    local version=$1
    local metrics_file=$2
    local analysis_file=$3

    echo -e "\n${BLUE}Analysis for ${version}:${NC}" | tee -a "$analysis_file"

    # Extract unique SLURM metrics (excluding exporter metrics)
    local unique_metrics=$(grep "^slurm_" "$metrics_file" | grep -v "^slurm_exporter" | sed 's/{.*//' | sort -u)
    local metric_count=$(echo "$unique_metrics" | wc -l)

    echo "Total unique metrics collected: $metric_count" | tee -a "$analysis_file"
    echo "" | tee -a "$analysis_file"

    # Analyze by collector
    echo -e "${YELLOW}Metrics by Collector:${NC}" | tee -a "$analysis_file"

    local total_expected=0
    local total_found=0

    for prefix in "${!METRIC_PREFIXES[@]}"; do
        local collector="${METRIC_PREFIXES[$prefix]}"
        local expected="${EXPECTED_METRICS[$collector]}"
        local found=$(echo "$unique_metrics" | grep "^${prefix}" | wc -l)

        total_expected=$((total_expected + expected))
        total_found=$((total_found + found))

        local status="✓"
        if [ "$found" -lt "$expected" ]; then
            status="${RED}✗${NC}"
        fi

        printf "  ${status} %-20s %2d/%2d metrics\n" "$collector:" "$found" "$expected" | tee -a "$analysis_file"
    done

    echo "" | tee -a "$analysis_file"
    local coverage=$((found * 100 / total_expected))
    echo -e "Total Coverage: ${BLUE}${total_found}/${total_expected} (${coverage}%)${NC}" | tee -a "$analysis_file"
    echo "" | tee -a "$analysis_file"

    # List missing metrics by collector
    echo -e "${YELLOW}Missing Metrics by Collector:${NC}" | tee -a "$analysis_file"

    local has_missing=false
    for prefix in "${!METRIC_PREFIXES[@]}"; do
        local collector="${METRIC_PREFIXES[$prefix]}"
        local expected="${EXPECTED_METRICS[$collector]}"
        local found=$(echo "$unique_metrics" | grep "^${prefix}" | wc -l)

        if [ "$found" -lt "$expected" ]; then
            has_missing=true
            echo "  ${RED}${collector}:${NC} missing $((expected - found)) metrics" | tee -a "$analysis_file"

            # Try to identify which specific metrics are missing
            case "$collector" in
                cluster)
                    check_metric "$unique_metrics" "slurm_cluster_info" "$analysis_file"
                    check_metric "$unique_metrics" "slurm_cluster_nodes_total" "$analysis_file"
                    ;;
                jobs)
                    check_metric "$unique_metrics" "slurm_job_state" "$analysis_file"
                    check_metric "$unique_metrics" "slurm_job_cpus" "$analysis_file"
                    check_metric "$unique_metrics" "slurm_job_run_time_seconds" "$analysis_file"
                    ;;
                shares)
                    check_metric "$unique_metrics" "slurm_shares_usage_factor" "$analysis_file"
                    ;;
                licenses)
                    check_metric "$unique_metrics" "slurm_licenses_total" "$analysis_file"
                    ;;
                tres)
                    check_metric "$unique_metrics" "slurm_tres_info" "$analysis_file"
                    ;;
            esac
        fi
    done

    if [ "$has_missing" = false ]; then
        echo "  ${GREEN}All expected metrics present!${NC}" | tee -a "$analysis_file"
    fi
    echo "" | tee -a "$analysis_file"
}

# Check if a specific metric exists
check_metric() {
    local metrics=$1
    local metric_name=$2
    local output_file=$3

    if echo "$metrics" | grep -q "^${metric_name}"; then
        printf "    ${GREEN}✓${NC} %s\n" "$metric_name" | tee -a "$output_file"
    else
        printf "    ${RED}✗${NC} %s (MISSING)\n" "$metric_name" | tee -a "$output_file"
    fi
}

# Main test execution
main() {
    echo -e "\nBuilding exporter with local slurm-client..."
    if ! go build -o slurm-exporter ./cmd/slurm-exporter 2>&1 | grep -i error; then
        echo -e "${GREEN}✓ Build successful${NC}\n"
    else
        echo -e "${RED}✗ Build failed${NC}"
        exit 1
    fi

    # Test each version
    for version in "${API_VERSIONS[@]}"; do
        echo ""
        test_version "$version" || true
        sleep 1
    done

    # Create summary report
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}Summary Report${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"

    # Compare all versions
    echo -e "${YELLOW}Coverage by Version:${NC}\n"
    for version in "${API_VERSIONS[@]}"; do
        if [ -f "${OUTPUT_DIR}/analysis-${version}.txt" ]; then
            local coverage=$(grep "Total Coverage:" "${OUTPUT_DIR}/analysis-${version}.txt" | tail -1)
            echo "  $version: $coverage"
        fi
    done

    echo -e "\n${YELLOW}Detailed analysis files saved to:${NC} $OUTPUT_DIR"
    echo -e "  - metrics-*.txt (raw metric output)"
    echo -e "  - analysis-*.txt (detailed analysis)\n"
}

main "$@"
