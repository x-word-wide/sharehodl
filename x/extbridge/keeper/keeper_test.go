package keeper_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cometbfttypes "github.com/cometbft/cometbft/api/cometbft/types/v2"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
)

// MockBankKeeper is a mock implementation of BankKeeper
type MockBankKeeper struct {
	balances map[string]sdk.Coins
	supply   map[string]sdk.Coin
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		balances: make(map[string]sdk.Coins),
		supply:   make(map[string]sdk.Coin),
	}
}

func (m *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coins := m.balances[addr.String()]
	return sdk.NewCoin(denom, coins.AmountOf(denom))
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return m.balances[addr.String()]
}

func (m *MockBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return m.balances[addr.String()]
}

func (m *MockBankKeeper) SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) GetSupply(ctx context.Context, denom string) sdk.Coin {
	if supply, ok := m.supply[denom]; ok {
		return supply
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

// MockAccountKeeper is a mock implementation of AccountKeeper
type MockAccountKeeper struct{}

func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m *MockAccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {
}

func (m *MockAccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m *MockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress("module_address")
}

func (m *MockAccountKeeper) GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI {
	return nil
}

// MockStakingKeeper is a mock implementation of StakingKeeper
type MockStakingKeeper struct {
	validators     map[string]bool
	validatorTiers map[string]int32
	totalVals      uint64
}

func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		validators:     make(map[string]bool),
		validatorTiers: make(map[string]int32),
		totalVals:      0,
	}
}

func (m *MockStakingKeeper) IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool {
	return m.validators[addr.String()]
}

func (m *MockStakingKeeper) GetValidatorTier(ctx sdk.Context, validator sdk.AccAddress) (int32, error) {
	tier, ok := m.validatorTiers[validator.String()]
	if !ok {
		return 0, nil
	}
	return tier, nil
}

func (m *MockStakingKeeper) GetTotalValidators(ctx sdk.Context) uint64 {
	return m.totalVals
}

func (m *MockStakingKeeper) GetValidatorsByTier(ctx sdk.Context, minTier int32) ([]sdk.AccAddress, error) {
	validators := []sdk.AccAddress{}
	for addr, tier := range m.validatorTiers {
		if tier >= minTier {
			// Try to parse as bech32 first, if it fails use raw bytes
			accAddr, err := sdk.AccAddressFromBech32(addr)
			if err != nil {
				// Use raw bytes for test addresses like "validator1"
				accAddr = sdk.AccAddress(addr)
			}
			validators = append(validators, accAddr)
		}
	}
	return validators, nil
}

// MockEscrowKeeper is a mock implementation of EscrowKeeper for ban checking
type MockEscrowKeeper struct {
	bannedAddresses map[string]string // address -> ban reason
}

func NewMockEscrowKeeper() *MockEscrowKeeper {
	return &MockEscrowKeeper{
		bannedAddresses: make(map[string]string),
	}
}

func (m *MockEscrowKeeper) IsAddressBanned(ctx sdk.Context, address string, blockTime time.Time) (bool, string) {
	if reason, banned := m.bannedAddresses[address]; banned {
		return true, reason
	}
	return false, ""
}

func (m *MockEscrowKeeper) BanAddress(address string, reason string) {
	m.bannedAddresses[address] = reason
}

func (m *MockEscrowKeeper) UnbanAddress(address string) {
	delete(m.bannedAddresses, address)
}

// KeeperTestSuite is the test suite for keeper tests
type KeeperTestSuite struct {
	suite.Suite
	keeper        *keeper.Keeper
	ctx           sdk.Context
	bankKeeper    *MockBankKeeper
	stakingKeeper *MockStakingKeeper
	escrowKeeper  *MockEscrowKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create mock keepers
	suite.bankKeeper = NewMockBankKeeper()
	suite.stakingKeeper = NewMockStakingKeeper()
	suite.escrowKeeper = NewMockEscrowKeeper()
	accountKeeper := &MockAccountKeeper{}

	// Create in-memory store with proper multistore
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	err := stateStore.LoadLatestVersion()
	suite.Require().NoError(err)

	// Create context with multistore
	header := cometbfttypes.Header{Height: 1, Time: time.Now()}
	suite.ctx = sdk.NewContext(stateStore, header, false, log.NewNopLogger())

	// Create keeper
	authority := "cosmos1authority"
	suite.keeper = keeper.NewKeeper(
		cdc,
		storeKey,
		memKey,
		suite.bankKeeper,
		accountKeeper,
		suite.stakingKeeper,
		suite.escrowKeeper,
		authority,
	)

	// Set default params
	params := types.DefaultParams()
	err = suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
}

