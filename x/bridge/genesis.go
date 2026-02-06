package bridge

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

// InitGenesis initializes the bridge module's state from a provided genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set params
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Set supported assets
	for _, asset := range genState.SupportedAssets {
		k.SetSupportedAsset(ctx, asset)
	}

	// Set reserves
	for _, reserve := range genState.Reserves {
		k.SetReserveBalance(ctx, reserve.Denom, reserve.Balance)
	}

	// Set swap records
	for _, record := range genState.SwapRecords {
		k.SetSwapRecord(ctx, record)
	}

	// Set withdrawal proposals
	for _, proposal := range genState.WithdrawalProposals {
		k.SetWithdrawalProposal(ctx, proposal)
	}

	// Set counters
	k.SetNextSwapID(ctx, genState.NextSwapID)
	k.SetNextProposalID(ctx, genState.NextProposalID)
}

// ExportGenesis returns the bridge module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:              k.GetParams(ctx),
		SupportedAssets:     k.GetAllSupportedAssets(ctx),
		Reserves:            k.GetAllReserveBalances(ctx),
		SwapRecords:         k.GetAllSwapRecords(ctx),
		WithdrawalProposals: k.GetAllWithdrawalProposals(ctx),
		NextSwapID:          k.GetNextSwapID(ctx),
		NextProposalID:      k.GetNextProposalID(ctx),
	}
}
