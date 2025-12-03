package security

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TransactionMonitor monitors transaction-related security events
type TransactionMonitor struct {
	name             string
	description      string
	monitoredEvents  []EventType
	isRunning        bool
	metrics          MonitorMetrics
	config           map[string]interface{}
	suspiciousAmounts map[string]float64
	transactionHistory map[string][]TransactionRecord
	mu               sync.RWMutex
}

type TransactionRecord struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

func NewTransactionMonitor() *TransactionMonitor {
	return &TransactionMonitor{
		name:        "transaction_monitor",
		description: "Monitors transactions for suspicious patterns and anomalies",
		monitoredEvents: []EventType{EventTypeTransaction},
		suspiciousAmounts: make(map[string]float64),
		transactionHistory: make(map[string][]TransactionRecord),
		metrics: MonitorMetrics{},
	}
}

func (tm *TransactionMonitor) GetName() string { return tm.name }
func (tm *TransactionMonitor) GetDescription() string { return tm.description }
func (tm *TransactionMonitor) GetMonitoredEvents() []EventType { return tm.monitoredEvents }

func (tm *TransactionMonitor) Configure(config map[string]interface{}) error {
	tm.config = config
	return nil
}

func (tm *TransactionMonitor) Start(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if tm.isRunning {
		return fmt.Errorf("transaction monitor already running")
	}
	
	tm.isRunning = true
	
	// Start background cleanup goroutine
	go tm.cleanupOldRecords(ctx)
	
	return nil
}

func (tm *TransactionMonitor) Stop() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	tm.isRunning = false
	return nil
}

func (tm *TransactionMonitor) ProcessEvent(event SecurityEvent) []SecurityAlert {
	startTime := time.Now()
	alerts := make([]SecurityAlert, 0)
	
	defer func() {
		tm.mu.Lock()
		tm.metrics.EventsProcessed++
		tm.metrics.ProcessingTime = time.Since(startTime)
		tm.mu.Unlock()
	}()

	if event.Type != EventTypeTransaction {
		return alerts
	}

	// Extract transaction details
	record := tm.extractTransactionRecord(event)
	if record == nil {
		return alerts
	}

	// Store transaction record
	tm.mu.Lock()
	if tm.transactionHistory[record.From] == nil {
		tm.transactionHistory[record.From] = make([]TransactionRecord, 0)
	}
	tm.transactionHistory[record.From] = append(tm.transactionHistory[record.From], *record)
	tm.mu.Unlock()

	// Check for suspicious patterns
	alerts = append(alerts, tm.checkLargeAmount(*record, event)...)
	alerts = append(alerts, tm.checkRapidTransactions(*record, event)...)
	alerts = append(alerts, tm.checkUnusualPatterns(*record, event)...)
	alerts = append(alerts, tm.checkCircularTransactions(*record, event)...)

	tm.mu.Lock()
	tm.metrics.AlertsGenerated += int64(len(alerts))
	tm.mu.Unlock()

	return alerts
}

func (tm *TransactionMonitor) GetMetrics() MonitorMetrics {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.metrics
}

func (tm *TransactionMonitor) extractTransactionRecord(event SecurityEvent) *TransactionRecord {
	// Extract transaction details from event data
	amountStr, exists := event.Data["amount"]
	if !exists {
		return nil
	}
	
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil
	}

	return &TransactionRecord{
		ID:        event.Data["transaction_id"],
		From:      event.Data["from"],
		To:        event.Data["to"],
		Amount:    amount,
		Timestamp: event.Timestamp,
		Type:      event.Data["type"],
	}
}

