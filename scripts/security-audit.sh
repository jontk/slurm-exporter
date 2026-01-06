#!/bin/bash

# SLURM Exporter Security Audit Script
# Comprehensive security assessment and hardening recommendations

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
AUDIT_DATE=$(date -u +"%Y%m%d_%H%M%S")
AUDIT_DIR="$PROJECT_ROOT/security-audit-$AUDIT_DATE"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Emojis
SHIELD='🛡️'
CHECK='✅'
WARNING='⚠️'
ERROR='❌'
INFO='ℹ️'
SEARCH='🔍'
LOCK='🔒'

# Helper functions
log_info() {
    echo -e "${BLUE}${INFO}${NC} $1"
}

log_success() {
    echo -e "${GREEN}${CHECK}${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}${WARNING}${NC} $1"
}

log_error() {
    echo -e "${RED}${ERROR}${NC} $1"
}

log_audit() {
    echo -e "${PURPLE}${SEARCH}${NC} $1"
}

print_header() {
    echo
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}            ${SHIELD} SLURM Exporter Security Audit ${SHIELD}              ${CYAN}║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo
}

print_section() {
    echo
    echo -e "${PURPLE}═══ $1 ═══${NC}"
    echo
}

create_audit_dir() {
    mkdir -p "$AUDIT_DIR"
    log_info "Audit results will be saved to: $AUDIT_DIR"
}

# Security audit functions
audit_file_permissions() {
    print_section "File Permissions Audit"
    
    local report="$AUDIT_DIR/file_permissions.txt"
    echo "File Permissions Security Audit - $(date)" > "$report"
    echo "=========================================" >> "$report"
    echo >> "$report"
    
    # Check for world-writable files
    log_audit "Checking for world-writable files..."
    if find "$PROJECT_ROOT" -type f -perm -002 -not -path "*/\.git/*" > "$AUDIT_DIR/world_writable.txt" 2>/dev/null; then
        local count=$(wc -l < "$AUDIT_DIR/world_writable.txt")
        if [ "$count" -gt 0 ]; then
            log_warning "Found $count world-writable files"
            echo "World-writable files found ($count):" >> "$report"
            cat "$AUDIT_DIR/world_writable.txt" >> "$report"
        else
            log_success "No world-writable files found"
            echo "✅ No world-writable files found" >> "$report"
        fi
    fi
    echo >> "$report"
    
    # Check for files with excessive permissions
    log_audit "Checking executable files..."
    find "$PROJECT_ROOT" -type f -executable -not -path "*/\.git/*" -not -name "*.sh" -not -name "*.py" > "$AUDIT_DIR/executables.txt" 2>/dev/null || true
    local exec_count=$(wc -l < "$AUDIT_DIR/executables.txt" 2>/dev/null || echo 0)
    echo "Executable files found ($exec_count):" >> "$report"
    if [ "$exec_count" -gt 0 ]; then
        cat "$AUDIT_DIR/executables.txt" >> "$report"
        log_info "Found $exec_count executable files (review recommended)"
    else
        echo "None" >> "$report"
        log_success "No unexpected executable files"
    fi
    echo >> "$report"
    
    # Check sensitive file permissions
    log_audit "Checking sensitive files permissions..."
    local sensitive_files=("*.key" "*.pem" "*.crt" "*password*" "*secret*" "*.env")
    local found_sensitive=false
    
    for pattern in "${sensitive_files[@]}"; do
        if find "$PROJECT_ROOT" -name "$pattern" -not -path "*/\.git/*" | head -5 > "$AUDIT_DIR/sensitive_temp.txt" 2>/dev/null; then
            if [ -s "$AUDIT_DIR/sensitive_temp.txt" ]; then
                if [ "$found_sensitive" = false ]; then
                    echo "Sensitive files permissions:" >> "$report"
                    found_sensitive=true
                fi
                while IFS= read -r file; do
                    if [ -f "$file" ]; then
                        local perms=$(stat -c "%a %n" "$file" 2>/dev/null || echo "unknown $file")
                        echo "  $perms" >> "$report"
                        local perm_num=$(echo "$perms" | cut -d' ' -f1)
                        if [[ "$perm_num" =~ [4567][4567][4567] ]]; then
                            log_warning "Sensitive file has broad permissions: $file ($perm_num)"
                        fi
                    fi
                done < "$AUDIT_DIR/sensitive_temp.txt"
            fi
        fi
    done
    
    if [ "$found_sensitive" = false ]; then
        echo "✅ No sensitive files with concerning permissions found" >> "$report"
        log_success "No sensitive files with concerning permissions"
    fi
    
    rm -f "$AUDIT_DIR/sensitive_temp.txt"
}

