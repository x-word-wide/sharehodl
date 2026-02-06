package types

import (
	"time"

	"cosmossdk.io/math"
)

// StakeTier represents the different staking tiers
// Order: Holder (lowest) → Validator (highest)
type StakeTier int32

const (
	TierHolder    StakeTier = iota // 100+ HODL - Basic: trade, vote
	TierKeeper                     // 10K+ HODL - Small biz verification, can lend
	TierWarden                     // 100K+ HODL - Medium biz verification
	TierSteward                    // 1M+ HODL - Large biz verification
	TierArchon                     // 10M+ HODL - All verification
	TierValidator                  // 50M+ HODL - Block production
)

// Default tier thresholds in uhodl (micro HODL)
// These are the DEFAULT values - actual values are governance-controlled via Params
const (
	HolderThresholdDefault    int64 = 100_000_000       // 100 HODL
	KeeperThresholdDefault    int64 = 10_000_000_000    // 10K HODL
	WardenThresholdDefault    int64 = 100_000_000_000   // 100K HODL
	StewardThresholdDefault   int64 = 1_000_000_000_000 // 1M HODL
	ArchonThresholdDefault    int64 = 10_000_000_000_000 // 10M HODL
	ValidatorThresholdDefault int64 = 50_000_000_000_000 // 50M HODL
)

// Legacy constants for backwards compatibility
const (
	HolderThreshold    = HolderThresholdDefault
	KeeperThreshold    = KeeperThresholdDefault
	WardenThreshold    = WardenThresholdDefault
	StewardThreshold   = StewardThresholdDefault
	ArchonThreshold    = ArchonThresholdDefault
	ValidatorThreshold = ValidatorThresholdDefault
)

// Testnet thresholds (1000x lower for testing) - used when isTestnet=true
const (
	TestnetHolderThreshold    int64 = 100_000           // 0.1 HODL
	TestnetKeeperThreshold    int64 = 10_000_000        // 10 HODL
	TestnetWardenThreshold    int64 = 100_000_000       // 100 HODL
	TestnetStewardThreshold   int64 = 1_000_000_000     // 1K HODL
	TestnetArchonThreshold    int64 = 10_000_000_000    // 10K HODL
	TestnetValidatorThreshold int64 = 50_000_000_000    // 50K HODL
)

// Default reward multipliers (basis points, 100 = 1x)
// These are the DEFAULT values - actual values are governance-controlled via Params
const (
	HolderMultiplierDefault    uint64 = 100 // 1.0x
	KeeperMultiplierDefault    uint64 = 150 // 1.5x
	WardenMultiplierDefault    uint64 = 200 // 2.0x
	StewardMultiplierDefault   uint64 = 250 // 2.5x
	ArchonMultiplierDefault    uint64 = 300 // 3.0x
	ValidatorMultiplierDefault uint64 = 400 // 4.0x
)

// Legacy constants for backwards compatibility
const (
	HolderMultiplier    = HolderMultiplierDefault
	KeeperMultiplier    = KeeperMultiplierDefault
	WardenMultiplier    = WardenMultiplierDefault
	StewardMultiplier   = StewardMultiplierDefault
	ArchonMultiplier    = ArchonMultiplierDefault
	ValidatorMultiplier = ValidatorMultiplierDefault
)

// Slash risk percentages (basis points)
const (
	HolderSlashRisk    uint64 = 100  // 1% max slash
	KeeperSlashRisk    uint64 = 300  // 3% max slash
	WardenSlashRisk    uint64 = 500  // 5% max slash
	StewardSlashRisk   uint64 = 1000 // 10% max slash
	ArchonSlashRisk    uint64 = 1500 // 15% max slash
	ValidatorSlashRisk uint64 = 2000 // 20% max slash
)

// String returns the tier name
func (t StakeTier) String() string {
	switch t {
	case TierHolder:
		return "Holder"
	case TierKeeper:
		return "Keeper"
	case TierWarden:
		return "Warden"
	case TierSteward:
		return "Steward"
	case TierArchon:
		return "Archon"
	case TierValidator:
		return "Validator"
	default:
		return "Unknown"
	}
}