func (tm *TransactionMonitor) checkLargeAmount(record TransactionRecord, event SecurityEvent) []SecurityAlert {
	alerts := make([]SecurityAlert, 0)
	
	// Define large amount threshold (would be configurable)
	largeAmountThreshold := 1000000.0
	
	if record.Amount > largeAmountThreshold {
		alert := SecurityAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Type:        AlertTypeAnomalousTransaction,
			Severity:    SeverityHigh,
			Title:       "Large Transaction Amount",
			Description: fmt.Sprintf("Transaction of %.2f detected from %s to %s", record.Amount, record.From, record.To),
			Source:      tm.name,
			TriggerEvent: event,
			RiskScore:   calculateRiskScore(record.Amount, largeAmountThreshold),
			Confidence:  85.0,
			Status:      StatusOpen,
			Evidence: []AlertEvidence{
				{
					Type:        "transaction_data",
					Description: "Large amount transaction",
					Data: map[string]string{
						"amount":         fmt.Sprintf("%.2f", record.Amount),
						"threshold":      fmt.Sprintf("%.2f", largeAmountThreshold),
						"transaction_id": record.ID,
					},
					Timestamp: time.Now(),
				},
			},
			Recommendations: []string{
				"Verify the legitimacy of this large transaction",
				"Check if the sender has authorization for such amounts",
				"Monitor for follow-up suspicious activity",
			},
			Tags:     []string{"large_amount", "anomaly", "transaction"},
			CreatedBy: tm.name,
		}
		alerts = append(alerts, alert)
	}
	
	return alerts
}

func (tm *TransactionMonitor) checkRapidTransactions(record TransactionRecord, event SecurityEvent) []SecurityAlert {
	alerts := make([]SecurityAlert, 0)
	
	tm.mu.RLock()
	history := tm.transactionHistory[record.From]
	tm.mu.RUnlock()
	
	if len(history) < 2 {
		return alerts
	}
	
	// Check for rapid transactions (more than 10 in 5 minutes)
	recentCount := 0
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	
	for _, tx := range history {
		if tx.Timestamp.After(fiveMinutesAgo) {
			recentCount++
		}
	}
	
	if recentCount > 10 {
		alert := SecurityAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Type:        AlertTypeSuspiciousActivity,
			Severity:    SeverityMedium,
			Title:       "Rapid Transaction Pattern",
			Description: fmt.Sprintf("Account %s performed %d transactions in the last 5 minutes", record.From, recentCount),
			Source:      tm.name,
			TriggerEvent: event,
			RiskScore:   float64(recentCount) * 5.0,
			Confidence:  75.0,
			Status:      StatusOpen,
			Evidence: []AlertEvidence{
				{
					Type:        "transaction_frequency",
					Description: "High frequency transaction pattern",
					Data: map[string]string{
						"transaction_count": fmt.Sprintf("%d", recentCount),
						"time_window":       "5 minutes",
						"account":          record.From,
					},
					Timestamp: time.Now(),
				},
			},
			Recommendations: []string{
				"Investigate the purpose of rapid transactions",
				"Check for automated bot activity",
				"Verify account security",
			},
			Tags:     []string{"rapid_transactions", "frequency", "suspicious"},
			CreatedBy: tm.name,
		}
		alerts = append(alerts, alert)
	}
	
	return alerts
}

func (tm *TransactionMonitor) checkUnusualPatterns(record TransactionRecord, event SecurityEvent) []SecurityAlert {
	alerts := make([]SecurityAlert, 0)
	
	tm.mu.RLock()
	history := tm.transactionHistory[record.From]
	tm.mu.RUnlock()
	
	if len(history) < 5 {
		return alerts
	}
	
	// Calculate average transaction amount
	totalAmount := 0.0
	for _, tx := range history {
		totalAmount += tx.Amount
	}
	avgAmount := totalAmount / float64(len(history))
	
	// Check if current transaction is significantly larger than average
	if record.Amount > avgAmount*10 {
		alert := SecurityAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Type:        AlertTypeAnomalousTransaction,
			Severity:    SeverityMedium,
			Title:       "Unusual Transaction Amount",
			Description: fmt.Sprintf("Transaction amount %.2f is %.1fx larger than historical average %.2f", record.Amount, record.Amount/avgAmount, avgAmount),
			Source:      tm.name,
			TriggerEvent: event,
			RiskScore:   (record.Amount / avgAmount) * 10,
			Confidence:  70.0,
			Status:      StatusOpen,
			Recommendations: []string{
				"Review transaction legitimacy",
				"Check account activity patterns",
				"Verify transaction authorization",
			},
			Tags:     []string{"unusual_amount", "deviation", "pattern"},
			CreatedBy: tm.name,
		}
		alerts = append(alerts, alert)
	}
	
	return alerts
}

