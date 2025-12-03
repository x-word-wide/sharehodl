package keeper

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/explorer/types"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	governancetypes "github.com/sharehodl/sharehodl-blockchain/x/governance/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
	validatortypes "github.com/sharehodl/sharehodl-blockchain/x/validator/types"
)

// Keeper handles ShareScan explorer operations
type Keeper struct {
	cdc          codec.Codec
	storeService storetypes.KVStoreService
	txDecoder    sdk.TxDecoder

	// Module keepers for data access
	equityKeeper    EquityKeeper
	hodlKeeper      HODLKeeper
	validatorKeeper ValidatorKeeper
	governanceKeeper GovernanceKeeper
	bankKeeper      BankKeeper

	// Indexer for real-time data processing
	indexer *Indexer

	// Configuration
	config types.ExplorerConfig
}

// Interface definitions for module dependencies
type EquityKeeper interface {
	GetCompany(ctx sdk.Context, id uint64) (equitytypes.Company, bool)
	GetCompanies(ctx sdk.Context) []equitytypes.Company
	GetShareholding(ctx sdk.Context, companyID uint64, shareholder string) (equitytypes.Shareholding, bool)
	GetDividend(ctx sdk.Context, id uint64) (equitytypes.Dividend, bool)
	GetDividends(ctx sdk.Context) []equitytypes.Dividend
	GetDividendPayment(ctx sdk.Context, dividendID uint64, shareholder string) (equitytypes.DividendPayment, bool)
	GetTrade(ctx sdk.Context, id uint64) (equitytypes.Trade, bool)
	GetTrades(ctx sdk.Context) []equitytypes.Trade
}

type HODLKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress) math.Int
	GetTotalSupply(ctx sdk.Context) math.Int
	GetTransactions(ctx sdk.Context, addr sdk.AccAddress) []hodltypes.Transaction
}

type ValidatorKeeper interface {
	GetValidator(ctx sdk.Context, addr string) (validatortypes.Validator, bool)
	GetValidators(ctx sdk.Context) []validatortypes.Validator
	GetValidatorTier(ctx sdk.Context, addr string) validatortypes.ValidatorTier
	IsValidatorActive(ctx sdk.Context, addr string) bool
}

type GovernanceKeeper interface {
	GetProposal(ctx sdk.Context, proposalID uint64) (governancetypes.Proposal, bool)
	GetProposals(ctx sdk.Context) []governancetypes.Proposal
	GetVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (governancetypes.Vote, bool)
	GetTallyResult(ctx sdk.Context, proposalID uint64) (governancetypes.TallyResult, bool)
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
}

// NewKeeper creates a new ShareScan explorer Keeper
func NewKeeper(
	cdc codec.Codec,
	storeService storetypes.KVStoreService,
	txDecoder sdk.TxDecoder,
	equityKeeper EquityKeeper,
	hodlKeeper HODLKeeper,
	validatorKeeper ValidatorKeeper,
	governanceKeeper GovernanceKeeper,
	bankKeeper BankKeeper,
) *Keeper {
	keeper := &Keeper{
		cdc:              cdc,
		storeService:     storeService,
		txDecoder:        txDecoder,
		equityKeeper:     equityKeeper,
		hodlKeeper:       hodlKeeper,
		validatorKeeper:  validatorKeeper,
		governanceKeeper: governanceKeeper,
		bankKeeper:       bankKeeper,
		config: types.ExplorerConfig{
			IndexingEnabled:      true,
			RealTimeUpdates:      true,
			AnalyticsEnabled:     true,
			NotificationsEnabled: true,
			MaxSearchResults:     100,
			CacheTimeout:         time.Hour,
			UpdateInterval:       time.Second * 10,
			HistoryDepth:         1000000, // 1M blocks
			EnabledModules:       []string{"hodl", "equity", "dex", "governance", "validator"},
		},
	}

	keeper.indexer = NewIndexer(keeper)
	return keeper
}

// Block Explorer Methods