// GetThreshold returns the minimum stake for this tier
func (t StakeTier) GetThreshold(isTestnet bool) math.Int {
	if isTestnet {
		switch t {
		case TierHolder:
			return math.NewInt(TestnetHolderThreshold)
		case TierKeeper:
			return math.NewInt(TestnetKeeperThreshold)
		case TierWarden:
			return math.NewInt(TestnetWardenThreshold)
		case TierSteward:
			return math.NewInt(TestnetStewardThreshold)
		case TierArchon:
			return math.NewInt(TestnetArchonThreshold)
		case TierValidator:
			return math.NewInt(TestnetValidatorThreshold)
		default:
			return math.ZeroInt()
		}
	}

	switch t {
	case TierHolder:
		return math.NewInt(HolderThreshold)
	case TierKeeper:
		return math.NewInt(KeeperThreshold)
	case TierWarden:
		return math.NewInt(WardenThreshold)
	case TierSteward:
		return math.NewInt(StewardThreshold)
	case TierArchon:
		return math.NewInt(ArchonThreshold)
	case TierValidator:
		return math.NewInt(ValidatorThreshold)
	default:
		return math.ZeroInt()
	}
}

// GetMultiplier returns the reward multiplier for this tier (basis points)
func (t StakeTier) GetMultiplier() uint64 {
	switch t {
	case TierHolder:
		return HolderMultiplier
	case TierKeeper:
		return KeeperMultiplier
	case TierWarden:
		return WardenMultiplier
	case TierSteward:
		return StewardMultiplier
	case TierArchon:
		return ArchonMultiplier
	case TierValidator:
		return ValidatorMultiplier
	default:
		return 100
	}
}

// GetSlashRisk returns the maximum slash percentage for this tier (basis points)
func (t StakeTier) GetSlashRisk() uint64 {
	switch t {
	case TierHolder:
		return HolderSlashRisk
	case TierKeeper:
		return KeeperSlashRisk
	case TierWarden:
		return WardenSlashRisk
	case TierSteward:
		return StewardSlashRisk
	case TierArchon:
		return ArchonSlashRisk
	case TierValidator:
		return ValidatorSlashRisk
	default:
		return 100
	}
}

// GetTierFromStake determines tier based on staked amount
func GetTierFromStake(amount math.Int, isTestnet bool) StakeTier {
	stake := amount.Int64()

	if isTestnet {
		if stake >= TestnetValidatorThreshold {
			return TierValidator
		} else if stake >= TestnetArchonThreshold {
			return TierArchon
		} else if stake >= TestnetStewardThreshold {
			return TierSteward
		} else if stake >= TestnetWardenThreshold {
			return TierWarden
		} else if stake >= TestnetKeeperThreshold {
			return TierKeeper
		} else if stake >= TestnetHolderThreshold {
			return TierHolder
		}
	} else {
		if stake >= ValidatorThreshold {
			return TierValidator
		} else if stake >= ArchonThreshold {
			return TierArchon
		} else if stake >= StewardThreshold {
			return TierSteward
		} else if stake >= WardenThreshold {
			return TierWarden
		} else if stake >= KeeperThreshold {
			return TierKeeper
		} else if stake >= HolderThreshold {
			return TierHolder
		}
	}

	return StakeTier(-1) // Below minimum
}

// GetTierFromStakeWithParams determines tier based on governance-controlled thresholds
// This is the preferred method as it uses the current governance-approved thresholds
func GetTierFromStakeWithParams(amount math.Int, params Params) StakeTier {
	if amount.GTE(params.ValidatorThreshold) {
		return TierValidator
	} else if amount.GTE(params.ArchonThreshold) {
		return TierArchon
	} else if amount.GTE(params.StewardThreshold) {
		return TierSteward
	} else if amount.GTE(params.WardenThreshold) {
		return TierWarden
	} else if amount.GTE(params.KeeperThreshold) {
		return TierKeeper
	} else if amount.GTE(params.HolderThreshold) {
		return TierHolder
	}
	return StakeTier(-1) // Below minimum
}

// GetMultiplierWithParams returns the reward multiplier from governance params
func (t StakeTier) GetMultiplierWithParams(params Params) uint64 {
	switch t {
	case TierHolder:
		return params.HolderMultiplier
	case TierKeeper:
		return params.KeeperMultiplier
	case TierWarden:
		return params.WardenMultiplier
	case TierSteward:
		return params.StewardMultiplier
	case TierArchon:
		return params.ArchonMultiplier
	case TierValidator:
		return params.ValidatorMultiplier
	default:
		return 100
	}
}

// UserStake represents a user's stake in the system
type UserStake struct {
	Owner           string         `json:"owner"`
	StakedAmount    math.Int       `json:"staked_amount"`
	Tier            StakeTier      `json:"tier"`
	ReputationScore math.LegacyDec `json:"reputation_score"` // 0-100
	StakedAt        time.Time      `json:"staked_at"`
	LastRewardClaim time.Time      `json:"last_reward_claim"`
	PendingRewards  math.Int       `json:"pending_rewards"`
	TotalRewardsClaimed math.Int   `json:"total_rewards_claimed"`
	IsValidator     bool           `json:"is_validator"` // Running a validator node
	ValidatorAddr   string         `json:"validator_addr,omitempty"`
}

