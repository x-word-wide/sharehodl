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

// Params defines the parameters for the hodl module.
type Params struct {
	MintingEnabled   bool               `json:"minting_enabled" yaml:"minting_enabled"`
	CollateralRatio  math.LegacyDec     `json:"collateral_ratio" yaml:"collateral_ratio"`
	StabilityFee     math.LegacyDec     `json:"stability_fee" yaml:"stability_fee"`
	LiquidationRatio math.LegacyDec     `json:"liquidation_ratio" yaml:"liquidation_ratio"`
	MintFee         math.LegacyDec      `json:"mint_fee" yaml:"mint_fee"`
	BurnFee         math.LegacyDec      `json:"burn_fee" yaml:"burn_fee"`
}

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