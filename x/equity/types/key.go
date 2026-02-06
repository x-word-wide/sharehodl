package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "equity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key  
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_equity"
)

// Store keys
var (
	// CompanyPrefix stores Company information
	CompanyPrefix = []byte{0x01}
	
	// ShareClassPrefix stores share class information
	ShareClassPrefix = []byte{0x02}
	
	// ShareholdingPrefix stores individual shareholdings
	ShareholdingPrefix = []byte{0x03}
	
	// VestingPrefix stores vesting schedules
	VestingPrefix = []byte{0x05}
	
	// CompanyCounterKey stores the global company counter
	CompanyCounterKey = []byte{0x06}
	
	// Business listing prefixes
	BusinessListingPrefix = []byte{0x07}
	ListingCounterKey     = []byte{0x08}
	ListingRequirementsKey = []byte{0x09}
	ValidatorAssignmentPrefix = []byte{0x0A}
	
	// Dividend system prefixes
	DividendPrefix = []byte{0x04}
	DividendCounterKey = []byte{0x0C}
	DividendPaymentPrefix = []byte{0x0D}
	DividendClaimPrefix = []byte{0x0E}
	DividendSnapshotPrefix = []byte{0x0F}
	ShareholderSnapshotPrefix = []byte{0x10}
	DividendPolicyPrefix = []byte{0x11}
	DividendStatisticsPrefix = []byte{0x12}
	DividendPaymentCounterKey = []byte{0x13}
	DividendClaimCounterKey = []byte{0x14}

	// Anti-dilution system prefixes
	AntiDilutionProvisionPrefix   = []byte{0x20}  // Store anti-dilution provisions
	IssuanceRecordPrefix          = []byte{0x21}  // Store historical issuance records
	IssuanceRecordCounterKey      = []byte{0x22}  // Counter for issuance record IDs
	AntiDilutionAdjustmentPrefix  = []byte{0x23}  // Store adjustment records
	AdjustmentCounterKey          = []byte{0x24}  // Counter for adjustment IDs
	LastIssuancePricePrefix       = []byte{0x25}  // Store last issuance price per class

	// Delisting and compensation prefixes
	DelistingCompensationPrefix   = []byte{0x30}  // Store delisting compensation records
	DelistingCompensationCounter  = []byte{0x31}  // Counter for compensation IDs
	CompensationClaimPrefix       = []byte{0x32}  // Store compensation claims
	CompensationClaimCounter      = []byte{0x33}  // Counter for claim IDs
	TradingHaltPrefix             = []byte{0x34}  // Store trading halts
	FrozenTreasuryPrefix          = []byte{0x35}  // Store frozen treasury records
	VerifierPenaltyPrefix         = []byte{0x36}  // Store verifier penalties
	VerifierPenaltyCounter        = []byte{0x37}  // Counter for penalty IDs
	CompensationByCompanyPrefix   = []byte{0x38}  // Index: company ID -> compensation ID
	ClaimsByCompensationPrefix    = []byte{0x39}  // Index: compensation ID -> claim IDs
	ClaimsByClaimantPrefix        = []byte{0x3A}  // Index: claimant -> claim IDs

	// Freeze warning prefixes
	FreezeWarningPrefix        = []byte{0x3B}  // Store freeze warnings
	FreezeWarningCounterKey    = []byte{0x3C}  // Counter for warning IDs
	FreezeWarningByCompanyKey  = []byte{0x3D}  // Index: company ID -> warning ID

	// Treasury withdrawal limit prefixes
	TreasuryWithdrawalLimitPrefix = []byte{0x3E}  // company_id -> TreasuryWithdrawalLimit

	// Company Treasury prefixes
	CompanyTreasuryPrefix         = []byte{0x40}  // company_id -> CompanyTreasury
	TreasuryWithdrawalPrefix      = []byte{0x41}  // withdrawal_id -> TreasuryWithdrawal
	TreasuryWithdrawalCounterKey  = []byte{0x42}  // global counter
	WithdrawalByCompanyPrefix     = []byte{0x43}  // company_id -> []withdrawal_id (index)
	WithdrawalByProposalPrefix    = []byte{0x44}  // proposal_id -> withdrawal_id (index)

	// Equity Withdrawal Request prefixes (for treasury investment holdings)
	EquityWithdrawalPrefix        = []byte{0x4B}  // request_id -> EquityWithdrawalRequest
	EquityWithdrawalCounterKey    = []byte{0x4C}  // global counter
	EquityWithdrawalByCompanyPfx  = []byte{0x4D}  // company_id -> []request_id (index)
	EquityWithdrawalByProposalPfx = []byte{0x4E}  // proposal_id -> request_id (index)

	// Primary Sale prefixes
	PrimarySaleOfferPrefix        = []byte{0x45}  // offer_id -> PrimarySaleOffer
	PrimarySaleOfferCounterKey    = []byte{0x46}  // global counter
	OffersByCompanyPrefix         = []byte{0x47}  // company_id -> []offer_id (index)
	ActiveOffersPrefix            = []byte{0x48}  // index for active offers
	OffersBySymbolPrefix          = []byte{0x49}  // symbol -> offer_id (for DEX lookup)
	BuyerPurchasePrefix           = []byte{0x4A}  // offer_id + buyer -> purchase amount

	// Company Authorization prefixes
	CompanyAuthorizationPrefix    = []byte{0x55}  // company_id -> CompanyAuthorization

	// Shareholder Blacklist prefixes
	ShareholderBlacklistPrefix   = []byte{0x62}  // company_id + address -> ShareholderBlacklist
	DividendRedirectionPrefix    = []byte{0x63}  // company_id -> DividendRedirection
	DefaultCharityWalletKey      = []byte{0x64}  // global default charity wallet

	// Beneficial Owner Registry prefixes
	BeneficialOwnerPrefix         = []byte{0x50}  // module + company + class + owner + ref -> BeneficialOwnership
	BeneficialOwnerByModulePrefix = []byte{0x51}  // module + company + class -> index for fast lookup

	// Dividend Audit prefixes
	DividendAuditPrefix           = []byte{0x70}  // audit_id -> DividendAudit
	DividendAuditCounterKey       = []byte{0x71}  // global counter for audit IDs
	AuditByDividendPrefix         = []byte{0x72}  // dividend_id -> audit_id (index)
	AuditByCompanyPrefix          = []byte{0x73}  // company_id -> []audit_id (index)
	AuditVerificationPrefix       = []byte{0x74}  // audit_id + verifier -> AuditVerification
	AuditContentHashPrefix        = []byte{0x75}  // content_hash -> audit_id (uniqueness index)

	// Shareholder Petition prefixes
	PetitionPrefix           = []byte{0x80}  // petition_id -> ShareholderPetition
	PetitionCounterKey       = []byte{0x81}  // global counter for petition IDs
	PetitionByCompanyPrefix  = []byte{0x82}  // company_id -> []petition_id (index)
	PetitionByCreatorPrefix  = []byte{0x83}  // creator -> []petition_id (index)
)