// TestGetSetParams tests parameter get/set functionality
func (suite *KeeperTestSuite) TestGetSetParams() {
	// Test default params
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().True(params.BridgingEnabled)
	suite.Require().Equal(math.LegacyNewDecWithPrec(67, 2), params.AttestationThreshold)

	// Test setting custom params
	customParams := types.Params{
		BridgingEnabled:        false,
		AttestationThreshold:   math.LegacyNewDecWithPrec(75, 2),
		MinValidatorTier:       3,
		WithdrawalTimelock:     7200,
		RateLimitWindow:        86400,
		MaxWithdrawalPerWindow: math.NewInt(500_000_000_000),
		BridgeFee:              math.LegacyNewDecWithPrec(2, 3),
		TSSThreshold:           math.LegacyNewDecWithPrec(75, 2),
		EmergencyPauseEnabled:  true,
	}

	err := suite.keeper.SetParams(suite.ctx, customParams)
	suite.Require().NoError(err)

	// Verify params were set
	retrievedParams := suite.keeper.GetParams(suite.ctx)
	suite.Require().False(retrievedParams.BridgingEnabled)
	suite.Require().Equal(math.LegacyNewDecWithPrec(75, 2), retrievedParams.AttestationThreshold)
	suite.Require().Equal(int32(3), retrievedParams.MinValidatorTier)
}

