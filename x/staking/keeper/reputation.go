package keeper

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/staking/types"
)

// =============================================================================
// REPUTATION MANAGEMENT
// =============================================================================

// UpdateReputation updates a user's reputation based on an action
func (k Keeper) UpdateReputation(
	ctx sdk.Context,
	addr sdk.AccAddress,
	action types.ReputationAction,
	reason string,
	relatedTx string,
) error {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return types.ErrStakeNotFound
	}

	oldScore := stake.ReputationScore
	scoreChange := action.GetScoreChange()
	newScore := oldScore.Add(scoreChange)

	// Clamp to [0, 100]
	if newScore.GT(math.LegacyNewDec(100)) {
		newScore = math.LegacyNewDec(100)
	}
	if newScore.LT(math.LegacyZeroDec()) {
		newScore = math.LegacyZeroDec()
	}

	// Update stake
	stake.ReputationScore = newScore
	if err := k.SetUserStake(ctx, stake); err != nil {
		return err
	}

	// Record history
	record := types.ReputationRecord{
		Owner:       addr.String(),
		Action:      action,
		ScoreChange: scoreChange,
		NewScore:    newScore,
		Reason:      reason,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime(),
		RelatedTx:   relatedTx,
	}
	k.addReputationRecord(ctx, addr, record)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReputationChange,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyReputation, newScore.String()),
			sdk.NewAttribute(types.AttributeKeyReputationChange, scoreChange.String()),
			sdk.NewAttribute(types.AttributeKeyAction, action.String()),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	// Call hooks
	if k.hooks != nil {
		if err := k.hooks.AfterReputationChange(ctx, addr, oldScore, newScore); err != nil {
			k.Logger(ctx).Error("reputation change hook failed", "error", err)
		}
	}

	k.Logger(ctx).Info("reputation updated",
		"user", addr.String(),
		"action", action.String(),
		"change", scoreChange.String(),
		"new_score", newScore.String(),
	)

	return nil
}

// GetReputation returns a user's current reputation score
func (k Keeper) GetReputation(ctx sdk.Context, addr sdk.AccAddress) math.LegacyDec {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return math.LegacyZeroDec()
	}
	return stake.ReputationScore
}

// GetReputationTier returns a user's reputation tier
func (k Keeper) GetReputationTier(ctx sdk.Context, addr sdk.AccAddress) types.ReputationTier {
	score := k.GetReputation(ctx, addr)
	return types.GetReputationTier(score)
}

// MeetsReputationRequirement checks if user meets reputation requirement for an action
func (k Keeper) MeetsReputationRequirement(ctx sdk.Context, addr sdk.AccAddress, action string) bool {
	score := k.GetReputation(ctx, addr)
	return types.MeetsReputationRequirement(score, action)
}

// =============================================================================
// REPUTATION HISTORY
// =============================================================================

// GetReputationHistory returns a user's reputation history
func (k Keeper) GetReputationHistory(ctx sdk.Context, addr sdk.AccAddress) types.ReputationHistory {
	store := ctx.KVStore(k.storeKey)
	key := types.GetReputationHistoryKey(addr)

	bz := store.Get(key)
	if bz == nil {
		return types.ReputationHistory{
			Owner:        addr.String(),
			Records:      []types.ReputationRecord{},
			TotalGains:   math.LegacyZeroDec(),
			TotalLosses:  math.LegacyZeroDec(),
			HighestScore: math.LegacyNewDec(100),
			LowestScore:  math.LegacyNewDec(100),
		}
	}

	var history types.ReputationHistory
	if err := json.Unmarshal(bz, &history); err != nil {
		return types.ReputationHistory{Owner: addr.String()}
	}
	return history
}

func (k Keeper) addReputationRecord(ctx sdk.Context, addr sdk.AccAddress, record types.ReputationRecord) {
	history := k.GetReputationHistory(ctx, addr)

	// Add record (keep last 100 records)
	history.Records = append(history.Records, record)
	if len(history.Records) > 100 {
		history.Records = history.Records[len(history.Records)-100:]
	}

	// Update totals
	if record.ScoreChange.IsPositive() {
		history.TotalGains = history.TotalGains.Add(record.ScoreChange)
	} else {
		history.TotalLosses = history.TotalLosses.Add(record.ScoreChange.Abs())
	}

	// Update high/low
	if record.NewScore.GT(history.HighestScore) {
		history.HighestScore = record.NewScore
	}
	if record.NewScore.LT(history.LowestScore) {
		history.LowestScore = record.NewScore
	}

	// Save
	store := ctx.KVStore(k.storeKey)
	key := types.GetReputationHistoryKey(addr)
	bz, _ := json.Marshal(history)
	store.Set(key, bz)
}

// =============================================================================
// REPUTATION ACTIONS (called by other modules)
// =============================================================================

// RewardSuccessfulVerification rewards a user for successful business verification
func (k Keeper) RewardSuccessfulVerification(ctx sdk.Context, addr sdk.AccAddress, businessID string) error {
	return k.UpdateReputation(ctx, addr, types.ActionSuccessfulVerification,
		"Successful business verification: "+businessID, "")
}

// PenalizeFailedVerification penalizes a user for failed verification
func (k Keeper) PenalizeFailedVerification(ctx sdk.Context, addr sdk.AccAddress, businessID string) error {
	return k.UpdateReputation(ctx, addr, types.ActionFailedVerification,
		"Failed business verification: "+businessID, "")
}

