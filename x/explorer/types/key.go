package types

import (
	"encoding/binary"
)

// Module constants
const (
	// ModuleName defines the module name
	ModuleName = "explorer"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_explorer"
)

// Storage key prefixes for ShareScan explorer data
var (
	// Block data
	BlockPrefix              = []byte{0x01}
	BlockHashPrefix          = []byte{0x02}
	BlockTransactionPrefix   = []byte{0x03}
	
	// Transaction data
	TransactionPrefix        = []byte{0x10}
	TransactionSenderPrefix  = []byte{0x11}
	TransactionTypePrefix    = []byte{0x12}
	
	// Account data
	AccountPrefix           = []byte{0x20}
	AccountTransactionPrefix = []byte{0x21}
	AccountBalancePrefix    = []byte{0x22}
	
	// Company data
	CompanyPrefix           = []byte{0x30}
	CompanySymbolPrefix     = []byte{0x31}
	CompanyShareholdersPrefix = []byte{0x32}
	
	// Trading data
	TradePrefix             = []byte{0x40}
	TradesByCompanyPrefix   = []byte{0x41}
	TradesByTraderPrefix    = []byte{0x42}
	TradingVolumePrefix     = []byte{0x43}
	
	// Dividend data
	DividendPrefix          = []byte{0x50}
	DividendPaymentPrefix   = []byte{0x51}
	DividendsByCompanyPrefix = []byte{0x52}
	DividendsByShareholderPrefix = []byte{0x53}
	
	// Governance data
	ProposalIndexPrefix     = []byte{0x60}
	VoteIndexPrefix         = []byte{0x61}
	DepositIndexPrefix      = []byte{0x62}
	GovernanceStatsPrefix   = []byte{0x63}
	
	// Validator data
	ValidatorIndexPrefix    = []byte{0x70}
	ValidatorStatsPrefix    = []byte{0x71}
	ValidatorHistoryPrefix  = []byte{0x72}
	
	// Analytics and statistics
	NetworkStatsPrefix      = []byte{0x80}
	AnalyticsPrefix         = []byte{0x81}
	MetricsPrefix          = []byte{0x82}
	ChartDataPrefix        = []byte{0x83}
	
	// Search and indexing
	SearchIndexPrefix       = []byte{0x90}
	EventIndexPrefix        = []byte{0x91}
	AttributeIndexPrefix    = []byte{0x92}
	
	// Real-time data
	LiveMetricsPrefix       = []byte{0xA0}
	NotificationPrefix      = []byte{0xA1}
	SubscriptionPrefix      = []byte{0xA2}
	
	// Cache and temporary data
	CachePrefix            = []byte{0xB0}
	TempDataPrefix         = []byte{0xB1}
	
	// Configuration
	ConfigPrefix           = []byte{0xC0}
)

// Key construction functions for blocks

// BlockKey returns the store key for a block by height
func BlockKey(height uint64) []byte {
	key := make([]byte, len(BlockPrefix)+8)
	copy(key, BlockPrefix)
	binary.BigEndian.PutUint64(key[len(BlockPrefix):], height)
	return key
}

// BlockHashKey returns the store key for block hash indexing
func BlockHashKey(hash string) []byte {
	return append(BlockHashPrefix, []byte(hash)...)
}

// BlockTransactionKey returns the store key for transactions in a block
func BlockTransactionKey(height uint64, txHash string) []byte {
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, height)
	return append(append(BlockTransactionPrefix, heightBytes...), []byte(txHash)...)
}

// Key construction functions for transactions

// TransactionKey returns the store key for a transaction by hash
func TransactionKey(hash string) []byte {
	return append(TransactionPrefix, []byte(hash)...)
}

// TransactionSenderKey returns the store key for transactions by sender
func TransactionSenderKey(sender string, txHash string) []byte {
	senderBytes := []byte(sender)
	return append(append(TransactionSenderPrefix, senderBytes...), []byte(txHash)...)
}

