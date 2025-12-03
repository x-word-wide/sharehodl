package keeper

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/explorer/types"
)

// Helper methods for the ShareScan explorer keeper

// getAccountShareholdings retrieves shareholdings for an account
func (k Keeper) getAccountShareholdings(ctx sdk.Context, address string) []types.ShareholdingInfo {
	shareholdings := make([]types.ShareholdingInfo, 0)
	
	// Get all companies and check if account has shares
	companies := k.equityKeeper.GetCompanies(ctx)
	for _, company := range companies {
		shareholding, found := k.equityKeeper.GetShareholding(ctx, company.ID, address)
		if found && shareholding.Shares.GT(math.ZeroInt()) {
			// Calculate market value (placeholder - would use real market data)
			marketValue := math.LegacyNewDecFromInt(shareholding.Shares).Mul(math.LegacyNewDec(100)) // $100 per share placeholder
			
			shareholdings = append(shareholdings, types.ShareholdingInfo{
				CompanyID:     company.ID,
				CompanyName:   company.Name,
				CompanySymbol: company.Symbol,
				ClassID:       shareholding.ClassID,
				ClassName:     shareholding.ClassID, // Would map to actual class name
				Shares:        shareholding.Shares,
				Percentage:    math.LegacyZeroDec(), // Would calculate percentage of total shares
				MarketValue:   marketValue,
				AcquiredAt:    shareholding.AcquiredAt,
				LastUpdated:   shareholding.UpdatedAt,
			})
		}
	}
	
	return shareholdings
}

// getAccountTransactions retrieves recent transactions for an account
func (k Keeper) getAccountTransactions(ctx sdk.Context, address string, limit int) []types.TransactionSummary {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	accountTxStore := prefix.NewStore(store, types.AccountTransactionPrefix)
	
	addrPrefix := []byte(address)
	iterator := accountTxStore.ReverseIterator(addrPrefix, sdk.PrefixEndBytes(addrPrefix))
	defer iterator.Close()
	
	transactions := make([]types.TransactionSummary, 0, limit)
	count := 0
	
	for ; iterator.Valid() && count < limit; iterator.Next() {
		// Extract transaction hash from key
		key := iterator.Key()
		txHash := string(key[len(addrPrefix):])
		
		// Get full transaction info
		if txInfo, found := k.GetTransactionInfo(ctx, txHash); found {
			// Create transaction summary
			summary := types.TransactionSummary{
				Hash:        txInfo.Hash,
				Type:        k.extractPrimaryMessageType(txInfo.Messages),
				Amount:      k.extractTotalAmount(txInfo.Messages),
				Fee:         txInfo.Fee,
				Success:     txInfo.Success,
				Timestamp:   txInfo.Timestamp,
				BlockHeight: txInfo.BlockHeight,
			}
			transactions = append(transactions, summary)
			count++
		}
	}
	
	return transactions
}

// getAccountVotes retrieves governance votes for an account
func (k Keeper) getAccountVotes(ctx sdk.Context, address string) []types.VoteInfo {
	votes := make([]types.VoteInfo, 0)
	
	// Get all proposals and check if account voted
	proposals := k.governanceKeeper.GetProposals(ctx)
	voterAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return votes
	}
	
	for _, proposal := range proposals {
		vote, found := k.governanceKeeper.GetVote(ctx, proposal.ID, voterAddr)
		if found {
			voteInfo := types.VoteInfo{
				Voter:       address,
				ProposalID:  proposal.ID,
				Option:      string(vote.Option),
				Weight:      vote.Weight,
				VotingPower: vote.VotingPower,
				Reason:      vote.Reason,
				SubmittedAt: vote.SubmittedAt,
				BlockHeight: 0, // Would need to track this
				TxHash:      "", // Would need to track this
				IsWeighted:  false, // Simple vote
			}
			votes = append(votes, voteInfo)
		}
	}
	
	return votes
}

