package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// GetOwnerAssets returns all assets owned by an address
func (k Keeper) GetOwnerAssets(ctx sdk.Context, owner sdk.AccAddress) (sdk.Coins, []types.EquityLock, error) {
	// Get token balances
	coins := k.bankKeeper.GetAllBalances(ctx, owner)

	// Get equity holdings (if equity keeper is available)
	var equityLocks []types.EquityLock
	// Note: This would require iterating through equity holdings
	// For now, we'll return empty equity locks as it requires equity keeper implementation

	return coins, equityLocks, nil
}

// TransferAssetsTobeneficiary transfers assets to a beneficiary
// This is the main entry point called by claim_processor.go
func (k Keeper) TransferAssetsTobeneficiary(ctx sdk.Context, plan types.InheritancePlan, beneficiary types.Beneficiary) ([]types.TransferredAsset, error) {
	var transferred []types.TransferredAsset

	// 1. Transfer bank balances (HODL and other tokens)
	bankAssets, err := k.transferBankBalances(ctx, plan.Owner, beneficiary.Address, beneficiary.Percentage)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer bank balances: %w", err)
	}
	transferred = append(transferred, bankAssets...)

	// 2. Transfer equity shares
	equityAssets, err := k.transferEquityShares(ctx, plan.Owner, beneficiary.Address, beneficiary.Percentage)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer equity shares: %w", err)
	}
	transferred = append(transferred, equityAssets...)

	// 3. Transfer STAKED HODL (initiate unstaking with beneficiary as recipient)
	if k.stakingKeeper != nil {
		stakedAmount, err := k.TransferStakedAssets(ctx, plan.Owner, beneficiary.Address, beneficiary.Percentage)
		if err != nil {
			k.Logger(ctx).Error("failed to transfer staked assets", "error", err)
			// Non-fatal: continue with other assets
		} else if stakedAmount.IsPositive() {
			transferred = append(transferred, types.TransferredAsset{
				AssetType:     types.AssetTypeStaked,
				Denom:         "uhodl",
				Amount:        stakedAmount,
				TransferredAt: ctx.BlockTime(),
			})
		}
	}

	// 4. Transfer LOAN POSITIONS (full position transfer)
	// Both borrower and lender positions are transferred to beneficiary
	if k.lendingKeeper != nil {
		if err := k.TransferLoanPositions(ctx, plan.Owner, beneficiary.Address); err != nil {
			k.Logger(ctx).Error("failed to transfer loan positions", "error", err)
			// Don't fail entire inheritance - loans can be handled separately
		}
	}

	// 5. Transfer ESCROW POSITIONS (full position transfer)
	// Both buyer and seller positions are transferred to beneficiary
	if k.escrowKeeper != nil {
		if err := k.TransferEscrowPositions(ctx, plan.Owner, beneficiary.Address); err != nil {
			k.Logger(ctx).Error("failed to transfer escrow positions", "error", err)
			// Don't fail entire inheritance
		}
	}

	// 6. Cancel DEX orders and transfer unlocked funds
	// This is handled separately as DEX integration is pending
	// if err := k.cancelDEXOrdersAndTransfer(ctx, plan.Owner, beneficiary.Address, beneficiary.Percentage); err != nil {
	// 	k.Logger(ctx).Error("failed to cancel DEX orders", "error", err)
	// }

	return transferred, nil
}

