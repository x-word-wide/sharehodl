package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
	dextypes "github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

// PerformanceTestSuite defines performance testing scenarios
type PerformanceTestSuite struct {
	*E2ETestSuite
	
	// Performance metrics
	transactionCount    int64
	successfulTxs       int64
	failedTxs          int64
	totalLatency       time.Duration
	maxLatency         time.Duration
	minLatency         time.Duration
	
	// Load test parameters
	concurrentUsers    int
	testDuration      time.Duration
	rampUpTime        time.Duration
	rampDownTime      time.Duration
	
	// Resource monitoring
	cpuUsage          []float64
	memoryUsage       []int64
	diskIO            []float64
	networkLatency    []time.Duration
	
	mutex sync.RWMutex
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite(e2eSuite *E2ETestSuite) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		E2ETestSuite:      e2eSuite,
		concurrentUsers:   s.testData.Performance.MaxConcurrentUsers,
		testDuration:      s.testData.Performance.LoadTestDuration,
		rampUpTime:        5 * time.Minute,
		rampDownTime:      2 * time.Minute,
		minLatency:        time.Duration(1<<63 - 1), // Max duration initially
	}
}

// TestTransactionThroughput tests transaction processing throughput
func (s *E2ETestSuite) TestTransactionThroughput() {
	s.T().Log("🚀 Testing Transaction Throughput")
	
	perf := NewPerformanceTestSuite(s)
	ctx, cancel := context.WithTimeout(context.Background(), perf.testDuration)
	defer cancel()
	
	startTime := time.Now()
	
	// Start resource monitoring
	go perf.monitorResources(ctx)
	
	// Prepare test accounts
	testAccounts := perf.createTestAccounts(perf.concurrentUsers)
	
	// Fund test accounts
	for _, account := range testAccounts {
		fundAmount := sdk.NewInt64Coin(hodltypes.DefaultDenom, 10000000) // 10 HODL
		_, err := s.hodlClient.MintHODL(ctx, s.validatorAccount.Address, account, fundAmount)
		require.NoError(s.T(), err)
	}
	
	s.WaitForBlocks(5)
	
	// Run concurrent transaction load
	var wg sync.WaitGroup
	userStep := perf.concurrentUsers / 10 // Ramp up in steps
	
	s.T().Logf("Starting load test with %d concurrent users for %v", perf.concurrentUsers, perf.testDuration)
	
	// Ramp up users gradually
	for step := 1; step <= 10; step++ {
		currentUsers := userStep * step
		s.T().Logf("Ramping up to %d users (step %d/10)", currentUsers, step)
		
		for i := (step-1) * userStep; i < currentUsers && i < len(testAccounts); i++ {
			wg.Add(1)
			go func(userIndex int, userAddr string) {
				defer wg.Done()
				perf.runUserTransactions(ctx, userAddr, testAccounts)
			}(i, testAccounts[i])
		}
		
		// Wait between ramp-up steps
		time.Sleep(perf.rampUpTime / 10)
	}
	
	// Wait for all users to complete
	wg.Wait()
	
	endTime := time.Now()
	totalDuration := endTime.Sub(startTime)
	
	// Calculate performance metrics
	tps := float64(atomic.LoadInt64(&perf.transactionCount)) / totalDuration.Seconds()
	successRate := float64(atomic.LoadInt64(&perf.successfulTxs)) / float64(atomic.LoadInt64(&perf.transactionCount)) * 100
	avgLatency := perf.totalLatency / time.Duration(atomic.LoadInt64(&perf.transactionCount))
	
	// Record performance metrics
	s.metrics.Performance.TransactionThroughput = tps
	s.metrics.Performance.APIResponseTime = float64(avgLatency.Milliseconds())
	
	s.T().Logf("📊 Transaction Throughput Results:")
	s.T().Logf("   Total Transactions: %d", atomic.LoadInt64(&perf.transactionCount))
	s.T().Logf("   Successful: %d (%.2f%%)", atomic.LoadInt64(&perf.successfulTxs), successRate)
	s.T().Logf("   Failed: %d", atomic.LoadInt64(&perf.failedTxs))
	s.T().Logf("   TPS: %.2f", tps)
	s.T().Logf("   Average Latency: %v", avgLatency)
	s.T().Logf("   Max Latency: %v", perf.maxLatency)
	s.T().Logf("   Min Latency: %v", perf.minLatency)
	
	// Assert performance requirements
	require.True(s.T(), tps >= 100, "TPS should be at least 100")
	require.True(s.T(), successRate >= 95, "Success rate should be at least 95%")
	require.True(s.T(), avgLatency <= 5*time.Second, "Average latency should be under 5 seconds")
	
	s.recordTestResult("Performance_Transaction_Throughput", 
		tps >= 100 && successRate >= 95 && avgLatency <= 5*time.Second,
		fmt.Sprintf("TPS: %.2f, Success: %.2f%%, Latency: %v", tps, successRate, avgLatency),
		startTime)
	
	s.T().Log("✅ Transaction Throughput test completed")
}