// ProtoMessage implements proto.Message interface
func (m *UserStake) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *UserStake) Reset() { *m = UserStake{} }

// String implements proto.Message interface
func (m *UserStake) String() string { return m.Owner }

// NewUserStake creates a new user stake
func NewUserStake(owner string, amount math.Int, tier StakeTier) UserStake {
	return UserStake{
		Owner:           owner,
		StakedAmount:    amount,
		Tier:            tier,
		ReputationScore: math.LegacyNewDec(100), // Start with perfect reputation
		StakedAt:        time.Now(),
		LastRewardClaim: time.Now(),
		PendingRewards:  math.ZeroInt(),
		TotalRewardsClaimed: math.ZeroInt(),
		IsValidator:     false,
		ValidatorAddr:   "",
	}
}

// CalculateWeight calculates the reward weight for this stake
// Weight = StakedAmount * TierMultiplier * (ReputationScore / 100)
func (s UserStake) CalculateWeight() math.LegacyDec {
	stakeDecimal := math.LegacyNewDecFromInt(s.StakedAmount)
	multiplier := math.LegacyNewDec(int64(s.Tier.GetMultiplier())).QuoInt64(100)
	reputationFactor := s.ReputationScore.QuoInt64(100)

	return stakeDecimal.Mul(multiplier).Mul(reputationFactor)
}

// CanVerifyBusiness checks if user can verify businesses of given size
func (s UserStake) CanVerifyBusiness(businessSize string) bool {
	switch businessSize {
	case "small":
		return s.Tier >= TierKeeper
	case "medium":
		return s.Tier >= TierWarden
	case "large":
		return s.Tier >= TierSteward
	case "enterprise":
		return s.Tier >= TierArchon
	default:
		return false
	}
}

// CanLend checks if user can participate in lending
func (s UserStake) CanLend() bool {
	return s.Tier >= TierKeeper
}

// CanBeLender checks if user can be a lender (provide liquidity)
func (s UserStake) CanBeLender() bool {
	return s.Tier >= TierKeeper && s.ReputationScore.GTE(math.LegacyNewDec(60))
}

// CanBeBorrower checks if user can borrow
func (s UserStake) CanBeBorrower() bool {
	return s.Tier >= TierKeeper && s.ReputationScore.GTE(math.LegacyNewDec(50))
}

// CanBeValidator checks if user meets validator requirements
func (s UserStake) CanBeValidator() bool {
	return s.Tier >= TierValidator && s.ReputationScore.GTE(math.LegacyNewDec(80))
}

// CanModerate checks if user can be a moderator (resolve disputes)
// Requires Warden+ tier (100K HODL)
func (s UserStake) CanModerate() bool {
	return s.Tier >= TierWarden
}

// CanModerateLarge checks if user can moderate large disputes
// Requires Steward+ tier (1M HODL)
func (s UserStake) CanModerateLarge() bool {
	return s.Tier >= TierSteward
}

// CanSlashModerators checks if user can slash other moderators
// Requires Archon+ tier (10M HODL)
func (s UserStake) CanSlashModerators() bool {
	return s.Tier >= TierArchon
}

// CanBlacklistModerators checks if user can blacklist moderators
// Requires Validator tier (50M HODL)
func (s UserStake) CanBlacklistModerators() bool {
	return s.Tier >= TierValidator
}

// EpochInfo tracks reward distribution epochs
type EpochInfo struct {
	CurrentEpoch      uint64    `json:"current_epoch"`
	EpochStartHeight  int64     `json:"epoch_start_height"`
	EpochStartTime    time.Time `json:"epoch_start_time"`
	TotalDistributed  math.Int  `json:"total_distributed"`
	LastDistribution  time.Time `json:"last_distribution"`
}

// ProtoMessage implements proto.Message interface
func (m *EpochInfo) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *EpochInfo) Reset() { *m = EpochInfo{} }

// String implements proto.Message interface
func (m *EpochInfo) String() string { return "epoch_info" }

// TierStats holds statistics for a tier
type TierStats struct {
	Tier         StakeTier `json:"tier"`
	MemberCount  uint64    `json:"member_count"`
	TotalStaked  math.Int  `json:"total_staked"`
	TotalWeight  math.LegacyDec `json:"total_weight"`
}
