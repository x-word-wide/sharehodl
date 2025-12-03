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
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	governancetypes "github.com/sharehodl/sharehodl-blockchain/x/governance/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// Indexer handles comprehensive blockchain data indexing for ShareScan explorer
type Indexer struct {
	keeper *Keeper
}

// NewIndexer creates a new indexer instance
func NewIndexer(keeper *Keeper) *Indexer {
	return &Indexer{
		keeper: keeper,
	}
}

// IndexBlock indexes a complete block with all transactions and events
func (idx *Indexer) IndexBlock(ctx sdk.Context, blockHeight int64) error {
	// Get block information
	blockInfo, err := idx.extractBlockInfo(ctx, blockHeight)
	if err != nil {
		return err
	}

	// Index block data
	idx.storeBlockInfo(ctx, blockInfo)

	// Index all transactions in the block
	txs := ctx.TxBytes()
	for i, txBytes := range txs {
		txInfo, err := idx.extractTransactionInfo(ctx, txBytes, blockHeight, uint32(i))
		if err != nil {
			continue // Log error but continue processing
		}
		idx.storeTransactionInfo(ctx, txInfo)
		idx.indexTransactionEvents(ctx, txInfo)
	}

	// Update network statistics
	idx.updateNetworkStats(ctx, blockInfo)

	// Update analytics data
	idx.updateAnalytics(ctx, blockInfo)

	// Trigger real-time notifications
	idx.triggerNotifications(ctx, blockInfo)

	return nil
}

// extractBlockInfo extracts comprehensive block information
func (idx *Indexer) extractBlockInfo(ctx sdk.Context, blockHeight int64) (types.BlockInfo, error) {
	block := ctx.BlockHeader()
	
	// Get validator information
	validators := idx.getBlockValidators(ctx)
	
	// Calculate module activity
	moduleActivity := idx.calculateModuleActivity(ctx, blockHeight)
	
	blockInfo := types.BlockInfo{
		Height:         blockHeight,
		Hash:           block.GetHash().String(),
		PreviousHash:   block.LastBlockId.Hash.String(),
		Timestamp:      block.Time,
		Proposer:       block.ProposerAddress.String(),
		TotalTxs:       uint64(len(ctx.TxBytes())),
		ValidTxs:       0, // Will be calculated
		FailedTxs:      0, // Will be calculated
		BlockSize:      uint64(len(ctx.TxBytes())), // Approximate
		GasUsed:        uint64(ctx.GasMeter().GasConsumed()),
		GasLimit:       uint64(ctx.GasMeter().Limit()),
		Evidence:       []string{}, // Extract if needed
		Validators:     validators,
		ModuleActivity: moduleActivity,
	}
	
	return blockInfo, nil
}

// extractTransactionInfo extracts comprehensive transaction information
func (idx *Indexer) extractTransactionInfo(ctx sdk.Context, txBytes []byte, blockHeight int64, txIndex uint32) (types.TransactionInfo, error) {
	// Parse transaction
	tx, err := idx.keeper.txDecoder(txBytes)
	if err != nil {
		return types.TransactionInfo{}, err
	}

	// Extract transaction hash
	txHash := sdk.Tx(tx).GetTxHash()
	
	// Extract messages
	messages := make([]types.MessageInfo, 0)
	for _, msg := range tx.GetMsgs() {
		msgInfo := idx.extractMessageInfo(ctx, msg, txHash)
		messages = append(messages, msgInfo)
	}

	// Extract events
	events := idx.extractEvents(ctx, txHash, blockHeight)

	// Calculate module impacts
	moduleImpacts := idx.calculateModuleImpacts(messages, events)

	// Get transaction result
	txResult := ctx.TxBytes() // This would need proper transaction result extraction
	
	txInfo := types.TransactionInfo{
		Hash:          txHash,
		BlockHeight:   blockHeight,
		BlockHash:     ctx.BlockHeader().GetHash().String(),
		TxIndex:       txIndex,
		Timestamp:     ctx.BlockTime(),
		Sender:        idx.extractSender(tx),
		Fee:           tx.GetFee(),
		GasWanted:     tx.GetGas(),
		GasUsed:       uint64(ctx.GasMeter().GasConsumed()), // Would need proper calculation
		Success:       true, // Would need proper success determination
		Code:          0,    // Would extract from tx result
		Log:           "",   // Would extract from tx result
		RawLog:        "",   // Would extract from tx result
		Messages:      messages,
		Events:        events,
		ModuleImpacts: moduleImpacts,
	}

	return txInfo, nil
}

