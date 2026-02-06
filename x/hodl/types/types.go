package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState defines the hodl module's genesis state.
type GenesisState struct {
	Params          Params               `json:"params" yaml:"params"`
	TotalSupply     math.Int             `json:"total_supply" yaml:"total_supply"`
	TotalCollateral sdk.Coins            `json:"total_collateral" yaml:"total_collateral"`
	Positions       []CollateralPosition `json:"positions" yaml:"positions"`
}

// ProtoMessage implements proto.Message interface
func (m *GenesisState) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *GenesisState) Reset() { *m = GenesisState{} }

// String implements proto.Message interface
func (m *GenesisState) String() string {
	return "hodl_genesis"
}

// Params defines the parameters for the hodl module.
// ALL parameters are governance-controllable via validator voting
type Params struct {
	// Core parameters
	MintingEnabled   bool           `json:"minting_enabled" yaml:"minting_enabled"`
	CollateralRatio  math.LegacyDec `json:"collateral_ratio" yaml:"collateral_ratio"`     // 150% default
	StabilityFee     math.LegacyDec `json:"stability_fee" yaml:"stability_fee"`           // 0.05% annual
	LiquidationRatio math.LegacyDec `json:"liquidation_ratio" yaml:"liquidation_ratio"`   // 130% default
	MintFee          math.LegacyDec `json:"mint_fee" yaml:"mint_fee"`                     // 0.01% default
	BurnFee          math.LegacyDec `json:"burn_fee" yaml:"burn_fee"`                     // 0.01% default

	// Liquidation parameters (governance-controllable)
	LiquidationPenalty math.LegacyDec `json:"liquidation_penalty" yaml:"liquidation_penalty"` // 10% default
	LiquidatorReward   math.LegacyDec `json:"liquidator_reward" yaml:"liquidator_reward"`     // 5% of penalty

	// Price oracle parameters (governance-controllable)
	MaxPriceDeviation math.LegacyDec `json:"max_price_deviation" yaml:"max_price_deviation"` // 20% max change

	// System parameters (governance-controllable)
	BlocksPerYear   uint64   `json:"blocks_per_year" yaml:"blocks_per_year"`       // ~5,256,000 at 6s blocks
	MaxBadDebtLimit math.Int `json:"max_bad_debt_limit" yaml:"max_bad_debt_limit"` // Circuit breaker
}

// ProtoMessage implements proto.Message interface
func (m *Params) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *Params) Reset() { *m = Params{} }

// MsgServer defines the Msg service.
type MsgServer interface {
	MintHODL(goCtx context.Context, msg *SimpleMsgMintHODL) (*MsgMintHODLResponse, error)
	BurnHODL(goCtx context.Context, msg *SimpleMsgBurnHODL) (*MsgBurnHODLResponse, error)
}

// MsgMintHODLResponse defines the response for MintHODL
type MsgMintHODLResponse struct {
	HodlMinted math.Int `json:"hodl_minted" yaml:"hodl_minted"`
	Fee        math.Int `json:"fee" yaml:"fee"`
}

// MsgBurnHODLResponse defines the response for BurnHODL
type MsgBurnHODLResponse struct {
	CollateralReturned sdk.Coins `json:"collateral_returned" yaml:"collateral_returned"`
	Fee               math.Int  `json:"fee" yaml:"fee"`
}

// QueryServer defines the Query service.
type QueryServer interface {
	Params(goCtx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	TotalSupply(goCtx context.Context, req *QueryTotalSupplyRequest) (*QueryTotalSupplyResponse, error)
	Position(goCtx context.Context, req *QueryPositionRequest) (*QueryPositionResponse, error)
}

// Query requests and responses
type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params" yaml:"params"`
}

type QueryTotalSupplyRequest struct{}

type QueryTotalSupplyResponse struct {
	TotalSupply math.Int `json:"total_supply" yaml:"total_supply"`
}

type QueryPositionRequest struct {
	Owner string `json:"owner" yaml:"owner"`
}

type QueryPositionResponse struct {
	Position CollateralPosition `json:"position" yaml:"position"`
}

// Supply represents the total supply of HODL tokens
type Supply struct {
	Amount math.Int `json:"amount" yaml:"amount"`
}

// ProtoMessage implements proto.Message interface
func (m *Supply) ProtoMessage() {}

// Reset implements proto.Message interface  
func (m *Supply) Reset() { *m = Supply{} }

// String implements proto.Message interface
func (m *Supply) String() string {
	return m.Amount.String()
}