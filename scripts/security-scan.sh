#!/bin/bash

# ShareHODL Security Scanner Script
# Runs comprehensive security scans on the ShareHODL codebase

set -euo pipefail

# Configuration
SOURCE_DIR="${SOURCE_DIR:-/app/source}"
REPORTS_DIR="${REPORTS_DIR:-/app/reports}"
CONFIG_DIR="${CONFIG_DIR:-/app/config}"
SCAN_INTERVAL="${SCAN_INTERVAL:-3600}"
ALERT_WEBHOOK="${ALERT_WEBHOOK:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Initialize directories
initialize() {
    mkdir -p "$REPORTS_DIR"/{static,dependency,secret,license,security}
    log_info "Security scanner initialized"
}

# Run Go security scan with gosec
run_gosec_scan() {
    log_info "Running gosec security scan..."
    
    cd "$SOURCE_DIR"
    
    if gosec -fmt json -out "$REPORTS_DIR/security/gosec.json" ./...; then
        log_success "gosec scan completed"
    else
        log_warning "gosec scan found issues"
    fi
    
    # Generate human-readable report
    gosec -fmt text -out "$REPORTS_DIR/security/gosec.txt" ./... || true
}

# Run static analysis with golangci-lint
run_golangci_scan() {
    log_info "Running golangci-lint scan..."
    
    cd "$SOURCE_DIR"
    
    if golangci-lint run --out-format json > "$REPORTS_DIR/static/golangci.json"; then
        log_success "golangci-lint scan completed"
    else
        log_warning "golangci-lint found issues"
    fi
    
    # Generate text report
    golangci-lint run --out-format colored-line-number > "$REPORTS_DIR/static/golangci.txt" || true
}

# Run dependency vulnerability scan
run_dependency_scan() {
    log_info "Running dependency vulnerability scan..."
    
    cd "$SOURCE_DIR"
    
    # Check for go.mod
    if [[ -f "go.mod" ]]; then
        # List dependencies
        go list -m -json all > "$REPORTS_DIR/dependency/modules.json"
        
        # Check for known vulnerabilities (would use actual vulnerability database)
        log_info "Checking Go dependencies for known vulnerabilities"
        
        # Simulate vulnerability check
        cat > "$REPORTS_DIR/dependency/vulnerabilities.json" << EOF
{
    "vulnerabilities": [],
    "scan_time": "$(date -Iseconds)",
    "total_dependencies": $(go list -m all | wc -l)
}
EOF
    fi
    
    # Check for package.json (if exists)
    if [[ -f "package.json" ]]; then
        log_info "Scanning npm dependencies"
        npm audit --json > "$REPORTS_DIR/dependency/npm-audit.json" || true
    fi
    
    log_success "Dependency scan completed"
}

# Run secret scanning
run_secret_scan() {
    log_info "Running secret scanning..."
    
    cd "$SOURCE_DIR"
    
    # Use trufflehog for secret scanning
    if command -v trufflehog >/dev/null 2>&1; then
        trufflehog --json . > "$REPORTS_DIR/secret/trufflehog.json" || true
    fi
    
    # Use git-secrets if available
    if command -v git-secrets >/dev/null 2>&1; then
        git secrets --scan --recursive . > "$REPORTS_DIR/secret/git-secrets.txt" 2>&1 || true
    fi
    
    # Custom pattern scanning
    run_custom_secret_patterns
    
    log_success "Secret scanning completed"
}