// extractMessageInfo extracts message information
func (idx *Indexer) extractMessageInfo(ctx sdk.Context, msg sdk.Msg, txHash string) types.MessageInfo {
	msgType := sdk.MsgTypeURL(msg)
	module := idx.extractModuleFromMsgType(msgType)
	
	return types.MessageInfo{
		Type:    msgType,
		Module:  module,
		Sender:  idx.extractMsgSender(msg),
		Content: msg,
		Success: true, // Would need proper success determination
		GasUsed: 0,    // Would need proper gas calculation
		Events:  []types.EventInfo{}, // Would extract message-specific events
	}
}

// extractEvents extracts events from the context
func (idx *Indexer) extractEvents(ctx sdk.Context, txHash string, blockHeight int64) []types.EventInfo {
	events := make([]types.EventInfo, 0)
	
	for _, event := range ctx.EventManager().Events() {
		eventInfo := types.EventInfo{
			Type:        event.Type,
			Module:      idx.extractModuleFromEventType(event.Type),
			Attributes:  event.Attributes,
			TxHash:      txHash,
			BlockHeight: blockHeight,
			Timestamp:   ctx.BlockTime(),
		}
		events = append(events, eventInfo)
	}
	
	return events
}

// getBlockValidators gets validator information for the block
func (idx *Indexer) getBlockValidators(ctx sdk.Context) []types.ValidatorInfo {
	validators := make([]types.ValidatorInfo, 0)
	
	// This would need integration with the validator keeper
	// For now, return empty slice
	
	return validators
}

// calculateModuleActivity calculates activity by module in the block
func (idx *Indexer) calculateModuleActivity(ctx sdk.Context, blockHeight int64) types.ModuleActivity {
	activity := types.ModuleActivity{
		HODL:       types.ModuleStats{},
		Equity:     types.ModuleStats{},
		DEX:        types.ModuleStats{},
		Governance: types.ModuleStats{},
		Validator:  types.ModuleStats{},
	}
	
	// Extract activity from events and messages
	events := ctx.EventManager().Events()
	
	for _, event := range events {
		module := idx.extractModuleFromEventType(event.Type)
		switch module {
		case "hodl":
			activity.HODL.EventCount++
		case "equity":
			activity.Equity.EventCount++
		case "dex":
			activity.DEX.EventCount++
		case "governance":
			activity.Governance.EventCount++
		case "validator":
			activity.Validator.EventCount++
		}
	}
	
	return activity
}

// calculateModuleImpacts calculates how modules were impacted by the transaction
func (idx *Indexer) calculateModuleImpacts(messages []types.MessageInfo, events []types.EventInfo) []types.ModuleImpact {
	impacts := make([]types.ModuleImpact, 0)
	moduleMap := make(map[string]*types.ModuleImpact)
	
	// Analyze messages
	for _, msg := range messages {
		if impact, exists := moduleMap[msg.Module]; exists {
			impact.Affected = append(impact.Affected, msg.Type)
		} else {
			moduleMap[msg.Module] = &types.ModuleImpact{
				Module:   msg.Module,
				Action:   msg.Type,
				Affected: []string{msg.Type},
				Volume:   math.LegacyZeroDec(),
				Success:  msg.Success,
			}
		}
	}
	
	// Analyze events for volume calculations
	for _, event := range events {
		if impact, exists := moduleMap[event.Module]; exists {
			// Extract volume from event attributes if applicable
			volume := idx.extractVolumeFromEvent(event)
			impact.Volume = impact.Volume.Add(volume)
		}
	}
	
	// Convert map to slice
	for _, impact := range moduleMap {
		impacts = append(impacts, *impact)
	}
	
	return impacts
}

