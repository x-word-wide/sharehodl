package types

import (
	"time"

	"cosmossdk.io/math"
)

// ModuleName is the name of the agent module
const ModuleName = "agent"

// StoreKey is the store key for the agent module
const StoreKey = ModuleName

// Agent represents an AI agent authorized to act on behalf of a user
type Agent struct {
	ID                uint64         `json:"id"`
	Owner             string         `json:"owner"`              // User who owns this agent
	AgentKey          string         `json:"agent_key"`          // Public key or API key for authentication
	Name              string         `json:"name"`               // Human-readable agent name
	Description       string         `json:"description"`        // Agent purpose description
	Permissions       []Permission   `json:"permissions"`        // What the agent can do
	SpendingLimits    SpendingLimits `json:"spending_limits"`    // Financial constraints
	RateLimits        RateLimits     `json:"rate_limits"`        // Action frequency limits
	IsActive          bool           `json:"is_active"`          // Whether agent is currently active
	CreatedAt         time.Time      `json:"created_at"`
	LastActiveAt      time.Time      `json:"last_active_at"`
	TotalActions      uint64         `json:"total_actions"`      // Total actions performed
	SubscriptionTier  string         `json:"subscription_tier"`  // basic, pro, enterprise
	ExpiresAt         time.Time      `json:"expires_at"`         // Subscription expiry
}

// Permission defines what actions an agent can perform
type Permission struct {
	Module    string   `json:"module"`    // Target module (hodl, dex, escrow, etc.)
	Actions   []string `json:"actions"`   // Allowed actions (mint, trade, etc.)
	MaxAmount math.Int `json:"max_amount"` // Max amount per action
}

// SpendingLimits defines financial constraints for agents
type SpendingLimits struct {
	DailyLimit     math.Int `json:"daily_limit"`      // Max HODL spend per day
	MonthlyLimit   math.Int `json:"monthly_limit"`    // Max HODL spend per month
	PerTxLimit     math.Int `json:"per_tx_limit"`     // Max HODL per transaction
	TotalSpentDay  math.Int `json:"total_spent_day"`  // Current day spend
	TotalSpentMonth math.Int `json:"total_spent_month"` // Current month spend
	ResetDay       time.Time `json:"reset_day"`         // When daily limit resets
	ResetMonth     time.Time `json:"reset_month"`       // When monthly limit resets
}

// RateLimits defines action frequency constraints
type RateLimits struct {
	MaxActionsPerMinute uint32 `json:"max_actions_per_minute"`
	MaxActionsPerHour   uint32 `json:"max_actions_per_hour"`
	MaxActionsPerDay    uint32 `json:"max_actions_per_day"`
	ActionsThisMinute   uint32 `json:"actions_this_minute"`
	ActionsThisHour     uint32 `json:"actions_this_hour"`
	ActionsThisDay      uint32 `json:"actions_this_day"`
	LastMinuteReset     time.Time `json:"last_minute_reset"`
	LastHourReset       time.Time `json:"last_hour_reset"`
	LastDayReset        time.Time `json:"last_day_reset"`
}

// AgentAction represents an action taken by an agent
type AgentAction struct {
	ID          uint64    `json:"id"`
	AgentID     uint64    `json:"agent_id"`
	Owner       string    `json:"owner"`
	Module      string    `json:"module"`
	Action      string    `json:"action"`
	Params      string    `json:"params"`        // JSON-encoded action parameters
	Amount      math.Int  `json:"amount"`        // HODL amount if applicable
	Status      string    `json:"status"`        // pending, executed, failed, rejected
	Result      string    `json:"result"`        // JSON-encoded result or error
	GasFee      math.Int  `json:"gas_fee"`
	Timestamp   time.Time `json:"timestamp"`
	BlockHeight int64     `json:"block_height"`
}

// SubscriptionTier defines subscription tier features
type SubscriptionTier struct {
	Name               string   `json:"name"`
	MaxAgents          uint32   `json:"max_agents"`
	MaxActionsPerDay   uint32   `json:"max_actions_per_day"`
	MaxDailySpend      math.Int `json:"max_daily_spend"`
	AllowedModules     []string `json:"allowed_modules"`
	PriorityExecution  bool     `json:"priority_execution"`
	AdvancedAnalytics  bool     `json:"advanced_analytics"`
	CustomWebhooks     bool     `json:"custom_webhooks"`
	MonthlyFeeHODL     math.Int `json:"monthly_fee_hodl"`
}