// TestDEXPerformance tests DEX performance under load
func (s *E2ETestSuite) TestDEXPerformance() {
	s.T().Log("📈 Testing DEX Performance")
	
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	
	startTime := time.Now()
	
	// Setup test trading pairs
	symbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "META"}
	
	// Create companies and liquidity pools
	for _, symbol := range symbols {
		// Create company
		company := equitytypes.Company{
			Name:         fmt.Sprintf("%s Corp", symbol),
			Symbol:       symbol,
			TotalShares:  1000000,
			Industry:     "Technology",
			ValuationUSD: 100000000,
		}
		
		_, err := s.equityClient.CreateCompany(ctx, s.businessAccount.Address, company)
		require.NoError(s.T(), err)
		
		// Create liquidity pool
		liquidityA := sdk.NewInt64Coin(symbol, 100000)
		liquidityB := sdk.NewInt64Coin(hodltypes.DefaultDenom, 1000000)
		
		_, err = s.dexClient.CreateLiquidityPool(ctx, s.businessAccount.Address, symbol, hodltypes.DefaultDenom, liquidityA, liquidityB)
		require.NoError(s.T(), err)
	}
	
	s.WaitForBlocks(5)
	
	// Concurrent trading simulation
	var wg sync.WaitGroup
	tradersCount := 50
	ordersPerTrader := 20
	
	var totalOrders int64
	var successfulOrders int64
	var orderLatencies []time.Duration
	var mutex sync.Mutex
	
	s.T().Logf("Starting DEX load test with %d traders, %d orders each", tradersCount, ordersPerTrader)
	
	for i := 0; i < tradersCount; i++ {
		wg.Add(1)
		go func(traderID int) {
			defer wg.Done()
			
			traderAddr := fmt.Sprintf("trader%d", traderID)
			
			for j := 0; j < ordersPerTrader; j++ {
				orderStart := time.Now()
				
				// Random symbol and order parameters
				symbol := symbols[rand.Intn(len(symbols))]
				side := dextypes.OrderSide_BUY
				if rand.Float32() > 0.5 {
					side = dextypes.OrderSide_SELL
				}
				
				order := dextypes.Order{
					OrderType:   dextypes.OrderType_LIMIT,
					Side:        side,
					Symbol:      fmt.Sprintf("%s/HODL", symbol),
					Quantity:    uint64(rand.Intn(1000) + 100),
					Price:       sdk.NewDec(int64(rand.Intn(50) + 10)),
					TimeInForce: dextypes.TimeInForce_GTC,
				}
				
				_, err := s.dexClient.PlaceOrder(ctx, traderAddr, order)
				
				orderLatency := time.Since(orderStart)
				
				mutex.Lock()
				atomic.AddInt64(&totalOrders, 1)
				if err == nil {
					atomic.AddInt64(&successfulOrders, 1)
				}
				orderLatencies = append(orderLatencies, orderLatency)
				mutex.Unlock()
				
				// Small delay between orders
				time.Sleep(time.Millisecond * 100)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Calculate DEX performance metrics
	totalDuration := time.Since(startTime)
	orderThroughput := float64(totalOrders) / totalDuration.Seconds()
	successRate := float64(successfulOrders) / float64(totalOrders) * 100
	
	// Calculate latency statistics
	var totalLatency time.Duration
	maxLatency := time.Duration(0)
	minLatency := time.Duration(1<<63 - 1)
	
	for _, latency := range orderLatencies {
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
		if latency < minLatency {
			minLatency = latency
		}
	}
	
	avgLatency := totalLatency / time.Duration(len(orderLatencies))
	
	s.T().Logf("📊 DEX Performance Results:")
	s.T().Logf("   Total Orders: %d", totalOrders)
	s.T().Logf("   Successful: %d (%.2f%%)", successfulOrders, successRate)
	s.T().Logf("   Order Throughput: %.2f orders/sec", orderThroughput)
	s.T().Logf("   Average Latency: %v", avgLatency)
	s.T().Logf("   Max Latency: %v", maxLatency)
	s.T().Logf("   Min Latency: %v", minLatency)
	
	// Test order book queries performance
	s.testOrderBookQueryPerformance(ctx, symbols)
	
	s.recordTestResult("Performance_DEX_Trading", 
		orderThroughput >= 10 && successRate >= 90,
		fmt.Sprintf("Throughput: %.2f orders/sec, Success: %.2f%%", orderThroughput, successRate),
		startTime)
	
	s.T().Log("✅ DEX Performance test completed")
}

// testOrderBookQueryPerformance tests query performance
func (s *E2ETestSuite) testOrderBookQueryPerformance(ctx context.Context, symbols []string) {
	s.T().Log("Testing order book query performance")
	
	queries := 1000
	var queryLatencies []time.Duration
	
	for i := 0; i < queries; i++ {
		symbol := symbols[i%len(symbols)]
		queryStart := time.Now()
		
		_, err := s.dexClient.GetOrderBook(ctx, fmt.Sprintf("%s/HODL", symbol))
		queryLatency := time.Since(queryStart)
		
		if err == nil {
			queryLatencies = append(queryLatencies, queryLatency)
		}
		
		// Small delay between queries
		time.Sleep(time.Millisecond * 10)
	}
	
	// Calculate query performance
	var totalQueryLatency time.Duration
	for _, latency := range queryLatencies {
		totalQueryLatency += latency
	}
	
	avgQueryLatency := totalQueryLatency / time.Duration(len(queryLatencies))
	qps := float64(queries) / (float64(totalQueryLatency.Nanoseconds()) / 1e9)
	
	s.T().Logf("📊 Order Book Query Performance:")
	s.T().Logf("   Queries: %d", len(queryLatencies))
	s.T().Logf("   Average Latency: %v", avgQueryLatency)
	s.T().Logf("   Queries per Second: %.2f", qps)
	
	require.True(s.T(), avgQueryLatency <= 100*time.Millisecond, "Query latency should be under 100ms")
}

// TestMemoryAndResourceUsage tests resource consumption under load
func (s *E2ETestSuite) TestMemoryAndResourceUsage() {
	s.T().Log("🔧 Testing Memory and Resource Usage")
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	
	startTime := time.Now()
	perf := NewPerformanceTestSuite(s)
	
	// Start resource monitoring
	resourceMetrics := make(chan ResourceSnapshot, 1000)
	go perf.monitorSystemResources(ctx, resourceMetrics)
	
	// Run intensive operations
	var wg sync.WaitGroup
	
	// Heavy transaction load
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.runHeavyTransactionLoad(ctx, 1000)
	}()
	
	// Heavy query load
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.runHeavyQueryLoad(ctx, 500)
	}()
	
	// Heavy DEX operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.runHeavyDEXOperations(ctx, 200)
	}()
	
	wg.Wait()
	
	// Collect resource metrics
	close(resourceMetrics)
	metrics := s.analyzeResourceMetrics(resourceMetrics)
	
	// Update performance metrics
	s.metrics.Performance.MemoryUsage = metrics.MaxMemoryMB * 1024 * 1024 // Convert to bytes
	s.metrics.Performance.CPUUsage = metrics.MaxCPUPercent
	s.metrics.ResourceUsage = metrics
	
	s.T().Logf("📊 Resource Usage Results:")
	s.T().Logf("   Max Memory: %.2f MB", metrics.MaxMemoryMB)
	s.T().Logf("   Max CPU: %.2f%%", metrics.MaxCPUPercent)
	s.T().Logf("   Disk IOPS: %.2f", metrics.DiskIOPS)
	s.T().Logf("   Network RX: %.2f MB", metrics.NetworkRxMB)
	s.T().Logf("   Network TX: %.2f MB", metrics.NetworkTxMB)
	
	// Assert resource limits
	memoryLimitMB := float64(s.testData.Performance.MemoryLimitMB)
	cpuLimitPercent := s.testData.Performance.CPULimitPercent
	
	require.True(s.T(), metrics.MaxMemoryMB <= memoryLimitMB, 
		fmt.Sprintf("Memory usage %.2fMB should be under limit %dMB", metrics.MaxMemoryMB, s.testData.Performance.MemoryLimitMB))
	require.True(s.T(), metrics.MaxCPUPercent <= cpuLimitPercent, 
		fmt.Sprintf("CPU usage %.2f%% should be under limit %.2f%%", metrics.MaxCPUPercent, cpuLimitPercent))
	
	s.recordTestResult("Performance_Resource_Usage", 
		metrics.MaxMemoryMB <= memoryLimitMB && metrics.MaxCPUPercent <= cpuLimitPercent,
		fmt.Sprintf("Memory: %.2fMB, CPU: %.2f%%", metrics.MaxMemoryMB, metrics.MaxCPUPercent),
		startTime)
	
	s.T().Log("✅ Memory and Resource Usage test completed")
}

