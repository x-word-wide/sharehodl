package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		SupportedAssets:    DefaultSupportedAssets(),
		Reserves:           []BridgeReserve{},
		SwapRecords:        []SwapRecord{},
		WithdrawalProposals: []ReserveWithdrawal{},
		NextSwapID:         1,
		NextProposalID:     1,
	}
}

// DefaultSupportedAssets returns the default supported assets for the bridge
func DefaultSupportedAssets() []SupportedAsset {
	return []SupportedAsset{
		{
			Denom:            "uusdt",
			SwapRateToHODL:   math.LegacyOneDec(),        // 1 USDT = 1 HODL
			SwapRateFromHODL: math.LegacyOneDec(),        // 1 HODL = 1 USDT
			MinSwapAmount:    math.NewInt(1_000_000),     // 1 USDT minimum
			MaxSwapAmount:    math.NewInt(1_000_000_000_000), // 1M USDT maximum
			Enabled:          true,
		},
		{
			Denom:            "uusdc",
			SwapRateToHODL:   math.LegacyOneDec(),
			SwapRateFromHODL: math.LegacyOneDec(),
			MinSwapAmount:    math.NewInt(1_000_000),
			MaxSwapAmount:    math.NewInt(1_000_000_000_000),
			Enabled:          true,
		},
		{
			Denom:            "udai",
			SwapRateToHODL:   math.LegacyOneDec(),
			SwapRateFromHODL: math.LegacyOneDec(),
			MinSwapAmount:    math.NewInt(1_000_000),
			MaxSwapAmount:    math.NewInt(1_000_000_000_000),
			Enabled:          true,
		},
		{
			Denom:            "ubtc",
			SwapRateToHODL:   math.LegacyNewDec(100000),  // 1 BTC = 100,000 HODL (assuming BTC ~$100k)
			SwapRateFromHODL: math.LegacyNewDecWithPrec(1, 5), // 1 HODL = 0.00001 BTC
			MinSwapAmount:    math.NewInt(100_000),       // 0.001 BTC minimum
			MaxSwapAmount:    math.NewInt(100_000_000_000), // 1000 BTC maximum
			Enabled:          true,
		},
		{
			Denom:            "ueth",
			SwapRateToHODL:   math.LegacyNewDec(3500),    // 1 ETH = 3,500 HODL (assuming ETH ~$3.5k)
			SwapRateFromHODL: math.LegacyNewDecWithPrec(2857, 7), // 1 HODL = 0.0002857 ETH
			MinSwapAmount:    math.NewInt(1_000_000),     // 0.001 ETH minimum
			MaxSwapAmount:    math.NewInt(10_000_000_000_000), // 10,000 ETH maximum
			Enabled:          true,
		},
	}
}

// GenesisState defines the bridge module's genesis state
type GenesisState struct {
	Params              Params             `json:"params" yaml:"params"`
	SupportedAssets     []SupportedAsset   `json:"supported_assets" yaml:"supported_assets"`
	Reserves            []BridgeReserve    `json:"reserves" yaml:"reserves"`
	SwapRecords         []SwapRecord       `json:"swap_records" yaml:"swap_records"`
	WithdrawalProposals []ReserveWithdrawal `json:"withdrawal_proposals" yaml:"withdrawal_proposals"`
	NextSwapID          uint64             `json:"next_swap_id" yaml:"next_swap_id"`
	NextProposalID      uint64             `json:"next_proposal_id" yaml:"next_proposal_id"`
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate supported assets
	seenDenoms := make(map[string]bool)
	for _, asset := range gs.SupportedAssets {
		if err := asset.Validate(); err != nil {
			return fmt.Errorf("invalid supported asset %s: %w", asset.Denom, err)
		}
		if seenDenoms[asset.Denom] {
			return fmt.Errorf("duplicate supported asset: %s", asset.Denom)
		}
		seenDenoms[asset.Denom] = true
	}

	// Validate reserves
	for _, reserve := range gs.Reserves {
		if err := reserve.Validate(); err != nil {
			return fmt.Errorf("invalid reserve %s: %w", reserve.Denom, err)
		}
	}

	// Validate swap records
	for _, record := range gs.SwapRecords {
		if err := record.Validate(); err != nil {
			return fmt.Errorf("invalid swap record %d: %w", record.ID, err)
		}
	}

	// Validate withdrawal proposals
	for _, proposal := range gs.WithdrawalProposals {
		if err := proposal.Validate(); err != nil {
			return fmt.Errorf("invalid withdrawal proposal %d: %w", proposal.ID, err)
		}
	}

	return nil
}