// GetCompanyKey returns the store key for a company
func GetCompanyKey(companyID uint64) []byte {
	return append(CompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetShareClassKey returns the store key for a share class
func GetShareClassKey(companyID uint64, classID string) []byte {
	key := append(ShareClassPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, []byte(classID)...)
}

// GetShareholdingKey returns the store key for a shareholding
func GetShareholdingKey(companyID uint64, classID string, owner string) []byte {
	key := append(ShareholdingPrefix, sdk.Uint64ToBigEndian(companyID)...)
	key = append(key, []byte(classID)...)
	key = append(key, []byte("|")...) // separator
	return append(key, []byte(owner)...)
}


// GetVestingKey returns the store key for a vesting schedule
func GetVestingKey(companyID uint64, classID string, owner string) []byte {
	key := append(VestingPrefix, sdk.Uint64ToBigEndian(companyID)...)
	key = append(key, []byte(classID)...)
	key = append(key, []byte("|")...) // separator
	return append(key, []byte(owner)...)
}

// GetBusinessListingKey returns the store key for a business listing
func GetBusinessListingKey(listingID uint64) []byte {
	return append(BusinessListingPrefix, sdk.Uint64ToBigEndian(listingID)...)
}

// GetValidatorAssignmentKey returns the store key for validator assignment
func GetValidatorAssignmentKey(listingID uint64, validator string) []byte {
	key := append(ValidatorAssignmentPrefix, sdk.Uint64ToBigEndian(listingID)...)
	return append(key, []byte(validator)...)
}

// GetDividendKey returns the store key for a dividend
func GetDividendKey(dividendID uint64) []byte {
	return append(DividendPrefix, sdk.Uint64ToBigEndian(dividendID)...)
}

// GetDividendPaymentKey returns the store key for a dividend payment
func GetDividendPaymentKey(paymentID uint64) []byte {
	return append(DividendPaymentPrefix, sdk.Uint64ToBigEndian(paymentID)...)
}

// GetDividendClaimKey returns the store key for a dividend claim
func GetDividendClaimKey(claimID uint64) []byte {
	return append(DividendClaimPrefix, sdk.Uint64ToBigEndian(claimID)...)
}

// GetDividendSnapshotKey returns the store key for a dividend snapshot
func GetDividendSnapshotKey(dividendID uint64) []byte {
	return append(DividendSnapshotPrefix, sdk.Uint64ToBigEndian(dividendID)...)
}

// GetShareholderSnapshotKey returns the store key for a shareholder snapshot
func GetShareholderSnapshotKey(dividendID uint64, shareholder string) []byte {
	key := append(ShareholderSnapshotPrefix, sdk.Uint64ToBigEndian(dividendID)...)
	return append(key, []byte(shareholder)...)
}

// GetDividendPolicyKey returns the store key for dividend policy
func GetDividendPolicyKey(companyID uint64) []byte {
	return append(DividendPolicyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetDividendStatisticsKey returns the store key for dividend statistics
func GetDividendStatisticsKey(companyID uint64, classID string) []byte {
	key := append(DividendStatisticsPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, []byte(classID)...)
}

// =============================================================================
// Anti-Dilution Key Functions
// =============================================================================

// GetAntiDilutionProvisionKey returns the store key for an anti-dilution provision
func GetAntiDilutionProvisionKey(companyID uint64, classID string) []byte {
	key := append(AntiDilutionProvisionPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, []byte(classID)...)
}

// GetIssuanceRecordKey returns the store key for an issuance record
func GetIssuanceRecordKey(recordID uint64) []byte {
	return append(IssuanceRecordPrefix, sdk.Uint64ToBigEndian(recordID)...)
}

// GetIssuanceRecordsByCompanyPrefix returns the prefix for iterating issuance records by company
func GetIssuanceRecordsByCompanyPrefix(companyID uint64) []byte {
	return append(IssuanceRecordPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetAntiDilutionAdjustmentKey returns the store key for an adjustment record
func GetAntiDilutionAdjustmentKey(adjustmentID uint64) []byte {
	return append(AntiDilutionAdjustmentPrefix, sdk.Uint64ToBigEndian(adjustmentID)...)
}

// GetAdjustmentsByShareholderPrefix returns prefix for iterating adjustments by shareholder
func GetAdjustmentsByShareholderPrefix(companyID uint64, classID string, shareholder string) []byte {
	key := append(AntiDilutionAdjustmentPrefix, sdk.Uint64ToBigEndian(companyID)...)
	key = append(key, []byte(classID)...)
	key = append(key, []byte("|")...)
	return append(key, []byte(shareholder)...)
}

// GetLastIssuancePriceKey returns the store key for last issuance price
func GetLastIssuancePriceKey(companyID uint64, classID string) []byte {
	key := append(LastIssuancePricePrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, []byte(classID)...)
}

// =============================================================================
// Delisting and Compensation Key Functions
// =============================================================================

// GetDelistingCompensationKey returns the store key for a delisting compensation
func GetDelistingCompensationKey(compensationID uint64) []byte {
	return append(DelistingCompensationPrefix, sdk.Uint64ToBigEndian(compensationID)...)
}

// GetCompensationByCompanyKey returns the index key for company -> compensation
func GetCompensationByCompanyKey(companyID uint64) []byte {
	return append(CompensationByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetCompensationClaimKey returns the store key for a compensation claim
func GetCompensationClaimKey(claimID uint64) []byte {
	return append(CompensationClaimPrefix, sdk.Uint64ToBigEndian(claimID)...)
}

// GetClaimsByCompensationKey returns the index key for compensation -> claims
func GetClaimsByCompensationKey(compensationID uint64, claimID uint64) []byte {
	key := append(ClaimsByCompensationPrefix, sdk.Uint64ToBigEndian(compensationID)...)
	return append(key, sdk.Uint64ToBigEndian(claimID)...)
}

// GetClaimsByCompensationPrefix returns the prefix for iterating claims by compensation
func GetClaimsByCompensationPrefix(compensationID uint64) []byte {
	return append(ClaimsByCompensationPrefix, sdk.Uint64ToBigEndian(compensationID)...)
}

// GetClaimsByClaimantKey returns the index key for claimant -> claims
func GetClaimsByClaimantKey(claimant string, claimID uint64) []byte {
	key := append(ClaimsByClaimantPrefix, []byte(claimant)...)
	key = append(key, []byte("|")...)
	return append(key, sdk.Uint64ToBigEndian(claimID)...)
}

// GetTradingHaltKey returns the store key for a trading halt
func GetTradingHaltKey(companyID uint64) []byte {
	return append(TradingHaltPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetFrozenTreasuryKey returns the store key for a frozen treasury
func GetFrozenTreasuryKey(companyID uint64) []byte {
	return append(FrozenTreasuryPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetVerifierPenaltyKey returns the store key for a verifier penalty
func GetVerifierPenaltyKey(penaltyID uint64) []byte {
	return append(VerifierPenaltyPrefix, sdk.Uint64ToBigEndian(penaltyID)...)
}

// =============================================================================
// Company Treasury Key Functions
// =============================================================================

// GetCompanyTreasuryKey returns the store key for a company treasury
func GetCompanyTreasuryKey(companyID uint64) []byte {
	return append(CompanyTreasuryPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetTreasuryWithdrawalKey returns the store key for a treasury withdrawal
func GetTreasuryWithdrawalKey(withdrawalID uint64) []byte {
	return append(TreasuryWithdrawalPrefix, sdk.Uint64ToBigEndian(withdrawalID)...)
}

// GetWithdrawalsByCompanyPrefix returns the prefix for iterating withdrawals by company
func GetWithdrawalsByCompanyPrefix(companyID uint64) []byte {
	return append(WithdrawalByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetWithdrawalByCompanyKey returns the index key for company -> withdrawal
func GetWithdrawalByCompanyKey(companyID uint64, withdrawalID uint64) []byte {
	key := append(WithdrawalByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, sdk.Uint64ToBigEndian(withdrawalID)...)
}

// GetWithdrawalByProposalKey returns the index key for proposal -> withdrawal
func GetWithdrawalByProposalKey(proposalID uint64) []byte {
	return append(WithdrawalByProposalPrefix, sdk.Uint64ToBigEndian(proposalID)...)
}

// =============================================================================
// Equity Withdrawal Key Functions (for treasury investment holdings)
// =============================================================================

// GetEquityWithdrawalKey returns the store key for an equity withdrawal request
func GetEquityWithdrawalKey(requestID uint64) []byte {
	return append(EquityWithdrawalPrefix, sdk.Uint64ToBigEndian(requestID)...)
}

// GetEquityWithdrawalsByCompanyPrefix returns prefix for iterating by company
func GetEquityWithdrawalsByCompanyPrefix(companyID uint64) []byte {
	return append(EquityWithdrawalByCompanyPfx, sdk.Uint64ToBigEndian(companyID)...)
}

// GetEquityWithdrawalByCompanyKey returns the index key for company -> request
func GetEquityWithdrawalByCompanyKey(companyID uint64, requestID uint64) []byte {
	key := append(EquityWithdrawalByCompanyPfx, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, sdk.Uint64ToBigEndian(requestID)...)
}

// GetEquityWithdrawalByProposalKey returns the index key for proposal -> request
func GetEquityWithdrawalByProposalKey(proposalID uint64) []byte {
	return append(EquityWithdrawalByProposalPfx, sdk.Uint64ToBigEndian(proposalID)...)
}

// =============================================================================
// Primary Sale Key Functions
// =============================================================================

// GetPrimarySaleOfferKey returns the store key for a primary sale offer
func GetPrimarySaleOfferKey(offerID uint64) []byte {
	return append(PrimarySaleOfferPrefix, sdk.Uint64ToBigEndian(offerID)...)
}

// GetOffersByCompanyPrefix returns the prefix for iterating offers by company
func GetOffersByCompanyPrefix(companyID uint64) []byte {
	return append(OffersByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetOfferByCompanyKey returns the index key for company -> offer
func GetOfferByCompanyKey(companyID uint64, offerID uint64) []byte {
	key := append(OffersByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, sdk.Uint64ToBigEndian(offerID)...)
}

// GetActiveOfferKey returns the store key for an active offer index
func GetActiveOfferKey(offerID uint64) []byte {
	return append(ActiveOffersPrefix, sdk.Uint64ToBigEndian(offerID)...)
}

// GetOfferBySymbolKey returns the index key for symbol -> offer
func GetOfferBySymbolKey(symbol string) []byte {
	return append(OffersBySymbolPrefix, []byte(symbol)...)
}

// GetBuyerPurchaseKey returns the store key for buyer purchase amount
func GetBuyerPurchaseKey(offerID uint64, buyer string) []byte {
	key := append(BuyerPurchasePrefix, sdk.Uint64ToBigEndian(offerID)...)
	return append(key, []byte(buyer)...)
}

// =============================================================================
// Company Authorization Key Functions
// =============================================================================

// GetCompanyAuthorizationKey returns the store key for a company authorization
func GetCompanyAuthorizationKey(companyID uint64) []byte {
	return append(CompanyAuthorizationPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// =============================================================================
// Shareholder Blacklist Key Functions
// =============================================================================

// GetShareholderBlacklistKey returns the store key for a shareholder blacklist entry
func GetShareholderBlacklistKey(companyID uint64, address string) []byte {
	key := append(ShareholderBlacklistPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, []byte(address)...)
}

// GetShareholderBlacklistsByCompanyPrefix returns the prefix for iterating blacklists by company
func GetShareholderBlacklistsByCompanyPrefix(companyID uint64) []byte {
	return append(ShareholderBlacklistPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetDividendRedirectionKey returns the store key for dividend redirection config
func GetDividendRedirectionKey(companyID uint64) []byte {
	return append(DividendRedirectionPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// =============================================================================
// Beneficial Owner Registry Key Functions
// =============================================================================

// GetBeneficialOwnerKey returns the store key for a beneficial ownership record
// Key format: prefix + module + "|" + company_id + class_id + "|" + owner + "|" + ref_id
func GetBeneficialOwnerKey(moduleAccount string, companyID uint64, classID string, owner string, refID uint64) []byte {
	key := append(BeneficialOwnerPrefix, []byte(moduleAccount)...)
	key = append(key, []byte("|")...)
	key = append(key, sdk.Uint64ToBigEndian(companyID)...)
	key = append(key, []byte(classID)...)
	key = append(key, []byte("|")...)
	key = append(key, []byte(owner)...)
	key = append(key, []byte("|")...)
	key = append(key, sdk.Uint64ToBigEndian(refID)...)
	return key
}

// GetBeneficialOwnerByModuleKey returns the index key for fast lookup by module + company + class
// Key format: prefix + module + "|" + company_id + class_id
func GetBeneficialOwnerByModuleKey(moduleAccount string, companyID uint64, classID string) []byte {
	key := append(BeneficialOwnerByModulePrefix, []byte(moduleAccount)...)
	key = append(key, []byte("|")...)
	key = append(key, sdk.Uint64ToBigEndian(companyID)...)
	key = append(key, []byte(classID)...)
	return key
}

// GetBeneficialOwnersByModulePrefix returns the prefix for iterating beneficial owners by module
func GetBeneficialOwnersByModulePrefix(moduleAccount string) []byte {
	return append(BeneficialOwnerPrefix, []byte(moduleAccount)...)
}

// =============================================================================
// Dividend Audit Key Functions
// =============================================================================

// GetDividendAuditKey returns the store key for a dividend audit
func GetDividendAuditKey(auditID uint64) []byte {
	return append(DividendAuditPrefix, sdk.Uint64ToBigEndian(auditID)...)
}

// GetAuditByDividendKey returns the index key for dividend -> audit
func GetAuditByDividendKey(dividendID uint64) []byte {
	return append(AuditByDividendPrefix, sdk.Uint64ToBigEndian(dividendID)...)
}

// GetAuditsByCompanyPrefix returns the prefix for iterating audits by company
func GetAuditsByCompanyPrefix(companyID uint64) []byte {
	return append(AuditByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetAuditByCompanyKey returns the index key for company -> audit
func GetAuditByCompanyKey(companyID uint64, auditID uint64) []byte {
	key := append(AuditByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, sdk.Uint64ToBigEndian(auditID)...)
}

// GetAuditVerificationKey returns the store key for an audit verification
func GetAuditVerificationKey(auditID uint64, verifier string) []byte {
	key := append(AuditVerificationPrefix, sdk.Uint64ToBigEndian(auditID)...)
	return append(key, []byte(verifier)...)
}

// GetAuditVerificationsByAuditPrefix returns the prefix for iterating verifications by audit
func GetAuditVerificationsByAuditPrefix(auditID uint64) []byte {
	return append(AuditVerificationPrefix, sdk.Uint64ToBigEndian(auditID)...)
}

// GetAuditContentHashKey returns the index key for content hash uniqueness check
func GetAuditContentHashKey(contentHash string) []byte {
	return append(AuditContentHashPrefix, []byte(contentHash)...)
}

// =============================================================================
// Shareholder Petition Key Functions
// =============================================================================

// GetPetitionKey returns the store key for a petition
func GetPetitionKey(petitionID uint64) []byte {
	return append(PetitionPrefix, sdk.Uint64ToBigEndian(petitionID)...)
}

// GetPetitionsByCompanyPrefix returns the prefix for iterating petitions by company
func GetPetitionsByCompanyPrefix(companyID uint64) []byte {
	return append(PetitionByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetPetitionByCompanyKey returns the index key for company -> petition
func GetPetitionByCompanyKey(companyID uint64, petitionID uint64) []byte {
	key := append(PetitionByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, sdk.Uint64ToBigEndian(petitionID)...)
}

// GetPetitionsByCreatorPrefix returns the prefix for iterating petitions by creator
func GetPetitionsByCreatorPrefix(creator string) []byte {
	return append(PetitionByCreatorPrefix, []byte(creator)...)
}

// GetPetitionByCreatorKey returns the index key for creator -> petition
func GetPetitionByCreatorKey(creator string, petitionID uint64) []byte {
	key := append(PetitionByCreatorPrefix, []byte(creator)...)
	key = append(key, []byte("|")...)
	return append(key, sdk.Uint64ToBigEndian(petitionID)...)
}

// =============================================================================
// Freeze Warning Key Functions
// =============================================================================

// GetFreezeWarningKey returns the store key for a freeze warning
func GetFreezeWarningKey(warningID uint64) []byte {
	return append(FreezeWarningPrefix, sdk.Uint64ToBigEndian(warningID)...)
}

// GetFreezeWarningByCompanyKey returns the index key for company -> warning
func GetFreezeWarningByCompanyKey(companyID uint64) []byte {
	return append(FreezeWarningByCompanyKey, sdk.Uint64ToBigEndian(companyID)...)
}

// GetTreasuryWithdrawalLimitKey returns the store key for a treasury withdrawal limit
func GetTreasuryWithdrawalLimitKey(companyID uint64) []byte {
	return append(TreasuryWithdrawalLimitPrefix, sdk.Uint64ToBigEndian(companyID)...)
}