audit_secrets_and_credentials() {
    print_section "Secrets and Credentials Audit"
    
    local report="$AUDIT_DIR/secrets_audit.txt"
    echo "Secrets and Credentials Audit - $(date)" > "$report"
    echo "=======================================" >> "$report"
    echo >> "$report"
    
    # Common secret patterns
    local secret_patterns=(
        "password\s*=\s*['\"][^'\"]{8,}['\"]"
        "api[_-]?key\s*=\s*['\"][^'\"]{20,}['\"]"
        "secret[_-]?key\s*=\s*['\"][^'\"]{20,}['\"]"
        "token\s*=\s*['\"][^'\"]{20,}['\"]"
        "-----BEGIN [A-Z ]+-----"
        "AKIA[0-9A-Z]{16}"
        "sk_live_[0-9a-zA-Z]{24}"
        "pk_live_[0-9a-zA-Z]{24}"
    )
    
    log_audit "Scanning for potential secrets in code..."
    local secrets_found=false
    
    for pattern in "${secret_patterns[@]}"; do
        log_audit "Checking pattern: ${pattern:0:20}..."
        if grep -r -E -n "$pattern" "$PROJECT_ROOT" \
            --exclude-dir=.git \
            --exclude-dir=node_modules \
            --exclude-dir=vendor \
            --exclude="*.log" \
            --exclude="security-audit.sh" \
            > "$AUDIT_DIR/secrets_temp.txt" 2>/dev/null; then
            
            if [ -s "$AUDIT_DIR/secrets_temp.txt" ]; then
                if [ "$secrets_found" = false ]; then
                    echo "⚠️ Potential secrets found:" >> "$report"
                    secrets_found=true
                fi
                echo "Pattern: $pattern" >> "$report"
                head -10 "$AUDIT_DIR/secrets_temp.txt" >> "$report"
                echo >> "$report"
                log_warning "Potential secrets found for pattern: ${pattern:0:30}..."
            fi
        fi
    done
    
    if [ "$secrets_found" = false ]; then
        echo "✅ No obvious secrets found in codebase" >> "$report"
        log_success "No obvious secrets detected in code"
    fi
    
    # Check for common credential files
    log_audit "Checking for credential files..."
    local cred_files=(".env" ".env.local" ".env.production" "credentials.yml" "secrets.yml" "config/secrets.yml")
    local creds_found=false
    
    for cred_file in "${cred_files[@]}"; do
        if find "$PROJECT_ROOT" -name "$cred_file" -not -path "*/\.git/*" | head -5 > "$AUDIT_DIR/creds_temp.txt" 2>/dev/null; then
            if [ -s "$AUDIT_DIR/creds_temp.txt" ]; then
                if [ "$creds_found" = false ]; then
                    echo "Credential files found:" >> "$report"
                    creds_found=true
                fi
                cat "$AUDIT_DIR/creds_temp.txt" >> "$report"
                log_warning "Found credential file: $cred_file"
            fi
        fi
    done
    
    if [ "$creds_found" = false ]; then
        echo "✅ No credential files found" >> "$report"
        log_success "No credential files detected"
    fi
    
    rm -f "$AUDIT_DIR/secrets_temp.txt" "$AUDIT_DIR/creds_temp.txt"
}