// TestParamsValidation tests parameter validation
func (suite *KeeperTestSuite) TestParamsValidation() {
	tests := []struct {
		name    string
		params  types.Params
		wantErr bool
	}{
		{
			name:    "default params are valid",
			params:  types.DefaultParams(),
			wantErr: false,
		},
		{
			name: "attestation threshold too low",
			params: types.Params{
				BridgingEnabled:        true,
				AttestationThreshold:   math.LegacyNewDecWithPrec(40, 2), // 40% < 50%
				MinValidatorTier:       4,
				WithdrawalTimelock:     3600,
				RateLimitWindow:        86400,
				MaxWithdrawalPerWindow: math.NewInt(1_000_000_000_000),
				BridgeFee:              math.LegacyNewDecWithPrec(1, 3),
				TSSThreshold:           math.LegacyNewDecWithPrec(67, 2),
				EmergencyPauseEnabled:  false,
			},
			wantErr: true,
		},
		{
			name: "zero withdrawal timelock",
			params: types.Params{
				BridgingEnabled:        true,
				AttestationThreshold:   math.LegacyNewDecWithPrec(67, 2),
				MinValidatorTier:       4,
				WithdrawalTimelock:     0, // Invalid
				RateLimitWindow:        86400,
				MaxWithdrawalPerWindow: math.NewInt(1_000_000_000_000),
				BridgeFee:              math.LegacyNewDecWithPrec(1, 3),
				TSSThreshold:           math.LegacyNewDecWithPrec(67, 2),
				EmergencyPauseEnabled:  false,
			},
			wantErr: true,
		},
		{
			name: "bridge fee >= 100%",
			params: types.Params{
				BridgingEnabled:        true,
				AttestationThreshold:   math.LegacyNewDecWithPrec(67, 2),
				MinValidatorTier:       4,
				WithdrawalTimelock:     3600,
				RateLimitWindow:        86400,
				MaxWithdrawalPerWindow: math.NewInt(1_000_000_000_000),
				BridgeFee:              math.LegacyOneDec(), // 100% is invalid
				TSSThreshold:           math.LegacyNewDecWithPrec(67, 2),
				EmergencyPauseEnabled:  false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.params.Validate()
			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestExternalChainManagement tests external chain CRUD operations
func (suite *KeeperTestSuite) TestExternalChainManagement() {
	chain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		true,
		12,
		12,
		"0x1234567890abcdef",
		math.NewInt(100_000),
		math.NewInt(10_000_000),
	)

	// Test SetExternalChain
	suite.keeper.SetExternalChain(suite.ctx, chain)

	// Test GetExternalChain
	retrieved, found := suite.keeper.GetExternalChain(suite.ctx, "ethereum-1")
	suite.Require().True(found)
	suite.Require().Equal(chain.ChainID, retrieved.ChainID)
	suite.Require().Equal(chain.ChainName, retrieved.ChainName)
	suite.Require().True(retrieved.Enabled)

	// Test GetAllExternalChains
	chains := suite.keeper.GetAllExternalChains(suite.ctx)
	suite.Require().Len(chains, 1)
	suite.Require().Equal("ethereum-1", chains[0].ChainID)

	// Test non-existent chain
	_, found = suite.keeper.GetExternalChain(suite.ctx, "bitcoin-1")
	suite.Require().False(found)
}

// TestExternalAssetManagement tests external asset CRUD operations
func (suite *KeeperTestSuite) TestExternalAssetManagement() {
	asset := types.NewExternalAsset(
		"ethereum-1",
		"USDT",
		"Tether USD",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7",
		6,
		true,
		math.LegacyOneDec(),
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)

	// Test SetExternalAsset
	suite.keeper.SetExternalAsset(suite.ctx, asset)

	// Test GetExternalAsset
	retrieved, found := suite.keeper.GetExternalAsset(suite.ctx, "ethereum-1", "USDT")
	suite.Require().True(found)
	suite.Require().Equal(asset.AssetSymbol, retrieved.AssetSymbol)
	suite.Require().Equal(asset.AssetName, retrieved.AssetName)
	suite.Require().True(retrieved.Enabled)

	// Test GetAllExternalAssets
	assets := suite.keeper.GetAllExternalAssets(suite.ctx)
	suite.Require().Len(assets, 1)
	suite.Require().Equal("USDT", assets[0].AssetSymbol)

	// Test non-existent asset
	_, found = suite.keeper.GetExternalAsset(suite.ctx, "ethereum-1", "USDC")
	suite.Require().False(found)
}

// TestIDManagement tests deposit/withdrawal/TSS ID counters
func (suite *KeeperTestSuite) TestIDManagement() {
	// Test deposit IDs
	id1 := suite.keeper.GetNextDepositID(suite.ctx)
	suite.Require().Equal(uint64(1), id1)

	id2 := suite.keeper.GetNextDepositID(suite.ctx)
	suite.Require().Equal(uint64(2), id2)

	// Test withdrawal IDs
	wid1 := suite.keeper.GetNextWithdrawalID(suite.ctx)
	suite.Require().Equal(uint64(1), wid1)

	wid2 := suite.keeper.GetNextWithdrawalID(suite.ctx)
	suite.Require().Equal(uint64(2), wid2)

	// Test TSS session IDs
	sid1 := suite.keeper.GetNextTSSSessionID(suite.ctx)
	suite.Require().Equal(uint64(1), sid1)

	sid2 := suite.keeper.GetNextTSSSessionID(suite.ctx)
	suite.Require().Equal(uint64(2), sid2)
}

// TestValidatorEligibility tests validator eligibility checks
func (suite *KeeperTestSuite) TestValidatorEligibility() {
	validator := sdk.AccAddress("validator1")

	// Test non-validator
	eligible, err := suite.keeper.IsValidatorEligible(suite.ctx, validator)
	suite.Require().Error(err)
	suite.Require().False(eligible)
	suite.Require().Equal(types.ErrNotValidator, err)

	// Set up validator with insufficient tier
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 2 // Below default tier 4

	eligible, err = suite.keeper.IsValidatorEligible(suite.ctx, validator)
	suite.Require().Error(err)
	suite.Require().False(eligible)
	suite.Require().Equal(types.ErrInsufficientTier, err)

	// Set up validator with sufficient tier
	suite.stakingKeeper.validatorTiers[validator.String()] = 4

	eligible, err = suite.keeper.IsValidatorEligible(suite.ctx, validator)
	suite.Require().NoError(err)
	suite.Require().True(eligible)
}

// TestCalculateRequiredAttestations tests attestation calculation
func (suite *KeeperTestSuite) TestCalculateRequiredAttestations() {
	// Test with no eligible validators
	suite.stakingKeeper.totalVals = 0
	suite.stakingKeeper.validatorTiers = make(map[string]int32) // Clear validators
	_, err := suite.keeper.CalculateRequiredAttestations(suite.ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "no eligible validators")

	// Test with 10 eligible validators (67% threshold = 7 required, ceiling of 6.7)
	suite.stakingKeeper.totalVals = 10
	for i := 1; i <= 10; i++ {
		v := sdk.AccAddress(fmt.Sprintf("validator%d", i))
		suite.stakingKeeper.validators[v.String()] = true
		suite.stakingKeeper.validatorTiers[v.String()] = 4 // Tier 4 meets min requirement
	}
	required, err := suite.keeper.CalculateRequiredAttestations(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(7), required) // Ceiling of 10 * 0.67 = 7

	// Test with 3 eligible validators (67% threshold = 3 required, ceiling of 2.01)
	suite.stakingKeeper.validatorTiers = make(map[string]int32) // Clear validators
	for i := 1; i <= 3; i++ {
		v := sdk.AccAddress(fmt.Sprintf("validator%d", i))
		suite.stakingKeeper.validators[v.String()] = true
		suite.stakingKeeper.validatorTiers[v.String()] = 4
	}
	suite.stakingKeeper.totalVals = 3
	required, err = suite.keeper.CalculateRequiredAttestations(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(3), required) // Ceiling of 3 * 0.67 = 3
}

// TestCircuitBreaker tests circuit breaker functionality
func (suite *KeeperTestSuite) TestCircuitBreaker() {
	// Initially, all operations should be allowed
	err := suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().NoError(err)

	err = suite.keeper.IsOperationAllowed(suite.ctx, "withdraw")
	suite.Require().NoError(err)

	// Enable circuit breaker with selective operations
	duration := 1 * time.Hour
	cb := types.NewCircuitBreaker(
		true,
		"Security incident",
		"cosmos1authority",
		false, // Deposits disabled
		true,  // Withdrawals allowed
		true,  // Attestations allowed
		&duration,
	)
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// Test deposit blocked
	err = suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrCircuitBreakerActive, err)

	// Test withdrawal allowed
	err = suite.keeper.IsOperationAllowed(suite.ctx, "withdraw")
	suite.Require().NoError(err)

	// Test circuit breaker retrieval
	retrieved := suite.keeper.GetCircuitBreaker(suite.ctx)
	suite.Require().True(retrieved.Enabled)
	suite.Require().Equal("Security incident", retrieved.Reason)
}

// TestCircuitBreakerExpiration tests circuit breaker auto-expiry
func (suite *KeeperTestSuite) TestCircuitBreakerExpiration() {
	// Enable circuit breaker with 1 hour expiry
	duration := 1 * time.Hour
	cb := types.NewCircuitBreaker(
		true,
		"Temporary pause",
		"cosmos1authority",
		false,
		false,
		false,
		&duration,
	)
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// Operations should be blocked
	err := suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().Error(err)

	// Move time forward past expiration
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(2 * time.Hour))

	// Operations should now be allowed (circuit breaker auto-disabled)
	err = suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().NoError(err)
}

// TestAssetConversion tests HODL conversion functions
func (suite *KeeperTestSuite) TestAssetConversion() {
	// Set up external asset with 1:1 conversion rate
	asset := types.NewExternalAsset(
		"ethereum-1",
		"USDT",
		"Tether USD",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7",
		6,
		true,
		math.LegacyOneDec(), // 1 USDT = 1 HODL
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, asset)

	// Test ConvertToHODL
	externalAmount := math.NewInt(1_000_000) // 1 USDT (6 decimals)
	hodlAmount, err := suite.keeper.ConvertToHODL(suite.ctx, "ethereum-1", "USDT", externalAmount)
	suite.Require().NoError(err)
	suite.Require().Equal(externalAmount, hodlAmount)

	// Test ConvertFromHODL
	convertedBack, err := suite.keeper.ConvertFromHODL(suite.ctx, "ethereum-1", "USDT", hodlAmount)
	suite.Require().NoError(err)
	suite.Require().Equal(externalAmount, convertedBack)

	// Test with non-existent asset
	_, err = suite.keeper.ConvertToHODL(suite.ctx, "ethereum-1", "BTC", externalAmount)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrAssetNotSupported, err)
}

