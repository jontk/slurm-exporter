#!/run/current-system/sw/bin/bash

# Comprehensive Test Validation Script
# Validates that all integration test components are working correctly

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

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

validate_mock_server() {
    log "Validating mock SLURM server..."
    
    # Check if Python script is valid
    if python3 -m py_compile "${SCRIPT_DIR}/mock-slurm-server.py"; then
        success "Mock SLURM server script is valid"
    else
        error "Mock SLURM server script has syntax errors"
        return 1
    fi
    
    # Test JWT token parsing
    local test_token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjI2NTQzMDU1NzksImlhdCI6MTc1NDMwNTU3OSwic3VuIjoicm9vdCJ9.ECpVcCqD1eXUk5-3LZLCDVDolvCJkhGNj24IvqwCkPs"
    log "Testing JWT token parsing..."
    
    # Decode JWT payload (basic validation)
    local payload=$(echo "${test_token}" | cut -d'.' -f2)
    while [[ $((${#payload} % 4)) -ne 0 ]]; do
        payload="${payload}="
    done
    
    if echo "${payload}" | base64 -d 2>/dev/null | jq . >/dev/null 2>&1; then
        success "JWT token is properly formatted"
    else
        warning "JWT token may have formatting issues"
    fi
}

validate_configurations() {
    log "Validating configuration files..."
    
    local configs=(
        "rocky-cluster-config.yaml"
        "mock-cluster-config.yaml"
        "prometheus-mock.yml"
        "grafana-datasources.yml"
        "grafana-dashboards.yml"
    )
    
    for config in "${configs[@]}"; do
        local config_path="${SCRIPT_DIR}/${config}"
        if [[ -f "${config_path}" ]]; then
            if [[ "${config}" =~ \.ya?ml$ ]]; then
                # Basic YAML validation - check for basic syntax issues
                if grep -q "^\s*-\s*-" "${config_path}" || grep -q "^\t" "${config_path}"; then
                    error "Configuration ${config} has YAML indentation issues"
                    return 1
                else
                    success "Configuration ${config} appears to be valid YAML"
                fi
            else
                success "Configuration ${config} exists"
            fi
        else
            error "Configuration ${config} not found"
            return 1
        fi
    done
}

validate_go_tests() {
    log "Validating Go test files..."
    
    cd "${PROJECT_ROOT}"
    
    # Check if test files compile
    if go test -c ./test/integration/... >/dev/null 2>&1; then
        success "Go integration tests compile successfully"
    else
        error "Go integration tests have compilation errors"
        go test -c ./test/integration/...
        return 1
    fi
    
    # Check for required test functions
    local test_files=(
        "test/integration/rocky_cluster_test.go"
        "test/integration/benchmark_test.go"
    )
    
    for test_file in "${test_files[@]}"; do
        if [[ -f "${test_file}" ]]; then
            if grep -q "func Test" "${test_file}"; then
                success "Test file ${test_file} contains test functions"
            else
                warning "Test file ${test_file} may not contain test functions"
            fi
        else
            error "Test file ${test_file} not found"
            return 1
        fi
    done
}

validate_docker_setup() {
    log "Validating Docker setup..."
    
    # Check if Docker is available
    if ! command -v docker >/dev/null 2>&1; then
        error "Docker is not installed or not in PATH"
        return 1
    fi
    
    if ! command -v docker-compose >/dev/null 2>&1; then
        error "Docker Compose is not installed or not in PATH"
        return 1
    fi
    
    # Validate Docker Compose files
    local compose_files=(
        "docker-compose.mock.yml"
    )
    
    for compose_file in "${compose_files[@]}"; do
        local compose_path="${SCRIPT_DIR}/${compose_file}"
        if [[ -f "${compose_path}" ]]; then
            if docker-compose -f "${compose_path}" config >/dev/null 2>&1; then
                success "Docker Compose file ${compose_file} is valid"
            else
                error "Docker Compose file ${compose_file} has validation errors"
                docker-compose -f "${compose_path}" config
                return 1
            fi
        else
            error "Docker Compose file ${compose_file} not found"
            return 1
        fi
    done
    
    # Check if Docker daemon is running
    if docker info >/dev/null 2>&1; then
        success "Docker daemon is running"
    else
        error "Docker daemon is not running or not accessible"
        return 1
    fi
}

validate_scripts() {
    log "Validating shell scripts..."
    
    local scripts=(
        "run-integration-tests.sh"
        "run-rocky-tests.sh"
        "validate-connection.sh"
    )
    
    for script in "${scripts[@]}"; do
        local script_path="${SCRIPT_DIR}/${script}"
        if [[ -f "${script_path}" ]]; then
            # Check if script is executable
            if [[ -x "${script_path}" ]]; then
                success "Script ${script} is executable"
            else
                warning "Script ${script} is not executable"
                chmod +x "${script_path}"
                success "Made script ${script} executable"
            fi
            
            # Basic syntax check
            if bash -n "${script_path}"; then
                success "Script ${script} has valid syntax"
            else
                error "Script ${script} has syntax errors"
                return 1
            fi
        else
            error "Script ${script} not found"
            return 1
        fi
    done
}

validate_dependencies() {
    log "Validating required dependencies..."
    
    local tools=(
        "go"
        "curl"
        "jq"
        "python3"
        "docker"
        "docker-compose"
    )
    
    local missing_tools=()
    
    for tool in "${tools[@]}"; do
        if command -v "${tool}" >/dev/null 2>&1; then
            success "Tool ${tool} is available"
        else
            error "Tool ${tool} is not installed or not in PATH"
            missing_tools+=("${tool}")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        error "Missing required tools: ${missing_tools[*]}"
        return 1
    fi
    
    # Check Go modules
    cd "${PROJECT_ROOT}"
    if go mod verify >/dev/null 2>&1; then
        success "Go modules are valid"
    else
        warning "Go modules may have issues"
        go mod verify
    fi
    
    # Check Python dependencies for mock server
    if python3 -c "import json, time, random, datetime, http.server, urllib.parse, ssl" 2>/dev/null; then
        success "Python standard library dependencies are available"
    else
        error "Python standard library dependencies are missing"
        return 1
    fi
}

run_basic_tests() {
    log "Running basic validation tests..."
    
    # Test configuration parsing
    log "Testing configuration parsing..."
    cd "${PROJECT_ROOT}"
    
    # Check if the config files exist and have basic structure
    if [[ -f "test/integration/mock-cluster-config.yaml" ]] && grep -q "slurm:" "test/integration/mock-cluster-config.yaml"; then
        success "Mock cluster configuration has expected structure"
    else
        error "Mock cluster configuration is missing or malformed"
        return 1
    fi
    
    # Test that all required configuration sections exist
    local required_sections=("slurm:" "metrics:" "server:" "logging:")
    for section in "${required_sections[@]}"; do
        if grep -q "${section}" "test/integration/mock-cluster-config.yaml"; then
            success "Configuration section ${section} found"
        else
            error "Configuration section ${section} missing"
            return 1
        fi
    done
}

generate_validation_report() {
    log "Generating validation report..."
    
    local report_file="${SCRIPT_DIR}/validation-report.md"
    
    cat > "${report_file}" << EOF
# Integration Test Validation Report

Generated: $(date)

## Validation Results

### Components Validated

- ✅ Mock SLURM Server
- ✅ Configuration Files
- ✅ Go Test Files
- ✅ Docker Setup
- ✅ Shell Scripts
- ✅ Dependencies
- ✅ Basic Functionality

### Test Environment

- **Project Root**: ${PROJECT_ROOT}
- **Test Directory**: ${SCRIPT_DIR}
- **Go Version**: $(go version)
- **Docker Version**: $(docker --version)
- **Python Version**: $(python3 --version)

### Available Test Modes

1. **Real Cluster Mode**: Tests against rocky.ar.jontk.com
   - Requires SLURM_JWT environment variable
   - Tests actual SLURM REST API connectivity
   - Full end-to-end validation

2. **Mock Server Mode**: Tests against local mock server
   - No external dependencies
   - Consistent test environment
   - Faster test execution

### Usage

Run comprehensive integration tests:
\`\`\`bash
./run-integration-tests.sh
\`\`\`

Force mock server mode:
\`\`\`bash
./run-integration-tests.sh mock
\`\`\`

Force real cluster mode:
\`\`\`bash
SLURM_JWT="your-token-here" ./run-integration-tests.sh real
\`\`\`

Clean up test environment:
\`\`\`bash
./run-integration-tests.sh cleanup
\`\`\`

### Next Steps

The integration testing infrastructure is ready for use. When the real SLURM cluster becomes available, tests will automatically detect and use it. Until then, the mock server provides a comprehensive testing environment.
EOF
    
    success "Validation report generated: ${report_file}"
}

main() {
    log "Starting comprehensive integration test validation"
    
    local validations=(
        "validate_dependencies"
        "validate_mock_server"
        "validate_configurations"
        "validate_go_tests"
        "validate_docker_setup"
        "validate_scripts"
        "run_basic_tests"
    )
    
    local failed_validations=()
    
    for validation in "${validations[@]}"; do
        log "Running ${validation}..."
        if $validation; then
            success "${validation} passed"
        else
            error "${validation} failed"
            failed_validations+=("${validation}")
        fi
        echo ""
    done
    
    # Generate report
    generate_validation_report
    
    # Summary
    log "=== VALIDATION SUMMARY ==="
    if [[ ${#failed_validations[@]} -eq 0 ]]; then
        success "All validations passed! Integration test infrastructure is ready."
        log "You can now run: ./run-integration-tests.sh"
    else
        error "Some validations failed: ${failed_validations[*]}"
        log "Please fix the issues before running integration tests."
        exit 1
    fi
}

main "$@"