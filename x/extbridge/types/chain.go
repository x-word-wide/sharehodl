package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// ExternalChain represents a supported external blockchain
type ExternalChain struct {
	ChainID          string         `json:"chain_id" yaml:"chain_id"`                     // e.g., "ethereum-1", "bitcoin-mainnet"
	ChainName        string         `json:"chain_name" yaml:"chain_name"`                 // e.g., "Ethereum", "Bitcoin"
	ChainType        string         `json:"chain_type" yaml:"chain_type"`                 // e.g., "evm", "utxo"
	Enabled          bool           `json:"enabled" yaml:"enabled"`                       // Whether bridging is enabled for this chain
	ConfirmationsReq uint64         `json:"confirmations_required" yaml:"confirmations_required"` // Min confirmations before accepting deposit
	BlockTime        uint64         `json:"block_time" yaml:"block_time"`                 // Average block time in seconds
	TSSAddress       string         `json:"tss_address" yaml:"tss_address"`               // TSS-controlled address on external chain
	MinDeposit       math.Int       `json:"min_deposit" yaml:"min_deposit"`               // Minimum deposit amount
	MaxDeposit       math.Int       `json:"max_deposit" yaml:"max_deposit"`               // Maximum deposit amount per transaction
}

// NewExternalChain creates a new ExternalChain
func NewExternalChain(
	chainID string,
	chainName string,
	chainType string,
	enabled bool,
	confirmationsReq uint64,
	blockTime uint64,
	tssAddress string,
	minDeposit math.Int,
	maxDeposit math.Int,
) ExternalChain {
	return ExternalChain{
		ChainID:          chainID,
		ChainName:        chainName,
		ChainType:        chainType,
		Enabled:          enabled,
		ConfirmationsReq: confirmationsReq,
		BlockTime:        blockTime,
		TSSAddress:       tssAddress,
		MinDeposit:       minDeposit,
		MaxDeposit:       maxDeposit,
	}
}

// Validate validates the external chain configuration
func (ec ExternalChain) Validate() error {
	if ec.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if ec.ChainName == "" {
		return fmt.Errorf("chain name cannot be empty")
	}
	if ec.ChainType == "" {
		return fmt.Errorf("chain type cannot be empty")
	}
	if ec.ConfirmationsReq == 0 {
		return fmt.Errorf("confirmations required must be positive")
	}
	if ec.BlockTime == 0 {
		return fmt.Errorf("block time must be positive")
	}
	if ec.TSSAddress == "" {
		return fmt.Errorf("TSS address cannot be empty")
	}
	if ec.MinDeposit.IsNil() || ec.MinDeposit.IsNegative() {
		return fmt.Errorf("min deposit must be non-negative")
	}
	if ec.MaxDeposit.IsNil() || !ec.MaxDeposit.IsPositive() {
		return fmt.Errorf("max deposit must be positive")
	}
	if ec.MaxDeposit.LT(ec.MinDeposit) {
		return fmt.Errorf("max deposit must be >= min deposit")
	}
	return nil
}

// ExternalAsset represents a supported asset on an external chain
type ExternalAsset struct {
	ChainID         string         `json:"chain_id" yaml:"chain_id"`           // Parent chain ID
	AssetSymbol     string         `json:"asset_symbol" yaml:"asset_symbol"`   // e.g., "USDT", "BTC", "ETH"
	AssetName       string         `json:"asset_name" yaml:"asset_name"`       // e.g., "Tether USD", "Bitcoin"
	ContractAddress string         `json:"contract_address" yaml:"contract_address"` // Token contract (for ERC20, etc.)
	Decimals        uint32         `json:"decimals" yaml:"decimals"`           // Decimals on external chain
	Enabled         bool           `json:"enabled" yaml:"enabled"`             // Whether this asset is enabled
	ConversionRate  math.LegacyDec `json:"conversion_rate" yaml:"conversion_rate"` // How many HODL per 1 unit of asset
	DailyLimit      math.Int       `json:"daily_limit" yaml:"daily_limit"`     // Max amount per day
	PerTxLimit      math.Int       `json:"per_tx_limit" yaml:"per_tx_limit"`   // Max amount per transaction
}

// NewExternalAsset creates a new ExternalAsset
func NewExternalAsset(
	chainID string,
	assetSymbol string,
	assetName string,
	contractAddress string,
	decimals uint32,
	enabled bool,
	conversionRate math.LegacyDec,
	dailyLimit math.Int,
	perTxLimit math.Int,
) ExternalAsset {
	return ExternalAsset{
		ChainID:         chainID,
		AssetSymbol:     assetSymbol,
		AssetName:       assetName,
		ContractAddress: contractAddress,
		Decimals:        decimals,
		Enabled:         enabled,
		ConversionRate:  conversionRate,
		DailyLimit:      dailyLimit,
		PerTxLimit:      perTxLimit,
	}
}

// Validate validates the external asset configuration
func (ea ExternalAsset) Validate() error {
	if ea.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if ea.AssetSymbol == "" {
		return fmt.Errorf("asset symbol cannot be empty")
	}
	if ea.AssetName == "" {
		return fmt.Errorf("asset name cannot be empty")
	}
	if ea.Decimals == 0 || ea.Decimals > 18 {
		return fmt.Errorf("decimals must be between 1 and 18")
	}
	if ea.ConversionRate.IsNil() || !ea.ConversionRate.IsPositive() {
		return fmt.Errorf("conversion rate must be positive")
	}
	if ea.DailyLimit.IsNil() || !ea.DailyLimit.IsPositive() {
		return fmt.Errorf("daily limit must be positive")
	}
	if ea.PerTxLimit.IsNil() || !ea.PerTxLimit.IsPositive() {
		return fmt.Errorf("per tx limit must be positive")
	}
	if ea.PerTxLimit.GT(ea.DailyLimit) {
		return fmt.Errorf("per tx limit cannot exceed daily limit")
	}
	return nil
}

// GetFullAssetID returns the full asset identifier
func (ea ExternalAsset) GetFullAssetID() string {
	return fmt.Sprintf("%s:%s", ea.ChainID, ea.AssetSymbol)
}
