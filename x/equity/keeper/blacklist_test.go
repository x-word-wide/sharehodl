package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

var (
	// Pre-computed valid test addresses for blacklist tests
	blacklistTestAddr1 = sdk.AccAddress([]byte("blacklist_addr_1____")).String()
	blacklistTestAddr2 = sdk.AccAddress([]byte("blacklist_addr_2____")).String()
	blacklistAuthAddr  = sdk.AccAddress([]byte("authority_addr______")).String()
)

// TestShareholderBlacklist tests the shareholder blacklist functionality
func TestShareholderBlacklist(t *testing.T) {
	// This is a placeholder test - in production, would use proper test suite
	// with keeper setup, etc.

	// Test ShareholderBlacklist validation
	t.Run("ShareholderBlacklist Validation", func(t *testing.T) {
		// Valid blacklist entry
		validBlacklist := types.NewShareholderBlacklist(
			1,
			blacklistTestAddr1,
			"Fraudulent activity",
			blacklistAuthAddr,
			24*time.Hour,
		)
		err := validBlacklist.Validate()
		require.NoError(t, err)

		// Invalid - zero company ID
		invalidBlacklist := validBlacklist
		invalidBlacklist.CompanyID = 0
		err = invalidBlacklist.Validate()
		require.Error(t, err)

		// Invalid - empty address
		invalidBlacklist = validBlacklist
		invalidBlacklist.Address = ""
		err = invalidBlacklist.Validate()
		require.Error(t, err)

		// Invalid - empty reason
		invalidBlacklist = validBlacklist
		invalidBlacklist.Reason = ""
		err = invalidBlacklist.Validate()
		require.Error(t, err)
	})

	// Test DividendRedirection validation
	t.Run("DividendRedirection Validation", func(t *testing.T) {
		// Valid redirection config
		validRedirect := types.NewDividendRedirection(
			1,
			blacklistTestAddr1,
			"community_pool",
			"Charity redirection",
			blacklistAuthAddr,
		)
		err := validRedirect.Validate()
		require.NoError(t, err)

		// Valid - no charity wallet, fallback only
		validRedirect2 := types.NewDividendRedirection(
			1,
			"",
			"burn",
			"Burn fallback",
			blacklistAuthAddr,
		)
		err = validRedirect2.Validate()
		require.NoError(t, err)

		// Invalid - zero company ID
		invalidRedirect := validRedirect
		invalidRedirect.CompanyID = 0
		err = invalidRedirect.Validate()
		require.Error(t, err)

		// Invalid - bad fallback action
		invalidRedirect = validRedirect
		invalidRedirect.FallbackAction = "invalid_action"
		err = invalidRedirect.Validate()
		require.Error(t, err)
	})

	// Test Message validation
	t.Run("Message Validation", func(t *testing.T) {
		// Valid blacklist message
		validMsg := types.SimpleMsgBlacklistShareholder{
			Authority: blacklistAuthAddr,
			CompanyID: 1,
			Address:   blacklistTestAddr1,
			Reason:    "Fraudulent activity",
			Duration:  86400, // 24 hours in seconds
		}
		err := validMsg.Validate()
		require.NoError(t, err)

		// Invalid - negative duration
		invalidMsg := validMsg
		invalidMsg.Duration = -1
		err = invalidMsg.Validate()
		require.Error(t, err)

		// Valid unblacklist message
		validUnblacklistMsg := types.SimpleMsgUnblacklistShareholder{
			Authority: blacklistAuthAddr,
			CompanyID: 1,
			Address:   blacklistTestAddr1,
		}
		err = validUnblacklistMsg.Validate()
		require.NoError(t, err)

		// Valid redirection message
		validRedirectMsg := types.SimpleMsgSetDividendRedirection{
			Authority:      blacklistAuthAddr,
			CompanyID:      1,
			CharityWallet:  blacklistTestAddr1,
			FallbackAction: "community_pool",
			Description:    "Send to charity",
		}
		err = validRedirectMsg.Validate()
		require.NoError(t, err)

		// Valid charity wallet message
		validCharityMsg := types.SimpleMsgSetCharityWallet{
			Authority: blacklistAuthAddr,
			Wallet:    blacklistTestAddr1,
		}
		err = validCharityMsg.Validate()
		require.NoError(t, err)
	})
}

// TestBlacklistExpiration tests blacklist expiration logic
func TestBlacklistExpiration(t *testing.T) {
	now := time.Now()

	// Permanent blacklist (no expiration)
	permanentBlacklist := types.NewShareholderBlacklist(
		1,
		blacklistTestAddr1,
		"Permanent ban",
		blacklistAuthAddr,
		0, // Permanent
	)
	require.False(t, permanentBlacklist.IsExpired(now.Add(100*365*24*time.Hour)), "Permanent blacklist should never expire")

	// Temporary blacklist (24 hours)
	tempBlacklist := types.NewShareholderBlacklist(
		1,
		blacklistTestAddr1,
		"Temporary ban",
		blacklistAuthAddr,
		24*time.Hour,
	)

	// Should not be expired immediately
	require.False(t, tempBlacklist.IsExpired(now), "Should not be expired immediately")

	// Should not be expired after 12 hours
	require.False(t, tempBlacklist.IsExpired(now.Add(12*time.Hour)), "Should not be expired after 12 hours")

	// Should be expired after 25 hours
	require.True(t, tempBlacklist.IsExpired(now.Add(25*time.Hour)), "Should be expired after 25 hours")
}