audit_dependencies() {
    print_section "Dependency Security Audit"
    
    local report="$AUDIT_DIR/dependencies_audit.txt"
    echo "Dependency Security Audit - $(date)" > "$report"
    echo "==================================" >> "$report"
    echo >> "$report"
    
    # Go dependencies audit
    if [ -f "$PROJECT_ROOT/go.mod" ]; then
        log_audit "Auditing Go dependencies..."
        echo "Go Dependencies:" >> "$report"
        echo "===============" >> "$report"
        
        cd "$PROJECT_ROOT"
        
        # List all dependencies
        go list -m all > "$AUDIT_DIR/go_deps.txt" 2>/dev/null || true
        local dep_count=$(wc -l < "$AUDIT_DIR/go_deps.txt" 2>/dev/null || echo 0)
        echo "Total dependencies: $dep_count" >> "$report"
        echo >> "$report"
        
        # Check for known vulnerable dependencies with nancy (if available)
        if command -v nancy >/dev/null 2>&1; then
            log_audit "Running Nancy vulnerability scan..."
            if go list -json -deps ./... | nancy sleuth --output=json > "$AUDIT_DIR/nancy_results.json" 2>/dev/null; then
                if [ -s "$AUDIT_DIR/nancy_results.json" ] && [ "$(cat "$AUDIT_DIR/nancy_results.json")" != "null" ]; then
                    echo "⚠️ Vulnerabilities found by Nancy - see nancy_results.json" >> "$report"
                    log_warning "Vulnerabilities detected by Nancy scanner"
                else
                    echo "✅ No vulnerabilities found by Nancy" >> "$report"
                    log_success "Nancy scan completed - no vulnerabilities"
                fi
            else
                echo "⚠️ Nancy scan failed or incomplete" >> "$report"
                log_warning "Nancy scan encountered issues"
            fi
        else
            echo "ℹ️ Nancy not available for vulnerability scanning" >> "$report"
            log_info "Install Nancy for dependency vulnerability scanning: go install github.com/sonatypecommunity/nancy@latest"
        fi
        
        # Check for outdated dependencies
        log_audit "Checking for outdated dependencies..."
        if go list -u -m all > "$AUDIT_DIR/outdated_deps.txt" 2>/dev/null; then
            local outdated=$(grep -c "\[" "$AUDIT_DIR/outdated_deps.txt" 2>/dev/null || echo 0)
            if [ "$outdated" -gt 0 ]; then
                echo "Outdated dependencies: $outdated" >> "$report"
                echo "See outdated_deps.txt for details" >> "$report"
                log_warning "$outdated outdated dependencies found"
            else
                echo "✅ All dependencies are up to date" >> "$report"
                log_success "All dependencies are current"
            fi
        fi
        
        cd - > /dev/null
    fi
    
    # Check for license compliance
    log_audit "Checking dependency licenses..."
    if command -v go-licenses >/dev/null 2>&1; then
        cd "$PROJECT_ROOT"
        if go-licenses report ./... > "$AUDIT_DIR/licenses.txt" 2>/dev/null; then
            # Check for problematic licenses
            local problematic_licenses=("GPL-2.0" "GPL-3.0" "AGPL-1.0" "AGPL-3.0")
            local license_issues=false
            
            for license in "${problematic_licenses[@]}"; do
                if grep -q "$license" "$AUDIT_DIR/licenses.txt"; then
                    if [ "$license_issues" = false ]; then
                        echo "⚠️ Problematic licenses found:" >> "$report"
                        license_issues=true
                    fi
                    echo "  - $license" >> "$report"
                    log_warning "Found problematic license: $license"
                fi
            done
            
            if [ "$license_issues" = false ]; then
                echo "✅ No problematic licenses detected" >> "$report"
                log_success "License compliance check passed"
            fi
            
            local total_licenses=$(wc -l < "$AUDIT_DIR/licenses.txt")
            echo "Total unique licenses: $total_licenses" >> "$report"
        else
            echo "⚠️ License check failed" >> "$report"
            log_warning "Could not check dependency licenses"
        fi
        cd - > /dev/null
    else
        echo "ℹ️ go-licenses not available for license checking" >> "$report"
        log_info "Install go-licenses for license compliance: go install github.com/google/go-licenses@latest"
    fi
}