// TransactionTypeKey returns the store key for transactions by type
func TransactionTypeKey(txType string, txHash string) []byte {
	typeBytes := []byte(txType)
	return append(append(TransactionTypePrefix, typeBytes...), []byte(txHash)...)
}

// Key construction functions for accounts

// AccountKey returns the store key for an account
func AccountKey(address string) []byte {
	return append(AccountPrefix, []byte(address)...)
}

// AccountTransactionKey returns the store key for account transactions
func AccountTransactionKey(address string, txHash string) []byte {
	addrBytes := []byte(address)
	return append(append(AccountTransactionPrefix, addrBytes...), []byte(txHash)...)
}

// AccountBalanceKey returns the store key for account balance history
func AccountBalanceKey(address string, timestamp uint64) []byte {
	addrBytes := []byte(address)
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, timestamp)
	return append(append(AccountBalancePrefix, addrBytes...), timestampBytes...)
}

// Key construction functions for companies

// CompanyKey returns the store key for a company
func CompanyKey(companyID uint64) []byte {
	key := make([]byte, len(CompanyPrefix)+8)
	copy(key, CompanyPrefix)
	binary.BigEndian.PutUint64(key[len(CompanyPrefix):], companyID)
	return key
}

// CompanySymbolKey returns the store key for company by symbol
func CompanySymbolKey(symbol string) []byte {
	return append(CompanySymbolPrefix, []byte(symbol)...)
}

// CompanyShareholdersKey returns the store key for company shareholders
func CompanyShareholdersKey(companyID uint64, shareholder string) []byte {
	companyBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(companyBytes, companyID)
	return append(append(CompanyShareholdersPrefix, companyBytes...), []byte(shareholder)...)
}

// Key construction functions for trading

// TradeKey returns the store key for a trade
func TradeKey(tradeID uint64) []byte {
	key := make([]byte, len(TradePrefix)+8)
	copy(key, TradePrefix)
	binary.BigEndian.PutUint64(key[len(TradePrefix):], tradeID)
	return key
}

// TradesByCompanyKey returns the store key for trades by company
func TradesByCompanyKey(companyID uint64, tradeID uint64) []byte {
	companyBytes := make([]byte, 8)
	tradeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(companyBytes, companyID)
	binary.BigEndian.PutUint64(tradeBytes, tradeID)
	return append(append(TradesByCompanyPrefix, companyBytes...), tradeBytes...)
}

// TradesByTraderKey returns the store key for trades by trader
func TradesByTraderKey(trader string, tradeID uint64) []byte {
	traderBytes := []byte(trader)
	tradeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tradeBytes, tradeID)
	return append(append(TradesByTraderPrefix, traderBytes...), tradeBytes...)
}

// TradingVolumeKey returns the store key for trading volume by time period
func TradingVolumeKey(companyID uint64, period string, timestamp uint64) []byte {
	companyBytes := make([]byte, 8)
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(companyBytes, companyID)
	binary.BigEndian.PutUint64(timestampBytes, timestamp)
	
	return append(append(append(TradingVolumePrefix, companyBytes...), []byte(period)...), timestampBytes...)
}

// Key construction functions for dividends

// DividendKey returns the store key for a dividend
func DividendKey(dividendID uint64) []byte {
	key := make([]byte, len(DividendPrefix)+8)
	copy(key, DividendPrefix)
	binary.BigEndian.PutUint64(key[len(DividendPrefix):], dividendID)
	return key
}

// DividendPaymentKey returns the store key for dividend payments
func DividendPaymentKey(dividendID uint64, shareholder string) []byte {
	dividendBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(dividendBytes, dividendID)
	return append(append(DividendPaymentPrefix, dividendBytes...), []byte(shareholder)...)
}

// DividendsByCompanyKey returns the store key for dividends by company
func DividendsByCompanyKey(companyID uint64, dividendID uint64) []byte {
	companyBytes := make([]byte, 8)
	dividendBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(companyBytes, companyID)
	binary.BigEndian.PutUint64(dividendBytes, dividendID)
	return append(append(DividendsByCompanyPrefix, companyBytes...), dividendBytes...)
}

