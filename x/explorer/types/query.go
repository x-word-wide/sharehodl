package types

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Query service interface for ShareScan explorer
type QueryServer interface {
	// Block queries
	Block(context.Context, *QueryBlockRequest) (*QueryBlockResponse, error)
	Blocks(context.Context, *QueryBlocksRequest) (*QueryBlocksResponse, error)
	
	// Transaction queries
	Transaction(context.Context, *QueryTransactionRequest) (*QueryTransactionResponse, error)
	BlockTransactions(context.Context, *QueryBlockTransactionsRequest) (*QueryBlockTransactionsResponse, error)
	TransactionsByAccount(context.Context, *QueryTransactionsByAccountRequest) (*QueryTransactionsByAccountResponse, error)
	
	// Account queries
	Account(context.Context, *QueryAccountRequest) (*QueryAccountResponse, error)
	
	// Company queries
	Company(context.Context, *QueryCompanyRequest) (*QueryCompanyResponse, error)
	Companies(context.Context, *QueryCompaniesRequest) (*QueryCompaniesResponse, error)
	TradesByCompany(context.Context, *QueryTradesByCompanyRequest) (*QueryTradesByCompanyResponse, error)
	DividendsByCompany(context.Context, *QueryDividendsByCompanyRequest) (*QueryDividendsByCompanyResponse, error)
	
	// Governance queries
	Proposal(context.Context, *QueryProposalRequest) (*QueryProposalResponse, error)
	VotesByProposal(context.Context, *QueryVotesByProposalRequest) (*QueryVotesByProposalResponse, error)
	
	// Analytics queries
	NetworkStats(context.Context, *QueryNetworkStatsRequest) (*QueryNetworkStatsResponse, error)
	Analytics(context.Context, *QueryAnalyticsRequest) (*QueryAnalyticsResponse, error)
	ChartData(context.Context, *QueryChartDataRequest) (*QueryChartDataResponse, error)
	
	// Search queries
	Search(context.Context, *QuerySearchRequest) (*QuerySearchResponse, error)
	
	// Real-time queries
	LiveMetrics(context.Context, *QueryLiveMetricsRequest) (*QueryLiveMetricsResponse, error)
}

// Block query messages

type QueryBlockRequest struct {
	Height int64 `json:"height"`
}

type QueryBlockResponse struct {
	Block *BlockInfo `json:"block"`
}