// Storage functions

// storeBlockInfo stores block information
func (idx *Indexer) storeBlockInfo(ctx sdk.Context, blockInfo types.BlockInfo) {
	store := runtime.KVStoreAdapter(idx.keeper.storeService.OpenKVStore(ctx))
	blockStore := prefix.NewStore(store, types.BlockPrefix)
	
	bz, err := json.Marshal(blockInfo)
	if err != nil {
		return
	}
	
	heightKey := sdk.Uint64ToBigEndian(uint64(blockInfo.Height))
	blockStore.Set(heightKey, bz)
	
	// Store block hash index
	hashStore := prefix.NewStore(store, types.BlockHashPrefix)
	hashStore.Set([]byte(blockInfo.Hash), heightKey)
}

// storeTransactionInfo stores transaction information
func (idx *Indexer) storeTransactionInfo(ctx sdk.Context, txInfo types.TransactionInfo) {
	store := runtime.KVStoreAdapter(idx.keeper.storeService.OpenKVStore(ctx))
	txStore := prefix.NewStore(store, types.TransactionPrefix)
	
	bz, err := json.Marshal(txInfo)
	if err != nil {
		return
	}
	
	// Store by hash
	txStore.Set([]byte(txInfo.Hash), bz)
	
	// Store by block height index
	blockTxStore := prefix.NewStore(store, types.BlockTransactionPrefix)
	blockKey := sdk.Uint64ToBigEndian(uint64(txInfo.BlockHeight))
	blockTxStore.Set(append(blockKey, []byte(txInfo.Hash)...), []byte{})
}

// indexTransactionEvents indexes transaction events for search
func (idx *Indexer) indexTransactionEvents(ctx sdk.Context, txInfo types.TransactionInfo) {
	store := runtime.KVStoreAdapter(idx.keeper.storeService.OpenKVStore(ctx))
	eventStore := prefix.NewStore(store, types.EventIndexPrefix)
	
	for _, event := range txInfo.Events {
		// Index by event type
		typeKey := append([]byte(event.Type), []byte(txInfo.Hash)...)
		eventStore.Set(typeKey, []byte{})
		
		// Index by module
		moduleKey := append([]byte(event.Module), []byte(txInfo.Hash)...)
		eventStore.Set(moduleKey, []byte{})
		
		// Index by attributes
		for _, attr := range event.Attributes {
			attrKey := append([]byte(attr.Key), []byte(txInfo.Hash)...)
			eventStore.Set(attrKey, []byte(attr.Value))
		}
	}
}

// updateNetworkStats updates overall network statistics
func (idx *Indexer) updateNetworkStats(ctx sdk.Context, blockInfo types.BlockInfo) {
	store := runtime.KVStoreAdapter(idx.keeper.storeService.OpenKVStore(ctx))
	statsStore := prefix.NewStore(store, types.NetworkStatsPrefix)
	
	// Get existing stats or initialize
	statsBytes := statsStore.Get([]byte("current"))
	var stats types.NetworkStats
	if statsBytes != nil {
		json.Unmarshal(statsBytes, &stats)
	} else {
		stats = types.NetworkStats{
			TotalBlocks:       0,
			TotalTransactions: 0,
			TotalAccounts:     0,
			ModuleActivity:    types.ModuleActivity{},
		}
	}
	
	// Update stats
	stats.LatestBlock = blockInfo
	stats.TotalBlocks++
	stats.TotalTransactions += blockInfo.TotalTxs
	stats.ModuleActivity = idx.aggregateModuleActivity(stats.ModuleActivity, blockInfo.ModuleActivity)
	
	// Calculate TPS (rough estimate)
	if blockInfo.Height > 1 {
		timeDiff := blockInfo.Timestamp.Sub(ctx.BlockTime())
		if timeDiff.Seconds() > 0 {
			stats.TPS = math.LegacyNewDec(int64(blockInfo.TotalTxs)).Quo(math.LegacyNewDecFromBigInt(timeDiff.Nanoseconds()))
		}
	}
	
	// Store updated stats
	updatedBytes, _ := json.Marshal(stats)
	statsStore.Set([]byte("current"), updatedBytes)
}

