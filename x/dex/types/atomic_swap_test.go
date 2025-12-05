package types

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestMsgPlaceAtomicSwapOrder tests atomic swap message validation
func TestMsgPlaceAtomicSwapOrder(t *testing.T) {
	validTrader := sdk.AccAddress("test_trader_addr___").String()

	tests := []struct {
		name    string
		msg     MsgPlaceAtomicSwapOrder
		wantErr bool
	}{
		{
			name: "valid atomic swap order",
			msg: MsgPlaceAtomicSwapOrder{
				Trader:      validTrader,
				FromSymbol:  "APPLE",
				ToSymbol:    "HODL",
				Quantity:    100,
				MaxSlippage: math.LegacyNewDecWithPrec(5, 2), // 5%
				TimeInForce: TimeInForceGTC,
			},
			wantErr: false,
		},
		{
			name: "invalid trader address",
			msg: MsgPlaceAtomicSwapOrder{
				Trader:      "invalid_address",
				FromSymbol:  "APPLE",
				ToSymbol:    "HODL",
				Quantity:    100,
				MaxSlippage: math.LegacyNewDecWithPrec(5, 2),
				TimeInForce: TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "same asset swap",
			msg: MsgPlaceAtomicSwapOrder{
				Trader:      validTrader,
				FromSymbol:  "APPLE",
				ToSymbol:    "APPLE",
				Quantity:    100,
				MaxSlippage: math.LegacyNewDecWithPrec(5, 2),
				TimeInForce: TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "zero quantity",
			msg: MsgPlaceAtomicSwapOrder{
				Trader:      validTrader,
				FromSymbol:  "APPLE",
				ToSymbol:    "HODL",
				Quantity:    0,
				MaxSlippage: math.LegacyNewDecWithPrec(5, 2),
				TimeInForce: TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "negative slippage",
			msg: MsgPlaceAtomicSwapOrder{
				Trader:      validTrader,
				FromSymbol:  "APPLE",
				ToSymbol:    "HODL",
				Quantity:    100,
				MaxSlippage: math.LegacyNewDec(-1),
				TimeInForce: TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "excessive slippage",
			msg: MsgPlaceAtomicSwapOrder{
				Trader:      validTrader,
				FromSymbol:  "APPLE",
				ToSymbol:    "HODL",
				Quantity:    100,
				MaxSlippage: math.LegacyNewDecWithPrec(60, 2), // 60% - excessive
				TimeInForce: TimeInForceGTC,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err, "Expected validation error for %s", tt.name)
			} else {
				require.NoError(t, err, "Expected no validation error for %s", tt.name)
			}
		})
	}
}

// TestMsgPlaceFractionalOrder tests fractional order message validation
func TestMsgPlaceFractionalOrder(t *testing.T) {
	validTrader := sdk.AccAddress("test_trader_addr___").String()

	tests := []struct {
		name    string
		msg     MsgPlaceFractionalOrder
		wantErr bool
	}{
		{
			name: "valid fractional buy order",
			msg: MsgPlaceFractionalOrder{
				Trader:             validTrader,
				Symbol:             "APPLE",
				Side:               OrderSideBuy,
				FractionalQuantity: math.LegacyNewDecWithPrec(5, 1), // 0.5 shares
				Price:              math.LegacyNewDec(150),
				MaxPrice:           math.LegacyNewDec(160),
				TimeInForce:        TimeInForceGTC,
			},
			wantErr: false,
		},
		{
			name: "valid fractional sell order",
			msg: MsgPlaceFractionalOrder{
				Trader:             validTrader,
				Symbol:             "TSLA",
				Side:               OrderSideSell,
				FractionalQuantity: math.LegacyNewDecWithPrec(25, 2), // 0.25 shares
				Price:              math.LegacyNewDec(800),
				MaxPrice:           math.LegacyNewDec(750), // Min price for sell
				TimeInForce:        TimeInForceGTC,
			},
			wantErr: false,
		},
		{
			name: "invalid trader address",
			msg: MsgPlaceFractionalOrder{
				Trader:             "invalid_address",
				Symbol:             "APPLE",
				Side:               OrderSideBuy,
				FractionalQuantity: math.LegacyNewDecWithPrec(5, 1),
				Price:              math.LegacyNewDec(150),
				MaxPrice:           math.LegacyNewDec(160),
				TimeInForce:        TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "zero fractional quantity",
			msg: MsgPlaceFractionalOrder{
				Trader:             validTrader,
				Symbol:             "APPLE",
				Side:               OrderSideBuy,
				FractionalQuantity: math.LegacyZeroDec(),
				Price:              math.LegacyNewDec(150),
				MaxPrice:           math.LegacyNewDec(160),
				TimeInForce:        TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "negative fractional quantity",
			msg: MsgPlaceFractionalOrder{
				Trader:             validTrader,
				Symbol:             "APPLE",
				Side:               OrderSideBuy,
				FractionalQuantity: math.LegacyNewDec(-1),
				Price:              math.LegacyNewDec(150),
				MaxPrice:           math.LegacyNewDec(160),
				TimeInForce:        TimeInForceGTC,
			},
			wantErr: true,
		},
		{
			name: "invalid price",
			msg: MsgPlaceFractionalOrder{
				Trader:             validTrader,
				Symbol:             "APPLE",
				Side:               OrderSideBuy,
				FractionalQuantity: math.LegacyNewDecWithPrec(5, 1),
				Price:              math.LegacyZeroDec(),
				MaxPrice:           math.LegacyNewDec(160),
				TimeInForce:        TimeInForceGTC,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err, "Expected validation error for %s", tt.name)
			} else {
				require.NoError(t, err, "Expected no validation error for %s", tt.name)
			}
		})
	}
}

// TestMsgCreateTradingStrategy tests trading strategy message validation  
func TestMsgCreateTradingStrategy(t *testing.T) {
	validOwner := sdk.AccAddress("test_creator_addr__").String()

	tests := []struct {
		name    string
		msg     MsgCreateTradingStrategy
		wantErr bool
	}{
		{
			name: "valid trading strategy",
			msg: MsgCreateTradingStrategy{
				Owner:       validOwner,
				Name:        "Momentum Strategy",
				Conditions:  []TriggerCondition{
					{
						Symbol:         "APPLE",
						PriceThreshold: math.LegacyNewDecWithPrec(105, 2), // 1.05x current price
						ConditionType:  "ABOVE",
						IsPercentage:   true,
					},
				},
				Actions:     []TradingAction{
					{
						ActionType:  OrderTypeMarket,
						Symbol:      "APPLE",
						Side:        OrderSideBuy,
						Quantity:    math.LegacyNewDec(100),
						PriceOffset: math.LegacyNewDec(5), // $5 above market
					},
				},
				MaxExposure: math.NewInt(10000),
			},
			wantErr: false,
		},
		{
			name: "invalid owner address",
			msg: MsgCreateTradingStrategy{
				Owner:       "invalid_address",
				Name:        "Test Strategy",
				Conditions:  []TriggerCondition{},
				Actions:     []TradingAction{},
				MaxExposure: math.NewInt(10000),
			},
			wantErr: true,
		},
		{
			name: "empty name",
			msg: MsgCreateTradingStrategy{
				Owner:       validOwner,
				Name:        "",
				Conditions:  []TriggerCondition{},
				Actions:     []TradingAction{},
				MaxExposure: math.NewInt(10000),
			},
			wantErr: true,
		},
		{
			name: "zero max exposure",
			msg: MsgCreateTradingStrategy{
				Owner:       validOwner,
				Name:        "Test Strategy",
				Conditions:  []TriggerCondition{},
				Actions:     []TradingAction{},
				MaxExposure: math.ZeroInt(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err, "Expected validation error for %s", tt.name)
			} else {
				require.NoError(t, err, "Expected no validation error for %s", tt.name)
			}
		})
	}
}

// TestSlippageCalculation tests slippage protection calculations
func TestSlippageCalculation(t *testing.T) {
	tests := []struct {
		name                   string
		currentRate            math.LegacyDec
		marketRate             math.LegacyDec
		maxSlippage            math.LegacyDec
		expectSlippageExceeded bool
	}{
		{
			name:                   "no slippage",
			currentRate:            math.LegacyNewDec(100),
			marketRate:             math.LegacyNewDec(100),
			maxSlippage:            math.LegacyNewDecWithPrec(5, 2), // 5%
			expectSlippageExceeded: false,
		},
		{
			name:                   "acceptable slippage",
			currentRate:            math.LegacyNewDec(102),
			marketRate:             math.LegacyNewDec(100),
			maxSlippage:            math.LegacyNewDecWithPrec(5, 2), // 5%
			expectSlippageExceeded: false,
		},
		{
			name:                   "excessive slippage",
			currentRate:            math.LegacyNewDec(110),
			marketRate:             math.LegacyNewDec(100),
			maxSlippage:            math.LegacyNewDecWithPrec(5, 2), // 5%
			expectSlippageExceeded: true,
		},
		{
			name:                   "better price (negative slippage)",
			currentRate:            math.LegacyNewDec(95),
			marketRate:             math.LegacyNewDec(100),
			maxSlippage:            math.LegacyNewDecWithPrec(5, 2), // 5%
			expectSlippageExceeded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate actual slippage: |currentRate - marketRate| / marketRate
			priceDiff := tt.currentRate.Sub(tt.marketRate).Abs()
			slippage := priceDiff.Quo(tt.marketRate)
			slippageExceeded := slippage.GT(tt.maxSlippage)

			require.Equal(t, tt.expectSlippageExceeded, slippageExceeded,
				"Slippage calculation for %s: rate=%.2f, market=%.2f, slippage=%.4f, max=%.4f",
				tt.name,
				float64(tt.currentRate.MustFloat64()),
				float64(tt.marketRate.MustFloat64()),
				float64(slippage.MustFloat64()),
				float64(tt.maxSlippage.MustFloat64()),
			)
		})
	}
}