// getAccountDividends retrieves dividend history for an account
func (k Keeper) getAccountDividends(ctx sdk.Context, address string) []types.DividendPayment {
	dividends := make([]types.DividendPayment, 0)
	
	// Get all dividends and check if account received payments
	allDividends := k.equityKeeper.GetDividends(ctx)
	for _, dividend := range allDividends {
		payment, found := k.equityKeeper.GetDividendPayment(ctx, dividend.ID, address)
		if found {
			dividendPayment := types.DividendPayment{
				ID:          payment.ID,
				DividendID:  dividend.ID,
				Shareholder: address,
				SharesHeld:  payment.SharesHeld,
				Amount:      payment.Amount,
				TaxWithheld: payment.TaxWithheld,
				NetAmount:   payment.NetAmount,
				Currency:    dividend.Currency,
				PaidAt:      payment.PaidAt,
				BlockHeight: 0, // Would need to track this
				TxHash:      payment.TxHash,
				Status:      payment.Status,
			}
			dividends = append(dividends, dividendPayment)
		}
	}
	
	return dividends
}

// getAccountTrades retrieves trading history for an account
func (k Keeper) getAccountTrades(ctx sdk.Context, address string, limit int) []types.TradeSummary {
	trades := make([]types.TradeSummary, 0)
	
	// Get all trades and filter by trader
	allTrades := k.equityKeeper.GetTrades(ctx)
	count := 0
	
	for i := len(allTrades) - 1; i >= 0 && count < limit; i-- {
		trade := allTrades[i]
		if trade.Trader == address {
			company, found := k.equityKeeper.GetCompany(ctx, trade.CompanyID)
			symbol := ""
			if found {
				symbol = company.Symbol
			}
			
			tradeSummary := types.TradeSummary{
				ID:             trade.ID,
				CompanyID:      trade.CompanyID,
				CompanySymbol:  symbol,
				OrderType:      string(trade.OrderType),
				Side:           string(trade.Side),
				Price:          trade.Price,
				Quantity:       trade.Quantity,
				FilledQuantity: trade.FilledQuantity,
				TotalValue:     trade.Price.MulInt(trade.FilledQuantity),
				Status:         string(trade.Status),
				Trader:         trade.Trader,
				CreatedAt:      trade.CreatedAt,
				UpdatedAt:      trade.UpdatedAt,
				BlockHeight:    0, // Would need to track this
				TxHash:         "", // Would need to track this
			}
			trades = append(trades, tradeSummary)
			count++
		}
	}
	
	return trades
}

// getCompanyShareClasses retrieves share classes for a company
func (k Keeper) getCompanyShareClasses(ctx sdk.Context, companyID uint64) []types.ShareClassInfo {
	shareClasses := make([]types.ShareClassInfo, 0)
	
	company, found := k.equityKeeper.GetCompany(ctx, companyID)
	if !found {
		return shareClasses
	}
	
	// For now, create a default common share class
	// In a full implementation, this would query actual share classes
	shareClasses = append(shareClasses, types.ShareClassInfo{
		ID:                "common",
		Name:              "Common Shares",
		VotingRights:      true,
		DividendRights:    true,
		TotalShares:       company.TotalShares,
		OutstandingShares: company.TotalShares,
		LastPrice:         math.LegacyNewDec(100), // Placeholder
		MarketCap:         math.LegacyNewDecFromInt(company.TotalShares).Mul(math.LegacyNewDec(100)),
	})
	
	return shareClasses
}

// getCompanyDividends retrieves recent dividends for a company
func (k Keeper) getCompanyDividends(ctx sdk.Context, companyID uint64, limit int) []types.DividendSummary {
	dividends := make([]types.DividendSummary, 0)
	
	allDividends := k.equityKeeper.GetDividends(ctx)
	count := 0
	
	for i := len(allDividends) - 1; i >= 0 && count < limit; i-- {
		dividend := allDividends[i]
		if dividend.CompanyID == companyID {
			company, found := k.equityKeeper.GetCompany(ctx, companyID)
			companyName := ""
			if found {
				companyName = company.Name
			}
			
			// Calculate progress
			progressPercent := math.LegacyZeroDec()
			if dividend.EligibleShares.GT(math.ZeroInt()) {
				progressPercent = math.LegacyNewDecFromInt(dividend.SharesProcessed).Quo(math.LegacyNewDecFromInt(dividend.EligibleShares))
			}
			
			dividendSummary := types.DividendSummary{
				ID:              dividend.ID,
				CompanyID:       dividend.CompanyID,
				CompanyName:     companyName,
				Type:            string(dividend.Type),
				AmountPerShare:  dividend.AmountPerShare,
				TotalAmount:     dividend.TotalAmount,
				Currency:        dividend.Currency,
				Status:          string(dividend.Status),
				DeclarationDate: dividend.DeclarationDate,
				ExDividendDate:  dividend.ExDividendDate,
				RecordDate:      dividend.RecordDate,
				PaymentDate:     dividend.PaymentDate,
				EligibleShares:  dividend.EligibleShares,
				SharesPaid:      dividend.SharesProcessed,
				AmountPaid:      dividend.PaidAmount,
				ProgressPercent: progressPercent,
			}
			dividends = append(dividends, dividendSummary)
			count++
		}
	}
	
	return dividends
}