// updateAnalytics updates time-series analytics data
func (idx *Indexer) updateAnalytics(ctx sdk.Context, blockInfo types.BlockInfo) {
	store := runtime.KVStoreAdapter(idx.keeper.storeService.OpenKVStore(ctx))
	analyticsStore := prefix.NewStore(store, types.AnalyticsPrefix)
	
	// Create analytics data point
	dataPoint := types.AnalyticsDataPoint{
		Timestamp:           blockInfo.Timestamp,
		BlockHeight:         blockInfo.Height,
		Transactions:        blockInfo.TotalTxs,
		UniqueAddresses:     0, // Would need to calculate from transactions
		TradingVolume:       math.LegacyZeroDec(), // Would extract from trading events
		MarketCap:           math.LegacyZeroDec(), // Would calculate from company data
		HODLPrice:           math.LegacyZeroDec(), // Would get from price feeds
		ActiveValidators:    uint64(len(blockInfo.Validators)),
		GovernanceActivity:  0, // Would extract from governance events
		DividendPayments:    math.LegacyZeroDec(), // Would extract from dividend events
		NewCompanies:        0, // Would extract from company events
		GasUsed:             blockInfo.GasUsed,
		AverageGasPrice:     math.LegacyZeroDec(), // Would calculate from transactions
	}
	
	// Store by timestamp for time-series queries
	timeKey := blockInfo.Timestamp.Unix()
	dataBytes, _ := json.Marshal(dataPoint)
	analyticsStore.Set(sdk.Uint64ToBigEndian(uint64(timeKey)), dataBytes)
}

// triggerNotifications sends real-time notifications for important events
func (idx *Indexer) triggerNotifications(ctx sdk.Context, blockInfo types.BlockInfo) {
	// Extract important events and create notifications
	events := ctx.EventManager().Events()
	
	for _, event := range events {
		notification := idx.createNotificationFromEvent(event, blockInfo)
		if notification != nil {
			idx.sendNotification(ctx, *notification)
		}
	}
}

// createNotificationFromEvent creates a notification from an event
func (idx *Indexer) createNotificationFromEvent(event sdk.Event, blockInfo types.BlockInfo) *types.NotificationEvent {
	switch event.Type {
	case "declare_dividend":
		return &types.NotificationEvent{
			ID:        "dividend_" + strconv.FormatInt(blockInfo.Height, 10),
			Type:      "dividend",
			Title:     "New Dividend Declared",
			Message:   "A company has declared a new dividend payment",
			Timestamp: blockInfo.Timestamp,
			Severity:  "info",
		}
	case "submit_proposal":
		return &types.NotificationEvent{
			ID:        "proposal_" + strconv.FormatInt(blockInfo.Height, 10),
			Type:      "proposal",
			Title:     "New Governance Proposal",
			Message:   "A new governance proposal has been submitted",
			Timestamp: blockInfo.Timestamp,
			Severity:  "info",
		}
	case "large_trade":
		return &types.NotificationEvent{
			ID:        "trade_" + strconv.FormatInt(blockInfo.Height, 10),
			Type:      "trade",
			Title:     "Large Trade Executed",
			Message:   "A significant trade has been executed",
			Timestamp: blockInfo.Timestamp,
			Severity:  "warning",
		}
	}
	
	return nil
}

// sendNotification sends a notification (placeholder implementation)
func (idx *Indexer) sendNotification(ctx sdk.Context, notification types.NotificationEvent) {
	store := runtime.KVStoreAdapter(idx.keeper.storeService.OpenKVStore(ctx))
	notificationStore := prefix.NewStore(store, types.NotificationPrefix)
	
	notificationBytes, _ := json.Marshal(notification)
	notificationStore.Set([]byte(notification.ID), notificationBytes)
}

// Helper functions

func (idx *Indexer) extractModuleFromMsgType(msgType string) string {
	parts := strings.Split(msgType, ".")
	if len(parts) >= 3 {
		return parts[len(parts)-2] // Usually the module name is second to last
	}
	return "unknown"
}

