# ShareHODL End-to-End Testing Suite

This directory contains the comprehensive end-to-end testing framework for the ShareHODL blockchain protocol. The testing suite validates all aspects of the system from basic functionality to complex integration scenarios.

## Overview

The E2E testing framework provides:

- **Module Integration Tests**: Individual module testing (HODL, Equity, DEX, Governance)
- **Cross-Module Integration**: Testing interactions between different modules  
- **Performance Testing**: Load testing, throughput analysis, and resource monitoring
- **Deployment Testing**: Validation of deployment procedures across environments
- **Security Testing**: Vulnerability scanning, penetration testing, and security validation
- **Protocol Validation**: Complete end-to-end protocol functionality verification

## Test Structure

```
tests/e2e/
├── e2e_test_framework.go      # Main test framework and suite setup
├── module_clients.go          # Module-specific test clients
├── integration_tests.go       # Module integration tests
├── performance_tests.go       # Performance and load testing
├── deployment_tests.go        # Deployment validation tests
├── security_validation_tests.go # Security testing suite
├── protocol_validation_tests.go # Complete protocol validation
├── main_test.go              # Test entry points and configuration
├── testdata/
│   └── test_data.json        # Test configuration and data
└── README.md                 # This documentation
```

## Prerequisites

### System Requirements

- Go 1.21+ 
- Docker and Docker Compose
- Kubernetes cluster (for K8s deployment tests)
- Terraform (for infrastructure tests)
- At least 16GB RAM and 50GB disk space

### Dependencies

Install required testing tools:

```bash
# Go testing tools
go install github.com/stretchr/testify@latest

# Security scanning tools (optional)
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install github.com/sonatypecommunity/nancy@latest

# Container scanning (optional)  
# Install trivy: https://github.com/aquasecurity/trivy

# Load testing (optional)
# Install hey: go install github.com/rakyll/hey@latest
```

### Environment Setup

Set required environment variables:

```bash
# Core test configuration
export RUN_E2E_TESTS=true
export CHAIN_ID=sharehodl-testnet
export VALIDATOR_MNEMONIC="your test validator mnemonic"

# Optional test categories  
export RUN_PERFORMANCE_TESTS=true
export RUN_DEPLOYMENT_TESTS=true
export RUN_SECURITY_TESTS=true
export RUN_PROTOCOL_VALIDATION=true
export RUN_BENCHMARKS=true

# AWS credentials (for Terraform tests)
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_DEFAULT_REGION=us-east-1

# Test environment overrides
export TEST_TIMEOUT=30m
export TEST_PARALLEL=4
```

## Running Tests

### Complete Test Suite

Run the full end-to-end test suite:

```bash
# Full test suite (can take 60-90 minutes)
make test-e2e

# Or directly with go test
go test -v ./tests/e2e/... -timeout=90m
```

### Individual Test Categories

Run specific test categories:

```bash
# Module integration tests only
go test -v ./tests/e2e -run TestModuleIntegration

# Performance tests only  
go test -v ./tests/e2e -run TestPerformance

# Deployment tests only
go test -v ./tests/e2e -run TestDeployment

# Security validation only
go test -v ./tests/e2e -run TestSecurity

# Complete protocol validation
go test -v ./tests/e2e -run TestProtocolValidation

# Business workflow tests
go test -v ./tests/e2e -run TestBusinessWorkflow
```

### Short Test Mode

For faster testing during development:

```bash
# Run only quick tests
go test -v ./tests/e2e -short

# Skip long-running tests
go test -v ./tests/e2e -run "Test.*" -skip "TestComplete|TestProtocol"
```

### Benchmarking

Run performance benchmarks:

```bash
# All benchmarks
go test -v ./tests/e2e -bench=. -benchmem

# Specific benchmark
go test -v ./tests/e2e -bench=BenchmarkTransactionProcessing
```

## Test Configuration

### Test Data

Modify `testdata/test_data.json` to customize test scenarios:

```json
{
  "companies": [...],
  "tokens": [...], 
  "orders": [...],
  "proposals": [...],
  "scenarios": [...],
  "performance": {
    "load_test_duration": "30m",
    "max_concurrent_users": 1000,
    "transaction_target": 10000,
    "memory_limit_mb": 8192,
    "cpu_limit_percent": 80.0
  }
}
```

### Environment-Specific Configuration

Create environment-specific test files:

```bash
# Development environment
cp testdata/test_data.json testdata/test_data.dev.json

# Staging environment  
cp testdata/test_data.json testdata/test_data.staging.json

# Use specific config
export TEST_DATA_FILE=testdata/test_data.dev.json
```

## Test Categories

### 1. Module Integration Tests

Tests individual ShareHODL modules:

- **HODL Module**: Stablecoin minting, burning, supply management
- **Equity Module**: Company registration, share issuance, transfers, dividends
- **DEX Module**: Order placement, matching, liquidity pools, trading
- **Governance Module**: Proposal submission, voting, execution

### 2. Performance Tests

Validates system performance under load:

- **Transaction Throughput**: Tests TPS limits and latency
- **DEX Performance**: Order processing and matching speed
- **Resource Usage**: Memory, CPU, and disk consumption
- **Block Production**: Block time consistency and finality

### 3. Deployment Tests  

Validates deployment procedures:

- **Docker Deployment**: Container-based deployment testing
- **Kubernetes Deployment**: K8s cluster deployment validation
- **Terraform Deployment**: Infrastructure as code testing
- **Script Validation**: Deployment script testing

### 4. Security Tests

Comprehensive security validation:

