package types

import (
	"cosmossdk.io/errors"
)

// x/lending module sentinel errors
var (
	// Loan errors
	ErrLoanNotFound           = errors.Register(ModuleName, 1, "loan not found")
	ErrLoanAlreadyExists      = errors.Register(ModuleName, 2, "loan already exists")
	ErrInvalidLoan            = errors.Register(ModuleName, 3, "invalid loan")
	ErrLoanNotActive          = errors.Register(ModuleName, 4, "loan is not active")
	ErrLoanAlreadyRepaid      = errors.Register(ModuleName, 5, "loan already repaid")
	ErrLoanDefaulted          = errors.Register(ModuleName, 6, "loan has defaulted")
	ErrLoanOverdue            = errors.Register(ModuleName, 7, "loan is overdue")
	ErrLoanNotLiquidatable    = errors.Register(ModuleName, 8, "loan is not liquidatable")

	// Collateral errors
	ErrInsufficientCollateral = errors.Register(ModuleName, 10, "insufficient collateral")
	ErrInvalidCollateral      = errors.Register(ModuleName, 11, "invalid collateral")
	ErrCollateralNotSupported = errors.Register(ModuleName, 12, "collateral type not supported")
	ErrCollateralLocked       = errors.Register(ModuleName, 13, "collateral is locked")
	ErrCollateralRatioTooLow  = errors.Register(ModuleName, 14, "collateral ratio too low")

	// Pool errors
	ErrPoolNotFound           = errors.Register(ModuleName, 20, "lending pool not found")
	ErrPoolAlreadyExists      = errors.Register(ModuleName, 21, "lending pool already exists")
	ErrInvalidPool            = errors.Register(ModuleName, 22, "invalid lending pool")
	ErrInsufficientLiquidity  = errors.Register(ModuleName, 23, "insufficient pool liquidity")
	ErrPoolPaused             = errors.Register(ModuleName, 24, "lending pool is paused")

	// Deposit/withdrawal errors
	ErrDepositNotFound        = errors.Register(ModuleName, 30, "deposit not found")
	ErrInvalidDeposit         = errors.Register(ModuleName, 31, "invalid deposit amount")
	ErrInvalidWithdrawal      = errors.Register(ModuleName, 32, "invalid withdrawal amount")
	ErrWithdrawalLimitExceeded = errors.Register(ModuleName, 33, "withdrawal limit exceeded")

	// Offer/request errors
	ErrOfferNotFound          = errors.Register(ModuleName, 40, "loan offer not found")
	ErrOfferExpired           = errors.Register(ModuleName, 41, "loan offer expired")
	ErrOfferNotActive         = errors.Register(ModuleName, 42, "loan offer not active")
	ErrRequestNotFound        = errors.Register(ModuleName, 43, "loan request not found")
	ErrRequestExpired         = errors.Register(ModuleName, 44, "loan request expired")

	// Repayment errors
	ErrInvalidRepayment       = errors.Register(ModuleName, 50, "invalid repayment amount")
	ErrRepaymentExceedsOwed   = errors.Register(ModuleName, 51, "repayment exceeds amount owed")

	// Interest errors
	ErrInvalidInterestRate    = errors.Register(ModuleName, 60, "invalid interest rate")
	ErrInterestAccrualFailed  = errors.Register(ModuleName, 61, "interest accrual failed")

	// Authorization errors
	ErrUnauthorized           = errors.Register(ModuleName, 70, "unauthorized")
	ErrNotBorrower            = errors.Register(ModuleName, 71, "not the loan borrower")
	ErrNotLender              = errors.Register(ModuleName, 72, "not the loan lender")

	// Asset errors
	ErrInsufficientFunds      = errors.Register(ModuleName, 80, "insufficient funds")
	ErrInvalidAsset           = errors.Register(ModuleName, 81, "invalid asset")
	ErrTransferFailed         = errors.Register(ModuleName, 82, "asset transfer failed")

	// Price errors
	ErrPriceNotAvailable = errors.Register(ModuleName, 90, "price not available")
	ErrPriceStale        = errors.Register(ModuleName, 91, "price data is stale")

	// Stake-as-Trust-Ceiling: P2P lending stake errors
	ErrOfferExceedsTrustCeiling   = errors.Register(ModuleName, 100, "offer value exceeds lender stake (trust ceiling)")
	ErrRequestExceedsTrustCeiling = errors.Register(ModuleName, 101, "request value exceeds borrower stake (trust ceiling)")
	ErrLenderStakeNotFound        = errors.Register(ModuleName, 102, "lender stake not found")
	ErrBorrowerStakeNotFound      = errors.Register(ModuleName, 103, "borrower stake not found")
	ErrInsufficientLenderStake    = errors.Register(ModuleName, 104, "insufficient lender stake")
	ErrInsufficientBorrowerStake  = errors.Register(ModuleName, 105, "insufficient borrower stake")
	ErrStakeUnbondingInProgress   = errors.Register(ModuleName, 106, "stake unbonding in progress")
	ErrActiveOffersExist          = errors.Register(ModuleName, 107, "cannot unstake with active offers")
	ErrActiveRequestsExist        = errors.Register(ModuleName, 108, "cannot unstake with active requests")
	ErrLenderBlacklisted          = errors.Register(ModuleName, 109, "lender is blacklisted")
	ErrBorrowerBlacklisted        = errors.Register(ModuleName, 110, "borrower is blacklisted")
	ErrStakeUnbondingNotComplete  = errors.Register(ModuleName, 111, "stake unbonding period not complete")
	ErrMinimumStakeRequired       = errors.Register(ModuleName, 112, "minimum stake amount required")

	// Validator oversight errors
	ErrUnauthorizedValidatorAction = errors.Register(ModuleName, 120, "unauthorized validator action")
	ErrValidatorNotActive          = errors.Register(ModuleName, 121, "validator is not active")
	ErrValidatorTierTooLow         = errors.Register(ModuleName, 122, "validator tier too low for this action")
	ErrAlreadyBlacklisted          = errors.Register(ModuleName, 123, "participant already blacklisted")
	ErrNotBlacklisted              = errors.Register(ModuleName, 124, "participant not blacklisted")
	ErrBanNotExpired               = errors.Register(ModuleName, 125, "temporary ban has not expired")
	ErrInvalidSlashFraction        = errors.Register(ModuleName, 126, "invalid slash fraction (must be 0-1)")

	// Universal staking tier requirement errors
	ErrInsufficientTierToLend   = errors.Register(ModuleName, 130, "must be Keeper tier or higher to lend")
	ErrInsufficientTierToBorrow = errors.Register(ModuleName, 131, "must be Keeper tier or higher to borrow")
	ErrStakingKeeperNotSet      = errors.Register(ModuleName, 132, "staking keeper not configured")
)