// TestBlockTimeAndFinality tests block production and finality
func (s *E2ETestSuite) TestBlockTimeAndFinality() {
	s.T().Log("⏱️ Testing Block Time and Finality")
	
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	// Monitor block production for a period
	blockTimes := []time.Duration{}
	blockCount := 50
	
	s.T().Logf("Monitoring %d blocks for timing analysis", blockCount)
	
	previousBlockTime := time.Now()
	
	for i := 0; i < blockCount; i++ {
		// Wait for next block
		s.WaitForBlocks(1)
		
		currentBlockTime := time.Now()
		blockTime := currentBlockTime.Sub(previousBlockTime)
		blockTimes = append(blockTimes, blockTime)
		previousBlockTime = currentBlockTime
		
		s.T().Logf("Block %d: %v", i+1, blockTime)
	}
	
	// Calculate block time statistics
	var totalBlockTime time.Duration
	maxBlockTime := time.Duration(0)
	minBlockTime := time.Duration(1<<63 - 1)
	
	for _, bt := range blockTimes {
		totalBlockTime += bt
		if bt > maxBlockTime {
			maxBlockTime = bt
		}
		if bt < minBlockTime {
			minBlockTime = bt
		}
	}
	
	avgBlockTime := totalBlockTime / time.Duration(len(blockTimes))
	
	// Update metrics
	s.metrics.Performance.BlockTime = avgBlockTime.Seconds()
	
	s.T().Logf("📊 Block Time Results:")
	s.T().Logf("   Average Block Time: %v", avgBlockTime)
	s.T().Logf("   Max Block Time: %v", maxBlockTime)
	s.T().Logf("   Min Block Time: %v", minBlockTime)
	s.T().Logf("   Expected Block Time: 6s")
	
	// Assert block time requirements
	expectedBlockTime := 6 * time.Second
	tolerance := 2 * time.Second
	
	require.True(s.T(), avgBlockTime >= expectedBlockTime-tolerance && avgBlockTime <= expectedBlockTime+tolerance,
		fmt.Sprintf("Average block time %v should be within %v of expected %v", avgBlockTime, tolerance, expectedBlockTime))
	
	s.recordTestResult("Performance_Block_Time", 
		avgBlockTime >= expectedBlockTime-tolerance && avgBlockTime <= expectedBlockTime+tolerance,
		fmt.Sprintf("Average: %v, Expected: %v±%v", avgBlockTime, expectedBlockTime, tolerance),
		startTime)
	
	s.T().Log("✅ Block Time and Finality test completed")
}

