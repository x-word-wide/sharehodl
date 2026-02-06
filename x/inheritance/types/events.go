package types

// Event types
const (
	EventTypePlanCreated       = "inheritance_plan_created"
	EventTypePlanUpdated       = "inheritance_plan_updated"
	EventTypePlanCancelled     = "inheritance_plan_cancelled"
	EventTypeSwitchTriggered   = "inheritance_switch_triggered"
	EventTypeSwitchCancelled   = "inheritance_switch_cancelled"
	EventTypeGracePeriodExpired = "inheritance_grace_period_expired"
	EventTypeClaimWindowOpened = "inheritance_claim_window_opened"
	EventTypeClaimWindowClosed = "inheritance_claim_window_closed"
	EventTypeAssetsClaimed     = "inheritance_assets_claimed"
	EventTypeBeneficiarySkipped = "inheritance_beneficiary_skipped"
	EventTypeAssetsToCharity   = "inheritance_assets_to_charity"
	EventTypePlanCompleted     = "inheritance_plan_completed"
	EventTypeActivityRecorded        = "inheritance_activity_recorded"
	EventTypeStakedAssetsTransferred = "inheritance_staked_assets_transferred"
	EventTypeLoanPositionTransferred = "inheritance_loan_position_transferred"
	EventTypeEscrowPositionTransferred = "inheritance_escrow_position_transferred"
)

// Event attribute keys
const (
	AttributeKeyPlanID          = "plan_id"
	AttributeKeyOwner           = "owner"
	AttributeKeyBeneficiary     = "beneficiary"
	AttributeKeyPriority        = "priority"
	AttributeKeyPercentage      = "percentage"
	AttributeKeyInactivityPeriod = "inactivity_period"
	AttributeKeyGracePeriod     = "grace_period"
	AttributeKeyClaimWindow     = "claim_window"
	AttributeKeyTriggeredAt     = "triggered_at"
	AttributeKeyGracePeriodEnd  = "grace_period_end"
	AttributeKeyClaimWindowStart = "claim_window_start"
	AttributeKeyClaimWindowEnd  = "claim_window_end"
	AttributeKeyAssetType       = "asset_type"
	AttributeKeyAmount          = "amount"
	AttributeKeyDenom           = "denom"
	AttributeKeyCompanyID       = "company_id"
	AttributeKeyClassID         = "class_id"
	AttributeKeyShares          = "shares"
	AttributeKeyReason          = "reason"
	AttributeKeyCharityAddress  = "charity_address"
	AttributeKeyStatus          = "status"
	AttributeKeyActivityType    = "activity_type"
	AttributeKeyLastActivity    = "last_activity"
)
