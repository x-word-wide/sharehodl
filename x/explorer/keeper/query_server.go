package keeper

import (
	"context"
	"strconv"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/explorer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryServer implements the Query service for ShareScan explorer
type QueryServer struct {
	Keeper
}

// NewQueryServerImpl creates a new QueryServer instance
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &QueryServer{Keeper: keeper}
}

var _ types.QueryServer = QueryServer{}

// Block queries

// Block returns detailed information about a specific block
func (q QueryServer) Block(c context.Context, req *types.QueryBlockRequest) (*types.QueryBlockResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	blockInfo, found := q.Keeper.GetBlockInfo(ctx, req.Height)
	if !found {
		return nil, status.Error(codes.NotFound, "block not found")
	}

	return &types.QueryBlockResponse{
		Block: &blockInfo,
	}, nil
}

// Blocks returns a list of recent blocks
func (q QueryServer) Blocks(c context.Context, req *types.QueryBlocksRequest) (*types.QueryBlocksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Set default limit if not provided
	limit := req.Limit
	if limit == 0 || limit > 100 {
		limit = 50 // Default to 50 blocks
	}

	blocks := q.Keeper.GetLatestBlocks(ctx, limit)

	return &types.QueryBlocksResponse{
		Blocks:     blocks,
		Pagination: nil, // Would implement pagination if needed
	}, nil
}

// Transaction queries

// Transaction returns detailed information about a specific transaction
func (q QueryServer) Transaction(c context.Context, req *types.QueryTransactionRequest) (*types.QueryTransactionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	txInfo, found := q.Keeper.GetTransactionInfo(ctx, req.Hash)
	if !found {
		return nil, status.Error(codes.NotFound, "transaction not found")
	}

	return &types.QueryTransactionResponse{
		Transaction: &txInfo,
	}, nil
}

// BlockTransactions returns all transactions in a specific block
func (q QueryServer) BlockTransactions(c context.Context, req *types.QueryBlockTransactionsRequest) (*types.QueryBlockTransactionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	transactions := q.Keeper.GetBlockTransactions(ctx, req.Height)

	return &types.QueryBlockTransactionsResponse{
		Transactions: transactions,
		Pagination:   nil,
	}, nil
}

// Account queries

// Account returns comprehensive account information
func (q QueryServer) Account(c context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	accountInfo := q.Keeper.GetAccountInfo(ctx, req.Address)

	return &types.QueryAccountResponse{
		Account: &accountInfo,
	}, nil
}

// Company queries

// Company returns comprehensive company information
func (q QueryServer) Company(c context.Context, req *types.QueryCompanyRequest) (*types.QueryCompanyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	companyInfo, found := q.Keeper.GetCompanyInfo(ctx, req.CompanyId)
	if !found {
		return nil, status.Error(codes.NotFound, "company not found")
	}

	return &types.QueryCompanyResponse{
		Company: &companyInfo,
	}, nil
}

// Companies returns all listed companies
func (q QueryServer) Companies(c context.Context, req *types.QueryCompaniesRequest) (*types.QueryCompaniesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	companies := q.Keeper.GetCompanies(ctx)

	return &types.QueryCompaniesResponse{
		Companies:  companies,
		Pagination: nil,
	}, nil
}

// Governance queries

