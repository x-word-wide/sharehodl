package types

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for business listing operations

// MsgSubmitBusinessListing represents a message to submit a business listing application
type MsgSubmitBusinessListing struct {
	Creator             string         `json:"creator"`
	CompanyName         string         `json:"company_name"`
	LegalEntityName     string         `json:"legal_entity_name"`
	RegistrationNumber  string         `json:"registration_number"`
	Jurisdiction        string         `json:"jurisdiction"`
	Industry            string         `json:"industry"`
	BusinessDescription string         `json:"business_description"`
	FoundedDate         time.Time      `json:"founded_date"`
	Website             string         `json:"website"`
	ContactEmail        string         `json:"contact_email"`
	TokenSymbol         string         `json:"token_symbol"`
	TokenName           string         `json:"token_name"`
	TotalShares         math.Int       `json:"total_shares"`
	InitialPrice        math.LegacyDec `json:"initial_price"`
	ListingFee          math.Int       `json:"listing_fee"`
	Documents           []BusinessDocumentUpload `json:"documents"`
}

// BusinessDocumentUpload represents a document upload in the message
type BusinessDocumentUpload struct {
	DocumentType string `json:"document_type"`
	FileName     string `json:"file_name"`
	FileHash     string `json:"file_hash"`
	FileSize     int64  `json:"file_size"`
}

// GetSigners implements the Msg interface
func (msg MsgSubmitBusinessListing) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic implements the Msg interface
func (msg MsgSubmitBusinessListing) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.CompanyName == "" {
		return ErrInvalidCompanyName
	}

	if msg.TokenSymbol == "" || len(msg.TokenSymbol) < 2 || len(msg.TokenSymbol) > 10 {
		return ErrInvalidTokenSymbol
	}

	if msg.TotalShares.IsZero() || msg.TotalShares.IsNegative() {
		return ErrInvalidTotalShares
	}

	if msg.InitialPrice.IsZero() || msg.InitialPrice.IsNegative() {
		return ErrInvalidInitialPrice
	}

	return nil
}

// MsgSubmitBusinessListingResponse is the response type for MsgSubmitBusinessListing
type MsgSubmitBusinessListingResponse struct {
	ListingID uint64 `json:"listing_id"`
	Success   bool   `json:"success"`
}

// MsgReviewBusinessListing represents a validator's review of a business listing
type MsgReviewBusinessListing struct {
	Validator string                 `json:"validator"`
	ListingID uint64                 `json:"listing_id"`
	Approved  bool                   `json:"approved"`
	Comments  string                 `json:"comments"`
	Score     int                    `json:"score"`
	Criteria  VerificationCriteria   `json:"criteria"`
	Data      map[string]interface{} `json:"data"`
}

// GetSigners implements the Msg interface
func (msg MsgReviewBusinessListing) GetSigners() []sdk.AccAddress {
	validator, err := sdk.AccAddressFromBech32(msg.Validator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{validator}
}

// ValidateBasic implements the Msg interface
func (msg MsgReviewBusinessListing) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Validator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.ListingID == 0 {
		return ErrInvalidListingID
	}

	if msg.Score < 0 || msg.Score > 100 {
		return ErrInvalidScore
	}

	return nil
}

// MsgReviewBusinessListingResponse is the response type for MsgReviewBusinessListing
type MsgReviewBusinessListingResponse struct {
	Success bool `json:"success"`
}

// MsgApproveListing represents a message to approve a business listing
type MsgApproveListing struct {
	Authority string `json:"authority"`
	ListingID uint64 `json:"listing_id"`
	Notes     string `json:"notes"`
}

// GetSigners implements the Msg interface
func (msg MsgApproveListing) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// ValidateBasic implements the Msg interface
func (msg MsgApproveListing) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.ListingID == 0 {
		return ErrInvalidListingID
	}

	return nil
}

// MsgApproveListingResponse is the response type for MsgApproveListing
type MsgApproveListingResponse struct {
	TokenID     uint64 `json:"token_id"`
	TokenSymbol string `json:"token_symbol"`
	Success     bool   `json:"success"`
}

// MsgRejectListing represents a message to reject a business listing
type MsgRejectListing struct {
	Authority string `json:"authority"`
	ListingID uint64 `json:"listing_id"`
	Reason    string `json:"reason"`
}

// GetSigners implements the Msg interface
func (msg MsgRejectListing) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// ValidateBasic implements the Msg interface
func (msg MsgRejectListing) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.ListingID == 0 {
		return ErrInvalidListingID
	}

	if msg.Reason == "" {
		return ErrInvalidReason
	}

	return nil
}

// MsgRejectListingResponse is the response type for MsgRejectListing
type MsgRejectListingResponse struct {
	Success bool `json:"success"`
}

// MsgSuspendListing represents a message to suspend a business listing
type MsgSuspendListing struct {
	Authority string `json:"authority"`
	ListingID uint64 `json:"listing_id"`
	Reason    string `json:"reason"`
}

// GetSigners implements the Msg interface
func (msg MsgSuspendListing) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// ValidateBasic implements the Msg interface
func (msg MsgSuspendListing) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.ListingID == 0 {
		return ErrInvalidListingID
	}

	return nil
}

// MsgSuspendListingResponse is the response type for MsgSuspendListing
type MsgSuspendListingResponse struct {
	Success bool `json:"success"`
}

// MsgUpdateListingRequirements represents a message to update listing requirements
type MsgUpdateListingRequirements struct {
	Authority    string              `json:"authority"`
	Requirements ListingRequirements `json:"requirements"`
}

// GetSigners implements the Msg interface
func (msg MsgUpdateListingRequirements) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// ValidateBasic implements the Msg interface
func (msg MsgUpdateListingRequirements) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return ErrUnauthorized
	}

	return nil
}

// MsgUpdateListingRequirementsResponse is the response type for MsgUpdateListingRequirements
type MsgUpdateListingRequirementsResponse struct {
	Success bool `json:"success"`
}

// NewMsgSubmitBusinessListing creates a new MsgSubmitBusinessListing instance
func NewMsgSubmitBusinessListing(
	creator, companyName, legalEntityName, registrationNumber, jurisdiction, industry string,
	businessDescription, website, contactEmail, tokenSymbol, tokenName string,
	foundedDate time.Time,
	totalShares math.Int,
	initialPrice math.LegacyDec,
	listingFee math.Int,
	documents []BusinessDocumentUpload,
) *MsgSubmitBusinessListing {
	return &MsgSubmitBusinessListing{
		Creator:             creator,
		CompanyName:         companyName,
		LegalEntityName:     legalEntityName,
		RegistrationNumber:  registrationNumber,
		Jurisdiction:        jurisdiction,
		Industry:            industry,
		BusinessDescription: businessDescription,
		Website:             website,
		ContactEmail:        contactEmail,
		TokenSymbol:         tokenSymbol,
		TokenName:           tokenName,
		FoundedDate:         foundedDate,
		TotalShares:         totalShares,
		InitialPrice:        initialPrice,
		ListingFee:          listingFee,
		Documents:           documents,
	}
}

// String returns a human readable string representation of the message
func (msg MsgSubmitBusinessListing) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// String returns a human readable string representation of the message
func (msg MsgReviewBusinessListing) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// String returns a human readable string representation of the message
func (msg MsgApproveListing) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}