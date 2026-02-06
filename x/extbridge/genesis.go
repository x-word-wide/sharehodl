package extbridge

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set params
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Set external chains
	for _, chain := range genState.ExternalChains {
		k.SetExternalChain(ctx, chain)
	}

	// Set external assets
	for _, asset := range genState.ExternalAssets {
		k.SetExternalAsset(ctx, asset)
	}

	// Set deposits
	for _, deposit := range genState.Deposits {
		k.SetDeposit(ctx, deposit)
	}

	// Set attestations
	for _, attestation := range genState.Attestations {
		k.SetAttestation(ctx, attestation)
	}

	// Set withdrawals
	for _, withdrawal := range genState.Withdrawals {
		k.SetWithdrawal(ctx, withdrawal)
	}

	// Set TSS sessions
	for _, session := range genState.TSSSessions {
		k.SetTSSSession(ctx, session)
	}

	// Set circuit breaker
	k.SetCircuitBreaker(ctx, genState.CircuitBreaker)

	// Initialize ID counters (handled by keeper via GetNext methods)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:           k.GetParams(ctx),
		ExternalChains:   k.GetAllExternalChains(ctx),
		ExternalAssets:   k.GetAllExternalAssets(ctx),
		Deposits:         k.GetAllDeposits(ctx),
		Attestations:     []types.DepositAttestation{}, // TODO: Export attestations
		Withdrawals:      k.GetAllWithdrawals(ctx),
		TSSSessions:      k.GetAllTSSSessions(ctx),
		CircuitBreaker:   k.GetCircuitBreaker(ctx),
		NextDepositID:    k.GetNextDepositID(ctx),
		NextWithdrawalID: k.GetNextWithdrawalID(ctx),
		NextTSSSessionID: k.GetNextTSSSessionID(ctx),
	}
}