// TestParseOrderID tests order ID parsing
func TestParseOrderID(t *testing.T) {
	tests := []struct {
		name     string
		orderStr string
		expected uint64
		wantErr  bool
	}{
		{
			name:     "valid order ID",
			orderStr: "12345",
			expected: 12345,
			wantErr:  false,
		},
		{
			name:     "zero order ID",
			orderStr: "0",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "large order ID",
			orderStr: "18446744073709551615", // max uint64
			expected: 18446744073709551615,
			wantErr:  false,
		},
		{
			name:     "invalid order ID - non-numeric",
			orderStr: "abc123",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid order ID - negative",
			orderStr: "-123",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid order ID - empty",
			orderStr: "",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseOrderID(tt.orderStr)
			if tt.wantErr {
				require.Error(t, err, "Expected parsing error for %s", tt.name)
			} else {
				require.NoError(t, err, "Expected no parsing error for %s", tt.name)
				require.Equal(t, tt.expected, result, "Parsed order ID should match expected")
			}
		})
	}
}

// TestOrderTypeValidation tests order type constants and validation
func TestOrderTypeValidation(t *testing.T) {
	// Test that our blockchain-native order types are properly defined
	require.Equal(t, OrderType(4), OrderTypeAtomicSwap, "AtomicSwap should be order type 4")
	require.Equal(t, OrderType(5), OrderTypeFractional, "Fractional should be order type 5")
	require.Equal(t, OrderType(6), OrderTypeProgrammatic, "Programmatic should be order type 6")

	// Test order type string representation
	require.NotEmpty(t, OrderTypeAtomicSwap.String(), "AtomicSwap should have string representation")
	require.NotEmpty(t, OrderTypeFractional.String(), "Fractional should have string representation")
	require.NotEmpty(t, OrderTypeProgrammatic.String(), "Programmatic should have string representation")
}

// TestOrderSideValidation tests order side validation
func TestOrderSideValidation(t *testing.T) {
	// Test valid order sides
	require.Equal(t, OrderSide(0), OrderSideBuy, "Buy should be order side 0")
	require.Equal(t, OrderSide(1), OrderSideSell, "Sell should be order side 1")

	// Test order side string representation
	require.NotEmpty(t, OrderSideBuy.String(), "Buy should have string representation")
	require.NotEmpty(t, OrderSideSell.String(), "Sell should have string representation")
}