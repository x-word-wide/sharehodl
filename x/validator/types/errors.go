package types

import (
	"cosmossdk.io/errors"
)

// x/validator module sentinel errors
var (
	ErrValidatorTierNotFound        = errors.Register(ModuleName, 1, "validator tier not found")
	ErrInsufficientStakeForTier     = errors.Register(ModuleName, 2, "insufficient stake for requested tier")
	ErrBusinessVerificationNotFound = errors.Register(ModuleName, 3, "business verification not found")
	ErrVerificationAlreadyCompleted = errors.Register(ModuleName, 4, "verification already completed")
	ErrValidatorNotAssigned         = errors.Register(ModuleName, 5, "validator not assigned to this verification")
	ErrInvalidVerificationResult    = errors.Register(ModuleName, 6, "invalid verification result")
	ErrBusinessAlreadyVerified      = errors.Register(ModuleName, 7, "business already verified")
	ErrInsufficientValidators       = errors.Register(ModuleName, 8, "insufficient validators available for verification")
	ErrVerificationExpired          = errors.Register(ModuleName, 9, "verification request has expired")
	ErrInvalidTierUpgrade          = errors.Register(ModuleName, 10, "invalid tier upgrade request")
	ErrBusinessStakeNotFound       = errors.Register(ModuleName, 11, "business stake not found")
	ErrInsufficientRewards         = errors.Register(ModuleName, 12, "insufficient rewards to claim")
	ErrValidatorReputationTooLow   = errors.Register(ModuleName, 13, "validator reputation too low")
	ErrBusinessNotApproved         = errors.Register(ModuleName, 14, "business not approved for staking")
	ErrInvalidStakeAmount          = errors.Register(ModuleName, 15, "invalid stake amount")
	ErrVestingPeriodActive         = errors.Register(ModuleName, 16, "vesting period still active")
	ErrMaxStakeExceeded            = errors.Register(ModuleName, 17, "maximum stake amount exceeded")
	ErrValidatorAlreadyRegistered  = errors.Register(ModuleName, 18, "validator already registered in tier")
	ErrInvalidBusinessDocument     = errors.Register(ModuleName, 19, "invalid business verification document")
	ErrVerificationFeeInsufficient = errors.Register(ModuleName, 20, "verification fee insufficient")
)