package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ShareScan Block Explorer Data Structures
// Comprehensive blockchain transparency and analytics

// BlockInfo represents detailed block information
type BlockInfo struct {
	Height          int64           `json:"height"`
	Hash            string          `json:"hash"`
	PreviousHash    string          `json:"previous_hash"`
	Timestamp       time.Time       `json:"timestamp"`
	Proposer        string          `json:"proposer"`
	TotalTxs        uint64          `json:"total_txs"`
	ValidTxs        uint64          `json:"valid_txs"`
	FailedTxs       uint64          `json:"failed_txs"`
	BlockSize       uint64          `json:"block_size"`
	GasUsed         uint64          `json:"gas_used"`
	GasLimit        uint64          `json:"gas_limit"`
	Evidence        []string        `json:"evidence"`
	Validators      []ValidatorInfo `json:"validators"`
	ModuleActivity  ModuleActivity  `json:"module_activity"`
}

// TransactionInfo represents detailed transaction information
type TransactionInfo struct {
	Hash            string              `json:"hash"`
	BlockHeight     int64               `json:"block_height"`
	BlockHash       string              `json:"block_hash"`
	TxIndex         uint32              `json:"tx_index"`
	Timestamp       time.Time           `json:"timestamp"`
	Sender          string              `json:"sender"`
	Fee             sdk.Coins           `json:"fee"`
	GasWanted       uint64              `json:"gas_wanted"`
	GasUsed         uint64              `json:"gas_used"`
	Success         bool                `json:"success"`
	Code            uint32              `json:"code"`
	Log             string              `json:"log"`
	RawLog          string              `json:"raw_log"`
	Messages        []MessageInfo       `json:"messages"`
	Events          []EventInfo         `json:"events"`
	ModuleImpacts   []ModuleImpact      `json:"module_impacts"`
}

// MessageInfo represents detailed message information
type MessageInfo struct {
	Type        string      `json:"type"`
	Module      string      `json:"module"`
	Sender      string      `json:"sender"`
	Content     interface{} `json:"content"`
	Success     bool        `json:"success"`
	GasUsed     uint64      `json:"gas_used"`
	Events      []EventInfo `json:"events"`
}

// EventInfo represents blockchain event information
type EventInfo struct {
	Type       string              `json:"type"`
	Module     string              `json:"module"`
	Attributes []sdk.Attribute     `json:"attributes"`
	TxHash     string              `json:"tx_hash"`
	BlockHeight int64              `json:"block_height"`
	Timestamp  time.Time           `json:"timestamp"`
}

// ValidatorInfo represents validator information in blocks
type ValidatorInfo struct {
	Address         string          `json:"address"`
	Moniker         string          `json:"moniker"`
	VotingPower     math.Int        `json:"voting_power"`
	ProposerPriority math.Int       `json:"proposer_priority"`
	Signed          bool            `json:"signed"`
	Tier            string          `json:"tier"`
	BusinessName    string          `json:"business_name"`
}

// CompanyInfo represents company information for explorer
type CompanyInfo struct {
	ID              uint64          `json:"id"`
	Name            string          `json:"name"`
	Symbol          string          `json:"symbol"`
	BusinessType    string          `json:"business_type"`
	Jurisdiction    string          `json:"jurisdiction"`
	ListingDate     time.Time       `json:"listing_date"`
	Status          string          `json:"status"`
	TotalShares     math.Int        `json:"total_shares"`
	ShareClasses    []ShareClassInfo `json:"share_classes"`
	MarketCap       math.LegacyDec  `json:"market_cap"`
	LastPrice       math.LegacyDec  `json:"last_price"`
	Volume24h       math.LegacyDec  `json:"volume_24h"`
	PriceChange24h  math.LegacyDec  `json:"price_change_24h"`
	Shareholders    uint64          `json:"shareholders"`
	RecentDividends []DividendSummary `json:"recent_dividends"`
	RecentTrades    []TradeSummary  `json:"recent_trades"`
}

// ShareClassInfo represents share class details
type ShareClassInfo struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	VotingRights    bool            `json:"voting_rights"`
	DividendRights  bool            `json:"dividend_rights"`
	TotalShares     math.Int        `json:"total_shares"`
	OutstandingShares math.Int      `json:"outstanding_shares"`
	LastPrice       math.LegacyDec  `json:"last_price"`
	MarketCap       math.LegacyDec  `json:"market_cap"`
}

