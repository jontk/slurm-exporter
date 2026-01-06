#!/run/current-system/sw/bin/bash

# Quick validation script for SLURM cluster connectivity
# Tests the connection to rocky.ar.jontk.com before running full integration tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_HOST="rocky9.ar.jontk.com"
SLURM_JWT="${SLURM_JWT:-}"

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

validate_prerequisites() {
    log "Validating prerequisites..."
    
    # Check if JWT token is provided
    if [[ -z "${SLURM_JWT}" ]]; then
        error "SLURM_JWT environment variable is required"
        echo "Usage: SLURM_JWT=your-token-here $0"
        exit 1
    fi
    
    # Check required tools
    local tools=("curl" "jq")
    for tool in "${tools[@]}"; do
        if ! command -v "${tool}" >/dev/null 2>&1; then
            error "Required tool '${tool}' is not installed"
            exit 1
        fi
    done
    
    success "Prerequisites validated"
}

test_network_connectivity() {
    log "Testing network connectivity to ${CLUSTER_HOST}..."
    
    # Test basic connectivity
    if ping -c 1 -W 5 "${CLUSTER_HOST}" >/dev/null 2>&1; then
        success "Host ${CLUSTER_HOST} is reachable"
    else
        warning "Host ${CLUSTER_HOST} is not responding to ping (may be firewalled)"
    fi
    
    # Test HTTPS port
    if timeout 10 bash -c "</dev/tcp/${CLUSTER_HOST}/6820" 2>/dev/null; then
        success "Port 6820 is open on ${CLUSTER_HOST}"
    else
        error "Cannot connect to port 6820 on ${CLUSTER_HOST}"
        return 1
    fi
}

test_slurm_api() {
    log "Testing SLURM REST API endpoints..."
    
    local auth_header="X-SLURM-USER-TOKEN: ${SLURM_JWT}"
    local base_url="http://${CLUSTER_HOST}:6820/slurm/v0.0.43"
    
    # Test ping endpoint
    log "Testing ping endpoint..."
    local ping_response=$(curl -k -s -H "${auth_header}" "${base_url}/ping" 2>/dev/null)
    if [[ -n "${ping_response}" ]]; then
        success "SLURM ping endpoint responding"
        echo "Response: ${ping_response}" | jq . 2>/dev/null || echo "Response: ${ping_response}"
    else
        error "SLURM ping endpoint not responding"
        return 1
    fi
    
    # Test diag endpoint
    log "Testing diag endpoint..."
    local diag_response=$(curl -k -s -H "${auth_header}" "${base_url}/diag" 2>/dev/null)
    if [[ -n "${diag_response}" ]]; then
        success "SLURM diag endpoint responding"
        
        # Extract some basic info
        local slurm_version=$(echo "${diag_response}" | jq -r '.meta.Slurm.version.major // "unknown"' 2>/dev/null)
        local cluster_name=$(echo "${diag_response}" | jq -r '.statistics.cluster // "unknown"' 2>/dev/null)
        
        log "SLURM Version: ${slurm_version}"
        log "Cluster Name: ${cluster_name}"
    else
        warning "SLURM diag endpoint not responding (may not be available)"
    fi
    
    # Test jobs endpoint
    log "Testing jobs endpoint..."
    local jobs_response=$(curl -k -s -H "${auth_header}" "${base_url}/jobs?limit=5" 2>/dev/null)
    if [[ -n "${jobs_response}" ]]; then
        success "SLURM jobs endpoint responding"
        
        local job_count=$(echo "${jobs_response}" | jq -r '.jobs | length' 2>/dev/null || echo "unknown")
        log "Found ${job_count} jobs"
    else
        warning "SLURM jobs endpoint not responding"
    fi
    
    # Test nodes endpoint
    log "Testing nodes endpoint..."
    local nodes_response=$(curl -k -s -H "${auth_header}" "${base_url}/nodes" 2>/dev/null)
    if [[ -n "${nodes_response}" ]]; then
        success "SLURM nodes endpoint responding"
        
        local node_count=$(echo "${nodes_response}" | jq -r '.nodes | length' 2>/dev/null || echo "unknown")
        log "Found ${node_count} nodes"
    else
        warning "SLURM nodes endpoint not responding"
    fi
    
    # Test partitions endpoint
    log "Testing partitions endpoint..."
    local partitions_response=$(curl -k -s -H "${auth_header}" "${base_url}/partitions" 2>/dev/null)
    if [[ -n "${partitions_response}" ]]; then
        success "SLURM partitions endpoint responding"
        
        local partition_count=$(echo "${partitions_response}" | jq -r '.partitions | length' 2>/dev/null || echo "unknown")
        log "Found ${partition_count} partitions"
    else
        warning "SLURM partitions endpoint not responding"
    fi
}

