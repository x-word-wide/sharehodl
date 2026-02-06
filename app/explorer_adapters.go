package app

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	equitykeeper "github.com/sharehodl/sharehodl-blockchain/x/equity/keeper"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	hodlkeeper "github.com/sharehodl/sharehodl-blockchain/x/hodl/keeper"
	universalstakingkeeper "github.com/sharehodl/sharehodl-blockchain/x/staking/keeper"
)

// ExplorerEquityKeeperAdapter wraps the equity keeper for explorer module
type ExplorerEquityKeeperAdapter struct {
	equityKeeper equitykeeper.Keeper
}

// NewExplorerEquityKeeperAdapter creates a new adapter for explorer equity keeper interface
func NewExplorerEquityKeeperAdapter(equityKeeper equitykeeper.Keeper) *ExplorerEquityKeeperAdapter {
	return &ExplorerEquityKeeperAdapter{equityKeeper: equityKeeper}
}

// GetCompany returns company information
func (a *ExplorerEquityKeeperAdapter) GetCompany(ctx sdk.Context, id uint64) (equitytypes.Company, bool) {
	company, found := a.equityKeeper.GetCompany(ctx, id)
	if !found {
		return equitytypes.Company{}, false
	}
	if c, ok := company.(equitytypes.Company); ok {
		return c, true
	}
	return equitytypes.Company{}, false
}

// GetCompanies returns all companies
func (a *ExplorerEquityKeeperAdapter) GetCompanies(ctx sdk.Context) []equitytypes.Company {
	return a.equityKeeper.GetAllCompanies(ctx)
}

// GetShareholding returns shareholding information
func (a *ExplorerEquityKeeperAdapter) GetShareholding(ctx sdk.Context, companyID uint64, shareholder string) (equitytypes.Shareholding, bool) {
	// Get all shareholdings for this owner and filter by companyID
	shareholdings := a.equityKeeper.GetOwnerShareholdings(ctx, shareholder)
	for _, sh := range shareholdings {
		if sh.CompanyID == companyID {
			return sh, true
		}
	}
	return equitytypes.Shareholding{}, false
}

// GetDividend returns dividend information
func (a *ExplorerEquityKeeperAdapter) GetDividend(ctx sdk.Context, id uint64) (equitytypes.Dividend, bool) {
	return a.equityKeeper.GetDividend(ctx, id)
}

// GetDividends returns all dividends
func (a *ExplorerEquityKeeperAdapter) GetDividends(ctx sdk.Context) []equitytypes.Dividend {
	return a.equityKeeper.GetAllDividends(ctx)
}

// GetDividendPayment returns dividend payment information
func (a *ExplorerEquityKeeperAdapter) GetDividendPayment(ctx sdk.Context, dividendID uint64, shareholder string) (equitytypes.DividendPayment, bool) {
	return a.equityKeeper.GetDividendPaymentForShareholder(ctx, dividendID, shareholder)
}

// ExplorerHODLKeeperAdapter wraps the HODL keeper for explorer module
type ExplorerHODLKeeperAdapter struct {
	hodlKeeper hodlkeeper.Keeper
}

// NewExplorerHODLKeeperAdapter creates a new adapter for explorer HODL keeper interface
func NewExplorerHODLKeeperAdapter(hodlKeeper hodlkeeper.Keeper) *ExplorerHODLKeeperAdapter {
	return &ExplorerHODLKeeperAdapter{hodlKeeper: hodlKeeper}
}

// GetBalance returns account HODL balance
func (a *ExplorerHODLKeeperAdapter) GetBalance(ctx sdk.Context, addr sdk.AccAddress) math.Int {
	return a.hodlKeeper.GetBalance(ctx, addr)
}

// GetTotalSupply returns total HODL supply
func (a *ExplorerHODLKeeperAdapter) GetTotalSupply(ctx sdk.Context) math.Int {
	supply := a.hodlKeeper.GetTotalSupply(ctx)
	// Convert interface{} to math.Int
	if supplyInt, ok := supply.(math.Int); ok {
		return supplyInt
	}
	return math.ZeroInt()
}

// ExplorerStakingKeeperAdapter wraps the staking keeper for explorer module
type ExplorerStakingKeeperAdapter struct {
	stakingKeeper *universalstakingkeeper.Keeper
}

// NewExplorerStakingKeeperAdapter creates a new adapter for explorer staking keeper interface
func NewExplorerStakingKeeperAdapter(stakingKeeper *universalstakingkeeper.Keeper) *ExplorerStakingKeeperAdapter {
	return &ExplorerStakingKeeperAdapter{stakingKeeper: stakingKeeper}
}

// GetUserStake returns user stake information as interface{}
func (a *ExplorerStakingKeeperAdapter) GetUserStake(ctx sdk.Context, addr sdk.AccAddress) (interface{}, bool) {
	if a.stakingKeeper == nil {
		return nil, false
	}
	return a.stakingKeeper.GetUserStake(ctx, addr)
}

// GetUserTier returns user tier
func (a *ExplorerStakingKeeperAdapter) GetUserTier(ctx sdk.Context, addr sdk.AccAddress) int {
	if a.stakingKeeper == nil {
		return 0
	}
	return int(a.stakingKeeper.GetUserTier(ctx, addr))
}

// GetReputation returns user reputation score
func (a *ExplorerStakingKeeperAdapter) GetReputation(ctx sdk.Context, addr sdk.AccAddress) math.LegacyDec {
	if a.stakingKeeper == nil {
		return math.LegacyZeroDec()
	}
	return a.stakingKeeper.GetReputation(ctx, addr)
}

// ExplorerBankKeeperAdapter wraps the bank keeper for explorer module
type ExplorerBankKeeperAdapter struct {
	bankKeeper bankkeeper.Keeper
}

// NewExplorerBankKeeperAdapter creates a new adapter for explorer bank keeper interface
func NewExplorerBankKeeperAdapter(bankKeeper bankkeeper.Keeper) *ExplorerBankKeeperAdapter {
	return &ExplorerBankKeeperAdapter{bankKeeper: bankKeeper}
}

// GetBalance returns account balance for a specific denom
func (a *ExplorerBankKeeperAdapter) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return a.bankKeeper.GetBalance(ctx, addr, denom)
}

// GetAllBalances returns all account balances
func (a *ExplorerBankKeeperAdapter) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return a.bankKeeper.GetAllBalances(ctx, addr)
}

// GetSupply returns total supply for a specific denom
func (a *ExplorerBankKeeperAdapter) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	return a.bankKeeper.GetSupply(ctx, denom)
}