// TestAssetConversionWithRate tests conversion with different rates
func (suite *KeeperTestSuite) TestAssetConversionWithRate() {
	// Set up external asset with 2:1 conversion rate (1 asset = 2 HODL)
	asset := types.NewExternalAsset(
		"ethereum-1",
		"ETH",
		"Ethereum",
		"native",
		18,
		true,
		math.LegacyNewDec(2), // 1 ETH = 2 HODL
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, asset)

	// Test ConvertToHODL
	externalAmount := math.NewInt(1_000_000_000_000_000_000) // 1 ETH (18 decimals)
	hodlAmount, err := suite.keeper.ConvertToHODL(suite.ctx, "ethereum-1", "ETH", externalAmount)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(2_000_000_000_000_000_000), hodlAmount) // 2 HODL

	// Test ConvertFromHODL
	convertedBack, err := suite.keeper.ConvertFromHODL(suite.ctx, "ethereum-1", "ETH", hodlAmount)
	suite.Require().NoError(err)
	suite.Require().Equal(externalAmount, convertedBack)
}

// TestBanChecking tests that banned addresses cannot use the bridge
func (suite *KeeperTestSuite) TestBanChecking() {
	// Initially, address should not be banned
	address := "cosmos1test123456789"
	banned, reason := suite.keeper.IsAddressBanned(suite.ctx, address)
	suite.Require().False(banned)
	suite.Require().Empty(reason)

	// Ban the address
	suite.escrowKeeper.BanAddress(address, "fraudulent activity")

	// Now address should be banned
	banned, reason = suite.keeper.IsAddressBanned(suite.ctx, address)
	suite.Require().True(banned)
	suite.Require().Equal("fraudulent activity", reason)

	// Unban the address
	suite.escrowKeeper.UnbanAddress(address)

	// Address should no longer be banned
	banned, reason = suite.keeper.IsAddressBanned(suite.ctx, address)
	suite.Require().False(banned)
	suite.Require().Empty(reason)
}