audit_configuration_security() {
    print_section "Configuration Security Audit"
    
    local report="$AUDIT_DIR/config_security.txt"
    echo "Configuration Security Audit - $(date)" > "$report"
    echo "=====================================" >> "$report"
    echo >> "$report"
    
    # Check configuration files
    local config_files=("*.yml" "*.yaml" "*.json" "*.toml" "*.conf" "config/*")
    local config_found=false
    
    log_audit "Scanning configuration files for security issues..."
    
    for pattern in "${config_files[@]}"; do
        find "$PROJECT_ROOT" -name "$pattern" -not -path "*/\.git/*" -type f > "$AUDIT_DIR/config_temp.txt" 2>/dev/null || true
        
        if [ -s "$AUDIT_DIR/config_temp.txt" ]; then
            while IFS= read -r config_file; do
                if [ "$config_found" = false ]; then
                    echo "Configuration files analyzed:" >> "$report"
                    config_found=true
                fi
                
                echo "File: $config_file" >> "$report"
                
                # Check for hardcoded credentials
                local cred_patterns=("password:" "secret:" "key:" "token:" "credential:")
                local creds_in_config=false
                
                for cred_pattern in "${cred_patterns[@]}"; do
                    if grep -i "$cred_pattern" "$config_file" >/dev/null 2>&1; then
                        if [ "$creds_in_config" = false ]; then
                            echo "  ⚠️ Potential credentials found" >> "$report"
                            creds_in_config=true
                        fi
                        log_warning "Potential credentials in config: $(basename "$config_file")"
                    fi
                done
                
                if [ "$creds_in_config" = false ]; then
                    echo "  ✅ No obvious credentials" >> "$report"
                fi
                
                # Check for insecure defaults
                local insecure_patterns=("debug.*true" "ssl.*false" "tls.*false" "secure.*false")
                local insecure_found=false
                
                for insecure_pattern in "${insecure_patterns[@]}"; do
                    if grep -i -E "$insecure_pattern" "$config_file" >/dev/null 2>&1; then
                        if [ "$insecure_found" = false ]; then
                            echo "  ⚠️ Potentially insecure settings" >> "$report"
                            insecure_found=true
                        fi
                        log_warning "Potentially insecure setting in: $(basename "$config_file")"
                    fi
                done
                
                if [ "$insecure_found" = false ]; then
                    echo "  ✅ No obviously insecure settings" >> "$report"
                fi
                
                echo >> "$report"
            done < "$AUDIT_DIR/config_temp.txt"
        fi
    done
    
    if [ "$config_found" = false ]; then
        echo "ℹ️ No configuration files found for analysis" >> "$report"
        log_info "No configuration files detected"
    fi
    
    rm -f "$AUDIT_DIR/config_temp.txt"
}