func (tm *TransactionMonitor) checkCircularTransactions(record TransactionRecord, event SecurityEvent) []SecurityAlert {
	alerts := make([]SecurityAlert, 0)
	
	// Check for potential circular transaction patterns (A -> B -> A)
	tm.mu.RLock()
	fromHistory := tm.transactionHistory[record.From]
	toHistory := tm.transactionHistory[record.To]
	tm.mu.RUnlock()
	
	// Look for recent transactions from 'to' back to 'from'
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	
	for _, tx := range toHistory {
		if tx.To == record.From && tx.Timestamp.After(oneHourAgo) && tx.Amount == record.Amount {
			alert := SecurityAlert{
				ID:          generateAlertID(),
				Timestamp:   time.Now(),
				Type:        AlertTypeSuspiciousActivity,
				Severity:    SeverityHigh,
				Title:       "Circular Transaction Pattern",
				Description: fmt.Sprintf("Detected circular transaction pattern between %s and %s with amount %.2f", record.From, record.To, record.Amount),
				Source:      tm.name,
				TriggerEvent: event,
				RiskScore:   80.0,
				Confidence:  90.0,
				Status:      StatusOpen,
				Evidence: []AlertEvidence{
					{
						Type:        "circular_pattern",
						Description: "Circular transaction detected",
						Data: map[string]string{
							"from":        record.From,
							"to":          record.To,
							"amount":      fmt.Sprintf("%.2f", record.Amount),
							"reverse_tx":  tx.ID,
						},
						Timestamp: time.Now(),
					},
				},
				Recommendations: []string{
					"Investigate potential money laundering",
					"Check for market manipulation",
					"Verify transaction purposes",
				},
				Tags:     []string{"circular", "money_laundering", "manipulation"},
				CreatedBy: tm.name,
			}
			alerts = append(alerts, alert)
			break
		}
	}
	
	return alerts
}

func (tm *TransactionMonitor) cleanupOldRecords(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.mu.Lock()
			cutoff := time.Now().Add(-24 * time.Hour)
			for account, history := range tm.transactionHistory {
				filtered := make([]TransactionRecord, 0)
				for _, record := range history {
					if record.Timestamp.After(cutoff) {
						filtered = append(filtered, record)
					}
				}
				tm.transactionHistory[account] = filtered
			}
			tm.mu.Unlock()
		}
	}
}

// AuthenticationMonitor monitors authentication-related security events
type AuthenticationMonitor struct {
	name            string
	description     string
	monitoredEvents []EventType
	isRunning       bool
	metrics         MonitorMetrics
	failedAttempts  map[string][]FailedAttempt
	mu              sync.RWMutex
}

