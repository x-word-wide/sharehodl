package types

import (
	"cosmossdk.io/errors"
)

// x/inheritance module sentinel errors
var (
	ErrInvalidPlan           = errors.Register(ModuleName, 1, "invalid inheritance plan")
	ErrPlanNotFound          = errors.Register(ModuleName, 2, "inheritance plan not found")
	ErrUnauthorized          = errors.Register(ModuleName, 3, "unauthorized operation")
	ErrInvalidBeneficiary    = errors.Register(ModuleName, 4, "invalid beneficiary")
	ErrBeneficiaryBanned     = errors.Register(ModuleName, 5, "beneficiary is banned")
	ErrOwnerBanned           = errors.Register(ModuleName, 6, "owner is banned - dead man switch disabled")
	ErrInvalidInactivity     = errors.Register(ModuleName, 7, "invalid inactivity period")
	ErrInvalidGracePeriod    = errors.Register(ModuleName, 8, "invalid grace period")
	ErrInvalidClaimWindow    = errors.Register(ModuleName, 9, "invalid claim window")
	ErrTriggerNotFound       = errors.Register(ModuleName, 10, "switch trigger not found")
	ErrTriggerNotExpired     = errors.Register(ModuleName, 11, "grace period has not expired")
	ErrClaimNotFound         = errors.Register(ModuleName, 12, "claim not found")
	ErrClaimWindowClosed     = errors.Register(ModuleName, 13, "claim window is closed")
	ErrClaimAlreadyProcessed = errors.Register(ModuleName, 14, "claim already processed")
	ErrInsufficientAssets    = errors.Register(ModuleName, 15, "insufficient assets for transfer")
	ErrAssetTransferFailed   = errors.Register(ModuleName, 16, "asset transfer failed")
	ErrTooManyBeneficiaries  = errors.Register(ModuleName, 17, "too many beneficiaries")
	ErrInvalidPercentage     = errors.Register(ModuleName, 18, "invalid percentage allocation")
	ErrDuplicatePriority     = errors.Register(ModuleName, 19, "duplicate beneficiary priority")
	ErrPlanAlreadyTriggered  = errors.Register(ModuleName, 20, "plan already triggered")
	ErrPlanNotActive         = errors.Register(ModuleName, 21, "plan is not active")
	ErrCannotModifyPlan      = errors.Register(ModuleName, 22, "cannot modify plan in current status")
	ErrInvalidAssetType      = errors.Register(ModuleName, 23, "invalid asset type")
	ErrOwnerStillActive      = errors.Register(ModuleName, 24, "owner is still active - cannot trigger")
	ErrClaimInProgress       = errors.Register(ModuleName, 25, "claim processing already in progress")
	ErrMaxCascadeDepthReached = errors.Register(ModuleName, 26, "maximum cascade depth reached, transferring to charity")
)