// DividendsByShareholderKey returns the store key for dividends by shareholder
func DividendsByShareholderKey(shareholder string, dividendID uint64) []byte {
	shareholderBytes := []byte(shareholder)
	dividendBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(dividendBytes, dividendID)
	return append(append(DividendsByShareholderPrefix, shareholderBytes...), dividendBytes...)
}

// Key construction functions for governance

// ProposalIndexKey returns the store key for proposal indexing
func ProposalIndexKey(proposalID uint64) []byte {
	key := make([]byte, len(ProposalIndexPrefix)+8)
	copy(key, ProposalIndexPrefix)
	binary.BigEndian.PutUint64(key[len(ProposalIndexPrefix):], proposalID)
	return key
}

// VoteIndexKey returns the store key for vote indexing
func VoteIndexKey(proposalID uint64, voter string) []byte {
	proposalBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(proposalBytes, proposalID)
	return append(append(VoteIndexPrefix, proposalBytes...), []byte(voter)...)
}

// DepositIndexKey returns the store key for deposit indexing
func DepositIndexKey(proposalID uint64, depositor string) []byte {
	proposalBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(proposalBytes, proposalID)
	return append(append(DepositIndexPrefix, proposalBytes...), []byte(depositor)...)
}

// GovernanceStatsKey returns the store key for governance statistics
func GovernanceStatsKey(period string) []byte {
	return append(GovernanceStatsPrefix, []byte(period)...)
}

// Key construction functions for validators

// ValidatorIndexKey returns the store key for validator indexing
func ValidatorIndexKey(validatorAddr string) []byte {
	return append(ValidatorIndexPrefix, []byte(validatorAddr)...)
}

// ValidatorStatsKey returns the store key for validator statistics
func ValidatorStatsKey(validatorAddr string, period string) []byte {
	validatorBytes := []byte(validatorAddr)
	return append(append(ValidatorStatsPrefix, validatorBytes...), []byte(period)...)
}

// ValidatorHistoryKey returns the store key for validator history
func ValidatorHistoryKey(validatorAddr string, timestamp uint64) []byte {
	validatorBytes := []byte(validatorAddr)
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, timestamp)
	return append(append(ValidatorHistoryPrefix, validatorBytes...), timestampBytes...)
}

// Key construction functions for analytics

// NetworkStatsKey returns the store key for network statistics
func NetworkStatsKey(period string) []byte {
	return append(NetworkStatsPrefix, []byte(period)...)
}

// AnalyticsKey returns the store key for analytics data
func AnalyticsKey(timestamp uint64, dataType string) []byte {
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, timestamp)
	return append(append(AnalyticsPrefix, timestampBytes...), []byte(dataType)...)
}

// MetricsKey returns the store key for metrics
func MetricsKey(metricType string, timestamp uint64) []byte {
	typeBytes := []byte(metricType)
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, timestamp)
	return append(append(MetricsPrefix, typeBytes...), timestampBytes...)
}

// ChartDataKey returns the store key for chart data
func ChartDataKey(chartType string, period string) []byte {
	typeBytes := []byte(chartType)
	return append(append(ChartDataPrefix, typeBytes...), []byte(period)...)
}

// Key construction functions for search and indexing

// SearchIndexKey returns the store key for search indexing
func SearchIndexKey(searchTerm string, resultType string, resultID string) []byte {
	termBytes := []byte(searchTerm)
	typeBytes := []byte(resultType)
	return append(append(append(SearchIndexPrefix, termBytes...), typeBytes...), []byte(resultID)...)
}

// EventIndexKey returns the store key for event indexing
func EventIndexKey(eventType string, txHash string) []byte {
	typeBytes := []byte(eventType)
	return append(append(EventIndexPrefix, typeBytes...), []byte(txHash)...)
}

// AttributeIndexKey returns the store key for attribute indexing
func AttributeIndexKey(attrKey string, attrValue string, txHash string) []byte {
	keyBytes := []byte(attrKey)
	valueBytes := []byte(attrValue)
	return append(append(append(AttributeIndexPrefix, keyBytes...), valueBytes...), []byte(txHash)...)
}