// GetBlockInfo retrieves detailed block information
func (k Keeper) GetBlockInfo(ctx sdk.Context, height int64) (types.BlockInfo, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	blockStore := prefix.NewStore(store, types.BlockPrefix)

	bz := blockStore.Get(sdk.Uint64ToBigEndian(uint64(height)))
	if bz == nil {
		return types.BlockInfo{}, false
	}

	var blockInfo types.BlockInfo
	err := json.Unmarshal(bz, &blockInfo)
	if err != nil {
		return types.BlockInfo{}, false
	}

	return blockInfo, true
}

// GetLatestBlocks retrieves the latest N blocks
func (k Keeper) GetLatestBlocks(ctx sdk.Context, limit uint32) []types.BlockInfo {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	blockStore := prefix.NewStore(store, types.BlockPrefix)

	// Get latest blocks in reverse order
	iterator := blockStore.ReverseIterator(nil, nil)
	defer iterator.Close()

	blocks := make([]types.BlockInfo, 0, limit)
	count := uint32(0)

	for ; iterator.Valid() && count < limit; iterator.Next() {
		var blockInfo types.BlockInfo
		err := json.Unmarshal(iterator.Value(), &blockInfo)
		if err != nil {
			continue
		}
		blocks = append(blocks, blockInfo)
		count++
	}

	return blocks
}

// GetTransactionInfo retrieves detailed transaction information
func (k Keeper) GetTransactionInfo(ctx sdk.Context, hash string) (types.TransactionInfo, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	txStore := prefix.NewStore(store, types.TransactionPrefix)

	bz := txStore.Get([]byte(hash))
	if bz == nil {
		return types.TransactionInfo{}, false
	}

	var txInfo types.TransactionInfo
	err := json.Unmarshal(bz, &txInfo)
	if err != nil {
		return types.TransactionInfo{}, false
	}

	return txInfo, true
}

// GetBlockTransactions retrieves all transactions in a block
func (k Keeper) GetBlockTransactions(ctx sdk.Context, height int64) []types.TransactionInfo {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	blockTxStore := prefix.NewStore(store, types.BlockTransactionPrefix)

	heightKey := sdk.Uint64ToBigEndian(uint64(height))
	iterator := blockTxStore.Iterator(heightKey, sdk.PrefixEndBytes(heightKey))
	defer iterator.Close()

	transactions := make([]types.TransactionInfo, 0)

	for ; iterator.Valid(); iterator.Next() {
		// Extract transaction hash from key
		key := iterator.Key()
		txHash := string(key[len(heightKey):])
		
		// Get full transaction info
		if txInfo, found := k.GetTransactionInfo(ctx, txHash); found {
			transactions = append(transactions, txInfo)
		}
	}

	return transactions
}

// Account Explorer Methods

// GetAccountInfo retrieves comprehensive account information
func (k Keeper) GetAccountInfo(ctx sdk.Context, address string) types.AccountInfo {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return types.AccountInfo{}
	}

	// Get basic account information
	allBalances := k.bankKeeper.GetAllBalances(ctx, addr)
	hodlBalance := k.hodlKeeper.GetBalance(ctx, addr)

	// Check if validator
	validator, isValidator := k.validatorKeeper.GetValidator(ctx, address)
	var validatorTier string
	var businessName string
	if isValidator {
		tier := k.validatorKeeper.GetValidatorTier(ctx, address)
		validatorTier = string(tier)
		businessName = validator.BusinessName
	}

	// Get shareholdings
	shareholdings := k.getAccountShareholdings(ctx, address)

	// Get recent transactions
	recentTxs := k.getAccountTransactions(ctx, address, 10)

	// Get governance votes
	governanceVotes := k.getAccountVotes(ctx, address)

	// Get dividend history
	dividendHistory := k.getAccountDividends(ctx, address)

	// Get trading history
	tradingHistory := k.getAccountTrades(ctx, address, 20)

	// Calculate totals
	totalDividends := k.calculateTotalDividends(dividendHistory)
	totalTrades := uint64(len(tradingHistory))
	tradingVolume := k.calculateTradingVolume(tradingHistory)

	return types.AccountInfo{
		Address:         address,
		Balance:         allBalances,
		HODLBalance:     hodlBalance,
		IsValidator:     isValidator,
		ValidatorTier:   validatorTier,
		BusinessName:    businessName,
		Shareholdings:   shareholdings,
		RecentTxs:       recentTxs,
		GovernanceVotes: governanceVotes,
		DividendHistory: dividendHistory,
		TradingHistory:  tradingHistory,
		TotalDividends:  totalDividends,
		TotalTrades:     totalTrades,
		TradingVolume:   tradingVolume,
	}
}