// transferBankBalances transfers percentage of bank balances to beneficiary
func (k Keeper) transferBankBalances(ctx sdk.Context, owner, beneficiary string, percentage math.LegacyDec) ([]types.TransferredAsset, error) {
	var transferred []types.TransferredAsset

	ownerAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return nil, err
	}

	beneficiaryAddr, err := sdk.AccAddressFromBech32(beneficiary)
	if err != nil {
		return nil, err
	}

	// Get all token balances
	balances := k.bankKeeper.GetAllBalances(ctx, ownerAddr)

	// Transfer percentage of each token
	for _, balance := range balances {
		if balance.Amount.IsZero() {
			continue
		}

		// CRITICAL FIX: Check for potential overflow before multiplication
		// For very large amounts (> 1e18), use a safer calculation method
		var transferAmount math.Int
		if balance.Amount.GT(math.NewInt(1e18)) {
			// Use division-first approach to prevent overflow
			// transferAmount = (balance.Amount / (1/percentage))
			// Which is: transferAmount = balance.Amount * percentage
			// But computed as: (balance.Amount * percentage_numerator) / percentage_denominator
			// For LegacyDec, we can use: balance * percentage / 1.0
			quotient := balance.Amount.Quo(math.NewInt(100))
			percentageInt := percentage.MulInt64(100).TruncateInt()
			transferAmount = quotient.Mul(percentageInt).Quo(math.NewInt(100))
		} else {
			// Standard calculation for normal amounts
			amountDec := math.LegacyNewDecFromInt(balance.Amount)
			transferAmount = percentage.Mul(amountDec).TruncateInt()
		}

		if transferAmount.IsZero() {
			continue
		}

		// Additional safety check: ensure transfer amount doesn't exceed balance
		if transferAmount.GT(balance.Amount) {
			k.Logger(ctx).Error("calculated transfer amount exceeds balance, using full balance",
				"balance", balance.Amount.String(),
				"calculated", transferAmount.String(),
				"percentage", percentage.String(),
			)
			transferAmount = balance.Amount
		}

		// Transfer tokens
		coins := sdk.NewCoins(sdk.NewCoin(balance.Denom, transferAmount))
		if err := k.bankKeeper.SendCoins(ctx, ownerAddr, beneficiaryAddr, coins); err != nil {
			k.Logger(ctx).Error("failed to transfer percentage of tokens",
				"from", owner,
				"to", beneficiary,
				"denom", balance.Denom,
				"amount", transferAmount,
				"error", err,
			)
			continue
		}

		transferred = append(transferred, types.TransferredAsset{
			AssetType:     types.AssetTypeHODL,
			Denom:         balance.Denom,
			Amount:        transferAmount,
			TransferredAt: ctx.BlockTime(),
		})

		k.Logger(ctx).Info("percentage of tokens transferred",
			"from", owner,
			"to", beneficiary,
			"denom", balance.Denom,
			"percentage", percentage,
			"amount", transferAmount,
		)
	}

	return transferred, nil
}