// Event types
const (
	EventTypeLoanCreated        = "loan_created"
	EventTypeLoanActivated      = "loan_activated"
	EventTypeLoanRepaid         = "loan_repaid"
	EventTypeLoanDefaulted      = "loan_defaulted"
	EventTypeLoanLiquidated     = "loan_liquidated"
	EventTypeInterestAccrued    = "interest_accrued"
	EventTypeCollateralAdded    = "collateral_added"
	EventTypeCollateralRemoved  = "collateral_removed"
	EventTypePoolDeposit        = "pool_deposit"
	EventTypePoolWithdrawal     = "pool_withdrawal"
	EventTypeLoanOfferCreated   = "loan_offer_created"
	EventTypeLoanRequestCreated = "loan_request_created"

	// Stake-as-Trust-Ceiling event types
	EventTypeLenderStakeDeposited   = "lender_stake_deposited"
	EventTypeLenderStakeWithdrawn   = "lender_stake_withdrawn"
	EventTypeLenderUnstakeRequested = "lender_unstake_requested"
	EventTypeBorrowerStakeDeposited = "borrower_stake_deposited"
	EventTypeBorrowerStakeWithdrawn = "borrower_stake_withdrawn"
	EventTypeBorrowerUnstakeRequested = "borrower_unstake_requested"
	EventTypeStakeLocked            = "stake_locked"
	EventTypeStakeReleased          = "stake_released"
	EventTypeOfferCancelled         = "loan_offer_cancelled"
	EventTypeRequestCancelled       = "loan_request_cancelled"

	// Validator oversight event types
	EventTypeLenderSlashed         = "lender_slashed"
	EventTypeBorrowerSlashed       = "borrower_slashed"
	EventTypeLenderBlacklisted     = "lender_blacklisted"
	EventTypeBorrowerBlacklisted   = "borrower_blacklisted"
	EventTypeLenderUnblacklisted   = "lender_unblacklisted"
	EventTypeBorrowerUnblacklisted = "borrower_unblacklisted"
	EventTypeValidatorAction       = "validator_lending_action"
)

// Attribute keys
const (
	AttributeKeyLoanID       = "loan_id"
	AttributeKeyPoolID       = "pool_id"
	AttributeKeyBorrower     = "borrower"
	AttributeKeyLender       = "lender"
	AttributeKeyAmount       = "amount"
	AttributeKeyInterestRate = "interest_rate"
	AttributeKeyCollateral   = "collateral"
	AttributeKeyStatus       = "status"

	// Stake-as-Trust-Ceiling attribute keys
	AttributeKeyOfferID        = "offer_id"
	AttributeKeyRequestID      = "request_id"
	AttributeKeyStakeAmount    = "stake_amount"
	AttributeKeyTrustCeiling   = "trust_ceiling"
	AttributeKeyAvailableStake = "available_stake"
	AttributeKeyLockedStake    = "locked_stake"
	AttributeKeyUnbondingAmount = "unbonding_amount"
	AttributeKeyCompletesAt    = "completes_at"

	// Validator oversight attribute keys
	AttributeKeyValidator     = "validator"
	AttributeKeyTarget        = "target"
	AttributeKeyReason        = "reason"
	AttributeKeySlashAmount   = "slash_amount"
	AttributeKeySlashFraction = "slash_fraction"
	AttributeKeyPermanent     = "permanent"
	AttributeKeyActionID      = "action_id"
	AttributeKeyAction        = "action"
)