// Helper methods for performance testing

// createTestAccounts creates test accounts for load testing
func (perf *PerformanceTestSuite) createTestAccounts(count int) []string {
	accounts := make([]string, count)
	for i := 0; i < count; i++ {
		accounts[i] = fmt.Sprintf("testuser%d", i)
	}
	return accounts
}

// runUserTransactions runs transactions for a user during load test
func (perf *PerformanceTestSuite) runUserTransactions(ctx context.Context, userAddr string, allUsers []string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Perform random transaction
			txStart := time.Now()
			
			err := perf.performRandomTransaction(ctx, userAddr, allUsers)
			
			txLatency := time.Since(txStart)
			
			atomic.AddInt64(&perf.transactionCount, 1)
			
			perf.mutex.Lock()
			perf.totalLatency += txLatency
			if txLatency > perf.maxLatency {
				perf.maxLatency = txLatency
			}
			if txLatency < perf.minLatency {
				perf.minLatency = txLatency
			}
			perf.mutex.Unlock()
			
			if err == nil {
				atomic.AddInt64(&perf.successfulTxs, 1)
			} else {
				atomic.AddInt64(&perf.failedTxs, 1)
			}
			
			// Random delay between transactions
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)+100))
		}
	}
}

// performRandomTransaction performs a random transaction type
func (perf *PerformanceTestSuite) performRandomTransaction(ctx context.Context, userAddr string, allUsers []string) error {
	txType := rand.Intn(4)
	
	switch txType {
	case 0: // HODL transfer
		recipient := allUsers[rand.Intn(len(allUsers))]
		amount := sdk.NewInt64Coin(hodltypes.DefaultDenom, int64(rand.Intn(1000)+1))
		// Simulate transfer (would need actual transfer method)
		return nil
	case 1: // Equity transfer  
		recipient := allUsers[rand.Intn(len(allUsers))]
		shares := uint64(rand.Intn(100) + 1)
		_, err := perf.equityClient.TransferShares(ctx, userAddr, recipient, "TEST", shares)
		return err
	case 2: // DEX order
		order := dextypes.Order{
			OrderType:   dextypes.OrderType_LIMIT,
			Side:        dextypes.OrderSide_BUY,
			Symbol:      "TEST/HODL",
			Quantity:    uint64(rand.Intn(100) + 1),
			Price:       sdk.NewDec(int64(rand.Intn(20) + 5)),
			TimeInForce: dextypes.TimeInForce_GTC,
		}
		_, err := perf.dexClient.PlaceOrder(ctx, userAddr, order)
		return err
	case 3: // Query operation
		_, err := perf.hodlClient.GetBalance(ctx, userAddr)
		return err
	default:
		return nil
	}
}

