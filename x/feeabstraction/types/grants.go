package types

import (
	"time"

	"cosmossdk.io/math"
)

// TreasuryFeeGrant represents a fee grant from protocol treasury
type TreasuryFeeGrant struct {
	// Grantee is the address that can use this grant
	Grantee string `json:"grantee" yaml:"grantee"`

	// AllowedFeeAmountUhodl is the maximum fee amount allowed
	AllowedFeeAmountUhodl math.Int `json:"allowed_fee_amount_uhodl" yaml:"allowed_fee_amount_uhodl"`

	// UsedFeeAmountUhodl is the amount already used
	UsedFeeAmountUhodl math.Int `json:"used_fee_amount_uhodl" yaml:"used_fee_amount_uhodl"`

	// ExpirationHeight is when the grant expires (0 = never)
	ExpirationHeight uint64 `json:"expiration_height" yaml:"expiration_height"`

	// AllowedEquitySymbols restricts which equities can be used (empty = all)
	AllowedEquitySymbols []string `json:"allowed_equity_symbols" yaml:"allowed_equity_symbols"`

	// GrantedAtHeight is when the grant was created
	GrantedAtHeight uint64 `json:"granted_at_height" yaml:"granted_at_height"`

	// GrantedByProposalID links to the governance proposal that created this grant
	GrantedByProposalID uint64 `json:"granted_by_proposal_id" yaml:"granted_by_proposal_id"`
}

// IsExpired checks if the grant has expired
func (g TreasuryFeeGrant) IsExpired(currentHeight uint64) bool {
	if g.ExpirationHeight == 0 {
		return false // Never expires
	}
	return currentHeight >= g.ExpirationHeight
}

// RemainingAllowance returns the remaining fee allowance
func (g TreasuryFeeGrant) RemainingAllowance() math.Int {
	if g.AllowedFeeAmountUhodl.IsNil() || g.UsedFeeAmountUhodl.IsNil() {
		return math.ZeroInt()
	}
	remaining := g.AllowedFeeAmountUhodl.Sub(g.UsedFeeAmountUhodl)
	if remaining.IsNegative() {
		return math.ZeroInt()
	}
	return remaining
}

// IsEquityAllowed checks if an equity symbol is allowed by this grant
func (g TreasuryFeeGrant) IsEquityAllowed(symbol string) bool {
	if len(g.AllowedEquitySymbols) == 0 {
		return true // Empty list means all equities allowed
	}
	for _, allowed := range g.AllowedEquitySymbols {
		if allowed == symbol {
			return true
		}
	}
	return false
}

// FeeAbstractionRecord stores an audit trail of fee abstraction usage
type FeeAbstractionRecord struct {
	// ID is the unique record identifier
	ID uint64 `json:"id" yaml:"id"`

	// TxHash is the transaction hash (if available)
	TxHash string `json:"tx_hash" yaml:"tx_hash"`

	// User is the address that used fee abstraction
	User string `json:"user" yaml:"user"`

	// EquitySymbol is the equity token that was swapped
	EquitySymbol string `json:"equity_symbol" yaml:"equity_symbol"`

	// EquitySwapped is the amount of equity swapped
	EquitySwapped math.Int `json:"equity_swapped" yaml:"equity_swapped"`

	// HODLReceived is the amount of HODL received from the swap
	HODLReceived math.Int `json:"hodl_received" yaml:"hodl_received"`

	// GasFeePaid is the actual gas fee paid
	GasFeePaid math.Int `json:"gas_fee_paid" yaml:"gas_fee_paid"`

	// MarkupPaid is the protocol markup fee
	MarkupPaid math.Int `json:"markup_paid" yaml:"markup_paid"`

	// TreasuryReturned is the amount returned to treasury
	TreasuryReturned math.Int `json:"treasury_returned" yaml:"treasury_returned"`

	// BlockHeight is the block height of the transaction
	BlockHeight uint64 `json:"block_height" yaml:"block_height"`

	// Timestamp is the time of the transaction
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}

// EligibleEquity represents an equity that can be used for fee abstraction
type EligibleEquity struct {
	// Symbol is the equity symbol
	Symbol string `json:"symbol" yaml:"symbol"`

	// Balance is the user's balance of this equity
	Balance math.Int `json:"balance" yaml:"balance"`

	// HODLValue is the estimated HODL value
	HODLValue math.Int `json:"hodl_value" yaml:"hodl_value"`

	// Price is the current price in HODL
	Price math.LegacyDec `json:"price" yaml:"price"`

	// MarketActive indicates if the market is active for trading
	MarketActive bool `json:"market_active" yaml:"market_active"`
}
