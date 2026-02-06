package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values - ALL governance-controllable via validator voting
const (
	DefaultEpochBlocks            uint64 = 1000  // Distribute rewards every 1000 blocks
	DefaultStakerPoolPercent      uint64 = 50    // 50% to all stakers
	DefaultValidatorPoolPercent   uint64 = 30    // 30% to validators (bonus)
	DefaultGovernancePoolPercent  uint64 = 20    // 20% to governance participants
	DefaultUnbondingPeriodBlocks  uint64 = 50000 // ~3.5 days at 6s blocks
	DefaultMaxReputationDecay     uint64 = 100   // 1% monthly decay (basis points)

	// Slashing defaults (basis points)
	DefaultDowntimeSlashBasisPoints         uint64 = 100  // 1% slash for downtime
	DefaultDoubleSignSlashBasisPoints       uint64 = 500  // 5% slash for double signing
	DefaultBadVerificationSlashBasisPoints  uint64 = 1000 // 10% slash for bad verification

	// Reputation threshold defaults
	DefaultGoodReputationThreshold uint64 = 70 // 70/100 minimum for governance rewards
)

// Params defines the parameters for the universal staking module
// ALL parameters are governance-controllable via validator voting
type Params struct {
	// EpochBlocks is the number of blocks per reward distribution epoch
	EpochBlocks uint64 `json:"epoch_blocks"`

	// Pool distribution percentages (must sum to 100)
	StakerPoolPercent     uint64 `json:"staker_pool_percent"`
	ValidatorPoolPercent  uint64 `json:"validator_pool_percent"`
	GovernancePoolPercent uint64 `json:"governance_pool_percent"`

	// UnbondingPeriodBlocks is the number of blocks to wait before unstaking
	UnbondingPeriodBlocks uint64 `json:"unbonding_period_blocks"`

	// MinStakeAmount is the minimum amount required to stake
	MinStakeAmount math.Int `json:"min_stake_amount"`

	// MaxReputationDecayBasisPoints is the monthly reputation decay rate
	MaxReputationDecayBasisPoints uint64 `json:"max_reputation_decay_basis_points"`

	// MinRewardsDistribution is the minimum treasury balance to trigger distribution
	MinRewardsDistribution math.Int `json:"min_rewards_distribution"`

	// ReputationRecoveryRate is how fast reputation recovers naturally (per epoch)
	ReputationRecoveryRate math.LegacyDec `json:"reputation_recovery_rate"`

	// ============ TIER THRESHOLDS (in uhodl) - Governance Controllable ============
	// Validators can vote to adjust these thresholds

	// HolderThreshold - minimum stake to be a Holder (basic tier)
	HolderThreshold math.Int `json:"holder_threshold"`

	// KeeperThreshold - minimum stake to be a Keeper (can lend/borrow)
	KeeperThreshold math.Int `json:"keeper_threshold"`

	// WardenThreshold - minimum stake to be a Warden (can moderate)
	WardenThreshold math.Int `json:"warden_threshold"`

	// StewardThreshold - minimum stake to be a Steward (large disputes)
	StewardThreshold math.Int `json:"steward_threshold"`

	// ArchonThreshold - minimum stake to be an Archon (slash moderators)
	ArchonThreshold math.Int `json:"archon_threshold"`

	// ValidatorThreshold - minimum stake to be a Validator
	ValidatorThreshold math.Int `json:"validator_threshold"`

	// ============ TIER MULTIPLIERS (basis points, 100 = 1.0x) ============

	// HolderMultiplier - reward multiplier for Holders
	HolderMultiplier uint64 `json:"holder_multiplier"`

	// KeeperMultiplier - reward multiplier for Keepers
	KeeperMultiplier uint64 `json:"keeper_multiplier"`

	// WardenMultiplier - reward multiplier for Wardens
	WardenMultiplier uint64 `json:"warden_multiplier"`

	// StewardMultiplier - reward multiplier for Stewards
	StewardMultiplier uint64 `json:"steward_multiplier"`

	// ArchonMultiplier - reward multiplier for Archons
	ArchonMultiplier uint64 `json:"archon_multiplier"`

	// ValidatorMultiplier - reward multiplier for Validators
	ValidatorMultiplier uint64 `json:"validator_multiplier"`

	// ============ SLASHING PARAMETERS (basis points) - Governance Controllable ============

	// DowntimeSlashBasisPoints - slash fraction for validator downtime (100 = 1%)
	DowntimeSlashBasisPoints uint64 `json:"downtime_slash_basis_points"`

	// DoubleSignSlashBasisPoints - slash fraction for double signing (500 = 5%)
	DoubleSignSlashBasisPoints uint64 `json:"double_sign_slash_basis_points"`

	// BadVerificationSlashBasisPoints - slash fraction for bad business verification (1000 = 10%)
	BadVerificationSlashBasisPoints uint64 `json:"bad_verification_slash_basis_points"`

	// ============ REPUTATION THRESHOLDS - Governance Controllable ============

	// GoodReputationThreshold - minimum score for governance rewards eligibility (0-100)
	GoodReputationThreshold uint64 `json:"good_reputation_threshold"`
}