// Company Explorer Methods

// GetCompanyInfo retrieves comprehensive company information
func (k Keeper) GetCompanyInfo(ctx sdk.Context, companyID uint64) (types.CompanyInfo, bool) {
	company, found := k.equityKeeper.GetCompany(ctx, companyID)
	if !found {
		return types.CompanyInfo{}, false
	}

	// Get share classes
	shareClasses := k.getCompanyShareClasses(ctx, companyID)

	// Calculate market metrics
	marketCap, lastPrice := k.calculateMarketMetrics(ctx, companyID)
	volume24h := k.getCompany24hVolume(ctx, companyID)
	priceChange24h := k.getCompanyPriceChange24h(ctx, companyID)

	// Get shareholders count
	shareholders := k.getShareholdersCount(ctx, companyID)

	// Get recent dividends and trades
	recentDividends := k.getCompanyDividends(ctx, companyID, 5)
	recentTrades := k.getCompanyTrades(ctx, companyID, 10)

	return types.CompanyInfo{
		ID:              companyID,
		Name:            company.Name,
		Symbol:          company.Symbol,
		BusinessType:    company.BusinessType,
		Jurisdiction:    company.Jurisdiction,
		ListingDate:     company.ListingDate,
		Status:          string(company.Status),
		TotalShares:     company.TotalShares,
		ShareClasses:    shareClasses,
		MarketCap:       marketCap,
		LastPrice:       lastPrice,
		Volume24h:       volume24h,
		PriceChange24h:  priceChange24h,
		Shareholders:    shareholders,
		RecentDividends: recentDividends,
		RecentTrades:    recentTrades,
	}, true
}

// GetCompanies retrieves all listed companies
func (k Keeper) GetCompanies(ctx sdk.Context) []types.CompanyInfo {
	companies := k.equityKeeper.GetCompanies(ctx)
	companyInfos := make([]types.CompanyInfo, 0, len(companies))

	for _, company := range companies {
		if info, found := k.GetCompanyInfo(ctx, company.ID); found {
			companyInfos = append(companyInfos, info)
		}
	}

	return companyInfos
}

// Governance Explorer Methods

// GetGovernanceProposalInfo retrieves comprehensive proposal information
func (k Keeper) GetGovernanceProposalInfo(ctx sdk.Context, proposalID uint64) (types.GovernanceProposalInfo, bool) {
	proposal, found := k.governanceKeeper.GetProposal(ctx, proposalID)
	if !found {
		return types.GovernanceProposalInfo{}, false
	}

	// Get deposits, votes, and tally
	deposits := k.getProposalDeposits(ctx, proposalID)
	votes := k.getProposalVotes(ctx, proposalID)
	tally, _ := k.governanceKeeper.GetTallyResult(ctx, proposalID)

	// Convert tally to explorer format
	tallyInfo := types.TallyInfo{
		ProposalID:         proposalID,
		YesVotes:          tally.YesVotes,
		NoVotes:           tally.NoVotes,
		AbstainVotes:      tally.AbstainVotes,
		VetoVotes:         tally.VetoVotes,
		TotalVotes:        tally.TotalVotes,
		TotalEligibleVotes: tally.TotalEligibleVotes,
		ParticipationRate: tally.ParticipationRate,
		YesPercentage:     tally.YesPercentage,
		NoPercentage:      tally.NoPercentage,
		AbstainPercentage: tally.AbstainPercentage,
		VetoPercentage:    tally.VetoPercentage,
		Outcome:           string(tally.Outcome),
		QuorumMet:         tally.ParticipationRate.GTE(proposal.QuorumRequired),
		ThresholdMet:      tally.YesPercentage.GTE(proposal.ThresholdRequired),
	}

	return types.GovernanceProposalInfo{
		ID:                proposal.ID,
		Type:              string(proposal.Type),
		Title:             proposal.Title,
		Description:       proposal.Description,
		Submitter:         proposal.Submitter,
		Status:            string(proposal.Status),
		SubmissionTime:    proposal.SubmissionTime,
		DepositEndTime:    proposal.DepositEndTime,
		VotingStartTime:   proposal.VotingStartTime,
		VotingEndTime:     proposal.VotingEndTime,
		TotalDeposit:      proposal.TotalDeposit,
		Deposits:          deposits,
		Votes:             votes,
		TallyResult:       tallyInfo,
		QuorumRequired:    proposal.QuorumRequired,
		ThresholdRequired: proposal.ThresholdRequired,
		Outcome:           string(tally.Outcome),
		ExecutionResult:   "", // Would be extracted from execution events
	}, true
}