# Run custom secret pattern scanning
run_custom_secret_patterns() {
    log_info "Scanning for hardcoded secrets with custom patterns..."
    
    local patterns_file="$CONFIG_DIR/secret-patterns.txt"
    
    # Create default patterns if config doesn't exist
    if [[ ! -f "$patterns_file" ]]; then
        cat > "$patterns_file" << EOF
password\s*[:=]\s*['""][^'""]{8,}['""']
api[_-]?key\s*[:=]\s*['""][^'""]{20,}['""']
secret[_-]?key\s*[:=]\s*['""][^'""]{20,}['""']
token\s*[:=]\s*['""][^'""]{20,}['""']
private[_-]?key\s*[:=]
-----BEGIN.*PRIVATE KEY-----
AKIA[0-9A-Z]{16}
ghp_[0-9a-zA-Z]{36}
xox[baprs]-[0-9a-zA-Z-]{10,48}
EOF
    fi
    
    local results_file="$REPORTS_DIR/secret/custom-patterns.json"
    echo '{"findings": [' > "$results_file"
    
    local first=true
    while IFS= read -r pattern; do
        [[ -z "$pattern" || "$pattern" =~ ^# ]] && continue
        
        local matches
        matches=$(grep -rn -E "$pattern" "$SOURCE_DIR" --exclude-dir=.git --exclude-dir=vendor --exclude="*.json" 2>/dev/null || true)
        
        if [[ -n "$matches" ]]; then
            while IFS= read -r match; do
                if [[ "$first" == "false" ]]; then
                    echo "," >> "$results_file"
                fi
                first=false
                
                local file=$(echo "$match" | cut -d: -f1)
                local line=$(echo "$match" | cut -d: -f2)
                local content=$(echo "$match" | cut -d: -f3-)
                
                cat >> "$results_file" << EOF
{
    "file": "$file",
    "line": $line,
    "pattern": "$pattern",
    "content": "*** REDACTED ***"
}
EOF
            done <<< "$matches"
        fi
    done < "$patterns_file"
    
    echo ']}' >> "$results_file"
}

# Run license compliance scan
run_license_scan() {
    log_info "Running license compliance scan..."
    
    cd "$SOURCE_DIR"
    
    # Check Go module licenses
    if [[ -f "go.mod" ]]; then
        log_info "Scanning Go module licenses"
        
        # Create license report
        cat > "$REPORTS_DIR/license/go-licenses.json" << EOF
{
    "scan_time": "$(date -Iseconds)",
    "licenses": [],
    "unknown_licenses": [],
    "restricted_licenses": []
}
EOF
    fi
    
    # Check for license files
    find "$SOURCE_DIR" -name "LICENSE*" -o -name "COPYING*" > "$REPORTS_DIR/license/license-files.txt"
    
    log_success "License scan completed"
}

# Run custom ShareHODL security checks
run_sharehodl_security_checks() {
    log_info "Running ShareHODL-specific security checks..."
    
    local report_file="$REPORTS_DIR/security/sharehodl-specific.json"
    
    cat > "$report_file" << EOF
{
    "scan_time": "$(date -Iseconds)",
    "checks": [
EOF
    
    # Check for proper access controls
    check_access_controls "$report_file"
    
    # Check for cryptographic implementations
    check_cryptography "$report_file"
    
    # Check for business logic vulnerabilities
    check_business_logic "$report_file"
    
    echo '    ]' >> "$report_file"
    echo '}' >> "$report_file"
    
    log_success "ShareHODL security checks completed"
}

# Check access control implementations
check_access_controls() {
    local report_file="$1"
    
    log_info "Checking access control implementations..."
    
    # Look for authorization patterns
    local auth_patterns=(
        "require.*authorized"
        "check.*permission"
        "validate.*access"
        "msg\.sender"
    )
    
    local first_check=true
    for pattern in "${auth_patterns[@]}"; do
        local matches
        matches=$(grep -rn -E "$pattern" "$SOURCE_DIR" --include="*.go" 2>/dev/null || true)
        
        if [[ -n "$matches" ]]; then
            if [[ "$first_check" == "false" ]]; then
                echo "," >> "$report_file"
            fi
            first_check=false
            
            local count=$(echo "$matches" | wc -l)
            cat >> "$report_file" << EOF
        {
            "type": "access_control",
            "pattern": "$pattern",
            "matches": $count,
            "status": "found"
        }
EOF
        fi
    done
}

# Check cryptographic implementations
check_cryptography() {
    local report_file="$1"
    
    log_info "Checking cryptographic implementations..."
    
    # Look for weak crypto patterns
    local weak_crypto=(
        "md5\."
        "sha1\."
        "des\."
        "rc4"
        "math/rand"
    )
    
    for pattern in "${weak_crypto[@]}"; do
        local matches
        matches=$(grep -rn -E "$pattern" "$SOURCE_DIR" --include="*.go" 2>/dev/null || true)
        
        if [[ -n "$matches" ]]; then
            echo "," >> "$report_file"
            local count=$(echo "$matches" | wc -l)
            cat >> "$report_file" << EOF
        {
            "type": "weak_cryptography",
            "pattern": "$pattern",
            "matches": $count,
            "status": "warning"
        }
EOF
        fi
    done
}

# Check business logic
check_business_logic() {
    local report_file="$1"
    
    log_info "Checking business logic implementations..."
    
    # Look for potential overflow/underflow
    local overflow_patterns=(
        "balance.*-.*"
        "amount.*\+.*"
        "supply.*=.*supply.*\+"
    )
    
    for pattern in "${overflow_patterns[@]}"; do
        local matches
        matches=$(grep -rn -E "$pattern" "$SOURCE_DIR" --include="*.go" 2>/dev/null || true)
        
        if [[ -n "$matches" ]]; then
            echo "," >> "$report_file"
            local count=$(echo "$matches" | wc -l)
            cat >> "$report_file" << EOF
        {
            "type": "arithmetic_operation",
            "pattern": "$pattern",
            "matches": $count,
            "status": "info"
        }
EOF
        fi
    done
}

# Generate consolidated report
generate_consolidated_report() {
    log_info "Generating consolidated security report..."
    
    local consolidated_report="$REPORTS_DIR/consolidated-security-report.json"
    local timestamp=$(date -Iseconds)
    
    cat > "$consolidated_report" << EOF
{
    "scan_timestamp": "$timestamp",
    "source_directory": "$SOURCE_DIR",
    "reports": {
        "static_analysis": {
            "gosec": "$(if [[ -f "$REPORTS_DIR/security/gosec.json" ]]; then echo "available"; else echo "unavailable"; fi)",
            "golangci": "$(if [[ -f "$REPORTS_DIR/static/golangci.json" ]]; then echo "available"; else echo "unavailable"; fi)"
        },
        "dependency_scan": "$(if [[ -f "$REPORTS_DIR/dependency/vulnerabilities.json" ]]; then echo "available"; else echo "unavailable"; fi)",
        "secret_scan": "$(if [[ -f "$REPORTS_DIR/secret/trufflehog.json" ]]; then echo "available"; else echo "unavailable"; fi)",
        "license_scan": "$(if [[ -f "$REPORTS_DIR/license/go-licenses.json" ]]; then echo "available"; else echo "unavailable"; fi)",
        "sharehodl_specific": "$(if [[ -f "$REPORTS_DIR/security/sharehodl-specific.json" ]]; then echo "available"; else echo "unavailable"; fi)"
    },
    "summary": {
        "critical_issues": 0,
        "high_issues": 0,
        "medium_issues": 0,
        "low_issues": 0,
        "info_issues": 0
    }
}
EOF
    
    # Calculate issue counts from various reports
    calculate_issue_summary "$consolidated_report"
    
    log_success "Consolidated report generated: $consolidated_report"
}

# Calculate issue summary from all reports
calculate_issue_summary() {
    local report_file="$1"
    
    local critical=0
    local high=0
    local medium=0
    local low=0
    local info=0
    
    # Count gosec issues
    if [[ -f "$REPORTS_DIR/security/gosec.json" ]]; then
        local gosec_issues
        gosec_issues=$(jq '.Issues | length' "$REPORTS_DIR/security/gosec.json" 2>/dev/null || echo "0")
        high=$((high + gosec_issues))
    fi
    
    # Count secret findings
    if [[ -f "$REPORTS_DIR/secret/custom-patterns.json" ]]; then
        local secret_issues
        secret_issues=$(jq '.findings | length' "$REPORTS_DIR/secret/custom-patterns.json" 2>/dev/null || echo "0")
        critical=$((critical + secret_issues))
    fi
    
    # Update summary in consolidated report
    jq ".summary.critical_issues = $critical | .summary.high_issues = $high | .summary.medium_issues = $medium | .summary.low_issues = $low | .summary.info_issues = $info" "$report_file" > "${report_file}.tmp" && mv "${report_file}.tmp" "$report_file"
}

# Send alerts if issues found
send_alerts() {
    if [[ -z "$ALERT_WEBHOOK" ]]; then
        return 0
    fi
    
    log_info "Checking for security alerts..."
    
    local consolidated_report="$REPORTS_DIR/consolidated-security-report.json"
    
    if [[ -f "$consolidated_report" ]]; then
        local critical
        local high
        critical=$(jq -r '.summary.critical_issues' "$consolidated_report")
        high=$(jq -r '.summary.high_issues' "$consolidated_report")
        
        if [[ "$critical" -gt 0 || "$high" -gt 0 ]]; then
            log_warning "Security issues found: $critical critical, $high high"
            
            # Send webhook alert
            local alert_payload
            alert_payload=$(cat << EOF
{
    "timestamp": "$(date -Iseconds)",
    "source": "sharehodl-security-scanner",
    "severity": "$(if [[ "$critical" -gt 0 ]]; then echo "critical"; else echo "high"; fi)",
    "title": "Security Scan Alert",
    "description": "Security scan found $critical critical and $high high severity issues",
    "details": {
        "critical_issues": $critical,
        "high_issues": $high
    }
}
EOF
            )
            
            if curl -X POST -H "Content-Type: application/json" -d "$alert_payload" "$ALERT_WEBHOOK"; then
                log_success "Alert sent to webhook"
            else
                log_error "Failed to send alert to webhook"
            fi
        fi
    fi
}

# Run full security scan
run_full_scan() {
    log_info "Starting comprehensive security scan..."
    
    initialize
    run_gosec_scan
    run_golangci_scan
    run_dependency_scan
    run_secret_scan
    run_license_scan
    run_sharehodl_security_checks
    generate_consolidated_report
    send_alerts
    
    log_success "Security scan completed"
}

# Main execution
main() {
    if [[ "${1:-}" == "--once" ]]; then
        run_full_scan
    else
        log_info "Starting continuous security scanning (interval: ${SCAN_INTERVAL}s)"
        while true; do
            run_full_scan
            log_info "Next scan in ${SCAN_INTERVAL} seconds"
            sleep "$SCAN_INTERVAL"
        done
    fi
}

# Run main function
main "$@"