// NewAgent creates a new agent
func NewAgent(id uint64, owner, agentKey, name, description string, tier string, expiry time.Time) Agent {
	return Agent{
		ID:               id,
		Owner:            owner,
		AgentKey:         agentKey,
		Name:             name,
		Description:      description,
		Permissions:      []Permission{},
		SpendingLimits:   DefaultSpendingLimits(),
		RateLimits:       DefaultRateLimits(),
		IsActive:         true,
		CreatedAt:        time.Now(),
		LastActiveAt:     time.Now(),
		TotalActions:     0,
		SubscriptionTier: tier,
		ExpiresAt:        expiry,
	}
}

// DefaultSpendingLimits returns default spending limits
func DefaultSpendingLimits() SpendingLimits {
	return SpendingLimits{
		DailyLimit:      math.NewInt(1000000000),  // 1000 HODL
		MonthlyLimit:    math.NewInt(10000000000), // 10000 HODL
		PerTxLimit:      math.NewInt(100000000),   // 100 HODL
		TotalSpentDay:   math.ZeroInt(),
		TotalSpentMonth: math.ZeroInt(),
		ResetDay:        time.Now(),
		ResetMonth:      time.Now(),
	}
}

// DefaultRateLimits returns default rate limits
func DefaultRateLimits() RateLimits {
	return RateLimits{
		MaxActionsPerMinute: 10,
		MaxActionsPerHour:   100,
		MaxActionsPerDay:    1000,
		ActionsThisMinute:   0,
		ActionsThisHour:     0,
		ActionsThisDay:      0,
		LastMinuteReset:     time.Now(),
		LastHourReset:       time.Now(),
		LastDayReset:        time.Now(),
	}
}

// Validate validates the agent
func (a Agent) Validate() error {
	if a.Owner == "" {
		return ErrInvalidOwner
	}
	if a.AgentKey == "" {
		return ErrInvalidAgentKey
	}
	return nil
}

// HasPermission checks if agent has permission for an action
func (a Agent) HasPermission(module, action string) bool {
	for _, perm := range a.Permissions {
		if perm.Module == module {
			for _, act := range perm.Actions {
				if act == action || act == "*" {
					return true
				}
			}
		}
	}
	return false
}

// CanSpend checks if agent can spend the given amount
func (a *Agent) CanSpend(amount math.Int) bool {
	if amount.GT(a.SpendingLimits.PerTxLimit) {
		return false
	}
	if a.SpendingLimits.TotalSpentDay.Add(amount).GT(a.SpendingLimits.DailyLimit) {
		return false
	}
	if a.SpendingLimits.TotalSpentMonth.Add(amount).GT(a.SpendingLimits.MonthlyLimit) {
		return false
	}
	return true
}

// CanExecute checks if agent can execute based on rate limits
func (a *Agent) CanExecute() bool {
	if a.RateLimits.ActionsThisMinute >= a.RateLimits.MaxActionsPerMinute {
		return false
	}
	if a.RateLimits.ActionsThisHour >= a.RateLimits.MaxActionsPerHour {
		return false
	}
	if a.RateLimits.ActionsThisDay >= a.RateLimits.MaxActionsPerDay {
		return false
	}
	return true
}

// DefaultSubscriptionTiers returns the default subscription tiers
func DefaultSubscriptionTiers() []SubscriptionTier {
	return []SubscriptionTier{
		{
			Name:               "basic",
			MaxAgents:          1,
			MaxActionsPerDay:   100,
			MaxDailySpend:      math.NewInt(1000000000),  // 1000 HODL
			AllowedModules:     []string{"dex"},
			PriorityExecution:  false,
			AdvancedAnalytics:  false,
			CustomWebhooks:     false,
			MonthlyFeeHODL:     math.NewInt(10000000), // 10 HODL
		},
		{
			Name:               "pro",
			MaxAgents:          5,
			MaxActionsPerDay:   1000,
			MaxDailySpend:      math.NewInt(10000000000), // 10000 HODL
			AllowedModules:     []string{"dex", "hodl", "escrow"},
			PriorityExecution:  true,
			AdvancedAnalytics:  true,
			CustomWebhooks:     false,
			MonthlyFeeHODL:     math.NewInt(100000000), // 100 HODL
		},
		{
			Name:               "enterprise",
			MaxAgents:          100,
			MaxActionsPerDay:   100000,
			MaxDailySpend:      math.NewInt(100000000000), // 100000 HODL
			AllowedModules:     []string{"dex", "hodl", "escrow", "lending", "governance"},
			PriorityExecution:  true,
			AdvancedAnalytics:  true,
			CustomWebhooks:     true,
			MonthlyFeeHODL:     math.NewInt(1000000000), // 1000 HODL
		},
	}
}

// GetSubscriptionTier returns the subscription tier by name
func GetSubscriptionTier(name string) (SubscriptionTier, bool) {
	for _, tier := range DefaultSubscriptionTiers() {
		if tier.Name == name {
			return tier, true
		}
	}
	return SubscriptionTier{}, false
}