// getCompanyTrades retrieves recent trades for a company
func (k Keeper) getCompanyTrades(ctx sdk.Context, companyID uint64, limit int) []types.TradeSummary {
	trades := make([]types.TradeSummary, 0)
	
	allTrades := k.equityKeeper.GetTrades(ctx)
	count := 0
	
	for i := len(allTrades) - 1; i >= 0 && count < limit; i-- {
		trade := allTrades[i]
		if trade.CompanyID == companyID {
			company, found := k.equityKeeper.GetCompany(ctx, companyID)
			symbol := ""
			if found {
				symbol = company.Symbol
			}
			
			tradeSummary := types.TradeSummary{
				ID:             trade.ID,
				CompanyID:      trade.CompanyID,
				CompanySymbol:  symbol,
				OrderType:      string(trade.OrderType),
				Side:           string(trade.Side),
				Price:          trade.Price,
				Quantity:       trade.Quantity,
				FilledQuantity: trade.FilledQuantity,
				TotalValue:     trade.Price.MulInt(trade.FilledQuantity),
				Status:         string(trade.Status),
				Trader:         trade.Trader,
				CreatedAt:      trade.CreatedAt,
				UpdatedAt:      trade.UpdatedAt,
				BlockHeight:    0, // Would need to track this
				TxHash:         "", // Would need to track this
			}
			trades = append(trades, tradeSummary)
			count++
		}
	}
	
	return trades
}

// getProposalDeposits retrieves deposits for a governance proposal
func (k Keeper) getProposalDeposits(ctx sdk.Context, proposalID uint64) []types.DepositInfo {
	deposits := make([]types.DepositInfo, 0)
	
	// This would iterate through deposits for the proposal
	// For now, return empty slice as placeholder
	
	return deposits
}

// getProposalVotes retrieves votes for a governance proposal
func (k Keeper) getProposalVotes(ctx sdk.Context, proposalID uint64) []types.VoteInfo {
	votes := make([]types.VoteInfo, 0)
	
	// This would iterate through votes for the proposal
	// For now, return empty slice as placeholder
	
	return votes
}

// calculateMarketMetrics calculates market cap and last price for a company
func (k Keeper) calculateMarketMetrics(ctx sdk.Context, companyID uint64) (math.LegacyDec, math.LegacyDec) {
	// Get recent trades to calculate last price
	allTrades := k.equityKeeper.GetTrades(ctx)
	lastPrice := math.LegacyNewDec(100) // Default price
	
	// Find most recent trade for this company
	for i := len(allTrades) - 1; i >= 0; i-- {
		trade := allTrades[i]
		if trade.CompanyID == companyID && trade.FilledQuantity.GT(math.ZeroInt()) {
			lastPrice = trade.Price
			break
		}
	}
	
	// Calculate market cap
	company, found := k.equityKeeper.GetCompany(ctx, companyID)
	if !found {
		return math.LegacyZeroDec(), lastPrice
	}
	
	marketCap := math.LegacyNewDecFromInt(company.TotalShares).Mul(lastPrice)
	return marketCap, lastPrice
}

// getCompany24hVolume calculates 24h trading volume for a company
func (k Keeper) getCompany24hVolume(ctx sdk.Context, companyID uint64) math.LegacyDec {
	volume := math.LegacyZeroDec()
	
	// Calculate from trades in last 24 hours
	cutoffTime := ctx.BlockTime().Add(-24 * time.Hour)
	allTrades := k.equityKeeper.GetTrades(ctx)
	
	for _, trade := range allTrades {
		if trade.CompanyID == companyID && 
		   trade.CreatedAt.After(cutoffTime) && 
		   trade.FilledQuantity.GT(math.ZeroInt()) {
			tradeValue := trade.Price.MulInt(trade.FilledQuantity)
			volume = volume.Add(tradeValue)
		}
	}
	
	return volume
}