- **Vulnerability Scanning**: Static analysis and dependency scanning
- **Penetration Testing**: Network, web app, and API security
- **Formal Verification**: Mathematical proof of security properties
- **Security Monitoring**: Intrusion detection and alerting
- **Cryptographic Security**: Key generation, signatures, encryption

### 5. Protocol Validation

End-to-end protocol verification:

- **Core Protocol**: Consensus, transactions, state management
- **Business Logic**: All business rules and workflows
- **Integration Scenarios**: Complex multi-step workflows
- **Stress Testing**: System limits and breaking points
- **Edge Cases**: Boundary conditions and error handling
- **Recovery Testing**: Failure recovery and resilience

## Test Results

### Output Formats

Tests produce multiple output formats:

```bash
# Console output with detailed logging
go test -v ./tests/e2e

# JSON output for CI/CD integration
go test -json ./tests/e2e > test_results.json

# Coverage report
go test -cover ./tests/e2e

# Benchmark results
go test -bench=. ./tests/e2e | tee benchmark_results.txt
```

### Result Analysis

Test metrics are automatically collected:

- **Performance Metrics**: TPS, latency, resource usage
- **Security Results**: Vulnerability scores, compliance status  
- **Deployment Status**: Success rates, deployment times
- **Protocol Validation**: Test pass/fail rates, error analysis

Results are saved to temporary directories for analysis:

```
/tmp/sharehodl-e2e-*/
├── test_results.json
├── performance_metrics.json  
├── security_scan_results.json
├── deployment_logs/
└── benchmark_results.txt
```

## Continuous Integration

### GitHub Actions Integration

The test suite integrates with CI/CD pipelines:

```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      - name: Run E2E Tests
        run: |
          export RUN_E2E_TESTS=true
          make test-e2e
        env:
          TEST_TIMEOUT: 60m
```

### Test Parallelization

Tests can be run in parallel for faster execution:

```bash
# Run tests in parallel
go test -v ./tests/e2e -parallel=4

# Limit resource usage
go test -v ./tests/e2e -parallel=2 -cpu=2
```

## Troubleshooting

### Common Issues

1. **Test Timeouts**
   ```bash
   # Increase timeout
   go test -v ./tests/e2e -timeout=120m
   ```

2. **Resource Constraints**
   ```bash
   # Reduce concurrent users
   export MAX_CONCURRENT_USERS=100
   
   # Skip resource-intensive tests
   go test -v ./tests/e2e -skip "TestMemoryPressure|TestLargeState"
   ```

3. **Network Issues**
   ```bash
   # Use local test network
   export USE_LOCAL_NETWORK=true
   
   # Increase network timeouts
   export NETWORK_TIMEOUT=60s
   ```

4. **Docker Issues**
   ```bash
   # Clean up containers
   docker system prune -f
   
   # Rebuild images
   docker build --no-cache -t sharehodl:test .
   ```

### Debug Mode

Enable verbose debugging:

```bash
# Full debug output
export DEBUG=true
export LOG_LEVEL=debug

# Test-specific debugging  
go test -v ./tests/e2e -run TestSpecific -args -debug
```

### Log Analysis

Analyze test logs:

```bash
# Filter for errors
go test -v ./tests/e2e 2>&1 | grep "ERROR\|FAIL"

# Save full logs
go test -v ./tests/e2e 2>&1 | tee full_test_log.txt

# Monitor resource usage during tests
htop &
go test -v ./tests/e2e
```

## Contributing

### Adding New Tests

1. **Module Tests**: Add to `integration_tests.go`
2. **Performance Tests**: Add to `performance_tests.go`  
3. **Security Tests**: Add to `security_validation_tests.go`
4. **Deployment Tests**: Add to `deployment_tests.go`

### Test Guidelines

- Use descriptive test names
- Add comprehensive logging
- Include proper cleanup
- Test both success and failure cases
- Document test requirements
- Add appropriate timeouts

### Example Test Structure

```go
func (s *E2ETestSuite) TestNewFeature() {
    s.T().Log("🧪 Testing New Feature")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    startTime := time.Now()
    defer func() {
        s.recordTestResult("New_Feature", 
            true, // success
            "Feature works correctly", 
            startTime)
    }()
    
    // Test implementation
    // ...
    
    s.T().Log("✅ New Feature test completed")
}
```

## Performance Baselines

### Expected Performance

- **Transaction Throughput**: 100+ TPS
- **Block Time**: 6 seconds ± 2 seconds  
- **API Response**: < 100ms average
- **Memory Usage**: < 8GB under load
- **CPU Usage**: < 80% under load

### Benchmark Targets

- **HODL Operations**: 1000+ ops/sec
- **Equity Operations**: 500+ ops/sec  
- **DEX Operations**: 200+ ops/sec
- **Query Operations**: 2000+ ops/sec

## Security Standards

### Security Test Coverage

- ✅ Static code analysis (gosec)
- ✅ Dependency vulnerability scanning  
- ✅ Container image scanning
- ✅ Network penetration testing
- ✅ Web application security testing
- ✅ API security validation
- ✅ Cryptographic security verification
- ✅ Formal verification protocols

### Compliance Requirements

- Security score ≥ 80/100
- Zero critical vulnerabilities
- All penetration tests pass
- Formal verification complete
- Security monitoring active

## Support

For issues with the testing framework:

1. Check this documentation
2. Review test logs and error messages
3. Search existing GitHub issues
4. Create new issue with:
   - Test command used
   - Full error output
   - Environment details
   - Steps to reproduce

## License

This testing framework is part of the ShareHODL project and follows the same licensing terms.