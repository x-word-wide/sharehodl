package types

import (
	"fmt"
)

// GenesisState defines the extbridge module's genesis state
type GenesisState struct {
	Params          Params           `json:"params" yaml:"params"`
	ExternalChains  []ExternalChain  `json:"external_chains" yaml:"external_chains"`
	ExternalAssets  []ExternalAsset  `json:"external_assets" yaml:"external_assets"`
	Deposits        []Deposit        `json:"deposits" yaml:"deposits"`
	Attestations    []DepositAttestation `json:"attestations" yaml:"attestations"`
	Withdrawals     []Withdrawal     `json:"withdrawals" yaml:"withdrawals"`
	TSSSessions     []TSSSession     `json:"tss_sessions" yaml:"tss_sessions"`
	CircuitBreaker  CircuitBreaker   `json:"circuit_breaker" yaml:"circuit_breaker"`
	NextDepositID   uint64           `json:"next_deposit_id" yaml:"next_deposit_id"`
	NextWithdrawalID uint64          `json:"next_withdrawal_id" yaml:"next_withdrawal_id"`
	NextTSSSessionID uint64          `json:"next_tss_session_id" yaml:"next_tss_session_id"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		ExternalChains:  []ExternalChain{},
		ExternalAssets:  []ExternalAsset{},
		Deposits:        []Deposit{},
		Attestations:    []DepositAttestation{},
		Withdrawals:     []Withdrawal{},
		TSSSessions:     []TSSSession{},
		CircuitBreaker:  CircuitBreaker{Enabled: false},
		NextDepositID:   1,
		NextWithdrawalID: 1,
		NextTSSSessionID: 1,
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate external chains
	chainIDs := make(map[string]bool)
	for _, chain := range gs.ExternalChains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("invalid external chain %s: %w", chain.ChainID, err)
		}
		if chainIDs[chain.ChainID] {
			return fmt.Errorf("duplicate chain ID: %s", chain.ChainID)
		}
		chainIDs[chain.ChainID] = true
	}

	// Validate external assets
	assetKeys := make(map[string]bool)
	for _, asset := range gs.ExternalAssets {
		if err := asset.Validate(); err != nil {
			return fmt.Errorf("invalid external asset %s:%s: %w", asset.ChainID, asset.AssetSymbol, err)
		}
		key := fmt.Sprintf("%s:%s", asset.ChainID, asset.AssetSymbol)
		if assetKeys[key] {
			return fmt.Errorf("duplicate asset: %s", key)
		}
		assetKeys[key] = true

		// Verify chain exists
		if !chainIDs[asset.ChainID] {
			return fmt.Errorf("asset %s references non-existent chain %s", asset.AssetSymbol, asset.ChainID)
		}
	}

	// Validate deposits
	depositIDs := make(map[uint64]bool)
	for _, deposit := range gs.Deposits {
		if err := deposit.Validate(); err != nil {
			return fmt.Errorf("invalid deposit %d: %w", deposit.ID, err)
		}
		if depositIDs[deposit.ID] {
			return fmt.Errorf("duplicate deposit ID: %d", deposit.ID)
		}
		depositIDs[deposit.ID] = true
	}

	// Validate attestations
	for _, attestation := range gs.Attestations {
		if err := attestation.Validate(); err != nil {
			return fmt.Errorf("invalid attestation: %w", err)
		}
		// Verify deposit exists
		if !depositIDs[attestation.DepositID] {
			return fmt.Errorf("attestation references non-existent deposit %d", attestation.DepositID)
		}
	}

	// Validate withdrawals
	withdrawalIDs := make(map[uint64]bool)
	for _, withdrawal := range gs.Withdrawals {
		if err := withdrawal.Validate(); err != nil {
			return fmt.Errorf("invalid withdrawal %d: %w", withdrawal.ID, err)
		}
		if withdrawalIDs[withdrawal.ID] {
			return fmt.Errorf("duplicate withdrawal ID: %d", withdrawal.ID)
		}
		withdrawalIDs[withdrawal.ID] = true
	}

	// Validate TSS sessions
	sessionIDs := make(map[uint64]bool)
	for _, session := range gs.TSSSessions {
		if err := session.Validate(); err != nil {
			return fmt.Errorf("invalid TSS session %d: %w", session.ID, err)
		}
		if sessionIDs[session.ID] {
			return fmt.Errorf("duplicate TSS session ID: %d", session.ID)
		}
		sessionIDs[session.ID] = true
	}

	// Validate circuit breaker
	if err := gs.CircuitBreaker.Validate(); err != nil {
		return fmt.Errorf("invalid circuit breaker: %w", err)
	}

	// Validate ID counters
	if gs.NextDepositID == 0 {
		return fmt.Errorf("next deposit ID must be positive")
	}
	if gs.NextWithdrawalID == 0 {
		return fmt.Errorf("next withdrawal ID must be positive")
	}
	if gs.NextTSSSessionID == 0 {
		return fmt.Errorf("next TSS session ID must be positive")
	}

	return nil
}