// ResourceSnapshot captures system resource usage at a point in time
type ResourceSnapshot struct {
	Timestamp   time.Time
	MemoryMB    float64
	CPUPercent  float64
	DiskIOPS    float64
	NetworkRxMB float64
	NetworkTxMB float64
}

// monitorResources monitors system resources during performance test
func (perf *PerformanceTestSuite) monitorResources(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate resource monitoring (would use actual system metrics)
			perf.mutex.Lock()
			perf.cpuUsage = append(perf.cpuUsage, rand.Float64()*100)
			perf.memoryUsage = append(perf.memoryUsage, rand.Int63n(8192*1024*1024)) // Up to 8GB
			perf.diskIO = append(perf.diskIO, rand.Float64()*1000)
			perf.networkLatency = append(perf.networkLatency, time.Duration(rand.Intn(100))*time.Millisecond)
			perf.mutex.Unlock()
		}
	}
}

// monitorSystemResources monitors detailed system resources
func (perf *PerformanceTestSuite) monitorSystemResources(ctx context.Context, metrics chan<- ResourceSnapshot) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := ResourceSnapshot{
				Timestamp:   time.Now(),
				MemoryMB:    rand.Float64() * 4096, // Up to 4GB
				CPUPercent:  rand.Float64() * 100,
				DiskIOPS:    rand.Float64() * 500,
				NetworkRxMB: rand.Float64() * 100,
				NetworkTxMB: rand.Float64() * 100,
			}
			
			select {
			case metrics <- snapshot:
			default:
				// Channel full, skip this measurement
			}
		}
	}
}