// Analytics Methods

// GetNetworkStats retrieves comprehensive network statistics
func (k Keeper) GetNetworkStats(ctx sdk.Context) types.NetworkStats {
	// Get latest block
	latestHeight := ctx.BlockHeight()
	latestBlock, _ := k.GetBlockInfo(ctx, latestHeight)

	// Get cached stats or calculate
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	statsStore := prefix.NewStore(store, types.NetworkStatsPrefix)
	
	statsBytes := statsStore.Get([]byte("current"))
	var stats types.NetworkStats
	if statsBytes != nil {
		json.Unmarshal(statsBytes, &stats)
		// Update with latest block
		stats.LatestBlock = latestBlock
	} else {
		// Calculate fresh stats
		stats = k.calculateNetworkStats(ctx)
	}

	return stats
}

// GetAnalytics retrieves time-series analytics data
func (k Keeper) GetAnalytics(ctx sdk.Context, timeFrame string, startTime, endTime time.Time) types.Analytics {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	analyticsStore := prefix.NewStore(store, types.AnalyticsPrefix)

	dataPoints := make([]types.AnalyticsDataPoint, 0)

	// Iterate through time range
	startKey := sdk.Uint64ToBigEndian(uint64(startTime.Unix()))
	endKey := sdk.Uint64ToBigEndian(uint64(endTime.Unix()))
	
	iterator := analyticsStore.Iterator(startKey, endKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var dataPoint types.AnalyticsDataPoint
		err := json.Unmarshal(iterator.Value(), &dataPoint)
		if err != nil {
			continue
		}
		dataPoints = append(dataPoints, dataPoint)
	}

	return types.Analytics{
		TimeFrame:  timeFrame,
		StartTime:  startTime,
		EndTime:    endTime,
		DataPoints: dataPoints,
	}
}

// Search Methods

