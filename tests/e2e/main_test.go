package e2e

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestMain is the main test entry point for E2E testing
func TestMain(m *testing.M) {
	// Setup code here if needed
	code := m.Run()
	// Cleanup code here if needed
	os.Exit(code)
}

// TestE2ESuite runs the complete E2E test suite
func TestE2ESuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	
	// Check if E2E tests should run
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("E2E tests disabled (set RUN_E2E_TESTS=true to enable)")
	}
	
	suite.Run(t, new(E2ETestSuite))
}

// TestModuleIntegration runs module integration tests only
func TestModuleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	testSuite := new(E2ETestSuite)
	suite.Run(t, testSuite)
	
	// Run individual module tests
	testSuite.TestHODLModule()
	testSuite.TestEquityModule()
	testSuite.TestDEXModule()
	testSuite.TestGovernanceModule()
	testSuite.TestCrossModuleIntegration()
}

// TestPerformanceTests runs performance tests only
func TestPerformanceTests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}
	
	if os.Getenv("RUN_PERFORMANCE_TESTS") != "true" {
		t.Skip("Performance tests disabled (set RUN_PERFORMANCE_TESTS=true to enable)")
	}
	
	testSuite := new(E2ETestSuite)
	suite.Run(t, testSuite)
	
	// Run performance tests
	testSuite.TestTransactionThroughput()
	testSuite.TestDEXPerformance()
	testSuite.TestMemoryAndResourceUsage()
	testSuite.TestBlockTimeAndFinality()
}

// TestDeployment runs deployment tests only
func TestDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deployment tests in short mode")
	}
	
	if os.Getenv("RUN_DEPLOYMENT_TESTS") != "true" {
		t.Skip("Deployment tests disabled (set RUN_DEPLOYMENT_TESTS=true to enable)")
	}
	
	testSuite := new(E2ETestSuite)
	suite.Run(t, testSuite)
	
	// Run deployment tests
	testSuite.TestDockerDeployment()
	testSuite.TestKubernetesDeployment()
	testSuite.TestTerraformDeployment()
	testSuite.TestDeploymentScripts()
}

// TestSecurity runs security validation tests only
func TestSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security tests in short mode")
	}
	
	if os.Getenv("RUN_SECURITY_TESTS") != "true" {
		t.Skip("Security tests disabled (set RUN_SECURITY_TESTS=true to enable)")
	}
	
	testSuite := new(E2ETestSuite)
	suite.Run(t, testSuite)
	
	// Run security tests
	testSuite.TestSecurityScanning()
	testSuite.TestPenetrationTesting()
	testSuite.TestFormalVerification()
	testSuite.TestSecurityMonitoring()
	testSuite.TestCryptographicSecurity()
}

// TestProtocolValidation runs complete protocol validation
func TestProtocolValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping protocol validation in short mode")
	}
	
	if os.Getenv("RUN_PROTOCOL_VALIDATION") != "true" {
		t.Skip("Protocol validation disabled (set RUN_PROTOCOL_VALIDATION=true to enable)")
	}
	
	testSuite := new(E2ETestSuite)
	suite.Run(t, testSuite)
	
	// Run complete protocol validation
	testSuite.TestCompleteProtocolValidation()
}

// TestBusinessWorkflow runs business workflow tests only
func TestBusinessWorkflow(t *testing.T) {
	testSuite := new(E2ETestSuite)
	suite.Run(t, testSuite)
	
	// Run business workflow test
	testSuite.TestBusinessWorkflow()
}

// Benchmark tests for performance measurement

// BenchmarkTransactionProcessing benchmarks transaction processing
func BenchmarkTransactionProcessing(b *testing.B) {
	if os.Getenv("RUN_BENCHMARKS") != "true" {
		b.Skip("Benchmarks disabled (set RUN_BENCHMARKS=true to enable)")
	}
	
	// Setup test suite
	testSuite := new(E2ETestSuite)
	testSuite.SetupSuite()
	defer testSuite.TearDownSuite()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Benchmark HODL minting transaction
		// This would contain the actual benchmarking logic
		b.StopTimer()
		// Setup for individual transaction
		b.StartTimer()
		
		// Execute transaction
		// testSuite.hodlClient.MintHODL(...)
		
		b.StopTimer()
		// Cleanup for individual transaction
		b.StartTimer()
	}
}

// BenchmarkDEXTrading benchmarks DEX trading operations
func BenchmarkDEXTrading(b *testing.B) {
	if os.Getenv("RUN_BENCHMARKS") != "true" {
		b.Skip("Benchmarks disabled (set RUN_BENCHMARKS=true to enable)")
	}
	
	testSuite := new(E2ETestSuite)
	testSuite.SetupSuite()
	defer testSuite.TearDownSuite()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Benchmark order placement
		// This would contain the actual benchmarking logic
	}
}

// BenchmarkQueryOperations benchmarks query operations
func BenchmarkQueryOperations(b *testing.B) {
	if os.Getenv("RUN_BENCHMARKS") != "true" {
		b.Skip("Benchmarks disabled (set RUN_BENCHMARKS=true to enable)")
	}
	
	testSuite := new(E2ETestSuite)
	testSuite.SetupSuite()
	defer testSuite.TearDownSuite()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Benchmark balance queries
		// This would contain the actual benchmarking logic
	}
}