audit_dockerfile_security() {
    print_section "Dockerfile Security Audit"
    
    local report="$AUDIT_DIR/dockerfile_security.txt"
    echo "Dockerfile Security Audit - $(date)" > "$report"
    echo "==================================" >> "$report"
    echo >> "$report"
    
    local dockerfile_found=false
    
    find "$PROJECT_ROOT" -name "Dockerfile*" -o -name "*.dockerfile" > "$AUDIT_DIR/dockerfiles.txt" 2>/dev/null || true
    
    if [ -s "$AUDIT_DIR/dockerfiles.txt" ]; then
        log_audit "Analyzing Dockerfile security..."
        
        while IFS= read -r dockerfile; do
            dockerfile_found=true
            echo "Analyzing: $dockerfile" >> "$report"
            echo "========================" >> "$report"
            
            # Check for root user
            if ! grep -q "USER.*[^0]" "$dockerfile"; then
                echo "⚠️ Running as root user (no USER directive)" >> "$report"
                log_warning "Dockerfile runs as root: $(basename "$dockerfile")"
            else
                echo "✅ Non-root user configured" >> "$report"
                log_success "Non-root user in: $(basename "$dockerfile")"
            fi
            
            # Check for secrets in build args
            if grep -E "ARG.*(PASSWORD|SECRET|KEY|TOKEN)" "$dockerfile" >/dev/null 2>&1; then
                echo "⚠️ Potential secrets in build arguments" >> "$report"
                log_warning "Potential secrets in build args: $(basename "$dockerfile")"
            else
                echo "✅ No obvious secrets in build arguments" >> "$report"
            fi
            
            # Check for package manager cache
            if grep -E "(apt-get|apk|yum).*(update|upgrade)" "$dockerfile" >/dev/null 2>&1; then
                if ! grep -E "(rm.*cache|clean)" "$dockerfile" >/dev/null 2>&1; then
                    echo "⚠️ Package manager cache not cleaned" >> "$report"
                    log_warning "Package cache not cleaned: $(basename "$dockerfile")"
                else
                    echo "✅ Package manager cache cleaned" >> "$report"
                fi
            fi
            
            # Check for specific versions
            if grep -E "FROM.*:latest" "$dockerfile" >/dev/null 2>&1; then
                echo "⚠️ Using 'latest' tag (not recommended)" >> "$report"
                log_warning "Using 'latest' tag: $(basename "$dockerfile")"
            else
                echo "✅ Specific image versions used" >> "$report"
            fi
            
            # Check for privileged operations
            if grep -E "(--privileged|SYS_ADMIN|--cap-add)" "$dockerfile" >/dev/null 2>&1; then
                echo "⚠️ Privileged operations detected" >> "$report"
                log_warning "Privileged operations: $(basename "$dockerfile")"
            else
                echo "✅ No privileged operations detected" >> "$report"
            fi
            
            echo >> "$report"
        done < "$AUDIT_DIR/dockerfiles.txt"
    fi
    
    if [ "$dockerfile_found" = false ]; then
        echo "ℹ️ No Dockerfiles found for analysis" >> "$report"
        log_info "No Dockerfiles detected"
    fi
    
    rm -f "$AUDIT_DIR/dockerfiles.txt"
}

audit_github_security() {
    print_section "GitHub Security Configuration Audit"
    
    local report="$AUDIT_DIR/github_security.txt"
    echo "GitHub Security Configuration Audit - $(date)" > "$report"
    echo "=============================================" >> "$report"
    echo >> "$report"
    
    # Check for security-related files
    local security_files=(
        "SECURITY.md:Security policy documentation"
        ".github/workflows/security.yml:Automated security scanning"
        ".github/dependabot.yml:Automated dependency updates"
        ".github/workflows/codeql.yml:Code scanning"
        "CODE_OF_CONDUCT.md:Community guidelines"
        "CONTRIBUTING.md:Contribution guidelines"
    )
    
    log_audit "Checking for required security files..."
    echo "Required security files:" >> "$report"
    echo "=======================" >> "$report"
    
    for file_info in "${security_files[@]}"; do
        local file_path="${file_info%%:*}"
        local description="${file_info##*:}"
        
        if [ -f "$PROJECT_ROOT/$file_path" ]; then
            echo "✅ $file_path - $description" >> "$report"
            log_success "Found: $file_path"
        else
            echo "❌ $file_path - $description (MISSING)" >> "$report"
            log_error "Missing: $file_path"
        fi
    done
    
    echo >> "$report"
    
    # Check workflow security
    if [ -d "$PROJECT_ROOT/.github/workflows" ]; then
        log_audit "Analyzing GitHub Actions workflows..."
        echo "GitHub Actions Security:" >> "$report"
        echo "=======================" >> "$report"
        
        find "$PROJECT_ROOT/.github/workflows" -name "*.yml" -o -name "*.yaml" > "$AUDIT_DIR/workflows.txt" 2>/dev/null || true
        
        if [ -s "$AUDIT_DIR/workflows.txt" ]; then
            while IFS= read -r workflow; do
                echo "Workflow: $(basename "$workflow")" >> "$report"
                
                # Check for security best practices
                if grep -q "permissions:" "$workflow"; then
                    echo "  ✅ Permissions defined" >> "$report"
                else
                    echo "  ⚠️ No explicit permissions (uses defaults)" >> "$report"
                    log_warning "No explicit permissions in: $(basename "$workflow")"
                fi
                
                # Check for secrets handling  
                if grep -E "\$\{\{\s*secrets\." "$workflow" >/dev/null 2>&1; then
                    echo "  ℹ️ Uses GitHub secrets" >> "$report"
                fi
                
                # Check for third-party actions
                local third_party=$(grep -E "uses:.*@" "$workflow" | grep -v "actions/" | wc -l)
                if [ "$third_party" -gt 0 ]; then
                    echo "  ⚠️ Uses $third_party third-party actions" >> "$report"
                    log_info "$third_party third-party actions in: $(basename "$workflow")"
                fi
                
                echo >> "$report"
            done < "$AUDIT_DIR/workflows.txt"
        fi
        
        rm -f "$AUDIT_DIR/workflows.txt"
    else
        echo "ℹ️ No GitHub Actions workflows found" >> "$report"
    fi
}