// Search performs comprehensive search across all indexed data
func (k Keeper) Search(ctx sdk.Context, query string, limit uint32) []types.SearchResult {
	results := make([]types.SearchResult, 0)
	
	// Convert query to lowercase for case-insensitive search
	query = strings.ToLower(query)
	
	// Search blocks by height
	if height, err := strconv.ParseInt(query, 10, 64); err == nil {
		if blockInfo, found := k.GetBlockInfo(ctx, height); found {
			results = append(results, types.SearchResult{
				Type:        "block",
				ID:          strconv.FormatInt(height, 10),
				Title:       "Block #" + strconv.FormatInt(height, 10),
				Description: "Block with " + strconv.FormatUint(blockInfo.TotalTxs, 10) + " transactions",
				URL:         "/block/" + strconv.FormatInt(height, 10),
				Timestamp:   blockInfo.Timestamp,
				Relevance:   1.0,
				Data:        blockInfo,
			})
		}
	}
	
	// Search transactions by hash
	if txInfo, found := k.GetTransactionInfo(ctx, query); found {
		results = append(results, types.SearchResult{
			Type:        "transaction",
			ID:          query,
			Title:       "Transaction " + query[:16] + "...",
			Description: "Transaction with " + strconv.FormatUint(uint64(len(txInfo.Messages)), 10) + " messages",
			URL:         "/tx/" + query,
			Timestamp:   txInfo.Timestamp,
			Relevance:   1.0,
			Data:        txInfo,
		})
	}
	
	// Search accounts by address
	if _, err := sdk.AccAddressFromBech32(query); err == nil {
		accountInfo := k.GetAccountInfo(ctx, query)
		results = append(results, types.SearchResult{
			Type:        "account",
			ID:          query,
			Title:       "Account " + query[:16] + "...",
			Description: "Account with " + accountInfo.HODLBalance.String() + " HODL",
			URL:         "/account/" + query,
			Timestamp:   time.Now(),
			Relevance:   1.0,
			Data:        accountInfo,
		})
	}
	
	// Search companies by symbol or name
	companies := k.GetCompanies(ctx)
	for _, company := range companies {
		if strings.Contains(strings.ToLower(company.Symbol), query) ||
		   strings.Contains(strings.ToLower(company.Name), query) {
			results = append(results, types.SearchResult{
				Type:        "company",
				ID:          strconv.FormatUint(company.ID, 10),
				Title:       company.Name + " (" + company.Symbol + ")",
				Description: company.BusinessType + " company with market cap " + company.MarketCap.String(),
				URL:         "/company/" + strconv.FormatUint(company.ID, 10),
				Timestamp:   company.ListingDate,
				Relevance:   0.8,
				Data:        company,
			})
		}
	}
	
	// Sort results by relevance and timestamp
	sort.Slice(results, func(i, j int) bool {
		if results[i].Relevance == results[j].Relevance {
			return results[i].Timestamp.After(results[j].Timestamp)
		}
		return results[i].Relevance > results[j].Relevance
	})
	
	// Limit results
	if uint32(len(results)) > limit {
		results = results[:limit]
	}
	
	return results
}

// Real-time Methods

// GetLiveMetrics retrieves current live metrics
func (k Keeper) GetLiveMetrics(ctx sdk.Context) types.LiveMetrics {
	return types.LiveMetrics{
		CurrentBlock:      ctx.BlockHeight(),
		BlockTime:         time.Second * 6, // Average block time
		TPS:               math.LegacyZeroDec(), // Would calculate from recent blocks
		PendingTxs:        0, // Would get from mempool
		MemPoolSize:       0, // Would get from mempool
		ActiveConnections: 0, // Would get from network stats
		NodeVersion:       "ShareHODL v1.0.0",
		ChainID:           ctx.ChainID(),
		LastUpdate:        ctx.BlockTime(),
	}
}

// SubscribeToNotifications subscribes an address to notifications
func (k Keeper) SubscribeToNotifications(ctx sdk.Context, subscriber string, notificationType string) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	subscriptionStore := prefix.NewStore(store, types.SubscriptionPrefix)
	
	key := types.SubscriptionKey(subscriber, notificationType)
	subscriptionStore.Set(key, []byte{})
	
	return nil
}

// BeginBlocker processes start of block for explorer
func (k Keeper) BeginBlocker(ctx sdk.Context) {
	if k.config.IndexingEnabled {
		// Index the current block
		err := k.indexer.IndexBlock(ctx, ctx.BlockHeight())
		if err != nil {
			ctx.Logger().Error("Failed to index block", "height", ctx.BlockHeight(), "error", err)
		}
	}
}

// EndBlocker processes end of block for explorer
func (k Keeper) EndBlocker(ctx sdk.Context) {
	if k.config.RealTimeUpdates {
		// Update live metrics
		k.updateLiveMetrics(ctx)
	}
}

// Private helper methods

func (k Keeper) calculateNetworkStats(ctx sdk.Context) types.NetworkStats {
	// This would calculate comprehensive network statistics
	// For now, return basic structure
	return types.NetworkStats{
		TotalBlocks:     uint64(ctx.BlockHeight()),
		TotalAccounts:   0, // Would count unique addresses
		TotalCompanies:  0, // Would count from equity module
		AverageBlockTime: time.Second * 6,
		TPS:             math.LegacyZeroDec(),
	}
}

func (k Keeper) updateLiveMetrics(ctx sdk.Context) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	metricsStore := prefix.NewStore(store, types.LiveMetricsPrefix)
	
	metrics := k.GetLiveMetrics(ctx)
	metricsBytes, _ := json.Marshal(metrics)
	metricsStore.Set([]byte("current"), metricsBytes)
}