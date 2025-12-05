package types

import (
	"time"

	"cosmossdk.io/math"
)

// TradingControls defines comprehensive trading safeguards
type TradingControls struct {
	// Circuit Breaker Configuration
	CircuitBreakerEnabled   bool               `json:"circuit_breaker_enabled"`
	PriceMovementThreshold  math.LegacyDec     `json:"price_movement_threshold"`    // e.g., 10% = 0.10
	VolumeSpikeTolerance    math.LegacyDec     `json:"volume_spike_tolerance"`      // e.g., 500% = 5.0
	HaltDuration           time.Duration       `json:"halt_duration"`               // e.g., 15 minutes
	CooldownPeriod         time.Duration       `json:"cooldown_period"`             // e.g., 1 hour
	
	// Order Size Limits (% of outstanding shares)
	MaxOrderSizePercent     math.LegacyDec     `json:"max_order_size_percent"`      // e.g., 5% = 0.05
	MaxDailyVolumePercent   math.LegacyDec     `json:"max_daily_volume_percent"`    // e.g., 20% = 0.20
	
	// Time-based Controls
	TradingHours           TradingHours        `json:"trading_hours"`
	SettlementPeriod       time.Duration       `json:"settlement_period"`           // e.g., T+0 (instant)
	
	// Anti-Manipulation
	MinimumHoldPeriod      time.Duration       `json:"minimum_hold_period"`         // e.g., 0 for liquid trading
	WashTradeDetection     bool                `json:"wash_trade_detection"`
	FrontRunningProtection bool                `json:"front_running_protection"`
	
	LastUpdated            time.Time           `json:"last_updated"`
}

// TradingHours defines when trading is allowed
type TradingHours struct {
	Enabled       bool      `json:"enabled"`         // If false, trading is 24/7
	OpenTime      string    `json:"open_time"`       // e.g., "09:30" (UTC)
	CloseTime     string    `json:"close_time"`      // e.g., "16:00" (UTC) 
	Timezone      string    `json:"timezone"`        // e.g., "UTC"
	TradingDays   []int     `json:"trading_days"`    // 0=Sunday, 1=Monday, etc.
}

// CircuitBreakerStatus tracks circuit breaker state
type CircuitBreakerStatus struct {
	MarketSymbol     string             `json:"market_symbol"`
	Active           bool               `json:"active"`
	Triggered        time.Time          `json:"triggered"`
	TriggerReason    string             `json:"trigger_reason"`
	TriggerPrice     math.LegacyDec     `json:"trigger_price"`
	ResumeTime       time.Time          `json:"resume_time"`
	HaltCount        uint32             `json:"halt_count"`        // Number of halts today
	LastReset        time.Time          `json:"last_reset"`        // When halt count was reset
}

// OrderExecutionRule defines how orders should be executed
type OrderExecutionRule struct {
	TimeInForce           TimeInForce        `json:"time_in_force"`
	AllowPartialFill      bool               `json:"allow_partial_fill"`
	MinimumFillQuantity   math.Int           `json:"minimum_fill_quantity"`
	MaximumSlippage       math.LegacyDec     `json:"maximum_slippage"`     // e.g., 5% = 0.05
	RequireImmediateExec  bool               `json:"require_immediate_exec"`
	HiddenOrder           bool               `json:"hidden_order"`         // Iceberg order visibility
	PostOnly              bool               `json:"post_only"`            // Only add liquidity, don't take
}

// OwnershipLimits defines ownership concentration controls
type OwnershipLimits struct {
	Enabled                    bool               `json:"enabled"`
	MaxOwnershipPercent        math.LegacyDec     `json:"max_ownership_percent"`        // e.g., 49.9% = 0.499
	MaxDailyAcquisitionPercent math.LegacyDec     `json:"max_daily_acquisition_percent"` // e.g., 2% = 0.02
	RequireDisclosureThreshold math.LegacyDec     `json:"require_disclosure_threshold"`  // e.g., 5% = 0.05
	BoardApprovalThreshold     math.LegacyDec     `json:"board_approval_threshold"`      // e.g., 25% = 0.25
	CoolingOffPeriod          time.Duration       `json:"cooling_off_period"`            // e.g., 6 months
}

// TransferRestrictions defines share transfer limitations
type TransferRestrictions struct {
	Enabled                bool               `json:"enabled"`
	RequireApproval        bool               `json:"require_approval"`        // Board/admin approval required
	RestrictedInvestors    []string           `json:"restricted_investors"`    // Addresses that cannot trade
	QualifiedInvestorsOnly bool               `json:"qualified_investors_only"`
	RightOfFirstRefusal    bool               `json:"right_of_first_refusal"`  // Company can match offers
	TagAlongRights         bool               `json:"tag_along_rights"`        // Minority protection
	DragAlongRights        bool               `json:"drag_along_rights"`       // Majority can force sale
	LockupPeriods          []LockupPeriod     `json:"lockup_periods"`
}

// LockupPeriod defines time-based transfer restrictions
type LockupPeriod struct {
	ShareholderType  string        `json:"shareholder_type"`  // e.g., "founder", "employee", "investor"
	StartDate        time.Time     `json:"start_date"`
	EndDate          time.Time     `json:"end_date"`
	UnlockSchedule   string        `json:"unlock_schedule"`   // e.g., "cliff_then_monthly"
	CliffPeriod      time.Duration `json:"cliff_period"`      // e.g., 12 months
	VestingDuration  time.Duration `json:"vesting_duration"`  // e.g., 48 months
}

// VotingControls defines voting rights and governance protections
type VotingControls struct {
	Enabled                bool               `json:"enabled"`
	VotingCap              math.LegacyDec     `json:"voting_cap"`              // Max voting % per shareholder
	SuperMajorityThreshold math.LegacyDec     `json:"super_majority_threshold"` // e.g., 67% = 0.67
	QuorumRequirement      math.LegacyDec     `json:"quorum_requirement"`      // e.g., 50% = 0.50
	ProxyVotingAllowed     bool               `json:"proxy_voting_allowed"`
	VotingDeadline         time.Duration      `json:"voting_deadline"`         // e.g., 30 days
	BoardSeatLimitation    math.LegacyDec     `json:"board_seat_limitation"`   // Max % of board per shareholder
}

// TakeoverDefenses defines anti-takeover mechanisms
type TakeoverDefenses struct {
	Enabled              bool               `json:"enabled"`
	PoisonPillTrigger    math.LegacyDec     `json:"poison_pill_trigger"`    // e.g., 15% = 0.15
	StaggeredBoard       bool               `json:"staggered_board"`        // Directors serve different terms
	GoldenParachute      bool               `json:"golden_parachute"`       // Executive protection
	WhiteKnightRights    bool               `json:"white_knight_rights"`    // Preferred bidder rights
	FairPriceProvision   bool               `json:"fair_price_provision"`   // Require fair price for all
	BusinessCombination  time.Duration      `json:"business_combination"`   // Delay after acquisition
	ShareholderApproval  math.LegacyDec     `json:"shareholder_approval"`   // Required approval % for takeover
}

// MarketMakingParams defines market making controls
type MarketMakingParams struct {
	Enabled                bool               `json:"enabled"`
	MinSpreadBps           math.Int           `json:"min_spread_bps"`           // Minimum spread in basis points
	MaxSpreadBps           math.Int           `json:"max_spread_bps"`           // Maximum spread in basis points
	MinLiquidityDepth      math.Int           `json:"min_liquidity_depth"`      // Minimum order book depth
	MarketMakerRebate      math.LegacyDec     `json:"market_maker_rebate"`      // Rebate for providing liquidity
	MinOrderSize           math.Int           `json:"min_order_size"`           // Minimum market making order
	MaxPositionSize        math.Int           `json:"max_position_size"`        // Maximum MM position
	InventoryLimits        math.LegacyDec     `json:"inventory_limits"`         // Max inventory %
}

// Validation methods

// ValidateTradingControls validates trading control parameters
func (tc TradingControls) Validate() error {
	if tc.PriceMovementThreshold.IsNil() || tc.PriceMovementThreshold.LTE(math.LegacyZeroDec()) {
		return ErrInvalidParameter.Wrap("price movement threshold must be positive")
	}
	if tc.PriceMovementThreshold.GTE(math.LegacyNewDec(1)) {
		return ErrInvalidParameter.Wrap("price movement threshold cannot be >= 100%")
	}
	if tc.VolumeSpikeTolerance.IsNil() || tc.VolumeSpikeTolerance.LTE(math.LegacyOneDec()) {
		return ErrInvalidParameter.Wrap("volume spike tolerance must be > 100%")
	}
	if tc.MaxOrderSizePercent.IsNil() || tc.MaxOrderSizePercent.LTE(math.LegacyZeroDec()) {
		return ErrInvalidParameter.Wrap("max order size percent must be positive")
	}
	if tc.MaxOrderSizePercent.GTE(math.LegacyNewDec(1)) {
		return ErrInvalidParameter.Wrap("max order size percent cannot be >= 100%")
	}
	return nil
}

// ValidateOwnershipLimits validates ownership limit parameters
func (ol OwnershipLimits) Validate() error {
	if !ol.Enabled {
		return nil
	}
	if ol.MaxOwnershipPercent.IsNil() || ol.MaxOwnershipPercent.LTE(math.LegacyZeroDec()) {
		return ErrInvalidParameter.Wrap("max ownership percent must be positive")
	}
	if ol.MaxOwnershipPercent.GT(math.LegacyOneDec()) {
		return ErrInvalidParameter.Wrap("max ownership percent cannot exceed 100%")
	}
	if ol.BoardApprovalThreshold.IsNil() || ol.BoardApprovalThreshold.LTE(math.LegacyZeroDec()) {
		return ErrInvalidParameter.Wrap("board approval threshold must be positive")
	}
	if ol.BoardApprovalThreshold.GT(ol.MaxOwnershipPercent) {
		return ErrInvalidParameter.Wrap("board approval threshold cannot exceed max ownership")
	}
	return nil
}

// Helper functions

// IsWithinTradingHours checks if trading is allowed at the current time
func (th TradingHours) IsWithinTradingHours(now time.Time) bool {
	if !th.Enabled {
		return true // 24/7 trading
	}
	
	// Check day of week
	weekday := int(now.Weekday())
	dayAllowed := false
	for _, allowedDay := range th.TradingDays {
		if weekday == allowedDay {
			dayAllowed = true
			break
		}
	}
	
	if !dayAllowed {
		return false
	}
	
	// Check time of day (simplified - would need proper timezone handling)
	currentTime := now.Format("15:04")
	return currentTime >= th.OpenTime && currentTime <= th.CloseTime
}

// ShouldTriggerCircuitBreaker checks if conditions warrant a trading halt
func (tc TradingControls) ShouldTriggerCircuitBreaker(
	currentPrice, referencePrice math.LegacyDec,
	currentVolume, avgVolume math.Int,
) (bool, string) {
	if !tc.CircuitBreakerEnabled {
		return false, ""
	}
	
	// Check price movement
	if referencePrice.IsPositive() {
		priceChange := currentPrice.Sub(referencePrice).Quo(referencePrice).Abs()
		if priceChange.GT(tc.PriceMovementThreshold) {
			return true, "excessive price movement"
		}
	}
	
	// Check volume spike
	if avgVolume.IsPositive() {
		volumeRatio := math.LegacyNewDecFromInt(currentVolume).Quo(math.LegacyNewDecFromInt(avgVolume))
		if volumeRatio.GT(tc.VolumeSpikeTolerance) {
			return true, "volume spike detected"
		}
	}
	
	return false, ""
}

// CalculateOwnershipPercent calculates ownership percentage for validation
func CalculateOwnershipPercent(sharesOwned, totalShares math.Int) math.LegacyDec {
	if totalShares.IsZero() {
		return math.LegacyZeroDec()
	}
	return math.LegacyNewDecFromInt(sharesOwned).Quo(math.LegacyNewDecFromInt(totalShares))
}

// DefaultTradingControls returns default trading control parameters
func DefaultTradingControls() TradingControls {
	return TradingControls{
		CircuitBreakerEnabled:   true,
		PriceMovementThreshold:  math.LegacyNewDecWithPrec(15, 2),  // 15%
		VolumeSpikeTolerance:    math.LegacyNewDec(5),               // 500%
		HaltDuration:           time.Minute * 15,                   // 15 minutes
		CooldownPeriod:         time.Hour * 1,                     // 1 hour
		MaxOrderSizePercent:    math.LegacyNewDecWithPrec(5, 2),    // 5%
		MaxDailyVolumePercent:  math.LegacyNewDecWithPrec(20, 2),   // 20%
		TradingHours: TradingHours{
			Enabled:     false, // 24/7 trading by default
			OpenTime:    "00:00",
			CloseTime:   "23:59",
			Timezone:    "UTC",
			TradingDays: []int{0, 1, 2, 3, 4, 5, 6}, // All days
		},
		SettlementPeriod:       0,    // Instant settlement
		MinimumHoldPeriod:      0,    // No minimum hold
		WashTradeDetection:     true,
		FrontRunningProtection: true,
		LastUpdated:            time.Now(),
	}
}

// DefaultOwnershipLimits returns default ownership limits for public companies
func DefaultOwnershipLimits() OwnershipLimits {
	return OwnershipLimits{
		Enabled:                    true,
		MaxOwnershipPercent:        math.LegacyNewDecWithPrec(499, 3), // 49.9%
		MaxDailyAcquisitionPercent: math.LegacyNewDecWithPrec(2, 2),   // 2%
		RequireDisclosureThreshold: math.LegacyNewDecWithPrec(5, 2),   // 5%
		BoardApprovalThreshold:     math.LegacyNewDecWithPrec(25, 2),  // 25%
		CoolingOffPeriod:          time.Hour * 24 * 180,              // 6 months
	}
}

// DefaultTakeoverDefenses returns default anti-takeover protections
func DefaultTakeoverDefenses() TakeoverDefenses {
	return TakeoverDefenses{
		Enabled:              true,
		PoisonPillTrigger:    math.LegacyNewDecWithPrec(15, 2),  // 15%
		StaggeredBoard:       true,
		GoldenParachute:      true,
		WhiteKnightRights:    true,
		FairPriceProvision:   true,
		BusinessCombination:  time.Hour * 24 * 365, // 1 year
		ShareholderApproval:  math.LegacyNewDecWithPrec(67, 2),  // 67%
	}
}