// TestBannedAddressCannotWithdraw tests that banned addresses cannot request withdrawals
func (suite *KeeperTestSuite) TestBannedAddressCannotWithdraw() {
	// Set up chain and asset
	chain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		true,
		6,   // confirmations
		12,  // blockTime
		"0xBridge...",
		math.NewInt(100_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalChain(suite.ctx, chain)

	asset := types.NewExternalAsset(
		"ethereum-1",
		"USDT",
		"Tether USD",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7",
		6,
		true,
		math.LegacyOneDec(),
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, asset)

	// Create sender address
	sender := sdk.AccAddress("testsender12345678")

	// Ban the sender
	suite.escrowKeeper.BanAddress(sender.String(), "scam activity")

	// Attempt to withdraw - should fail
	_, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xRecipient...",
		math.NewInt(1_000_000),
	)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "banned")

	// Unban the sender
	suite.escrowKeeper.UnbanAddress(sender.String())

	// Give sender balance
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin("uhodl", math.NewInt(10_000_000)))

	// Now withdrawal should be allowed (may fail for other reasons, but not ban)
	_, err = suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xRecipient...",
		math.NewInt(1_000_000),
	)
	// If it errors, it should NOT be because of a ban
	if err != nil {
		suite.Require().NotContains(err.Error(), "banned")
	}
}