// transferEquityShares transfers percentage of equity shares to beneficiary
func (k Keeper) transferEquityShares(ctx sdk.Context, owner, beneficiary string, percentage math.LegacyDec) ([]types.TransferredAsset, error) {
	var transferred []types.TransferredAsset

	// Check if equity keeper is available
	if k.equityKeeper == nil {
		k.Logger(ctx).Debug("equity keeper not available, skipping equity transfer")
		return transferred, nil
	}

	// Get all equity holdings for owner
	holdingsInterface := k.equityKeeper.GetAllHoldingsByAddress(ctx, owner)
	if len(holdingsInterface) == 0 {
		k.Logger(ctx).Debug("no equity holdings found for owner", "owner", owner)
		return transferred, nil
	}

	// Process each holding
	for _, holdingInterface := range holdingsInterface {
		// Type assert to extract holding details
		holding, ok := holdingInterface.(interface {
			GetCompanyID() uint64
			GetClassID() string
			GetShares() math.Int
		})

		// If type assertion fails, try struct field access
		if !ok {
			// Try direct field access via reflection-like pattern
			type AddressHolding struct {
				CompanyID uint64
				ClassID   string
				Shares    math.Int
			}

			if h, ok := holdingInterface.(AddressHolding); ok {
				companyID := h.CompanyID
				classID := h.ClassID
				shares := h.Shares

				// Calculate transfer amount
				shareAmount := percentage.MulInt(shares).TruncateInt()
				if shareAmount.IsZero() {
					continue
				}

				// Transfer shares
				err := k.equityKeeper.TransferShares(ctx, companyID, classID, owner, beneficiary, shareAmount)
				if err != nil {
					k.Logger(ctx).Error("failed to transfer equity shares",
						"company_id", companyID,
						"class_id", classID,
						"amount", shareAmount.String(),
						"error", err,
					)
					continue
				}

				// Record transfer
				transferred = append(transferred, types.TransferredAsset{
					AssetType:     types.AssetTypeEquity,
					CompanyID:     companyID,
					ClassID:       classID,
					Shares:        shareAmount,
					TransferredAt: ctx.BlockTime(),
				})

				k.Logger(ctx).Info("equity shares transferred",
					"from", owner,
					"to", beneficiary,
					"company_id", companyID,
					"class_id", classID,
					"percentage", percentage.String(),
					"shares", shareAmount.String(),
				)
			} else {
				k.Logger(ctx).Error("invalid holding type", "holding", holdingInterface)
				continue
			}
			continue
		}

		// Extract holding details using interface
		companyID := holding.GetCompanyID()
		classID := holding.GetClassID()
		shares := holding.GetShares()

		// Calculate transfer amount
		shareAmount := percentage.MulInt(shares).TruncateInt()
		if shareAmount.IsZero() {
			continue
		}

		// Transfer shares
		err := k.equityKeeper.TransferShares(ctx, companyID, classID, owner, beneficiary, shareAmount)
		if err != nil {
			k.Logger(ctx).Error("failed to transfer equity shares",
				"company_id", companyID,
				"class_id", classID,
				"amount", shareAmount.String(),
				"error", err,
			)
			continue
		}

		// Record transfer
		transferred = append(transferred, types.TransferredAsset{
			AssetType:     types.AssetTypeEquity,
			CompanyID:     companyID,
			ClassID:       classID,
			Shares:        shareAmount,
			TransferredAt: ctx.BlockTime(),
		})

		k.Logger(ctx).Info("equity shares transferred",
			"from", owner,
			"to", beneficiary,
			"company_id", companyID,
			"class_id", classID,
			"percentage", percentage.String(),
			"shares", shareAmount.String(),
		)
	}

	return transferred, nil
}

// TransferStakedAssets handles staked HODL inheritance
// Initiates unstaking with beneficiary as the recipient of unstaked funds
func (k Keeper) TransferStakedAssets(ctx sdk.Context, owner string, beneficiary string, percentage math.LegacyDec) (math.Int, error) {
	ownerAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Get owner's stake
	stake, found := k.stakingKeeper.GetUserStake(ctx, ownerAddr)
	if !found {
		return math.ZeroInt(), nil // No stake to transfer
	}

	// Type assert to get staked amount
	// The stake interface should have a StakedAmount field
	stakeAny, ok := stake.(interface{ GetStakedAmount() math.Int })
	if !ok {
		k.Logger(ctx).Error("stake does not implement GetStakedAmount")
		return math.ZeroInt(), fmt.Errorf("invalid stake type")
	}

	stakedAmount := stakeAny.GetStakedAmount()
	if stakedAmount.IsZero() {
		return math.ZeroInt(), nil
	}

	// Calculate beneficiary's share
	shareAmount := percentage.MulInt(stakedAmount).TruncateInt()
	if shareAmount.IsZero() {
		return math.ZeroInt(), nil
	}

	// Initiate unstaking with beneficiary as recipient
	// The unbonding period still applies, but funds will go to beneficiary
	if err := k.stakingKeeper.UnstakeForInheritance(ctx, ownerAddr, beneficiary, shareAmount); err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to initiate unstaking: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStakedAssetsTransferred,
			sdk.NewAttribute(types.AttributeKeyOwner, owner),
			sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary),
			sdk.NewAttribute(types.AttributeKeyAmount, shareAmount.String()),
			sdk.NewAttribute(types.AttributeKeyPercentage, percentage.String()),
		),
	)

	k.Logger(ctx).Info("staked assets transferred",
		"from", owner,
		"to", beneficiary,
		"amount", shareAmount.String(),
		"note", "unbonding period applies, funds will go to beneficiary",
	)

	return shareAmount, nil
}