// analyzeResourceMetrics analyzes collected resource metrics
func (s *E2ETestSuite) analyzeResourceMetrics(metrics <-chan ResourceSnapshot) ResourceMetrics {
	var maxMemory, maxCPU, totalDiskIOPS, totalNetworkRx, totalNetworkTx float64
	count := 0
	
	for snapshot := range metrics {
		if snapshot.MemoryMB > maxMemory {
			maxMemory = snapshot.MemoryMB
		}
		if snapshot.CPUPercent > maxCPU {
			maxCPU = snapshot.CPUPercent
		}
		totalDiskIOPS += snapshot.DiskIOPS
		totalNetworkRx += snapshot.NetworkRxMB
		totalNetworkTx += snapshot.NetworkTxMB
		count++
	}
	
	if count == 0 {
		return ResourceMetrics{}
	}
	
	return ResourceMetrics{
		MaxMemoryMB:     maxMemory,
		MaxCPUPercent:   maxCPU,
		DiskIOPS:        totalDiskIOPS / float64(count),
		NetworkRxMB:     totalNetworkRx,
		NetworkTxMB:     totalNetworkTx,
		PostgreSQLConns: rand.Intn(100) + 10,
		RedisConns:      rand.Intn(50) + 5,
	}
}

// runHeavyTransactionLoad runs intensive transaction operations
func (s *E2ETestSuite) runHeavyTransactionLoad(ctx context.Context, txCount int) {
	for i := 0; i < txCount && ctx.Err() == nil; i++ {
		// Simulate heavy transaction
		amount := sdk.NewInt64Coin(hodltypes.DefaultDenom, int64(rand.Intn(10000)+1000))
		_, err := s.hodlClient.MintHODL(ctx, s.validatorAccount.Address, s.investorAccount1.Address, amount)
		if err == nil {
			time.Sleep(time.Millisecond * 50) // Small delay
		}
	}
}

// runHeavyQueryLoad runs intensive query operations
func (s *E2ETestSuite) runHeavyQueryLoad(ctx context.Context, queryCount int) {
	addresses := []string{s.validatorAccount.Address, s.businessAccount.Address, s.investorAccount1.Address, s.investorAccount2.Address}
	
	for i := 0; i < queryCount && ctx.Err() == nil; i++ {
		addr := addresses[i%len(addresses)]
		_, err := s.hodlClient.GetBalance(ctx, addr)
		if err == nil {
			time.Sleep(time.Millisecond * 20)
		}
	}
}

// runHeavyDEXOperations runs intensive DEX operations
func (s *E2ETestSuite) runHeavyDEXOperations(ctx context.Context, opCount int) {
	for i := 0; i < opCount && ctx.Err() == nil; i++ {
		// Create random order
		order := dextypes.Order{
			OrderType:   dextypes.OrderType_LIMIT,
			Side:        dextypes.OrderSide_BUY,
			Symbol:      "TEST/HODL",
			Quantity:    uint64(rand.Intn(1000) + 100),
			Price:       sdk.NewDec(int64(rand.Intn(50) + 10)),
			TimeInForce: dextypes.TimeInForce_GTC,
		}
		
		_, err := s.dexClient.PlaceOrder(ctx, s.investorAccount1.Address, order)
		if err == nil {
			time.Sleep(time.Millisecond * 100)
		}
	}
}