// DividendSummary represents dividend information for explorer
type DividendSummary struct {
	ID              uint64          `json:"id"`
	CompanyID       uint64          `json:"company_id"`
	CompanyName     string          `json:"company_name"`
	Type            string          `json:"type"`
	AmountPerShare  math.LegacyDec  `json:"amount_per_share"`
	TotalAmount     math.LegacyDec  `json:"total_amount"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	DeclarationDate time.Time       `json:"declaration_date"`
	ExDividendDate  time.Time       `json:"ex_dividend_date"`
	RecordDate      time.Time       `json:"record_date"`
	PaymentDate     time.Time       `json:"payment_date"`
	EligibleShares  math.Int        `json:"eligible_shares"`
	SharesPaid      math.Int        `json:"shares_paid"`
	AmountPaid      math.LegacyDec  `json:"amount_paid"`
	ProgressPercent math.LegacyDec  `json:"progress_percent"`
}

// TradeSummary represents trading information for explorer
type TradeSummary struct {
	ID              uint64          `json:"id"`
	CompanyID       uint64          `json:"company_id"`
	CompanySymbol   string          `json:"company_symbol"`
	OrderType       string          `json:"order_type"`
	Side            string          `json:"side"`
	Price           math.LegacyDec  `json:"price"`
	Quantity        math.Int        `json:"quantity"`
	FilledQuantity  math.Int        `json:"filled_quantity"`
	TotalValue      math.LegacyDec  `json:"total_value"`
	Status          string          `json:"status"`
	Trader          string          `json:"trader"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	BlockHeight     int64           `json:"block_height"`
	TxHash          string          `json:"tx_hash"`
}

