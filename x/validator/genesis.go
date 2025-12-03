package validator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/validator/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/validator/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set params
	k.SetParams(ctx, genState.Params)
	
	// Set validator tiers
	for _, tierInfo := range genState.ValidatorTiers {
		k.SetValidatorTierInfo(ctx, tierInfo)
	}
	
	// Set business verifications
	for _, verification := range genState.BusinessVerifications {
		k.SetBusinessVerification(ctx, verification)
	}
	
	// Set validator business stakes
	for _, stake := range genState.ValidatorBusinessStakes {
		k.SetValidatorBusinessStake(ctx, stake)
	}
	
	// Set verification rewards
	for _, rewards := range genState.VerificationRewards {
		k.SetVerificationRewards(ctx, rewards)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.ValidatorTiers = k.GetAllValidatorTiers(ctx)
	genesis.BusinessVerifications = k.GetAllBusinessVerifications(ctx)
	
	// Export validator business stakes
	validatorBusinessStakes := []types.ValidatorBusinessStake{}
	for _, tierInfo := range genesis.ValidatorTiers {
		valAddr, _ := sdk.ValAddressFromBech32(tierInfo.ValidatorAddress)
		stakes := k.GetValidatorBusinessStakes(ctx, valAddr)
		validatorBusinessStakes = append(validatorBusinessStakes, stakes...)
	}
	genesis.ValidatorBusinessStakes = validatorBusinessStakes
	
	// Export verification rewards
	verificationRewards := []types.VerificationRewards{}
	for _, tierInfo := range genesis.ValidatorTiers {
		valAddr, _ := sdk.ValAddressFromBech32(tierInfo.ValidatorAddress)
		rewards, exists := k.GetVerificationRewards(ctx, valAddr)
		if exists {
			verificationRewards = append(verificationRewards, rewards)
		}
	}
	genesis.VerificationRewards = verificationRewards
	
	return genesis
}