generate_security_report() {
    print_section "Generating Security Report"
    
    local main_report="$AUDIT_DIR/SECURITY_AUDIT_REPORT.md"
    
    log_info "Generating comprehensive security report..."
    
    cat > "$main_report" << 'EOF'
# SLURM Exporter Security Audit Report

## Executive Summary

This report provides a comprehensive security assessment of the SLURM Exporter project, including code analysis, dependency scanning, configuration review, and security best practices evaluation.

### Audit Information
- **Date**: 
- **Scope**: Complete codebase and configuration analysis
- **Tools Used**: Custom security audit script, static analysis
- **Methodology**: OWASP security guidelines and best practices

## Findings Summary

### Critical Issues 🔴
- [ ] Issues requiring immediate attention

### High Priority Issues 🟠  
- [ ] Issues requiring attention within 7 days

### Medium Priority Issues 🟡
- [ ] Issues requiring attention within 30 days

### Low Priority Issues 🟢
- [ ] Recommendations for improvement

## Detailed Findings

### 1. File Permissions Security
EOF
    
    # Add audit date
    sed -i "s/- \*\*Date\*\*:/- **Date**: $(date -u '+%Y-%m-%d %H:%M:%S UTC')/" "$main_report"
    
    # Append individual audit results
    if [ -f "$AUDIT_DIR/file_permissions.txt" ]; then
        echo >> "$main_report"
        echo '```' >> "$main_report"
        cat "$AUDIT_DIR/file_permissions.txt" >> "$main_report"
        echo '```' >> "$main_report"
    fi
    
    echo >> "$main_report"
    echo "### 2. Secrets and Credentials" >> "$main_report"
    if [ -f "$AUDIT_DIR/secrets_audit.txt" ]; then
        echo >> "$main_report"
        echo '```' >> "$main_report"
        cat "$AUDIT_DIR/secrets_audit.txt" >> "$main_report"
        echo '```' >> "$main_report"
    fi
    
    echo >> "$main_report"
    echo "### 3. Dependencies Security" >> "$main_report"
    if [ -f "$AUDIT_DIR/dependencies_audit.txt" ]; then
        echo >> "$main_report"
        echo '```' >> "$main_report"
        cat "$AUDIT_DIR/dependencies_audit.txt" >> "$main_report"
        echo '```' >> "$main_report"
    fi
    
    echo >> "$main_report"
    echo "### 4. Configuration Security" >> "$main_report"
    if [ -f "$AUDIT_DIR/config_security.txt" ]; then
        echo >> "$main_report"
        echo '```' >> "$main_report"
        cat "$AUDIT_DIR/config_security.txt" >> "$main_report"
        echo '```' >> "$main_report"
    fi
    
    echo >> "$main_report"
    echo "### 5. Container Security" >> "$main_report"
    if [ -f "$AUDIT_DIR/dockerfile_security.txt" ]; then
        echo >> "$main_report"
        echo '```' >> "$main_report"
        cat "$AUDIT_DIR/dockerfile_security.txt" >> "$main_report"
        echo '```' >> "$main_report"
    fi
    
    echo >> "$main_report"
    echo "### 6. GitHub Security Configuration" >> "$main_report"
    if [ -f "$AUDIT_DIR/github_security.txt" ]; then
        echo >> "$main_report"
        echo '```' >> "$main_report"
        cat "$AUDIT_DIR/github_security.txt" >> "$main_report"
        echo '```' >> "$main_report"
    fi
    
    # Add recommendations section
    cat >> "$main_report" << 'EOF'

## Recommendations

### Immediate Actions Required
1. **Review and rotate any exposed credentials** found in the audit
2. **Fix file permission issues** for world-writable files
3. **Update vulnerable dependencies** identified by security scanners

### Short-term Improvements (1-4 weeks)
1. **Implement automated security scanning** in CI/CD pipeline
2. **Add security-focused unit tests** for authentication and authorization
3. **Create incident response procedures** for security events
4. **Set up monitoring and alerting** for security metrics

### Long-term Security Enhancements (1-3 months)
1. **Conduct regular penetration testing** of deployed instances
2. **Implement comprehensive audit logging** for all security events
3. **Create security training materials** for contributors
4. **Establish security code review processes**

## Security Tools Integration

### Recommended Tools
- **Static Analysis**: gosec, StaticCheck
- **Dependency Scanning**: nancy, Snyk
- **Secret Detection**: TruffleHog, GitLeaks
- **Container Scanning**: Trivy, Grype
- **License Compliance**: go-licenses

### CI/CD Integration
```yaml
# Add to .github/workflows/security.yml
- name: Security Scan
  uses: securecodewarrior/github-action-gosec@master
  with:
    args: './...'
```

## Compliance Considerations

### Industry Standards
- **OWASP Top 10**: Address web application security risks
- **NIST Cybersecurity Framework**: Implement comprehensive security controls
- **SOC 2 Type II**: Consider compliance for enterprise deployments

### Regulatory Requirements
- Review data handling practices for GDPR compliance
- Implement appropriate logging for audit requirements
- Ensure encryption standards meet regulatory requirements

## Next Steps

1. **Prioritize critical and high-priority findings** for immediate remediation
2. **Create tracking issues** for each security finding
3. **Schedule regular security audits** (quarterly recommended)
4. **Update security documentation** based on findings
5. **Train development team** on identified security practices

---

**Report Generated**: Automated Security Audit Script
**Next Audit Recommended**: 3 months from report date
**Questions or Concerns**: Contact the security team

EOF
    
    log_success "Security audit report generated: $main_report"
}