type FailedAttempt struct {
	Actor     string    `json:"actor"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
}

func NewAuthenticationMonitor() *AuthenticationMonitor {
	return &AuthenticationMonitor{
		name:        "authentication_monitor",
		description: "Monitors authentication events for brute force and anomalous patterns",
		monitoredEvents: []EventType{EventTypeAuthentication},
		failedAttempts: make(map[string][]FailedAttempt),
		metrics: MonitorMetrics{},
	}
}

func (am *AuthenticationMonitor) GetName() string { return am.name }
func (am *AuthenticationMonitor) GetDescription() string { return am.description }
func (am *AuthenticationMonitor) GetMonitoredEvents() []EventType { return am.monitoredEvents }

func (am *AuthenticationMonitor) Configure(config map[string]interface{}) error {
	return nil
}

func (am *AuthenticationMonitor) Start(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.isRunning = true
	go am.cleanupFailedAttempts(ctx)
	return nil
}

func (am *AuthenticationMonitor) Stop() error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.isRunning = false
	return nil
}

func (am *AuthenticationMonitor) ProcessEvent(event SecurityEvent) []SecurityAlert {
	startTime := time.Now()
	alerts := make([]SecurityAlert, 0)
	
	defer func() {
		am.mu.Lock()
		am.metrics.EventsProcessed++
		am.metrics.ProcessingTime = time.Since(startTime)
		am.mu.Unlock()
	}()

	if event.Type != EventTypeAuthentication {
		return alerts
	}

	if event.Result == "failed" {
		am.recordFailedAttempt(event)
		alerts = append(alerts, am.checkBruteForce(event)...)
	}

	am.mu.Lock()
	am.metrics.AlertsGenerated += int64(len(alerts))
	am.mu.Unlock()

	return alerts
}

func (am *AuthenticationMonitor) GetMetrics() MonitorMetrics {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.metrics
}

func (am *AuthenticationMonitor) recordFailedAttempt(event SecurityEvent) {
	attempt := FailedAttempt{
		Actor:     event.Actor,
		IPAddress: event.Metadata.IPAddress,
		Timestamp: event.Timestamp,
		Reason:    event.Data["reason"],
	}

	am.mu.Lock()
	key := event.Actor + "_" + event.Metadata.IPAddress
	if am.failedAttempts[key] == nil {
		am.failedAttempts[key] = make([]FailedAttempt, 0)
	}
	am.failedAttempts[key] = append(am.failedAttempts[key], attempt)
	am.mu.Unlock()
}

func (am *AuthenticationMonitor) checkBruteForce(event SecurityEvent) []SecurityAlert {
	alerts := make([]SecurityAlert, 0)

	am.mu.RLock()
	key := event.Actor + "_" + event.Metadata.IPAddress
	attempts := am.failedAttempts[key]
	am.mu.RUnlock()

	// Count recent failed attempts (last 15 minutes)
	recentCount := 0
	fifteenMinutesAgo := time.Now().Add(-15 * time.Minute)

	for _, attempt := range attempts {
		if attempt.Timestamp.After(fifteenMinutesAgo) {
			recentCount++
		}
	}

	// Generate alert if threshold exceeded
	if recentCount >= 5 {
		alert := SecurityAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Type:        AlertTypeSuspiciousActivity,
			Severity:    SeverityHigh,
			Title:       "Potential Brute Force Attack",
			Description: fmt.Sprintf("Detected %d failed login attempts for user %s from IP %s in the last 15 minutes", recentCount, event.Actor, event.Metadata.IPAddress),
			Source:      am.name,
			TriggerEvent: event,
			RiskScore:   float64(recentCount) * 15.0,
			Confidence:  90.0,
			Status:      StatusOpen,
			Evidence: []AlertEvidence{
				{
					Type:        "failed_attempts",
					Description: "Multiple failed authentication attempts",
					Data: map[string]string{
						"attempt_count": fmt.Sprintf("%d", recentCount),
						"time_window":   "15 minutes",
						"user":         event.Actor,
						"ip_address":   event.Metadata.IPAddress,
					},
					Timestamp: time.Now(),
				},
			},
			Recommendations: []string{
				"Block suspicious IP address",
				"Implement account lockout",
				"Notify user of suspicious activity",
				"Review authentication logs",
			},
			Tags:     []string{"brute_force", "authentication", "failed_login"},
			CreatedBy: am.name,
		}
		alerts = append(alerts, alert)
	}

	return alerts
}

func (am *AuthenticationMonitor) cleanupFailedAttempts(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			am.mu.Lock()
			cutoff := time.Now().Add(-1 * time.Hour)
			for key, attempts := range am.failedAttempts {
				filtered := make([]FailedAttempt, 0)
				for _, attempt := range attempts {
					if attempt.Timestamp.After(cutoff) {
						filtered = append(filtered, attempt)
					}
				}
				if len(filtered) == 0 {
					delete(am.failedAttempts, key)
				} else {
					am.failedAttempts[key] = filtered
				}
			}
			am.mu.Unlock()
		}
	}
}

// GovernanceMonitor monitors governance-related security events
type GovernanceMonitor struct {
	name            string
	description     string
	monitoredEvents []EventType
	isRunning       bool
	metrics         MonitorMetrics
}

func NewGovernanceMonitor() *GovernanceMonitor {
	return &GovernanceMonitor{
		name:        "governance_monitor",
		description: "Monitors governance events for unauthorized actions and manipulation",
		monitoredEvents: []EventType{EventTypeGovernance},
		metrics:     MonitorMetrics{},
	}
}

func (gm *GovernanceMonitor) GetName() string { return gm.name }
func (gm *GovernanceMonitor) GetDescription() string { return gm.description }
func (gm *GovernanceMonitor) GetMonitoredEvents() []EventType { return gm.monitoredEvents }

func (gm *GovernanceMonitor) Configure(config map[string]interface{}) error { return nil }
func (gm *GovernanceMonitor) Start(ctx context.Context) error {
	gm.isRunning = true
	return nil
}
func (gm *GovernanceMonitor) Stop() error {
	gm.isRunning = false
	return nil
}

func (gm *GovernanceMonitor) ProcessEvent(event SecurityEvent) []SecurityAlert {
	startTime := time.Now()
	alerts := make([]SecurityAlert, 0)
	
	defer func() {
		gm.metrics.EventsProcessed++
		gm.metrics.ProcessingTime = time.Since(startTime)
	}()

	if event.Type != EventTypeGovernance {
		return alerts
	}

	// Check for unauthorized governance actions
	if event.Data["authorized"] == "false" {
		alert := SecurityAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Type:        AlertTypeUnauthorizedAccess,
			Severity:    SeverityCritical,
			Title:       "Unauthorized Governance Action",
			Description: fmt.Sprintf("Unauthorized governance action attempted by %s: %s", event.Actor, event.Action),
			Source:      gm.name,
			TriggerEvent: event,
			RiskScore:   95.0,
			Confidence:  95.0,
			Status:      StatusOpen,
			Recommendations: []string{
				"Investigate attempted unauthorized access",
				"Review governance permissions",
				"Check for account compromise",
			},
			Tags:     []string{"governance", "unauthorized", "critical"},
			CreatedBy: gm.name,
		}
		alerts = append(alerts, alert)
	}

	gm.metrics.AlertsGenerated += int64(len(alerts))
	return alerts
}

func (gm *GovernanceMonitor) GetMetrics() MonitorMetrics {
	return gm.metrics
}

// Additional monitor implementations would follow similar patterns...

// AnomalyDetectionMonitor uses statistical analysis to detect anomalies
type AnomalyDetectionMonitor struct {
	name            string
	description     string
	monitoredEvents []EventType
	isRunning       bool
	metrics         MonitorMetrics
}

func NewAnomalyDetectionMonitor() *AnomalyDetectionMonitor {
	return &AnomalyDetectionMonitor{
		name:        "anomaly_detection_monitor",
		description: "Uses statistical analysis to detect anomalous behavior patterns",
		monitoredEvents: []EventType{EventTypeTransaction, EventTypeTrade, EventTypeNetworkActivity},
		metrics:     MonitorMetrics{},
	}
}

func (adm *AnomalyDetectionMonitor) GetName() string { return adm.name }
func (adm *AnomalyDetectionMonitor) GetDescription() string { return adm.description }
func (adm *AnomalyDetectionMonitor) GetMonitoredEvents() []EventType { return adm.monitoredEvents }
func (adm *AnomalyDetectionMonitor) Configure(config map[string]interface{}) error { return nil }
func (adm *AnomalyDetectionMonitor) Start(ctx context.Context) error {
	adm.isRunning = true
	return nil
}
func (adm *AnomalyDetectionMonitor) Stop() error {
	adm.isRunning = false
	return nil
}
func (adm *AnomalyDetectionMonitor) ProcessEvent(event SecurityEvent) []SecurityAlert {
	// Simplified anomaly detection - would use ML models in real implementation
	return []SecurityAlert{}
}
func (adm *AnomalyDetectionMonitor) GetMetrics() MonitorMetrics {
	return adm.metrics
}

// NetworkSecurityMonitor monitors network-related security events
type NetworkSecurityMonitor struct {
	name            string
	description     string
	monitoredEvents []EventType
	isRunning       bool
	metrics         MonitorMetrics
}

func NewNetworkSecurityMonitor() *NetworkSecurityMonitor {
	return &NetworkSecurityMonitor{
		name:        "network_security_monitor",
		description: "Monitors network activity for security threats and anomalies",
		monitoredEvents: []EventType{EventTypeNetworkActivity},
		metrics:     MonitorMetrics{},
	}
}

func (nsm *NetworkSecurityMonitor) GetName() string { return nsm.name }
func (nsm *NetworkSecurityMonitor) GetDescription() string { return nsm.description }
func (nsm *NetworkSecurityMonitor) GetMonitoredEvents() []EventType { return nsm.monitoredEvents }
func (nsm *NetworkSecurityMonitor) Configure(config map[string]interface{}) error { return nil }
func (nsm *NetworkSecurityMonitor) Start(ctx context.Context) error {
	nsm.isRunning = true
	return nil
}
func (nsm *NetworkSecurityMonitor) Stop() error {
	nsm.isRunning = false
	return nil
}
func (nsm *NetworkSecurityMonitor) ProcessEvent(event SecurityEvent) []SecurityAlert {
	// Network security monitoring implementation
	return []SecurityAlert{}
}
func (nsm *NetworkSecurityMonitor) GetMetrics() MonitorMetrics {
	return nsm.metrics
}

// Mock implementations for interfaces

type MockAlertManager struct{}

func NewMockAlertManager() *MockAlertManager {
	return &MockAlertManager{}
}

func (mam *MockAlertManager) SendAlert(alert SecurityAlert) error {
	fmt.Printf("ALERT [%s]: %s - %s\n", alert.Severity, alert.Title, alert.Description)
	return nil
}

func (mam *MockAlertManager) EscalateAlert(alertID string) error {
	fmt.Printf("Escalating alert: %s\n", alertID)
	return nil
}

func (mam *MockAlertManager) ResolveAlert(alertID string, resolution string) error {
	fmt.Printf("Resolving alert %s: %s\n", alertID, resolution)
	return nil
}

func (mam *MockAlertManager) SuppressAlert(alertID string, reason string) error {
	fmt.Printf("Suppressing alert %s: %s\n", alertID, reason)
	return nil
}

func (mam *MockAlertManager) GetAlerts(criteria AlertCriteria) ([]SecurityAlert, error) {
	return []SecurityAlert{}, nil
}

type MockRuleEngine struct {
	rules []SecurityRule
	mu    sync.RWMutex
}

func NewMockRuleEngine() *MockRuleEngine {
	return &MockRuleEngine{
		rules: make([]SecurityRule, 0),
	}
}

func (mre *MockRuleEngine) AddRule(rule SecurityRule) error {
	mre.mu.Lock()
	defer mre.mu.Unlock()
	mre.rules = append(mre.rules, rule)
	return nil
}

func (mre *MockRuleEngine) RemoveRule(ruleID string) error {
	mre.mu.Lock()
	defer mre.mu.Unlock()
	
	for i, rule := range mre.rules {
		if rule.ID == ruleID {
			mre.rules = append(mre.rules[:i], mre.rules[i+1:]...)
			break
		}
	}
	return nil
}

func (mre *MockRuleEngine) EvaluateEvent(event SecurityEvent) []SecurityAlert {
	alerts := make([]SecurityAlert, 0)
	
	mre.mu.RLock()
	defer mre.mu.RUnlock()
	
	for _, rule := range mre.rules {
		if !rule.Enabled {
			continue
		}
		
		if mre.evaluateCondition(rule.Condition, event) {
			alert := SecurityAlert{
				ID:          generateAlertID(),
				Timestamp:   time.Now(),
				Type:        AlertTypePolicyViolation,
				Severity:    rule.Severity,
				Title:       rule.Name,
				Description: rule.Description,
				Source:      "rule_engine",
				TriggerEvent: event,
				RiskScore:   50.0,
				Confidence:  80.0,
				Status:      StatusOpen,
				Tags:        rule.Tags,
				CreatedBy:   "rule_engine",
			}
			alerts = append(alerts, alert)
		}
	}
	
	return alerts
}

func (mre *MockRuleEngine) GetRules() []SecurityRule {
	mre.mu.RLock()
	defer mre.mu.RUnlock()
	return append([]SecurityRule{}, mre.rules...)
}

func (mre *MockRuleEngine) evaluateCondition(condition string, event SecurityEvent) bool {
	// Simplified condition evaluation
	// In a real implementation, this would parse and evaluate complex logical expressions
	
	if strings.Contains(condition, "event.type = 'transaction'") && event.Type == EventTypeTransaction {
		if strings.Contains(condition, "event.data.amount > 1000000") {
			if amountStr, exists := event.Data["amount"]; exists {
				if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
					return amount > 1000000
				}
			}
		}
		return true
	}
	
	if strings.Contains(condition, "event.type = 'authentication'") && event.Type == EventTypeAuthentication {
		if strings.Contains(condition, "event.result = 'failed'") {
			return event.Result == "failed"
		}
		return true
	}
	
	if strings.Contains(condition, "event.type = 'governance'") && event.Type == EventTypeGovernance {
		if strings.Contains(condition, "event.data.authorized = 'false'") {
			return event.Data["authorized"] == "false"
		}
		return true
	}
	
	return false
}

// Helper functions

func calculateRiskScore(value, threshold float64) float64 {
	ratio := value / threshold
	return math.Min(100.0, ratio * 50.0)
}