func (idx *Indexer) extractModuleFromEventType(eventType string) string {
	// Extract module name from event type
	if strings.Contains(eventType, "_") {
		parts := strings.Split(eventType, "_")
		return parts[0]
	}
	return "unknown"
}

func (idx *Indexer) extractSender(tx sdk.Tx) string {
	msgs := tx.GetMsgs()
	if len(msgs) > 0 {
		signers := msgs[0].GetSigners()
		if len(signers) > 0 {
			return signers[0].String()
		}
	}
	return ""
}

func (idx *Indexer) extractMsgSender(msg sdk.Msg) string {
	signers := msg.GetSigners()
	if len(signers) > 0 {
		return signers[0].String()
	}
	return ""
}

func (idx *Indexer) extractVolumeFromEvent(event types.EventInfo) math.LegacyDec {
	for _, attr := range event.Attributes {
		if attr.Key == "amount" || attr.Key == "volume" {
			if dec, err := math.LegacyNewDecFromStr(attr.Value); err == nil {
				return dec
			}
		}
	}
	return math.LegacyZeroDec()
}

func (idx *Indexer) aggregateModuleActivity(existing, new types.ModuleActivity) types.ModuleActivity {
	return types.ModuleActivity{
		HODL: types.ModuleStats{
			TotalTxs:   existing.HODL.TotalTxs + new.HODL.TotalTxs,
			SuccessTxs: existing.HODL.SuccessTxs + new.HODL.SuccessTxs,
			FailedTxs:  existing.HODL.FailedTxs + new.HODL.FailedTxs,
			GasUsed:    existing.HODL.GasUsed + new.HODL.GasUsed,
			EventCount: existing.HODL.EventCount + new.HODL.EventCount,
			Volume:     existing.HODL.Volume.Add(new.HODL.Volume),
		},
		Equity: types.ModuleStats{
			TotalTxs:   existing.Equity.TotalTxs + new.Equity.TotalTxs,
			SuccessTxs: existing.Equity.SuccessTxs + new.Equity.SuccessTxs,
			FailedTxs:  existing.Equity.FailedTxs + new.Equity.FailedTxs,
			GasUsed:    existing.Equity.GasUsed + new.Equity.GasUsed,
			EventCount: existing.Equity.EventCount + new.Equity.EventCount,
			Volume:     existing.Equity.Volume.Add(new.Equity.Volume),
		},
		DEX: types.ModuleStats{
			TotalTxs:   existing.DEX.TotalTxs + new.DEX.TotalTxs,
			SuccessTxs: existing.DEX.SuccessTxs + new.DEX.SuccessTxs,
			FailedTxs:  existing.DEX.FailedTxs + new.DEX.FailedTxs,
			GasUsed:    existing.DEX.GasUsed + new.DEX.GasUsed,
			EventCount: existing.DEX.EventCount + new.DEX.EventCount,
			Volume:     existing.DEX.Volume.Add(new.DEX.Volume),
		},
		Governance: types.ModuleStats{
			TotalTxs:   existing.Governance.TotalTxs + new.Governance.TotalTxs,
			SuccessTxs: existing.Governance.SuccessTxs + new.Governance.SuccessTxs,
			FailedTxs:  existing.Governance.FailedTxs + new.Governance.FailedTxs,
			GasUsed:    existing.Governance.GasUsed + new.Governance.GasUsed,
			EventCount: existing.Governance.EventCount + new.Governance.EventCount,
			Volume:     existing.Governance.Volume.Add(new.Governance.Volume),
		},
		Validator: types.ModuleStats{
			TotalTxs:   existing.Validator.TotalTxs + new.Validator.TotalTxs,
			SuccessTxs: existing.Validator.SuccessTxs + new.Validator.SuccessTxs,
			FailedTxs:  existing.Validator.FailedTxs + new.Validator.FailedTxs,
			GasUsed:    existing.Validator.GasUsed + new.Validator.GasUsed,
			EventCount: existing.Validator.EventCount + new.Validator.EventCount,
			Volume:     existing.Validator.Volume.Add(new.Validator.Volume),
		},
	}
}