// RewardGoodVote rewards a user for voting with majority on passed proposal
func (k Keeper) RewardGoodVote(ctx sdk.Context, addr sdk.AccAddress, proposalID uint64) error {
	return k.UpdateReputation(ctx, addr, types.ActionGoodVote,
		"Voted with majority on passed proposal", "")
}

// PenalizeBadVote penalizes a user for voting against community
func (k Keeper) PenalizeBadVote(ctx sdk.Context, addr sdk.AccAddress, proposalID uint64) error {
	return k.UpdateReputation(ctx, addr, types.ActionBadVote,
		"Voted against passed proposal", "")
}

// RewardValidatorUptime rewards a validator for uptime
func (k Keeper) RewardValidatorUptime(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.UpdateReputation(ctx, addr, types.ActionUptimeBonus,
		"Validator uptime bonus", "")
}

// PenalizeDowntime penalizes a validator for downtime
func (k Keeper) PenalizeDowntime(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.UpdateReputation(ctx, addr, types.ActionDowntime,
		"Validator downtime penalty", "")
}

// RewardSuccessfulLoan rewards a user for repaying a loan on time
func (k Keeper) RewardSuccessfulLoan(ctx sdk.Context, addr sdk.AccAddress, loanID string) error {
	return k.UpdateReputation(ctx, addr, types.ActionSuccessfulLoan,
		"Loan repaid on time: "+loanID, "")
}

// PenalizeLoanDefault penalizes a user for defaulting on a loan
func (k Keeper) PenalizeLoanDefault(ctx sdk.Context, addr sdk.AccAddress, loanID string) error {
	return k.UpdateReputation(ctx, addr, types.ActionLoanDefault,
		"Loan default: "+loanID, "")
}

// RewardLiquidityProvided rewards a user for providing liquidity
func (k Keeper) RewardLiquidityProvided(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.UpdateReputation(ctx, addr, types.ActionLiquidityProvided,
		"Liquidity provided to DEX", "")
}

// PenalizeSlash penalizes a user for being slashed
func (k Keeper) PenalizeSlash(ctx sdk.Context, addr sdk.AccAddress, reason string) error {
	return k.UpdateReputation(ctx, addr, types.ActionSlashed,
		"Slashed: "+reason, "")
}

// PenalizeFraudAttempt penalizes a user for attempted fraud
func (k Keeper) PenalizeFraudAttempt(ctx sdk.Context, addr sdk.AccAddress, details string) error {
	return k.UpdateReputation(ctx, addr, types.ActionFraudAttempt,
		"Fraud attempt: "+details, "")
}

// =============================================================================
// ESCROW MODERATOR REPUTATION ACTIONS
// =============================================================================

// RewardSuccessfulDispute rewards a moderator for fair dispute resolution
func (k Keeper) RewardSuccessfulDispute(ctx sdk.Context, addr sdk.AccAddress, disputeID string) error {
	return k.UpdateReputation(ctx, addr, types.ActionSuccessfulDispute,
		"Fair dispute resolution: "+disputeID, "")
}

// PenalizeBadDispute penalizes a moderator for unfair resolution
func (k Keeper) PenalizeBadDispute(ctx sdk.Context, addr sdk.AccAddress, disputeID string) error {
	return k.UpdateReputation(ctx, addr, types.ActionUnfairDispute,
		"Unfair dispute resolution overturned: "+disputeID, "")
}

// PenalizeBadModeration penalizes a moderator for pattern of bad moderation
func (k Keeper) PenalizeBadModeration(ctx sdk.Context, addr sdk.AccAddress, reason string) error {
	return k.UpdateReputation(ctx, addr, types.ActionBadModeration,
		"Pattern of unfair moderation: "+reason, "")
}

// =============================================================================
// REPUTATION DECAY & RECOVERY
// =============================================================================

// ApplyReputationDecay applies monthly reputation decay (called from EndBlocker)
func (k Keeper) ApplyReputationDecay(ctx sdk.Context) {
	params := k.GetParams(ctx)

	// Only apply decay once per day (~14400 blocks)
	if ctx.BlockHeight()%14400 != 0 {
		return
	}

	// Monthly decay rate converted to daily
	dailyDecayRate := math.LegacyNewDec(int64(params.MaxReputationDecayBasisPoints)).QuoInt64(10000).QuoInt64(30)

	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		// Only decay if above 100 (perfect score)
		if stake.ReputationScore.GT(math.LegacyNewDec(100)) {
			newScore := stake.ReputationScore.Sub(dailyDecayRate)
			if newScore.LT(math.LegacyNewDec(100)) {
				newScore = math.LegacyNewDec(100)
			}
			stake.ReputationScore = newScore
			k.SetUserStake(ctx, stake)
		}
		return false
	})
}

// ApplyReputationRecovery allows reputation to naturally recover over time
func (k Keeper) ApplyReputationRecovery(ctx sdk.Context) {
	params := k.GetParams(ctx)

	// Only apply recovery once per epoch
	if !params.ShouldDistribute(uint64(ctx.BlockHeight())) {
		return
	}

	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		// Only recover if below 100
		if stake.ReputationScore.LT(math.LegacyNewDec(100)) {
			// Check last activity - only recover if user has been active
			if time.Since(stake.LastRewardClaim) < 7*24*time.Hour {
				newScore := stake.ReputationScore.Add(params.ReputationRecoveryRate)
				if newScore.GT(math.LegacyNewDec(100)) {
					newScore = math.LegacyNewDec(100)
				}
				stake.ReputationScore = newScore
				k.SetUserStake(ctx, stake)
			}
		}
		return false
	})
}