// TransferLoanPositions handles loan inheritance
// Transfers both borrower and lender positions to beneficiary
func (k Keeper) TransferLoanPositions(ctx sdk.Context, owner string, beneficiary string) error {
	// SECURITY FIX: Check if beneficiary is banned before transferring loan positions
	_, err := sdk.AccAddressFromBech32(beneficiary)
	if err != nil {
		return err
	}

	if k.IsBeneficiaryValid(ctx, beneficiary) != nil {
		k.Logger(ctx).Warn("cannot transfer loan positions to banned beneficiary",
			"owner", owner,
			"beneficiary", beneficiary,
		)
		return types.ErrBeneficiaryBanned
	}

	// Get all loans where owner is borrower or lender
	loans := k.lendingKeeper.GetUserLoans(ctx, owner)
	if len(loans) == 0 {
		return nil // No loans to transfer
	}

	for _, loanInterface := range loans {
		// Type assert to get loan details
		loan, ok := loanInterface.(interface {
			GetID() uint64
			GetBorrower() string
			GetLender() string
			GetStatus() string
		})
		if !ok {
			k.Logger(ctx).Error("invalid loan type", "loan", loanInterface)
			continue
		}

		loanID := loan.GetID()
		status := loan.GetStatus()

		// Only transfer active or pending loans
		if status != "active" && status != "pending" {
			continue
		}

		// Transfer borrower position if owner is borrower
		if loan.GetBorrower() == owner {
			if err := k.lendingKeeper.TransferBorrowerPosition(ctx, loanID, owner, beneficiary); err != nil {
				k.Logger(ctx).Error("failed to transfer borrower position",
					"loan_id", loanID,
					"from", owner,
					"to", beneficiary,
					"error", err,
				)
				continue
			}

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeLoanPositionTransferred,
					sdk.NewAttribute("loan_id", fmt.Sprintf("%d", loanID)),
					sdk.NewAttribute("position", "borrower"),
					sdk.NewAttribute(types.AttributeKeyOwner, owner),
					sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary),
				),
			)

			k.Logger(ctx).Info("loan borrower position inherited",
				"loan_id", loanID,
				"new_owner", beneficiary,
				"note", "beneficiary now owns collateral AND owes the debt",
			)
		}

		// Transfer lender position if owner is lender
		if loan.GetLender() == owner {
			if err := k.lendingKeeper.TransferLenderPosition(ctx, loanID, owner, beneficiary); err != nil {
				k.Logger(ctx).Error("failed to transfer lender position",
					"loan_id", loanID,
					"from", owner,
					"to", beneficiary,
					"error", err,
				)
				continue
			}

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeLoanPositionTransferred,
					sdk.NewAttribute("loan_id", fmt.Sprintf("%d", loanID)),
					sdk.NewAttribute("position", "lender"),
					sdk.NewAttribute(types.AttributeKeyOwner, owner),
					sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary),
				),
			)

			k.Logger(ctx).Info("loan lender position inherited",
				"loan_id", loanID,
				"new_owner", beneficiary,
				"note", "beneficiary will receive future repayments",
			)
		}
	}

	return nil
}

