package keeper

import (
	"encoding/json"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/types"
)

// =============================================================================
// TREASURY GRANT MANAGEMENT
// =============================================================================

// SetTreasuryGrant stores a treasury grant
func (k Keeper) SetTreasuryGrant(ctx sdk.Context, grant types.TreasuryFeeGrant) error {
	if grant.Grantee == "" {
		return types.ErrInvalidGrant
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(grant)
	if err != nil {
		return err
	}
	store.Set(types.TreasuryGrantKey(grant.Grantee), bz)
	return nil
}

// GetTreasuryGrant retrieves a treasury grant for a grantee
func (k Keeper) GetTreasuryGrant(ctx sdk.Context, grantee string) (types.TreasuryFeeGrant, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TreasuryGrantKey(grantee))
	if bz == nil {
		return types.TreasuryFeeGrant{}, false
	}

	var grant types.TreasuryFeeGrant
	if err := json.Unmarshal(bz, &grant); err != nil {
		return types.TreasuryFeeGrant{}, false
	}
	return grant, true
}

// DeleteTreasuryGrant removes a treasury grant
func (k Keeper) DeleteTreasuryGrant(ctx sdk.Context, grantee string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.TreasuryGrantKey(grantee))
}

// IterateTreasuryGrants iterates over all treasury grants
func (k Keeper) IterateTreasuryGrants(ctx sdk.Context, fn func(grant types.TreasuryFeeGrant) bool) {
	store := ctx.KVStore(k.storeKey)
	grantStore := prefix.NewStore(store, types.TreasuryGrantPrefix)
	iterator := grantStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var grant types.TreasuryFeeGrant
		if err := json.Unmarshal(iterator.Value(), &grant); err != nil {
			continue
		}
		if fn(grant) {
			break
		}
	}
}

// GetAllTreasuryGrants returns all treasury grants
func (k Keeper) GetAllTreasuryGrants(ctx sdk.Context) []types.TreasuryFeeGrant {
	grants := []types.TreasuryFeeGrant{}
	k.IterateTreasuryGrants(ctx, func(grant types.TreasuryFeeGrant) bool {
		grants = append(grants, grant)
		return false
	})
	return grants
}

// =============================================================================
// TREASURY BALANCE MANAGEMENT
// =============================================================================

// GetTreasuryBalance returns the current treasury balance
func (k Keeper) GetTreasuryBalance(ctx sdk.Context) math.Int {
	treasuryAddr := k.accountKeeper.GetModuleAddress(types.TreasuryPoolName)
	if treasuryAddr == nil {
		return math.ZeroInt()
	}
	balance := k.bankKeeper.GetBalance(ctx, treasuryAddr, "uhodl")
	return balance.Amount
}

// FundTreasury adds funds to the treasury from a funder
func (k Keeper) FundTreasury(ctx sdk.Context, funder sdk.AccAddress, amount math.Int) error {
	if amount.IsZero() || amount.IsNegative() {
		return types.ErrZeroAmount
	}

	coins := sdk.NewCoins(sdk.NewCoin("uhodl", amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, funder, types.TreasuryPoolName, coins); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_funded",
			sdk.NewAttribute("funder", funder.String()),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("new_balance", k.GetTreasuryBalance(ctx).String()),
		),
	)

	return nil
}

// WithdrawFromTreasury withdraws funds from the treasury (governance only)
func (k Keeper) WithdrawFromTreasury(ctx sdk.Context, recipient sdk.AccAddress, amount math.Int) error {
	if amount.IsZero() || amount.IsNegative() {
		return types.ErrZeroAmount
	}

	// Check treasury has sufficient funds
	balance := k.GetTreasuryBalance(ctx)
	if balance.LT(amount) {
		return types.ErrInsufficientTreasuryFunds
	}

	coins := sdk.NewCoins(sdk.NewCoin("uhodl", amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.TreasuryPoolName, recipient, coins); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_withdrawal",
			sdk.NewAttribute("recipient", recipient.String()),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("remaining_balance", k.GetTreasuryBalance(ctx).String()),
		),
	)

	return nil
}

// =============================================================================
// GRANT USAGE
// =============================================================================

// UseGrant uses a portion of a treasury grant for fee abstraction
func (k Keeper) UseGrant(ctx sdk.Context, grantee string, amount math.Int) error {
	grant, found := k.GetTreasuryGrant(ctx, grantee)
	if !found {
		return types.ErrGrantNotFound
	}

	// Check if grant is expired
	if grant.IsExpired(uint64(ctx.BlockHeight())) {
		k.DeleteTreasuryGrant(ctx, grantee)
		return types.ErrGrantExpired
	}

	// Check if grant has sufficient allowance
	remaining := grant.RemainingAllowance()
	if remaining.LT(amount) {
		return types.ErrGrantExhausted
	}

	// Update usage
	grant.UsedFeeAmountUhodl = grant.UsedFeeAmountUhodl.Add(amount)
	if err := k.SetTreasuryGrant(ctx, grant); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"grant_used",
			sdk.NewAttribute("grantee", grantee),
			sdk.NewAttribute("amount_used", amount.String()),
			sdk.NewAttribute("remaining", grant.RemainingAllowance().String()),
		),
	)

	return nil
}

// CreateGrant creates a new treasury grant (governance only)
func (k Keeper) CreateGrant(
	ctx sdk.Context,
	grantee string,
	allowedAmount math.Int,
	expirationHeight uint64,
	allowedEquities []string,
	proposalID uint64,
) error {
	grant := types.TreasuryFeeGrant{
		Grantee:              grantee,
		AllowedFeeAmountUhodl: allowedAmount,
		UsedFeeAmountUhodl:   math.ZeroInt(),
		ExpirationHeight:     expirationHeight,
		AllowedEquitySymbols: allowedEquities,
		GrantedAtHeight:      uint64(ctx.BlockHeight()),
		GrantedByProposalID:  proposalID,
	}

	if err := k.SetTreasuryGrant(ctx, grant); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"grant_created",
			sdk.NewAttribute("grantee", grantee),
			sdk.NewAttribute("allowed_amount", allowedAmount.String()),
			sdk.NewAttribute("expiration_height", string(rune(expirationHeight))),
			sdk.NewAttribute("proposal_id", string(rune(proposalID))),
		),
	)

	return nil
}

// RevokeGrant revokes a treasury grant (governance only)
func (k Keeper) RevokeGrant(ctx sdk.Context, grantee string) error {
	grant, found := k.GetTreasuryGrant(ctx, grantee)
	if !found {
		return types.ErrGrantNotFound
	}

	k.DeleteTreasuryGrant(ctx, grantee)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"grant_revoked",
			sdk.NewAttribute("grantee", grantee),
			sdk.NewAttribute("remaining_at_revocation", grant.RemainingAllowance().String()),
		),
	)

	return nil
}