// ProtoMessage implements proto.Message interface
func (m *Params) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *Params) Reset() { *m = Params{} }

// String implements proto.Message interface
func (m *Params) String() string { return "staking_params" }

// DefaultParams returns default staking parameters
func DefaultParams() Params {
	return Params{
		EpochBlocks:                   DefaultEpochBlocks,
		StakerPoolPercent:             DefaultStakerPoolPercent,
		ValidatorPoolPercent:          DefaultValidatorPoolPercent,
		GovernancePoolPercent:         DefaultGovernancePoolPercent,
		UnbondingPeriodBlocks:         DefaultUnbondingPeriodBlocks,
		MinStakeAmount:                math.NewInt(HolderThreshold),
		MaxReputationDecayBasisPoints: DefaultMaxReputationDecay,
		MinRewardsDistribution:        math.NewInt(1_000_000_000), // 1000 HODL minimum
		ReputationRecoveryRate:        math.LegacyNewDecWithPrec(1, 1), // 0.1 points per epoch

		// Tier thresholds (in uhodl) - governance controllable
		HolderThreshold:    math.NewInt(HolderThresholdDefault),    // 100 HODL
		KeeperThreshold:    math.NewInt(KeeperThresholdDefault),    // 10K HODL
		WardenThreshold:    math.NewInt(WardenThresholdDefault),    // 100K HODL
		StewardThreshold:   math.NewInt(StewardThresholdDefault),   // 1M HODL
		ArchonThreshold:    math.NewInt(ArchonThresholdDefault),    // 10M HODL
		ValidatorThreshold: math.NewInt(ValidatorThresholdDefault), // 50M HODL

		// Tier multipliers (basis points, 100 = 1.0x) - governance controllable
		HolderMultiplier:    HolderMultiplierDefault,    // 1.0x
		KeeperMultiplier:    KeeperMultiplierDefault,    // 1.5x
		WardenMultiplier:    WardenMultiplierDefault,    // 2.0x
		StewardMultiplier:   StewardMultiplierDefault,   // 2.5x
		ArchonMultiplier:    ArchonMultiplierDefault,    // 3.0x
		ValidatorMultiplier: ValidatorMultiplierDefault, // 4.0x

		// Slashing parameters (basis points) - governance controllable
		DowntimeSlashBasisPoints:        DefaultDowntimeSlashBasisPoints,
		DoubleSignSlashBasisPoints:      DefaultDoubleSignSlashBasisPoints,
		BadVerificationSlashBasisPoints: DefaultBadVerificationSlashBasisPoints,

		// Reputation thresholds - governance controllable
		GoodReputationThreshold: DefaultGoodReputationThreshold,
	}
}

