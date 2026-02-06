package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// =============================================================================
// Company Authorization - Co-Founder and Proposer System
// =============================================================================

// CompanyRole defines permission levels for company governance
type CompanyRole int32

const (
	RoleNone     CompanyRole = 0 // No special permissions
	RoleProposer CompanyRole = 1 // Can propose (treasury withdrawals, dilution, primary sales)
	RoleOwner    CompanyRole = 2 // Full control (add/remove proposers, transfer ownership)
)

// String returns the string representation of CompanyRole
func (r CompanyRole) String() string {
	switch r {
	case RoleNone:
		return "none"
	case RoleProposer:
		return "proposer"
	case RoleOwner:
		return "owner"
	default:
		return "unknown"
	}
}

// CompanyAuthorization tracks who can do what for a company
// This enables co-founder scenarios and delegated proposal creation
type CompanyAuthorization struct {
	CompanyID            uint64    `json:"company_id"`
	Owner                string    `json:"owner"`                           // Current owner (founder initially)
	AuthorizedProposers  []string  `json:"authorized_proposers"`            // Addresses with proposer role
	PendingOwnerTransfer string    `json:"pending_owner_transfer,omitempty"` // Two-step transfer for safety
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Validate validates a CompanyAuthorization
func (a CompanyAuthorization) Validate() error {
	if a.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if a.Owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(a.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	// Validate all proposer addresses
	for i, proposer := range a.AuthorizedProposers {
		if proposer == "" {
			return fmt.Errorf("proposer at index %d cannot be empty", i)
		}
		if _, err := sdk.AccAddressFromBech32(proposer); err != nil {
			return fmt.Errorf("invalid proposer address at index %d: %v", i, err)
		}
	}
	// Validate pending transfer address if set
	if a.PendingOwnerTransfer != "" {
		if _, err := sdk.AccAddressFromBech32(a.PendingOwnerTransfer); err != nil {
			return fmt.Errorf("invalid pending owner transfer address: %v", err)
		}
	}
	return nil
}

// IsOwner checks if the given address is the company owner
func (a CompanyAuthorization) IsOwner(address string) bool {
	return a.Owner == address
}

// IsProposer checks if the given address is an authorized proposer
func (a CompanyAuthorization) IsProposer(address string) bool {
	for _, proposer := range a.AuthorizedProposers {
		if proposer == address {
			return true
		}
	}
	return false
}

// CanPropose checks if the address can propose (owner OR proposer)
func (a CompanyAuthorization) CanPropose(address string) bool {
	return a.IsOwner(address) || a.IsProposer(address)
}

// AddProposer adds a proposer to the authorized list
func (a *CompanyAuthorization) AddProposer(address string) error {
	// Validate address
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		return fmt.Errorf("invalid address: %v", err)
	}

	// Check if already a proposer
	if a.IsProposer(address) {
		return ErrProposerAlreadyExists
	}

	// Check if trying to add owner as proposer (redundant)
	if a.IsOwner(address) {
		return fmt.Errorf("owner already has proposer privileges")
	}

	a.AuthorizedProposers = append(a.AuthorizedProposers, address)
	a.UpdatedAt = time.Now()
	return nil
}

// RemoveProposer removes a proposer from the authorized list
func (a *CompanyAuthorization) RemoveProposer(address string) error {
	// Find and remove the proposer
	for i, proposer := range a.AuthorizedProposers {
		if proposer == address {
			// Remove by replacing with last element and truncating
			a.AuthorizedProposers[i] = a.AuthorizedProposers[len(a.AuthorizedProposers)-1]
			a.AuthorizedProposers = a.AuthorizedProposers[:len(a.AuthorizedProposers)-1]
			a.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrProposerNotFound
}

// HasPendingTransfer checks if there's a pending ownership transfer
func (a CompanyAuthorization) HasPendingTransfer() bool {
	return a.PendingOwnerTransfer != ""
}

// InitiateOwnershipTransfer starts the two-step ownership transfer
func (a *CompanyAuthorization) InitiateOwnershipTransfer(newOwner string) error {
	// Validate new owner address
	if _, err := sdk.AccAddressFromBech32(newOwner); err != nil {
		return fmt.Errorf("invalid new owner address: %v", err)
	}

	// Cannot transfer to current owner
	if a.Owner == newOwner {
		return fmt.Errorf("new owner cannot be current owner")
	}

	// Check if already has pending transfer
	if a.HasPendingTransfer() {
		return ErrOwnershipTransferPending
	}

	a.PendingOwnerTransfer = newOwner
	a.UpdatedAt = time.Now()
	return nil
}

// AcceptOwnershipTransfer completes the ownership transfer (called by new owner)
func (a *CompanyAuthorization) AcceptOwnershipTransfer(acceptor string) error {
	// Check if there's a pending transfer
	if !a.HasPendingTransfer() {
		return ErrNoOwnershipTransferPending
	}

	// Check if acceptor is the pending owner
	if a.PendingOwnerTransfer != acceptor {
		return fmt.Errorf("only pending owner can accept transfer")
	}

	// Transfer ownership
	a.Owner = acceptor
	a.PendingOwnerTransfer = ""
	a.UpdatedAt = time.Now()
	return nil
}

// CancelOwnershipTransfer cancels a pending ownership transfer
func (a *CompanyAuthorization) CancelOwnershipTransfer() error {
	if !a.HasPendingTransfer() {
		return ErrNoOwnershipTransferPending
	}

	a.PendingOwnerTransfer = ""
	a.UpdatedAt = time.Now()
	return nil
}

// GetRole returns the role for the given address
func (a CompanyAuthorization) GetRole(address string) CompanyRole {
	if a.IsOwner(address) {
		return RoleOwner
	}
	if a.IsProposer(address) {
		return RoleProposer
	}
	return RoleNone
}

// NewCompanyAuthorization creates a new company authorization with owner
func NewCompanyAuthorization(companyID uint64, owner string) CompanyAuthorization {
	now := time.Now()
	return CompanyAuthorization{
		CompanyID:            companyID,
		Owner:                owner,
		AuthorizedProposers:  []string{}, // Empty initially
		PendingOwnerTransfer: "",
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// =============================================================================
// Authorization Message Types
// =============================================================================

// MsgAddAuthorizedProposer adds a new authorized proposer
type MsgAddAuthorizedProposer struct {
	Creator     string `json:"creator"`      // Must be company owner
	CompanyID   uint64 `json:"company_id"`
	NewProposer string `json:"new_proposer"` // Address to add as proposer
}

// MsgRemoveAuthorizedProposer removes an authorized proposer
type MsgRemoveAuthorizedProposer struct {
	Creator         string `json:"creator"`          // Must be company owner
	CompanyID       uint64 `json:"company_id"`
	ProposerToRemove string `json:"proposer_to_remove"` // Address to remove
}

// MsgInitiateOwnershipTransfer initiates a two-step ownership transfer
type MsgInitiateOwnershipTransfer struct {
	Creator  string `json:"creator"`   // Must be current owner
	CompanyID uint64 `json:"company_id"`
	NewOwner string `json:"new_owner"` // Proposed new owner
}

// MsgAcceptOwnershipTransfer accepts a pending ownership transfer
type MsgAcceptOwnershipTransfer struct {
	Creator   string `json:"creator"`    // Must be pending owner
	CompanyID uint64 `json:"company_id"`
}

// MsgCancelOwnershipTransfer cancels a pending ownership transfer
type MsgCancelOwnershipTransfer struct {
	Creator   string `json:"creator"`    // Must be current owner
	CompanyID uint64 `json:"company_id"`
}

// Response types

type MsgAddAuthorizedProposerResponse struct {
	Success bool `json:"success"`
}

type MsgRemoveAuthorizedProposerResponse struct {
	Success bool `json:"success"`
}

type MsgInitiateOwnershipTransferResponse struct {
	Success bool `json:"success"`
}

type MsgAcceptOwnershipTransferResponse struct {
	Success  bool   `json:"success"`
	NewOwner string `json:"new_owner"`
}

type MsgCancelOwnershipTransferResponse struct {
	Success bool `json:"success"`
}

// Validation methods

func (msg MsgAddAuthorizedProposer) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.NewProposer == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.NewProposer); err != nil {
		return ErrUnauthorized
	}
	if msg.Creator == msg.NewProposer {
		return fmt.Errorf("cannot add yourself as proposer (owner already has privileges)")
	}
	return nil
}

func (msg MsgRemoveAuthorizedProposer) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.ProposerToRemove == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.ProposerToRemove); err != nil {
		return ErrUnauthorized
	}
	return nil
}

func (msg MsgInitiateOwnershipTransfer) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.NewOwner == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.NewOwner); err != nil {
		return ErrUnauthorized
	}
	if msg.Creator == msg.NewOwner {
		return fmt.Errorf("cannot transfer ownership to yourself")
	}
	return nil
}

func (msg MsgAcceptOwnershipTransfer) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	return nil
}

func (msg MsgCancelOwnershipTransfer) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	return nil
}

// =============================================================================
// Event Types
// =============================================================================

const (
	EventTypeProposerAdded             = "proposer_added"
	EventTypeProposerRemoved           = "proposer_removed"
	EventTypeOwnershipTransferInitiated = "ownership_transfer_initiated"
	EventTypeOwnershipTransferAccepted  = "ownership_transfer_accepted"
	EventTypeOwnershipTransferCancelled = "ownership_transfer_cancelled"
)

// Attribute keys
const (
	AttributeKeyProposer       = "proposer"
	AttributeKeyOldOwner       = "old_owner"
	AttributeKeyNewOwner       = "new_owner"
	AttributeKeyPendingOwner   = "pending_owner"
	AttributeKeyProposerCount  = "proposer_count"
)