// Proposal returns comprehensive governance proposal information
func (q QueryServer) Proposal(c context.Context, req *types.QueryProposalRequest) (*types.QueryProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	proposalInfo, found := q.Keeper.GetGovernanceProposalInfo(ctx, req.ProposalId)
	if !found {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	return &types.QueryProposalResponse{
		Proposal: &proposalInfo,
	}, nil
}

// Analytics queries

// NetworkStats returns comprehensive network statistics
func (q QueryServer) NetworkStats(c context.Context, req *types.QueryNetworkStatsRequest) (*types.QueryNetworkStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	stats := q.Keeper.GetNetworkStats(ctx)

	return &types.QueryNetworkStatsResponse{
		Stats: &stats,
	}, nil
}

// Analytics returns time-series analytics data
func (q QueryServer) Analytics(c context.Context, req *types.QueryAnalyticsRequest) (*types.QueryAnalyticsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	analytics := q.Keeper.GetAnalytics(ctx, req.TimeFrame, req.StartTime.AsTime(), req.EndTime.AsTime())

	return &types.QueryAnalyticsResponse{
		Analytics: &analytics,
	}, nil
}

// Search queries

// Search performs comprehensive search across all indexed data
func (q QueryServer) Search(c context.Context, req *types.QuerySearchRequest) (*types.QuerySearchResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Set default limit if not provided
	limit := req.Limit
	if limit == 0 || limit > 100 {
		limit = 20 // Default to 20 results
	}

	results := q.Keeper.Search(ctx, req.Query, limit)

	return &types.QuerySearchResponse{
		Results: results,
	}, nil
}

// Real-time queries

// LiveMetrics returns current live metrics
func (q QueryServer) LiveMetrics(c context.Context, req *types.QueryLiveMetricsRequest) (*types.QueryLiveMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	metrics := q.Keeper.GetLiveMetrics(ctx)

	return &types.QueryLiveMetricsResponse{
		Metrics: &metrics,
	}, nil
}

// Advanced queries with pagination

// TransactionsByAccount returns transactions for a specific account with pagination
func (q QueryServer) TransactionsByAccount(c context.Context, req *types.QueryTransactionsByAccountRequest) (*types.QueryTransactionsByAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := runtime.KVStoreAdapter(q.Keeper.storeService.OpenKVStore(ctx))
	accountTxStore := prefix.NewStore(store, types.AccountTransactionPrefix)

	addrPrefix := []byte(req.Address)
	
	transactions, pageRes, err := query.GenericFilteredPaginate(
		q.Keeper.cdc,
		accountTxStore,
		req.Pagination,
		func(key []byte, value []byte) (*types.TransactionInfo, error) {
			// Extract transaction hash from key
			txHash := string(key[len(addrPrefix):])
			
			// Get full transaction info
			if txInfo, found := q.Keeper.GetTransactionInfo(ctx, txHash); found {
				return &txInfo, nil
			}
			return nil, status.Error(codes.NotFound, "transaction not found")
		},
		func() *types.TransactionInfo {
			return &types.TransactionInfo{}
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTransactionsByAccountResponse{
		Transactions: transactions,
		Pagination:   pageRes,
	}, nil
}

// TradesByCompany returns trades for a specific company with pagination
func (q QueryServer) TradesByCompany(c context.Context, req *types.QueryTradesByCompanyRequest) (*types.QueryTradesByCompanyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Get company trades (simplified implementation)
	trades := q.Keeper.getCompanyTrades(ctx, req.CompanyId, 100) // Get up to 100 trades

	return &types.QueryTradesByCompanyResponse{
		Trades:     trades,
		Pagination: nil,
	}, nil
}

// DividendsByCompany returns dividends for a specific company
func (q QueryServer) DividendsByCompany(c context.Context, req *types.QueryDividendsByCompanyRequest) (*types.QueryDividendsByCompanyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Get company dividends (simplified implementation)
	dividends := q.Keeper.getCompanyDividends(ctx, req.CompanyId, 50) // Get up to 50 dividends

	return &types.QueryDividendsByCompanyResponse{
		Dividends:  dividends,
		Pagination: nil,
	}, nil
}

// VotesByProposal returns votes for a specific governance proposal
func (q QueryServer) VotesByProposal(c context.Context, req *types.QueryVotesByProposalRequest) (*types.QueryVotesByProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Get proposal votes (simplified implementation)
	votes := q.Keeper.getProposalVotes(ctx, req.ProposalId)

	return &types.QueryVotesByProposalResponse{
		Votes:      votes,
		Pagination: nil,
	}, nil
}

// Chart data queries

// ChartData returns chart data for visualizations
func (q QueryServer) ChartData(c context.Context, req *types.QueryChartDataRequest) (*types.QueryChartDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Generate chart data based on type
	chartData := q.generateChartData(ctx, req.ChartType, req.TimeFrame, req.CompanyId)

	return &types.QueryChartDataResponse{
		ChartData: &chartData,
	}, nil
}

// generateChartData generates chart data for different visualization types
func (q QueryServer) generateChartData(ctx sdk.Context, chartType string, timeFrame string, companyID uint64) types.ChartData {
	switch chartType {
	case "price":
		return q.generatePriceChartData(ctx, companyID, timeFrame)
	case "volume":
		return q.generateVolumeChartData(ctx, companyID, timeFrame)
	case "network_activity":
		return q.generateNetworkActivityChartData(ctx, timeFrame)
	case "governance_activity":
		return q.generateGovernanceActivityChartData(ctx, timeFrame)
	default:
		return types.ChartData{
			Labels:   []string{},
			Datasets: []types.Dataset{},
		}
	}
}

// generatePriceChartData generates price chart data for a company
func (q QueryServer) generatePriceChartData(ctx sdk.Context, companyID uint64, timeFrame string) types.ChartData {
	// Placeholder implementation - would generate real price data
	labels := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	data := []float64{100, 105, 98, 110, 115, 108}

	dataset := types.Dataset{
		Label:           "Price",
		Data:            data,
		BackgroundColor: "rgba(54, 162, 235, 0.2)",
		BorderColor:     "rgba(54, 162, 235, 1)",
		BorderWidth:     2,
	}

	return types.ChartData{
		Labels:   labels,
		Datasets: []types.Dataset{dataset},
	}
}

// generateVolumeChartData generates volume chart data for a company
func (q QueryServer) generateVolumeChartData(ctx sdk.Context, companyID uint64, timeFrame string) types.ChartData {
	// Placeholder implementation - would generate real volume data
	labels := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	data := []float64{1000000, 1200000, 800000, 1500000, 1800000, 1300000}

	dataset := types.Dataset{
		Label:           "Volume",
		Data:            data,
		BackgroundColor: "rgba(255, 99, 132, 0.2)",
		BorderColor:     "rgba(255, 99, 132, 1)",
		BorderWidth:     2,
	}

	return types.ChartData{
		Labels:   labels,
		Datasets: []types.Dataset{dataset},
	}
}

// generateNetworkActivityChartData generates network activity chart data
func (q QueryServer) generateNetworkActivityChartData(ctx sdk.Context, timeFrame string) types.ChartData {
	// Placeholder implementation - would generate real network activity data
	labels := []string{"00:00", "04:00", "08:00", "12:00", "16:00", "20:00"}
	transactionData := []float64{50, 80, 120, 200, 180, 90}
	blockData := []float64{10, 16, 24, 40, 36, 18}

	transactionDataset := types.Dataset{
		Label:           "Transactions",
		Data:            transactionData,
		BackgroundColor: "rgba(75, 192, 192, 0.2)",
		BorderColor:     "rgba(75, 192, 192, 1)",
		BorderWidth:     2,
	}

	blockDataset := types.Dataset{
		Label:           "Blocks",
		Data:            blockData,
		BackgroundColor: "rgba(153, 102, 255, 0.2)",
		BorderColor:     "rgba(153, 102, 255, 1)",
		BorderWidth:     2,
	}

	return types.ChartData{
		Labels:   labels,
		Datasets: []types.Dataset{transactionDataset, blockDataset},
	}
}

// generateGovernanceActivityChartData generates governance activity chart data
func (q QueryServer) generateGovernanceActivityChartData(ctx sdk.Context, timeFrame string) types.ChartData {
	// Placeholder implementation - would generate real governance activity data
	labels := []string{"Q1", "Q2", "Q3", "Q4"}
	proposalData := []float64{5, 8, 12, 6}
	voteData := []float64{150, 240, 360, 180}

	proposalDataset := types.Dataset{
		Label:           "Proposals",
		Data:            proposalData,
		BackgroundColor: "rgba(255, 206, 86, 0.2)",
		BorderColor:     "rgba(255, 206, 86, 1)",
		BorderWidth:     2,
	}

	voteDataset := types.Dataset{
		Label:           "Votes",
		Data:            voteData,
		BackgroundColor: "rgba(54, 162, 235, 0.2)",
		BorderColor:     "rgba(54, 162, 235, 1)",
		BorderWidth:     2,
	}

	return types.ChartData{
		Labels:   labels,
		Datasets: []types.Dataset{proposalDataset, voteDataset},
	}
}