// Validate validates the params
func (p Params) Validate() error {
	// Pool percentages must sum to 100
	total := p.StakerPoolPercent + p.ValidatorPoolPercent + p.GovernancePoolPercent
	if total != 100 {
		return fmt.Errorf("pool percentages must sum to 100, got %d", total)
	}

	if p.EpochBlocks == 0 {
		return fmt.Errorf("epoch blocks must be positive")
	}

	if p.UnbondingPeriodBlocks == 0 {
		return fmt.Errorf("unbonding period must be positive")
	}

	if !p.MinStakeAmount.IsPositive() {
		return fmt.Errorf("minimum stake amount must be positive")
	}

	if p.MaxReputationDecayBasisPoints > 1000 {
		return fmt.Errorf("max reputation decay cannot exceed 10%%")
	}

	if p.ReputationRecoveryRate.IsNegative() {
		return fmt.Errorf("reputation recovery rate must be non-negative")
	}

	// Validate tier thresholds are positive and in ascending order
	if !p.HolderThreshold.IsPositive() {
		return fmt.Errorf("holder threshold must be positive")
	}
	if !p.KeeperThreshold.GT(p.HolderThreshold) {
		return fmt.Errorf("keeper threshold must be greater than holder threshold")
	}
	if !p.WardenThreshold.GT(p.KeeperThreshold) {
		return fmt.Errorf("warden threshold must be greater than keeper threshold")
	}
	if !p.StewardThreshold.GT(p.WardenThreshold) {
		return fmt.Errorf("steward threshold must be greater than warden threshold")
	}
	if !p.ArchonThreshold.GT(p.StewardThreshold) {
		return fmt.Errorf("archon threshold must be greater than steward threshold")
	}
	if !p.ValidatorThreshold.GT(p.ArchonThreshold) {
		return fmt.Errorf("validator threshold must be greater than archon threshold")
	}

	// Validate multipliers are reasonable (at least 1x, no more than 10x)
	if p.HolderMultiplier < 100 || p.HolderMultiplier > 1000 {
		return fmt.Errorf("holder multiplier must be between 100 (1x) and 1000 (10x)")
	}
	if p.KeeperMultiplier < 100 || p.KeeperMultiplier > 1000 {
		return fmt.Errorf("keeper multiplier must be between 100 (1x) and 1000 (10x)")
	}
	if p.WardenMultiplier < 100 || p.WardenMultiplier > 1000 {
		return fmt.Errorf("warden multiplier must be between 100 (1x) and 1000 (10x)")
	}
	if p.StewardMultiplier < 100 || p.StewardMultiplier > 1000 {
		return fmt.Errorf("steward multiplier must be between 100 (1x) and 1000 (10x)")
	}
	if p.ArchonMultiplier < 100 || p.ArchonMultiplier > 1000 {
		return fmt.Errorf("archon multiplier must be between 100 (1x) and 1000 (10x)")
	}
	if p.ValidatorMultiplier < 100 || p.ValidatorMultiplier > 1000 {
		return fmt.Errorf("validator multiplier must be between 100 (1x) and 1000 (10x)")
	}

	// Validate slashing parameters (cannot exceed 100%)
	if p.DowntimeSlashBasisPoints > 10000 {
		return fmt.Errorf("downtime slash cannot exceed 100%%")
	}
	if p.DoubleSignSlashBasisPoints > 10000 {
		return fmt.Errorf("double sign slash cannot exceed 100%%")
	}
	if p.BadVerificationSlashBasisPoints > 10000 {
		return fmt.Errorf("bad verification slash cannot exceed 100%%")
	}

	// Validate reputation threshold (0-100)
	if p.GoodReputationThreshold > 100 {
		return fmt.Errorf("good reputation threshold cannot exceed 100")
	}

	return nil
}

// GetStakerPoolMultiplier returns the staker pool percentage as a decimal
func (p Params) GetStakerPoolMultiplier() math.LegacyDec {
	return math.LegacyNewDec(int64(p.StakerPoolPercent)).QuoInt64(100)
}

// GetValidatorPoolMultiplier returns the validator pool percentage as a decimal
func (p Params) GetValidatorPoolMultiplier() math.LegacyDec {
	return math.LegacyNewDec(int64(p.ValidatorPoolPercent)).QuoInt64(100)
}

// GetGovernancePoolMultiplier returns the governance pool percentage as a decimal
func (p Params) GetGovernancePoolMultiplier() math.LegacyDec {
	return math.LegacyNewDec(int64(p.GovernancePoolPercent)).QuoInt64(100)
}

// ShouldDistribute checks if it's time to distribute rewards
func (p Params) ShouldDistribute(blockHeight uint64) bool {
	return blockHeight > 0 && blockHeight%p.EpochBlocks == 0
}