// GovernanceProposalInfo represents governance proposal information
type GovernanceProposalInfo struct {
	ID               uint64              `json:"id"`
	Type             string              `json:"type"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	Submitter        string              `json:"submitter"`
	Status           string              `json:"status"`
	SubmissionTime   time.Time           `json:"submission_time"`
	DepositEndTime   time.Time           `json:"deposit_end_time"`
	VotingStartTime  time.Time           `json:"voting_start_time"`
	VotingEndTime    time.Time           `json:"voting_end_time"`
	TotalDeposit     math.Int            `json:"total_deposit"`
	Deposits         []DepositInfo       `json:"deposits"`
	Votes            []VoteInfo          `json:"votes"`
	TallyResult      TallyInfo           `json:"tally_result"`
	QuorumRequired   math.LegacyDec      `json:"quorum_required"`
	ThresholdRequired math.LegacyDec     `json:"threshold_required"`
	Outcome          string              `json:"outcome"`
	ExecutionResult  string              `json:"execution_result"`
}

// VoteInfo represents individual vote information
type VoteInfo struct {
	Voter           string              `json:"voter"`
	ProposalID      uint64              `json:"proposal_id"`
	Option          string              `json:"option"`
	Weight          math.LegacyDec      `json:"weight"`
	VotingPower     math.LegacyDec      `json:"voting_power"`
	Reason          string              `json:"reason"`
	SubmittedAt     time.Time           `json:"submitted_at"`
	BlockHeight     int64               `json:"block_height"`
	TxHash          string              `json:"tx_hash"`
	IsWeighted      bool                `json:"is_weighted"`
	WeightedOptions []WeightedOption    `json:"weighted_options,omitempty"`
}

// WeightedOption represents weighted voting options
type WeightedOption struct {
	Option string          `json:"option"`
	Weight math.LegacyDec  `json:"weight"`
}

// DepositInfo represents proposal deposit information
type DepositInfo struct {
	Depositor   string      `json:"depositor"`
	ProposalID  uint64      `json:"proposal_id"`
	Amount      math.Int    `json:"amount"`
	DepositedAt time.Time   `json:"deposited_at"`
	BlockHeight int64       `json:"block_height"`
	TxHash      string      `json:"tx_hash"`
}

// TallyInfo represents proposal tally information
type TallyInfo struct {
	ProposalID           uint64          `json:"proposal_id"`
	YesVotes            math.LegacyDec  `json:"yes_votes"`
	NoVotes             math.LegacyDec  `json:"no_votes"`
	AbstainVotes        math.LegacyDec  `json:"abstain_votes"`
	VetoVotes           math.LegacyDec  `json:"veto_votes"`
	TotalVotes          math.LegacyDec  `json:"total_votes"`
	TotalEligibleVotes  math.LegacyDec  `json:"total_eligible_votes"`
	ParticipationRate   math.LegacyDec  `json:"participation_rate"`
	YesPercentage       math.LegacyDec  `json:"yes_percentage"`
	NoPercentage        math.LegacyDec  `json:"no_percentage"`
	AbstainPercentage   math.LegacyDec  `json:"abstain_percentage"`
	VetoPercentage      math.LegacyDec  `json:"veto_percentage"`
	Outcome             string          `json:"outcome"`
	QuorumMet           bool            `json:"quorum_met"`
	ThresholdMet        bool            `json:"threshold_met"`
}

// AccountInfo represents account information for explorer
type AccountInfo struct {
	Address         string              `json:"address"`
	Balance         sdk.Coins           `json:"balance"`
	HODLBalance     math.Int            `json:"hodl_balance"`
	IsValidator     bool                `json:"is_validator"`
	ValidatorTier   string              `json:"validator_tier"`
	BusinessName    string              `json:"business_name"`
	Shareholdings   []ShareholdingInfo  `json:"shareholdings"`
	RecentTxs       []TransactionSummary `json:"recent_txs"`
	GovernanceVotes []VoteInfo          `json:"governance_votes"`
	DividendHistory []DividendPayment   `json:"dividend_history"`
	TradingHistory  []TradeSummary      `json:"trading_history"`
	TotalDividends  math.LegacyDec      `json:"total_dividends"`
	TotalTrades     uint64              `json:"total_trades"`
	TradingVolume   math.LegacyDec      `json:"trading_volume"`
}

// ShareholdingInfo represents shareholding details
type ShareholdingInfo struct {
	CompanyID       uint64          `json:"company_id"`
	CompanyName     string          `json:"company_name"`
	CompanySymbol   string          `json:"company_symbol"`
	ClassID         string          `json:"class_id"`
	ClassName       string          `json:"class_name"`
	Shares          math.Int        `json:"shares"`
	Percentage      math.LegacyDec  `json:"percentage"`
	MarketValue     math.LegacyDec  `json:"market_value"`
	AcquiredAt      time.Time       `json:"acquired_at"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// DividendPayment represents individual dividend payment
type DividendPayment struct {
	ID              uint64          `json:"id"`
	DividendID      uint64          `json:"dividend_id"`
	Shareholder     string          `json:"shareholder"`
	SharesHeld      math.Int        `json:"shares_held"`
	Amount          math.LegacyDec  `json:"amount"`
	TaxWithheld     math.LegacyDec  `json:"tax_withheld"`
	NetAmount       math.LegacyDec  `json:"net_amount"`
	Currency        string          `json:"currency"`
	PaidAt          time.Time       `json:"paid_at"`
	BlockHeight     int64           `json:"block_height"`
	TxHash          string          `json:"tx_hash"`
	Status          string          `json:"status"`
}

// TransactionSummary represents summarized transaction info
type TransactionSummary struct {
	Hash        string          `json:"hash"`
	Type        string          `json:"type"`
	Amount      math.LegacyDec  `json:"amount"`
	Fee         sdk.Coins       `json:"fee"`
	Success     bool            `json:"success"`
	Timestamp   time.Time       `json:"timestamp"`
	BlockHeight int64           `json:"block_height"`
}

// ModuleActivity represents activity by module in a block
type ModuleActivity struct {
	HODL        ModuleStats `json:"hodl"`
	Equity      ModuleStats `json:"equity"`
	DEX         ModuleStats `json:"dex"`
	Governance  ModuleStats `json:"governance"`
	Validator   ModuleStats `json:"validator"`
}

// ModuleStats represents statistics for a module
type ModuleStats struct {
	TotalTxs    uint64          `json:"total_txs"`
	SuccessTxs  uint64          `json:"success_txs"`
	FailedTxs   uint64          `json:"failed_txs"`
	GasUsed     uint64          `json:"gas_used"`
	EventCount  uint64          `json:"event_count"`
	Volume      math.LegacyDec  `json:"volume"`
}

// ModuleImpact represents how a transaction impacted different modules
type ModuleImpact struct {
	Module      string          `json:"module"`
	Action      string          `json:"action"`
	Affected    []string        `json:"affected"`
	Volume      math.LegacyDec  `json:"volume"`
	Success     bool            `json:"success"`
}

// NetworkStats represents overall network statistics
type NetworkStats struct {
	LatestBlock         BlockInfo           `json:"latest_block"`
	TotalBlocks         uint64              `json:"total_blocks"`
	TotalTransactions   uint64              `json:"total_transactions"`
	TotalAccounts       uint64              `json:"total_accounts"`
	TotalValidators     uint64              `json:"total_validators"`
	ActiveValidators    uint64              `json:"active_validators"`
	TotalCompanies      uint64              `json:"total_companies"`
	ListedCompanies     uint64              `json:"listed_companies"`
	TotalShareClasses   uint64              `json:"total_share_classes"`
	TotalMarketCap      math.LegacyDec      `json:"total_market_cap"`
	Total24hVolume      math.LegacyDec      `json:"total_24h_volume"`
	TotalHODLSupply     math.Int            `json:"total_hodl_supply"`
	HODLStaked          math.Int            `json:"hodl_staked"`
	ActiveProposals     uint64              `json:"active_proposals"`
	TotalProposals      uint64              `json:"total_proposals"`
	PassedProposals     uint64              `json:"passed_proposals"`
	RejectedProposals   uint64              `json:"rejected_proposals"`
	AverageBlockTime    time.Duration       `json:"average_block_time"`
	NetworkHashRate     math.LegacyDec      `json:"network_hash_rate"`
	TPS                 math.LegacyDec      `json:"transactions_per_second"`
	ModuleActivity      ModuleActivity      `json:"module_activity"`
}

// SearchResult represents search results from the explorer
type SearchResult struct {
	Type        string      `json:"type"`        // "block", "transaction", "account", "company", "proposal"
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	URL         string      `json:"url"`
	Timestamp   time.Time   `json:"timestamp"`
	Relevance   float64     `json:"relevance"`
	Data        interface{} `json:"data"`
}

// Analytics represents time-series analytics data
type Analytics struct {
	TimeFrame   string                   `json:"time_frame"`   // "1h", "1d", "1w", "1m", "1y"
	StartTime   time.Time                `json:"start_time"`
	EndTime     time.Time                `json:"end_time"`
	DataPoints  []AnalyticsDataPoint     `json:"data_points"`
}

// AnalyticsDataPoint represents a single analytics data point
type AnalyticsDataPoint struct {
	Timestamp           time.Time       `json:"timestamp"`
	BlockHeight         int64           `json:"block_height"`
	Transactions        uint64          `json:"transactions"`
	UniqueAddresses     uint64          `json:"unique_addresses"`
	TradingVolume       math.LegacyDec  `json:"trading_volume"`
	MarketCap           math.LegacyDec  `json:"market_cap"`
	HODLPrice           math.LegacyDec  `json:"hodl_price"`
	ActiveValidators    uint64          `json:"active_validators"`
	GovernanceActivity  uint64          `json:"governance_activity"`
	DividendPayments    math.LegacyDec  `json:"dividend_payments"`
	NewCompanies        uint64          `json:"new_companies"`
	GasUsed             uint64          `json:"gas_used"`
	AverageGasPrice     math.LegacyDec  `json:"average_gas_price"`
}

// ChartData represents chart data for visualizations
type ChartData struct {
	Labels   []string        `json:"labels"`
	Datasets []Dataset       `json:"datasets"`
}

// Dataset represents a chart dataset
type Dataset struct {
	Label           string          `json:"label"`
	Data            []float64       `json:"data"`
	BackgroundColor string          `json:"background_color"`
	BorderColor     string          `json:"border_color"`
	BorderWidth     int             `json:"border_width"`
}

// LiveMetrics represents real-time metrics for the explorer
type LiveMetrics struct {
	CurrentBlock        int64           `json:"current_block"`
	BlockTime           time.Duration   `json:"block_time"`
	TPS                 math.LegacyDec  `json:"tps"`
	PendingTxs          uint64          `json:"pending_txs"`
	MemPoolSize         uint64          `json:"mempool_size"`
	ActiveConnections   uint64          `json:"active_connections"`
	NodeVersion         string          `json:"node_version"`
	ChainID             string          `json:"chain_id"`
	LastUpdate          time.Time       `json:"last_update"`
}

// NotificationEvent represents events for real-time notifications
type NotificationEvent struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`          // "block", "transaction", "trade", "dividend", "proposal"
	Title       string          `json:"title"`
	Message     string          `json:"message"`
	Data        interface{}     `json:"data"`
	Timestamp   time.Time       `json:"timestamp"`
	Severity    string          `json:"severity"`      // "info", "warning", "error", "success"
	Recipients  []string        `json:"recipients"`    // Addresses to notify
}

// ExplorerConfig represents configuration for the explorer
type ExplorerConfig struct {
	IndexingEnabled     bool            `json:"indexing_enabled"`
	RealTimeUpdates     bool            `json:"real_time_updates"`
	AnalyticsEnabled    bool            `json:"analytics_enabled"`
	NotificationsEnabled bool           `json:"notifications_enabled"`
	MaxSearchResults    uint64          `json:"max_search_results"`
	CacheTimeout        time.Duration   `json:"cache_timeout"`
	UpdateInterval      time.Duration   `json:"update_interval"`
	HistoryDepth        uint64          `json:"history_depth"`
	EnabledModules      []string        `json:"enabled_modules"`
}