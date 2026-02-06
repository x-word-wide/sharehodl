package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// Shareholder Blacklist Management

// BlacklistShareholder adds a shareholder to the blacklist
func (k Keeper) BlacklistShareholder(
	ctx sdk.Context,
	companyID uint64,
	address string,
	reason string,
	blacklistedBy string,
	duration time.Duration,
) error {
	// Validate inputs
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		return types.ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(blacklistedBy); err != nil {
		return types.ErrUnauthorized
	}

	// Check if already blacklisted
	if k.IsShareholderBlacklisted(ctx, companyID, address) {
		return types.ErrBlacklistAlreadyExists
	}

	// Create blacklist entry
	blacklist := types.NewShareholderBlacklist(
		companyID,
		address,
		reason,
		blacklistedBy,
		duration,
	)

	// Validate
	if err := blacklist.Validate(); err != nil {
		return err
	}

	// Store blacklist entry
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholderBlacklistKey(companyID, address)
	bz, err := json.Marshal(blacklist)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeShareholderBlacklisted,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyShareholder, address),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyBlacklistedBy, blacklistedBy),
		),
	)

	return nil
}

// UnblacklistShareholder removes a shareholder from the blacklist
func (k Keeper) UnblacklistShareholder(ctx sdk.Context, companyID uint64, address string) error {
	// Validate address
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		return types.ErrUnauthorized
	}

	// Check if blacklisted
	if !k.IsShareholderBlacklisted(ctx, companyID, address) {
		return types.ErrNotBlacklisted
	}

	// Remove blacklist entry
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholderBlacklistKey(companyID, address)
	store.Delete(key)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeShareholderUnblacklisted,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyShareholder, address),
		),
	)

	return nil
}

// IsBlacklisted checks if an address is blacklisted from trading a company's shares
// This is the interface method called by DEX module for trade validation
func (k Keeper) IsBlacklisted(ctx sdk.Context, companyID uint64, address string) bool {
	return k.IsShareholderBlacklisted(ctx, companyID, address)
}

// IsShareholderBlacklisted checks if a shareholder is currently blacklisted
func (k Keeper) IsShareholderBlacklisted(ctx sdk.Context, companyID uint64, address string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholderBlacklistKey(companyID, address)
	bz := store.Get(key)
	if bz == nil {
		return false
	}

	var blacklist types.ShareholderBlacklist
	if err := json.Unmarshal(bz, &blacklist); err != nil {
		return false
	}

	// Check if blacklist is active
	if !blacklist.IsActive {
		return false
	}

	// Check if expired
	if blacklist.IsExpired(ctx.BlockTime()) {
		// Auto-remove expired blacklist
		store.Delete(key)
		return false
	}

	return true
}

// GetShareholderBlacklist returns the blacklist entry for a shareholder
func (k Keeper) GetShareholderBlacklist(ctx sdk.Context, companyID uint64, address string) (types.ShareholderBlacklist, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholderBlacklistKey(companyID, address)
	bz := store.Get(key)
	if bz == nil {
		return types.ShareholderBlacklist{}, false
	}

	var blacklist types.ShareholderBlacklist
	if err := json.Unmarshal(bz, &blacklist); err != nil {
		return types.ShareholderBlacklist{}, false
	}

	return blacklist, true
}

// GetBlacklistedShareholders returns all blacklisted shareholders for a company
func (k Keeper) GetBlacklistedShareholders(ctx sdk.Context, companyID uint64) []types.ShareholderBlacklist {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ShareholderBlacklistPrefix)
	prefixBytes := sdk.Uint64ToBigEndian(companyID)
	iterator := store.Iterator(prefixBytes, append(prefixBytes, 0xFF))
	defer iterator.Close()

	blacklists := []types.ShareholderBlacklist{}
	for ; iterator.Valid(); iterator.Next() {
		var blacklist types.ShareholderBlacklist
		if err := json.Unmarshal(iterator.Value(), &blacklist); err == nil {
			// Only include active and non-expired blacklists
			if blacklist.IsActive && !blacklist.IsExpired(ctx.BlockTime()) {
				blacklists = append(blacklists, blacklist)
			}
		}
	}

	return blacklists
}

// Dividend Redirection Management

// SetDividendRedirection configures where blocked dividends should be redirected
func (k Keeper) SetDividendRedirection(ctx sdk.Context, redirect types.DividendRedirection) error {
	// Validate
	if err := redirect.Validate(); err != nil {
		return err
	}

	// Store redirection config
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendRedirectionKey(redirect.CompanyID)
	bz, err := json.Marshal(redirect)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

// GetDividendRedirection returns the dividend redirection config for a company
func (k Keeper) GetDividendRedirection(ctx sdk.Context, companyID uint64) (types.DividendRedirection, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendRedirectionKey(companyID)
	bz := store.Get(key)
	if bz == nil {
		return types.DividendRedirection{}, false
	}

	var redirect types.DividendRedirection
	if err := json.Unmarshal(bz, &redirect); err != nil {
		return types.DividendRedirection{}, false
	}

	return redirect, true
}

// Default Charity Wallet Management

// SetDefaultCharityWallet sets the protocol-level default charity wallet
func (k Keeper) SetDefaultCharityWallet(ctx sdk.Context, address string) error {
	// Validate address
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		return types.ErrInvalidCharityWallet
	}

	// Store default charity wallet
	store := ctx.KVStore(k.storeKey)
	store.Set(types.DefaultCharityWalletKey, []byte(address))

	return nil
}

// GetDefaultCharityWallet returns the protocol-level default charity wallet
func (k Keeper) GetDefaultCharityWallet(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DefaultCharityWalletKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// GetCommunityPoolAddress returns the community pool address from the governance module
func (k Keeper) GetCommunityPoolAddress(ctx sdk.Context) string {
	// In Cosmos SDK, the community pool is typically managed by the distribution module
	// For ShareHODL, we'll use a well-known address or the governance module account
	// This is a placeholder - in production, coordinate with governance module
	govModuleAddr := sdk.AccAddress([]byte("governance_module"))
	return govModuleAddr.String()
}