// getCompanyPriceChange24h calculates 24h price change for a company
func (k Keeper) getCompanyPriceChange24h(ctx sdk.Context, companyID uint64) math.LegacyDec {
	// Get current price and price 24h ago
	currentPrice := math.LegacyNewDec(100) // Placeholder
	price24hAgo := math.LegacyNewDec(95)   // Placeholder
	
	// Calculate percentage change
	if !price24hAgo.IsZero() {
		change := currentPrice.Sub(price24hAgo).Quo(price24hAgo).Mul(math.LegacyNewDec(100))
		return change
	}
	
	return math.LegacyZeroDec()
}

// getShareholdersCount counts unique shareholders for a company
func (k Keeper) getShareholdersCount(ctx sdk.Context, companyID uint64) uint64 {
	// This would iterate through all shareholdings and count unique addresses
	// For now, return placeholder
	return 42
}

// calculateTotalDividends calculates total dividends received by an account
func (k Keeper) calculateTotalDividends(dividends []types.DividendPayment) math.LegacyDec {
	total := math.LegacyZeroDec()
	
	for _, dividend := range dividends {
		total = total.Add(dividend.NetAmount)
	}
	
	return total
}

// calculateTradingVolume calculates total trading volume for an account
func (k Keeper) calculateTradingVolume(trades []types.TradeSummary) math.LegacyDec {
	volume := math.LegacyZeroDec()
	
	for _, trade := range trades {
		volume = volume.Add(trade.TotalValue)
	}
	
	return volume
}

// extractPrimaryMessageType extracts the primary message type from a transaction
func (k Keeper) extractPrimaryMessageType(messages []types.MessageInfo) string {
	if len(messages) == 0 {
		return "unknown"
	}
	
	// Use the first message type as primary
	msgType := messages[0].Type
	
	// Extract readable type name
	parts := strings.Split(msgType, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1] // Return the last part
	}
	
	return msgType
}

// extractTotalAmount extracts total amount from transaction messages
func (k Keeper) extractTotalAmount(messages []types.MessageInfo) math.LegacyDec {
	total := math.LegacyZeroDec()
	
	// This would analyze message content to extract amounts
	// For now, return placeholder
	
	return total
}

// storeAccountIndex stores account indexing information
func (k Keeper) storeAccountIndex(ctx sdk.Context, address string, txHash string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	accountTxStore := prefix.NewStore(store, types.AccountTransactionPrefix)
	
	key := types.AccountTransactionKey(address, txHash)
	accountTxStore.Set(key, []byte{})
}

// storeCompanyIndex stores company indexing information
func (k Keeper) storeCompanyIndex(ctx sdk.Context, companyID uint64, dataType string, dataID string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	
	var storePrefix []byte
	var key []byte
	
	switch dataType {
	case "trade":
		storePrefix = types.TradesByCompanyPrefix
		if tradeID, err := strconv.ParseUint(dataID, 10, 64); err == nil {
			key = types.TradesByCompanyKey(companyID, tradeID)
		}
	case "dividend":
		storePrefix = types.DividendsByCompanyPrefix
		if dividendID, err := strconv.ParseUint(dataID, 10, 64); err == nil {
			key = types.DividendsByCompanyKey(companyID, dividendID)
		}
	}
	
	if key != nil {
		companyStore := prefix.NewStore(store, storePrefix)
		companyStore.Set(key, []byte{})
	}
}

// getCachedData retrieves cached data if available and not expired
func (k Keeper) getCachedData(ctx sdk.Context, cacheType string, cacheID string) ([]byte, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	cacheStore := prefix.NewStore(store, types.CachePrefix)
	
	key := types.CacheKey(cacheType, cacheID)
	data := cacheStore.Get(key)
	
	if data == nil {
		return nil, false
	}
	
	// Check expiration (simplified - would need proper cache metadata)
	return data, true
}

// setCachedData stores data in cache
func (k Keeper) setCachedData(ctx sdk.Context, cacheType string, cacheID string, data []byte) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	cacheStore := prefix.NewStore(store, types.CachePrefix)
	
	key := types.CacheKey(cacheType, cacheID)
	cacheStore.Set(key, data)
}