type QueryBlocksRequest struct {
	Limit      uint32              `json:"limit"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryBlocksResponse struct {
	Blocks     []BlockInfo          `json:"blocks"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// Transaction query messages

type QueryTransactionRequest struct {
	Hash string `json:"hash"`
}

type QueryTransactionResponse struct {
	Transaction *TransactionInfo `json:"transaction"`
}

type QueryBlockTransactionsRequest struct {
	Height     int64               `json:"height"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryBlockTransactionsResponse struct {
	Transactions []TransactionInfo    `json:"transactions"`
	Pagination   *query.PageResponse `json:"pagination,omitempty"`
}

type QueryTransactionsByAccountRequest struct {
	Address    string              `json:"address"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryTransactionsByAccountResponse struct {
	Transactions []*TransactionInfo   `json:"transactions"`
	Pagination   *query.PageResponse `json:"pagination,omitempty"`
}

// Account query messages

type QueryAccountRequest struct {
	Address string `json:"address"`
}

type QueryAccountResponse struct {
	Account *AccountInfo `json:"account"`
}

// Company query messages

type QueryCompanyRequest struct {
	CompanyId uint64 `json:"company_id"`
}

type QueryCompanyResponse struct {
	Company *CompanyInfo `json:"company"`
}

type QueryCompaniesRequest struct {
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryCompaniesResponse struct {
	Companies  []CompanyInfo       `json:"companies"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

type QueryTradesByCompanyRequest struct {
	CompanyId  uint64              `json:"company_id"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryTradesByCompanyResponse struct {
	Trades     []TradeSummary      `json:"trades"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

type QueryDividendsByCompanyRequest struct {
	CompanyId  uint64              `json:"company_id"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryDividendsByCompanyResponse struct {
	Dividends  []DividendSummary   `json:"dividends"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// Governance query messages

type QueryProposalRequest struct {
	ProposalId uint64 `json:"proposal_id"`
}

type QueryProposalResponse struct {
	Proposal *GovernanceProposalInfo `json:"proposal"`
}

type QueryVotesByProposalRequest struct {
	ProposalId uint64              `json:"proposal_id"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryVotesByProposalResponse struct {
	Votes      []VoteInfo          `json:"votes"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// Analytics query messages

type QueryNetworkStatsRequest struct{}

type QueryNetworkStatsResponse struct {
	Stats *NetworkStats `json:"stats"`
}

type QueryAnalyticsRequest struct {
	TimeFrame string                 `json:"time_frame"`
	StartTime *timestamppb.Timestamp `json:"start_time"`
	EndTime   *timestamppb.Timestamp `json:"end_time"`
}

type QueryAnalyticsResponse struct {
	Analytics *Analytics `json:"analytics"`
}

type QueryChartDataRequest struct {
	ChartType string `json:"chart_type"`
	TimeFrame string `json:"time_frame"`
	CompanyId uint64 `json:"company_id,omitempty"`
}

type QueryChartDataResponse struct {
	ChartData *ChartData `json:"chart_data"`
}

// Search query messages

type QuerySearchRequest struct {
	Query string `json:"query"`
	Limit uint32 `json:"limit"`
}

type QuerySearchResponse struct {
	Results []SearchResult `json:"results"`
}

// Real-time query messages

type QueryLiveMetricsRequest struct{}

type QueryLiveMetricsResponse struct {
	Metrics *LiveMetrics `json:"metrics"`
}

// Advanced query messages with filters

type QueryAdvancedTransactionsRequest struct {
	Filters    TransactionFilters  `json:"filters"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryAdvancedTransactionsResponse struct {
	Transactions []TransactionInfo    `json:"transactions"`
	Pagination   *query.PageResponse `json:"pagination,omitempty"`
}

type TransactionFilters struct {
	FromAddress   string    `json:"from_address,omitempty"`
	ToAddress     string    `json:"to_address,omitempty"`
	MessageType   string    `json:"message_type,omitempty"`
	MinAmount     string    `json:"min_amount,omitempty"`
	MaxAmount     string    `json:"max_amount,omitempty"`
	StartTime     time.Time `json:"start_time,omitempty"`
	EndTime       time.Time `json:"end_time,omitempty"`
	Success       *bool     `json:"success,omitempty"`
}

type QueryAdvancedCompaniesRequest struct {
	Filters    CompanyFilters      `json:"filters"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryAdvancedCompaniesResponse struct {
	Companies  []CompanyInfo       `json:"companies"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

type CompanyFilters struct {
	BusinessType    string `json:"business_type,omitempty"`
	Jurisdiction    string `json:"jurisdiction,omitempty"`
	Status          string `json:"status,omitempty"`
	MinMarketCap    string `json:"min_market_cap,omitempty"`
	MaxMarketCap    string `json:"max_market_cap,omitempty"`
	MinVolume24h    string `json:"min_volume_24h,omitempty"`
	MaxVolume24h    string `json:"max_volume_24h,omitempty"`
}

type QueryAdvancedTradesRequest struct {
	Filters    TradeFilters        `json:"filters"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryAdvancedTradesResponse struct {
	Trades     []TradeSummary      `json:"trades"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

type TradeFilters struct {
	CompanyId     *uint64   `json:"company_id,omitempty"`
	Trader        string    `json:"trader,omitempty"`
	Side          string    `json:"side,omitempty"`          // "buy" or "sell"
	OrderType     string    `json:"order_type,omitempty"`    // "market", "limit", etc.
	MinPrice      string    `json:"min_price,omitempty"`
	MaxPrice      string    `json:"max_price,omitempty"`
	MinQuantity   string    `json:"min_quantity,omitempty"`
	MaxQuantity   string    `json:"max_quantity,omitempty"`
	MinValue      string    `json:"min_value,omitempty"`
	MaxValue      string    `json:"max_value,omitempty"`
	StartTime     time.Time `json:"start_time,omitempty"`
	EndTime       time.Time `json:"end_time,omitempty"`
	Status        string    `json:"status,omitempty"`
}

// Aggregated data queries

type QueryMarketSummaryRequest struct {
	CompanyId *uint64 `json:"company_id,omitempty"` // If nil, returns overall market summary
}

type QueryMarketSummaryResponse struct {
	Summary *MarketSummary `json:"summary"`
}

type MarketSummary struct {
	TotalMarketCap      string `json:"total_market_cap"`
	TotalVolume24h      string `json:"total_volume_24h"`
	TotalCompanies      uint64 `json:"total_companies"`
	ActiveCompanies     uint64 `json:"active_companies"`
	TotalTrades24h      uint64 `json:"total_trades_24h"`
	AveragePriceChange  string `json:"average_price_change"`
	TopGainers          []CompanyPerformance `json:"top_gainers"`
	TopLosers           []CompanyPerformance `json:"top_losers"`
	MostActive          []CompanyActivity    `json:"most_active"`
}

type CompanyPerformance struct {
	CompanyId     uint64 `json:"company_id"`
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Price         string `json:"price"`
	PriceChange   string `json:"price_change"`
	PercentChange string `json:"percent_change"`
	Volume24h     string `json:"volume_24h"`
}

type CompanyActivity struct {
	CompanyId   uint64 `json:"company_id"`
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Volume24h   string `json:"volume_24h"`
	Trades24h   uint64 `json:"trades_24h"`
	Price       string `json:"price"`
}

type QueryGovernanceSummaryRequest struct{}

type QueryGovernanceSummaryResponse struct {
	Summary *GovernanceSummary `json:"summary"`
}

type GovernanceSummary struct {
	TotalProposals      uint64                  `json:"total_proposals"`
	ActiveProposals     uint64                  `json:"active_proposals"`
	PassedProposals     uint64                  `json:"passed_proposals"`
	RejectedProposals   uint64                  `json:"rejected_proposals"`
	TotalVotes          uint64                  `json:"total_votes"`
	TotalDeposits       string                  `json:"total_deposits"`
	AverageParticipation string                 `json:"average_participation"`
	RecentProposals     []GovernanceProposalInfo `json:"recent_proposals"`
}

type QueryValidatorSummaryRequest struct{}

type QueryValidatorSummaryResponse struct {
	Summary *ValidatorSummary `json:"summary"`
}

type ValidatorSummary struct {
	TotalValidators     uint64           `json:"total_validators"`
	ActiveValidators    uint64           `json:"active_validators"`
	TierDistribution    map[string]uint64 `json:"tier_distribution"`
	TotalStake          string           `json:"total_stake"`
	AverageCommission   string           `json:"average_commission"`
	TopValidators       []ValidatorInfo  `json:"top_validators"`
}

// Historical data queries

type QueryHistoricalDataRequest struct {
	DataType    string                 `json:"data_type"`    // "prices", "volumes", "network_activity", etc.
	CompanyId   *uint64                `json:"company_id,omitempty"`
	Granularity string                 `json:"granularity"`  // "1m", "5m", "1h", "1d", etc.
	StartTime   *timestamppb.Timestamp `json:"start_time"`
	EndTime     *timestamppb.Timestamp `json:"end_time"`
}

type QueryHistoricalDataResponse struct {
	Data []HistoricalDataPoint `json:"data"`
}

type HistoricalDataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Values    map[string]interface{} `json:"values"`
}

// Notification and subscription queries

type QueryNotificationsRequest struct {
	Address    string              `json:"address"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type QueryNotificationsResponse struct {
	Notifications []NotificationEvent `json:"notifications"`
	Pagination    *query.PageResponse `json:"pagination,omitempty"`
}

type QuerySubscriptionsRequest struct {
	Address string `json:"address"`
}

type QuerySubscriptionsResponse struct {
	Subscriptions []string `json:"subscriptions"`
}

// Export and reporting queries

type QueryExportDataRequest struct {
	DataType   string                 `json:"data_type"`   // "transactions", "trades", "dividends", etc.
	Format     string                 `json:"format"`      // "csv", "json", "xlsx"
	Filters    map[string]interface{} `json:"filters"`
	StartTime  *timestamppb.Timestamp `json:"start_time"`
	EndTime    *timestamppb.Timestamp `json:"end_time"`
}

type QueryExportDataResponse struct {
	ExportId    string `json:"export_id"`
	DownloadUrl string `json:"download_url"`
	Status      string `json:"status"`
	FileSize    uint64 `json:"file_size"`
}

type QueryReportRequest struct {
	ReportType string                 `json:"report_type"` // "financial", "governance", "network", etc.
	Period     string                 `json:"period"`      // "daily", "weekly", "monthly", etc.
	StartTime  *timestamppb.Timestamp `json:"start_time"`
	EndTime    *timestamppb.Timestamp `json:"end_time"`
	Params     map[string]interface{} `json:"params"`
}

type QueryReportResponse struct {
	Report *Report `json:"report"`
}

type Report struct {
	Id          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Period      string                 `json:"period"`
	GeneratedAt time.Time              `json:"generated_at"`
	Summary     map[string]interface{} `json:"summary"`
	Sections    []ReportSection        `json:"sections"`
}

type ReportSection struct {
	Title   string                 `json:"title"`
	Type    string                 `json:"type"`    // "text", "chart", "table", etc.
	Content map[string]interface{} `json:"content"`
}

// Health and status queries

type QueryHealthRequest struct{}

type QueryHealthResponse struct {
	Status        string            `json:"status"`
	Version       string            `json:"version"`
	Uptime        string            `json:"uptime"`
	LastIndexed   time.Time         `json:"last_indexed"`
	IndexingLag   string            `json:"indexing_lag"`
	DatabaseStats DatabaseStats     `json:"database_stats"`
	ModuleStatus  map[string]string `json:"module_status"`
}

type DatabaseStats struct {
	TotalRecords    uint64 `json:"total_records"`
	IndexedBlocks   uint64 `json:"indexed_blocks"`
	IndexedTxs      uint64 `json:"indexed_txs"`
	DatabaseSize    uint64 `json:"database_size"`
	LastBackup      time.Time `json:"last_backup"`
	BackupSize      uint64 `json:"backup_size"`
}

type QueryConfigRequest struct{}

type QueryConfigResponse struct {
	Config *ExplorerConfig `json:"config"`
}

// Utility functions for creating query requests

func NewQueryBlockRequest(height int64) *QueryBlockRequest {
	return &QueryBlockRequest{Height: height}
}

func NewQueryTransactionRequest(hash string) *QueryTransactionRequest {
	return &QueryTransactionRequest{Hash: hash}
}

func NewQueryAccountRequest(address string) *QueryAccountRequest {
	return &QueryAccountRequest{Address: address}
}

func NewQueryCompanyRequest(companyID uint64) *QueryCompanyRequest {
	return &QueryCompanyRequest{CompanyId: companyID}
}

func NewQuerySearchRequest(query string, limit uint32) *QuerySearchRequest {
	return &QuerySearchRequest{Query: query, Limit: limit}
}

func NewQueryAnalyticsRequest(timeFrame string, startTime, endTime time.Time) *QueryAnalyticsRequest {
	return &QueryAnalyticsRequest{
		TimeFrame: timeFrame,
		StartTime: timestamppb.New(startTime),
		EndTime:   timestamppb.New(endTime),
	}
}

// RegisterQueryServer is a placeholder stub for gRPC registration
// This will be replaced when protobuf generation is available
func RegisterQueryServer(server interface{}, impl QueryServer) {
	// Placeholder - will be implemented via proto generation
}

// RegisterQueryHandlerClient is a placeholder stub for gRPC gateway registration
func RegisterQueryHandlerClient(ctx context.Context, mux interface{}, client interface{}) error {
	return nil
}

// NewQueryClient creates a query client (placeholder)
func NewQueryClient(cc interface{}) QueryClient {
	return &queryClient{}
}

type queryClient struct{}

type QueryClient interface {}