print_summary() {
    print_section "Security Audit Summary"
    
    log_info "Security audit completed successfully!"
    echo
    echo "📁 Audit results location: $AUDIT_DIR"
    echo "📋 Main report: $AUDIT_DIR/SECURITY_AUDIT_REPORT.md"
    echo
    echo "📊 Generated files:"
    find "$AUDIT_DIR" -type f -name "*.txt" -o -name "*.md" -o -name "*.json" | while read -r file; do
        echo "  • $(basename "$file")"
    done
    echo
    echo "🔄 Next steps:"
    echo "  1. Review the main security report"
    echo "  2. Address critical and high-priority findings"
    echo "  3. Create tracking issues for security improvements"
    echo "  4. Schedule regular security audits"
    echo
    echo "💡 For questions about this audit, consult the SECURITY.md file"
}

# Main execution
main() {
    print_header
    
    log_info "Starting comprehensive security audit..."
    log_info "Project: $(basename "$PROJECT_ROOT")"
    log_info "Date: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
    
    create_audit_dir
    
    # Run all audit functions
    audit_file_permissions
    audit_secrets_and_credentials
    audit_dependencies
    audit_configuration_security
    audit_dockerfile_security
    audit_github_security
    
    # Generate comprehensive report
    generate_security_report
    
    # Print summary
    print_summary
    
    log_success "Security audit completed! 🛡️"
}

# Only run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi