package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values - ALL governance-controllable via validator voting
const (
	DefaultOrderBookDepth       uint64 = 50       // Return top 50 levels in aggregated order book
	DefaultBaseFee              uint64 = 1000     // 0.001 HODL base transaction fee
	DefaultFeePerByte           uint64 = 10       // 10 uhodl per byte
	DefaultMinFee               uint64 = 500      // 0.0005 HODL minimum fee
	DefaultMaxFee               uint64 = 10000000 // 10 HODL maximum fee
	DefaultValidatorFeeShare    uint64 = 70       // 70% of fees to validators

	// Trader tier thresholds (in HODL)
	DefaultDiamondTierThreshold  uint64 = 10000000 // 10M HODL volume
	DefaultPlatinumTierThreshold uint64 = 1000000  // 1M HODL volume
	DefaultGoldTierThreshold     uint64 = 100000   // 100K HODL volume
	DefaultSilverTierThreshold   uint64 = 10000    // 10K HODL volume

	// Trading guardrails defaults
	DefaultMaxOrdersPerBlock uint32 = 100
)

var (
	// Fee percentages (basis points where 1000 = 10%)
	DefaultAtomicSwapFee    = math.LegacyNewDecWithPrec(1, 3)  // 0.1%
	DefaultEquitySwapFee    = math.LegacyNewDecWithPrec(3, 3)  // 0.3%
	DefaultAutoSwapSlippage = math.LegacyNewDecWithPrec(5, 2)  // 5% max slippage
	DefaultTradingFee       = math.LegacyNewDecWithPrec(3, 3)  // 0.3% trading fee
	DefaultMakerFee         = math.LegacyNewDecWithPrec(1, 3)  // 0.1% maker fee
	DefaultTakerFee         = math.LegacyNewDecWithPrec(2, 3)  // 0.2% taker fee

	// Trading guardrails
	DefaultMaxOrderSize         = math.LegacyNewDec(1000000)        // 1M HODL max order
	DefaultMaxPositionSize      = math.LegacyNewDec(10000000)       // 10M HODL max position
	DefaultMaxDailyVolume       = math.LegacyNewDec(50000000)       // 50M HODL daily limit
	DefaultPriceCircuitBreaker  = math.LegacyNewDecWithPrec(10, 2)  // 10% price change triggers halt
	DefaultVolumeCircuitBreaker = math.LegacyNewDec(10)             // 10x normal volume triggers halt
	DefaultMinimumOrderValue    = math.LegacyNewDec(1)              // 1 HODL minimum order
)

// Params defines the parameters for the DEX module
// ALL parameters are governance-controllable via validator voting
type Params struct {
	// Order book configuration
	OrderBookDepth uint64 `json:"order_book_depth" yaml:"order_book_depth"` // Max levels to return

	// Transaction fees (in uhodl)
	BaseFee    uint64 `json:"base_fee" yaml:"base_fee"`         // Base transaction fee
	FeePerByte uint64 `json:"fee_per_byte" yaml:"fee_per_byte"` // Fee per byte of tx
	MinFee     uint64 `json:"min_fee" yaml:"min_fee"`           // Minimum transaction fee
	MaxFee     uint64 `json:"max_fee" yaml:"max_fee"`           // Maximum transaction fee

	// Trading fees (percentages)
	TradingFee       math.LegacyDec `json:"trading_fee" yaml:"trading_fee"`               // Standard trading fee
	AtomicSwapFee    math.LegacyDec `json:"atomic_swap_fee" yaml:"atomic_swap_fee"`       // Atomic swap fee
	EquitySwapFee    math.LegacyDec `json:"equity_swap_fee" yaml:"equity_swap_fee"`       // Equity-to-HODL swap fee
	AutoSwapSlippage math.LegacyDec `json:"auto_swap_slippage" yaml:"auto_swap_slippage"` // Max slippage for auto-swap
	MakerFee         math.LegacyDec `json:"maker_fee" yaml:"maker_fee"`                   // Maker (liquidity provider) fee
	TakerFee         math.LegacyDec `json:"taker_fee" yaml:"taker_fee"`                   // Taker (liquidity consumer) fee

	// Fee distribution
	ValidatorFeeShare uint64 `json:"validator_fee_share" yaml:"validator_fee_share"` // % of fees to validators

	// Trading controls
	TradingEnabled bool `json:"trading_enabled" yaml:"trading_enabled"` // Master toggle for trading

	// Trader tier volume thresholds (in HODL)
	DiamondTierThreshold  uint64 `json:"diamond_tier_threshold" yaml:"diamond_tier_threshold"`   // Volume for diamond tier
	PlatinumTierThreshold uint64 `json:"platinum_tier_threshold" yaml:"platinum_tier_threshold"` // Volume for platinum tier
	GoldTierThreshold     uint64 `json:"gold_tier_threshold" yaml:"gold_tier_threshold"`         // Volume for gold tier
	SilverTierThreshold   uint64 `json:"silver_tier_threshold" yaml:"silver_tier_threshold"`     // Volume for silver tier

	// Trading guardrails (governance-controllable)
	MaxOrderSize         math.LegacyDec `json:"max_order_size" yaml:"max_order_size"`                   // Maximum order size
	MaxPositionSize      math.LegacyDec `json:"max_position_size" yaml:"max_position_size"`             // Maximum position per trader
	MaxDailyVolume       math.LegacyDec `json:"max_daily_volume" yaml:"max_daily_volume"`               // Maximum daily volume per trader
	PriceCircuitBreaker  math.LegacyDec `json:"price_circuit_breaker" yaml:"price_circuit_breaker"`     // % change to trigger halt
	VolumeCircuitBreaker math.LegacyDec `json:"volume_circuit_breaker" yaml:"volume_circuit_breaker"`   // Volume spike to trigger halt
	MinimumOrderValue    math.LegacyDec `json:"minimum_order_value" yaml:"minimum_order_value"`         // Minimum order value
	MaxOrdersPerBlock    uint32         `json:"max_orders_per_block" yaml:"max_orders_per_block"`       // Max orders per block per user
}

// ProtoMessage implements proto.Message interface
func (m *Params) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *Params) Reset() { *m = Params{} }

// String implements proto.Message interface
func (m *Params) String() string { return "dex_params" }

// DefaultParams returns default DEX parameters
// ALL values can be changed via governance proposals
func DefaultParams() Params {
	return Params{
		OrderBookDepth:        DefaultOrderBookDepth,
		BaseFee:               DefaultBaseFee,
		FeePerByte:            DefaultFeePerByte,
		MinFee:                DefaultMinFee,
		MaxFee:                DefaultMaxFee,
		TradingFee:            DefaultTradingFee,
		AtomicSwapFee:         DefaultAtomicSwapFee,
		EquitySwapFee:         DefaultEquitySwapFee,
		AutoSwapSlippage:      DefaultAutoSwapSlippage,
		MakerFee:              DefaultMakerFee,
		TakerFee:              DefaultTakerFee,
		ValidatorFeeShare:     DefaultValidatorFeeShare,
		TradingEnabled:        true,
		DiamondTierThreshold:  DefaultDiamondTierThreshold,
		PlatinumTierThreshold: DefaultPlatinumTierThreshold,
		GoldTierThreshold:     DefaultGoldTierThreshold,
		SilverTierThreshold:   DefaultSilverTierThreshold,
		MaxOrderSize:          DefaultMaxOrderSize,
		MaxPositionSize:       DefaultMaxPositionSize,
		MaxDailyVolume:        DefaultMaxDailyVolume,
		PriceCircuitBreaker:   DefaultPriceCircuitBreaker,
		VolumeCircuitBreaker:  DefaultVolumeCircuitBreaker,
		MinimumOrderValue:     DefaultMinimumOrderValue,
		MaxOrdersPerBlock:     DefaultMaxOrdersPerBlock,
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.OrderBookDepth == 0 {
		return fmt.Errorf("order book depth must be positive")
	}
	if p.MinFee > p.MaxFee {
		return fmt.Errorf("min fee cannot exceed max fee")
	}
	if !p.TradingFee.IsNil() && (p.TradingFee.IsNegative() || p.TradingFee.GT(math.LegacyOneDec())) {
		return fmt.Errorf("trading fee must be between 0 and 100%%")
	}
	if !p.AtomicSwapFee.IsNil() && (p.AtomicSwapFee.IsNegative() || p.AtomicSwapFee.GT(math.LegacyOneDec())) {
		return fmt.Errorf("atomic swap fee must be between 0 and 100%%")
	}
	if !p.EquitySwapFee.IsNil() && (p.EquitySwapFee.IsNegative() || p.EquitySwapFee.GT(math.LegacyOneDec())) {
		return fmt.Errorf("equity swap fee must be between 0 and 100%%")
	}
	if !p.AutoSwapSlippage.IsNil() && (p.AutoSwapSlippage.IsNegative() || p.AutoSwapSlippage.GT(math.LegacyOneDec())) {
		return fmt.Errorf("auto swap slippage must be between 0 and 100%%")
	}
	if !p.MakerFee.IsNil() && (p.MakerFee.IsNegative() || p.MakerFee.GT(math.LegacyOneDec())) {
		return fmt.Errorf("maker fee must be between 0 and 100%%")
	}
	if !p.TakerFee.IsNil() && (p.TakerFee.IsNegative() || p.TakerFee.GT(math.LegacyOneDec())) {
		return fmt.Errorf("taker fee must be between 0 and 100%%")
	}
	if p.ValidatorFeeShare > 100 {
		return fmt.Errorf("validator fee share cannot exceed 100%%")
	}
	// Tier thresholds should be in descending order
	if p.DiamondTierThreshold <= p.PlatinumTierThreshold {
		return fmt.Errorf("diamond tier threshold must exceed platinum")
	}
	if p.PlatinumTierThreshold <= p.GoldTierThreshold {
		return fmt.Errorf("platinum tier threshold must exceed gold")
	}
	if p.GoldTierThreshold <= p.SilverTierThreshold {
		return fmt.Errorf("gold tier threshold must exceed silver")
	}
	// Guardrails validation
	if !p.MaxOrderSize.IsNil() && p.MaxOrderSize.IsNegative() {
		return fmt.Errorf("max order size cannot be negative")
	}
	if !p.MinimumOrderValue.IsNil() && p.MinimumOrderValue.IsNegative() {
		return fmt.Errorf("minimum order value cannot be negative")
	}
	return nil
}

// Note: ParamsKey is defined in key.go to avoid duplication