// TransferEscrowPositions handles escrow inheritance
// Transfers both buyer and seller positions to beneficiary
func (k Keeper) TransferEscrowPositions(ctx sdk.Context, owner string, beneficiary string) error {
	// SECURITY FIX: Check if beneficiary is banned before transferring escrow positions
	_, err := sdk.AccAddressFromBech32(beneficiary)
	if err != nil {
		return err
	}

	if k.IsBeneficiaryValid(ctx, beneficiary) != nil {
		k.Logger(ctx).Warn("cannot transfer escrow positions to banned beneficiary",
			"owner", owner,
			"beneficiary", beneficiary,
		)
		return types.ErrBeneficiaryBanned
	}

	// Get all escrows
	escrows := k.escrowKeeper.GetAllEscrows(ctx)
	if len(escrows) == 0 {
		return nil // No escrows to transfer
	}

	for _, escrowInterface := range escrows {
		// Type assert to get escrow details
		escrow, ok := escrowInterface.(interface {
			GetID() uint64
			GetSender() string
			GetRecipient() string
			GetStatus() string
		})
		if !ok {
			k.Logger(ctx).Error("invalid escrow type", "escrow", escrowInterface)
			continue
		}

		escrowID := escrow.GetID()
		status := escrow.GetStatus()

		// Only transfer active escrows (disputed escrows are blocked by escrow keeper)
		if status != "active" && status != "funded" {
			continue
		}

		// Transfer buyer position if owner is sender (buyer)
		if escrow.GetSender() == owner {
			if err := k.escrowKeeper.TransferBuyerPosition(ctx, escrowID, owner, beneficiary); err != nil {
				k.Logger(ctx).Error("failed to transfer buyer position",
					"escrow_id", escrowID,
					"from", owner,
					"to", beneficiary,
					"error", err,
				)
				continue
			}

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEscrowPositionTransferred,
					sdk.NewAttribute("escrow_id", fmt.Sprintf("%d", escrowID)),
					sdk.NewAttribute("position", "buyer"),
					sdk.NewAttribute(types.AttributeKeyOwner, owner),
					sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary),
				),
			)

			k.Logger(ctx).Info("escrow buyer position inherited",
				"escrow_id", escrowID,
				"new_owner", beneficiary,
				"note", "beneficiary expects delivery and can release/dispute",
			)
		}

		// Transfer seller position if owner is recipient (seller)
		if escrow.GetRecipient() == owner {
			if err := k.escrowKeeper.TransferSellerPosition(ctx, escrowID, owner, beneficiary); err != nil {
				k.Logger(ctx).Error("failed to transfer seller position",
					"escrow_id", escrowID,
					"from", owner,
					"to", beneficiary,
					"error", err,
				)
				continue
			}

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEscrowPositionTransferred,
					sdk.NewAttribute("escrow_id", fmt.Sprintf("%d", escrowID)),
					sdk.NewAttribute("position", "seller"),
					sdk.NewAttribute(types.AttributeKeyOwner, owner),
					sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary),
				),
			)

			k.Logger(ctx).Info("escrow seller position inherited",
				"escrow_id", escrowID,
				"new_owner", beneficiary,
				"note", "beneficiary must deliver and receives payment on completion",
			)
		}
	}

	return nil
}

// TransferSpecificAsset transfers a specific asset to a beneficiary
func (k Keeper) TransferSpecificAsset(ctx sdk.Context, owner, beneficiary sdk.AccAddress, asset types.SpecificAsset) ([]types.TransferredAsset, error) {
	var transferred []types.TransferredAsset

	switch asset.AssetType {
	case types.AssetTypeHODL, types.AssetTypeCustomToken:
		// Transfer tokens
		if asset.Amount.IsPositive() {
			coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, asset.Amount))
			err := k.bankKeeper.SendCoins(ctx, owner, beneficiary, coins)
			if err != nil {
				return nil, fmt.Errorf("failed to transfer tokens: %w", err)
			}

			transferred = append(transferred, types.TransferredAsset{
				AssetType:     asset.AssetType,
				Denom:         asset.Denom,
				Amount:        asset.Amount,
				TransferredAt: ctx.BlockTime(),
			})

			k.Logger(ctx).Info("tokens transferred",
				"from", owner.String(),
				"to", beneficiary.String(),
				"denom", asset.Denom,
				"amount", asset.Amount,
			)
		}

	case types.AssetTypeEquity:
		// Transfer equity shares
		if k.equityKeeper != nil && asset.Shares.IsPositive() {
			err := k.equityKeeper.TransferShares(ctx, asset.CompanyID, asset.ClassID, owner.String(), beneficiary.String(), asset.Shares)
			if err != nil {
				return nil, fmt.Errorf("failed to transfer equity: %w", err)
			}

			transferred = append(transferred, types.TransferredAsset{
				AssetType:     types.AssetTypeEquity,
				CompanyID:     asset.CompanyID,
				ClassID:       asset.ClassID,
				Shares:        asset.Shares,
				TransferredAt: ctx.BlockTime(),
			})

			k.Logger(ctx).Info("equity transferred",
				"from", owner.String(),
				"to", beneficiary.String(),
				"company_id", asset.CompanyID,
				"class_id", asset.ClassID,
				"shares", asset.Shares,
			)
		}
	}

	return transferred, nil
}