test_jwt_token() {
    log "Validating JWT token..."
    
    # Decode JWT payload (without verification for basic info)
    local payload=$(echo "${SLURM_JWT}" | cut -d'.' -f2)
    
    # Add padding if needed
    while [[ $((${#payload} % 4)) -ne 0 ]]; do
        payload="${payload}="
    done
    
    local decoded=$(echo "${payload}" | base64 -d 2>/dev/null || echo "{}")
    
    if [[ "${decoded}" != "{}" ]]; then
        success "JWT token is properly formatted"
        
        local exp=$(echo "${decoded}" | jq -r '.exp // empty' 2>/dev/null)
        local iat=$(echo "${decoded}" | jq -r '.iat // empty' 2>/dev/null)
        local sub=$(echo "${decoded}" | jq -r '.sun // .sub // empty' 2>/dev/null)
        
        if [[ -n "${exp}" ]]; then
            local exp_date=$(date -d "@${exp}" 2>/dev/null || echo "unknown")
            log "Token expires: ${exp_date}"
            
            local current_time=$(date +%s)
            if [[ "${exp}" -lt "${current_time}" ]]; then
                error "JWT token has expired!"
                return 1
            else
                success "JWT token is not expired"
            fi
        fi
        
        if [[ -n "${iat}" ]]; then
            local iat_date=$(date -d "@${iat}" 2>/dev/null || echo "unknown")
            log "Token issued: ${iat_date}"
        fi
        
        if [[ -n "${sub}" ]]; then
            log "Token subject: ${sub}"
        fi
    else
        warning "Could not decode JWT token payload"
    fi
}

generate_summary_report() {
    log "Generating validation summary..."
    
    cat << EOF

=== SLURM CLUSTER VALIDATION SUMMARY ===

Cluster: ${CLUSTER_HOST}
Test Date: $(date)
JWT Token: $(echo "${SLURM_JWT}" | cut -c1-20)...

Connection Status:
- Network connectivity: $(test_network_connectivity >/dev/null 2>&1 && echo "✓ PASS" || echo "✗ FAIL")
- SLURM API access: $(test_slurm_api >/dev/null 2>&1 && echo "✓ PASS" || echo "✗ FAIL")
- JWT token valid: $(test_jwt_token >/dev/null 2>&1 && echo "✓ PASS" || echo "✗ FAIL")

Ready for integration testing: $(
    test_network_connectivity >/dev/null 2>&1 && 
    test_slurm_api >/dev/null 2>&1 && 
    test_jwt_token >/dev/null 2>&1 && 
    echo "✓ YES" || echo "✗ NO"
)

EOF
}

main() {
    log "Starting SLURM cluster validation for rocky.ar.jontk.com"
    
    validate_prerequisites
    test_jwt_token
    test_network_connectivity
    test_slurm_api
    
    generate_summary_report
    
    success "Validation completed successfully!"
    log "You can now run the full integration test suite:"
    log "  SLURM_JWT=${SLURM_JWT} ./run-rocky-tests.sh"
}

main "$@"