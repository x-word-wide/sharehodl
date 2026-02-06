package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Event types for shareholder blacklist
const (
	EventTypeShareholderBlacklisted   = "shareholder_blacklisted"
	EventTypeShareholderUnblacklisted = "shareholder_unblacklisted"
	EventTypeDividendRedirected       = "dividend_redirected"
)

// Attribute keys for shareholder blacklist events
// Note: AttributeKeyCompanyID and AttributeKeyReason are already defined in delisting.go
const (
	AttributeKeyShareholder    = "shareholder"
	AttributeKeyBlacklistedBy  = "blacklisted_by"
	AttributeKeyRedirectTo     = "redirect_to"
	AttributeKeyRedirectReason = "redirect_reason"
	AttributeKeyRedirectAmount = "redirect_amount"
)

// ShareholderBlacklist tracks blacklisted addresses per company
type ShareholderBlacklist struct {
	CompanyID     uint64    `json:"company_id"`
	Address       string    `json:"address"`
	Reason        string    `json:"reason"`
	BlacklistedBy string    `json:"blacklisted_by"`  // Who blacklisted (governance/owner)
	BlacklistedAt time.Time `json:"blacklisted_at"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"` // Zero = permanent
	IsActive      bool      `json:"is_active"`
}

// Validate validates a ShareholderBlacklist
func (sb ShareholderBlacklist) Validate() error {
	if sb.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if sb.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(sb.Address); err != nil {
		return fmt.Errorf("invalid address: %v", err)
	}
	if sb.Reason == "" {
		return fmt.Errorf("reason cannot be empty")
	}
	if sb.BlacklistedBy == "" {
		return fmt.Errorf("blacklisted_by cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(sb.BlacklistedBy); err != nil {
		return fmt.Errorf("invalid blacklisted_by address: %v", err)
	}
	return nil
}

// IsExpired checks if the blacklist entry has expired
func (sb ShareholderBlacklist) IsExpired(currentTime time.Time) bool {
	if sb.ExpiresAt.IsZero() {
		return false // Permanent blacklist
	}
	return currentTime.After(sb.ExpiresAt)
}

// DividendRedirection defines where blocked dividends go
type DividendRedirection struct {
	CompanyID       uint64 `json:"company_id"`
	CharityWallet   string `json:"charity_wallet,omitempty"`   // Primary destination
	FallbackAction  string `json:"fallback_action"`            // "community_pool", "burn", "pro_rata"
	Description     string `json:"description,omitempty"`
	SetBy           string `json:"set_by"`                     // Who set this configuration
	SetAt           time.Time `json:"set_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Validate validates a DividendRedirection
func (dr DividendRedirection) Validate() error {
	if dr.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}

	// Validate charity wallet if provided
	if dr.CharityWallet != "" {
		if _, err := sdk.AccAddressFromBech32(dr.CharityWallet); err != nil {
			return ErrInvalidCharityWallet
		}
	}

	// Validate fallback action
	if dr.FallbackAction != "community_pool" &&
	   dr.FallbackAction != "burn" &&
	   dr.FallbackAction != "pro_rata" &&
	   dr.FallbackAction != "" {
		return fmt.Errorf("invalid fallback action: must be 'community_pool', 'burn', 'pro_rata', or empty")
	}

	if dr.SetBy != "" {
		if _, err := sdk.AccAddressFromBech32(dr.SetBy); err != nil {
			return fmt.Errorf("invalid set_by address: %v", err)
		}
	}

	return nil
}

// NewShareholderBlacklist creates a new shareholder blacklist entry
func NewShareholderBlacklist(
	companyID uint64,
	address string,
	reason string,
	blacklistedBy string,
	duration time.Duration,
) ShareholderBlacklist {
	now := time.Now()

	var expiresAt time.Time
	if duration > 0 {
		expiresAt = now.Add(duration)
	}

	return ShareholderBlacklist{
		CompanyID:     companyID,
		Address:       address,
		Reason:        reason,
		BlacklistedBy: blacklistedBy,
		BlacklistedAt: now,
		ExpiresAt:     expiresAt,
		IsActive:      true,
	}
}

// NewDividendRedirection creates a new dividend redirection configuration
func NewDividendRedirection(
	companyID uint64,
	charityWallet string,
	fallbackAction string,
	description string,
	setBy string,
) DividendRedirection {
	now := time.Now()

	return DividendRedirection{
		CompanyID:      companyID,
		CharityWallet:  charityWallet,
		FallbackAction: fallbackAction,
		Description:    description,
		SetBy:          setBy,
		SetAt:          now,
		UpdatedAt:      now,
	}
}