// TransferPercentageOfAssets transfers a percentage of all assets to a beneficiary
func (k Keeper) TransferPercentageOfAssets(ctx sdk.Context, owner, beneficiary sdk.AccAddress, percentage math.LegacyDec) ([]types.TransferredAsset, error) {
	return k.transferBankBalances(ctx, owner.String(), beneficiary.String(), percentage)
}

// TransferToCharity transfers unclaimed assets to charity
func (k Keeper) TransferToCharity(ctx sdk.Context, plan types.InheritancePlan) error {
	owner, _ := sdk.AccAddressFromBech32(plan.Owner)

	var charityAddr sdk.AccAddress
	var err error

	// Determine charity address
	if plan.CharityAddress != "" {
		charityAddr, err = sdk.AccAddressFromBech32(plan.CharityAddress)
		if err != nil {
			return fmt.Errorf("invalid charity address: %w", err)
		}
	} else {
		// Use default charity address (community pool module)
		params := k.GetParams(ctx)
		if params.DefaultCharityAddress != "" {
			charityAddr, err = sdk.AccAddressFromBech32(params.DefaultCharityAddress)
			if err != nil {
				// Fall back to community pool
				charityAddr = k.accountKeeper.GetModuleAddress("distribution")
			}
		} else {
			charityAddr = k.accountKeeper.GetModuleAddress("distribution")
		}
	}

	if charityAddr == nil {
		return fmt.Errorf("no valid charity address available")
	}

	// Get all remaining balances
	balances := k.bankKeeper.GetAllBalances(ctx, owner)

	// Transfer all tokens to charity
	for _, balance := range balances {
		if balance.Amount.IsZero() {
			continue
		}

		coins := sdk.NewCoins(balance)
		err := k.bankKeeper.SendCoins(ctx, owner, charityAddr, coins)
		if err != nil {
			k.Logger(ctx).Error("failed to transfer to charity",
				"from", owner.String(),
				"to", charityAddr.String(),
				"denom", balance.Denom,
				"amount", balance.Amount,
				"error", err,
			)
			continue
		}

		k.Logger(ctx).Info("assets transferred to charity",
			"from", owner.String(),
			"to", charityAddr.String(),
			"denom", balance.Denom,
			"amount", balance.Amount,
		)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAssetsToCharity,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(plan.PlanID)),
			sdk.NewAttribute(types.AttributeKeyOwner, plan.Owner),
			sdk.NewAttribute(types.AttributeKeyCharityAddress, charityAddr.String()),
		),
	)

	return nil
}

// HandleLockedAssets locks assets during the inheritance process
func (k Keeper) HandleLockedAssets(ctx sdk.Context, planID uint64, owner sdk.AccAddress) error {
	// Get current assets
	coins, equity, err := k.GetOwnerAssets(ctx, owner)
	if err != nil {
		return err
	}

	// Create locked asset record
	locked := types.LockedAssetDuringInheritance{
		PlanID:       planID,
		Owner:        owner.String(),
		LockedCoins:  coins,
		LockedEquity: equity,
		LockedAt:     ctx.BlockTime(),
	}

	// Store locked assets
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&locked)
	store.Set(types.LockedAssetKey(planID), bz)

	k.Logger(ctx).Info("assets locked for inheritance",
		"plan_id", planID,
		"owner", owner.String(),
		"coins", coins,
	)

	return nil
}

// GetLockedAssets returns locked assets for a plan
func (k Keeper) GetLockedAssets(ctx sdk.Context, planID uint64) (types.LockedAssetDuringInheritance, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LockedAssetKey(planID))
	if bz == nil {
		return types.LockedAssetDuringInheritance{}, false
	}

	var locked types.LockedAssetDuringInheritance
	k.cdc.MustUnmarshal(bz, &locked)
	return locked, true
}

// UnlockAssets releases locked assets
func (k Keeper) UnlockAssets(ctx sdk.Context, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.LockedAssetKey(planID))

	k.Logger(ctx).Info("assets unlocked",
		"plan_id", planID,
	)
}
