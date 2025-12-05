package hodl

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/hodl/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set params
	k.SetParams(ctx, genState.Params)
	
	// Set total supply
	k.SetTotalSupply(ctx, genState.TotalSupply)
	
	// Set collateral positions
	for _, position := range genState.Positions {
		k.SetCollateralPosition(ctx, position)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.TotalSupply = k.GetTotalSupply(ctx).(math.Int)
	genesis.Positions = k.GetAllCollateralPositions(ctx)
	
	// Calculate total collateral from positions
	totalCollateral := sdk.NewCoins()
	for _, position := range genesis.Positions {
		totalCollateral = totalCollateral.Add(position.Collateral...)
	}
	genesis.TotalCollateral = totalCollateral
	
	return genesis
}