// Key construction functions for real-time data

// LiveMetricsKey returns the store key for live metrics
func LiveMetricsKey() []byte {
	return append(LiveMetricsPrefix, []byte("current")...)
}

// NotificationKey returns the store key for notifications
func NotificationKey(notificationID string) []byte {
	return append(NotificationPrefix, []byte(notificationID)...)
}

// SubscriptionKey returns the store key for subscriptions
func SubscriptionKey(subscriber string, subscriptionType string) []byte {
	subscriberBytes := []byte(subscriber)
	return append(append(SubscriptionPrefix, subscriberBytes...), []byte(subscriptionType)...)
}

// Key construction functions for cache

// CacheKey returns the store key for cached data
func CacheKey(cacheType string, cacheID string) []byte {
	typeBytes := []byte(cacheType)
	return append(append(CachePrefix, typeBytes...), []byte(cacheID)...)
}

// TempDataKey returns the store key for temporary data
func TempDataKey(dataType string, dataID string) []byte {
	typeBytes := []byte(dataType)
	return append(append(TempDataPrefix, typeBytes...), []byte(dataID)...)
}

// Key construction functions for configuration

// ConfigKey returns the store key for configuration
func ConfigKey(configType string) []byte {
	return append(ConfigPrefix, []byte(configType)...)
}

// Iterator key helpers

// BlockIteratorKey returns the prefix for iterating blocks
func BlockIteratorKey() []byte {
	return BlockPrefix
}

// TransactionIteratorKey returns the prefix for iterating transactions
func TransactionIteratorKey() []byte {
	return TransactionPrefix
}

// AccountIteratorKey returns the prefix for iterating accounts
func AccountIteratorKey() []byte {
	return AccountPrefix
}

// CompanyIteratorKey returns the prefix for iterating companies
func CompanyIteratorKey() []byte {
	return CompanyPrefix
}

// TradeIteratorKey returns the prefix for iterating trades
func TradeIteratorKey() []byte {
	return TradePrefix
}

// DividendIteratorKey returns the prefix for iterating dividends
func DividendIteratorKey() []byte {
	return DividendPrefix
}

// ProposalIteratorKey returns the prefix for iterating proposals
func ProposalIteratorKey() []byte {
	return ProposalIndexPrefix
}

// ValidatorIteratorKey returns the prefix for iterating validators
func ValidatorIteratorKey() []byte {
	return ValidatorIndexPrefix
}

// AnalyticsIteratorKey returns the prefix for iterating analytics data
func AnalyticsIteratorKey() []byte {
	return AnalyticsPrefix
}

// Key extraction helpers

// ExtractHeightFromBlockKey extracts block height from block key
func ExtractHeightFromBlockKey(key []byte) uint64 {
	if len(key) < len(BlockPrefix)+8 {
		return 0
	}
	return binary.BigEndian.Uint64(key[len(BlockPrefix):])
}

// ExtractCompanyIDFromKey extracts company ID from various keys
func ExtractCompanyIDFromKey(key []byte, prefixLen int) uint64 {
	if len(key) < prefixLen+8 {
		return 0
	}
	return binary.BigEndian.Uint64(key[prefixLen:prefixLen+8])
}

// ExtractTimestampFromKey extracts timestamp from analytics keys
func ExtractTimestampFromKey(key []byte, prefixLen int) uint64 {
	if len(key) < prefixLen+8 {
		return 0
	}
	return binary.BigEndian.Uint64(key[prefixLen:prefixLen+8])
}

// SplitCompoundKey splits compound keys into components
func SplitCompoundKey(key []byte, prefixLen int, componentSizes []int) [][]byte {
	components := make([][]byte, 0, len(componentSizes))
	offset := prefixLen
	
	for _, size := range componentSizes {
		if offset+size > len(key) {
			break
		}
		components = append(components, key[offset:offset+size])
		offset += size
	}
	
	return components
}