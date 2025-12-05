"use client";

import { useState } from "react";

interface OrderFormProps {
  symbol: string;
}

export default function AdvancedTrading({ symbol }: OrderFormProps) {
  const [orderType, setOrderType] = useState("limit");
  const [timeInForce, setTimeInForce] = useState("gtc");
  const [side, setSide] = useState("buy");
  const [price, setPrice] = useState("");
  const [quantity, setQuantity] = useState("");
  const [stopPrice, setStopPrice] = useState("");

  const marketData = {
    "APPLE/HODL": { lastPrice: 185.25, bid: 185.20, ask: 185.30, volume: 12500000 },
    "TSLA/HODL": { lastPrice: 245.80, bid: 245.75, ask: 245.85, volume: 8900000 },
    "GOOGL/HODL": { lastPrice: 2750.00, bid: 2749.50, ask: 2750.50, volume: 3200000 },
    "MSFT/HODL": { lastPrice: 385.60, bid: 385.55, ask: 385.65, volume: 6700000 }
  };

  const orderBook = {
    bids: [
      { price: 185.20, size: 1500, total: 1500 },
      { price: 185.15, size: 2300, total: 3800 },
      { price: 185.10, size: 1800, total: 5600 },
      { price: 185.05, size: 3200, total: 8800 },
      { price: 185.00, size: 2100, total: 10900 }
    ],
    asks: [
      { price: 185.30, size: 1200, total: 1200 },
      { price: 185.35, size: 1900, total: 3100 },
      { price: 185.40, size: 2500, total: 5600 },
      { price: 185.45, size: 1700, total: 7300 },
      { price: 185.50, size: 2800, total: 10100 }
    ]
  };

  const recentTrades = [
    { time: "14:23:45", price: 185.25, size: 150, side: "buy" },
    { time: "14:23:44", price: 185.23, size: 89, side: "sell" },
    { time: "14:23:43", price: 185.24, size: 234, side: "buy" },
    { time: "14:23:42", price: 185.22, size: 167, side: "sell" },
    { time: "14:23:41", price: 185.26, size: 312, side: "buy" }
  ];

  const myOrders = [
    { id: "ORD-001", side: "buy", type: "limit", quantity: 100, price: 185.00, tif: "gtc", status: "open" },
    { id: "ORD-002", side: "sell", type: "limit", quantity: 50, price: 185.50, tif: "gtc", status: "open" },
    { id: "ORD-003", side: "buy", type: "fok", quantity: 200, price: 185.10, tif: "fok", status: "rejected" }
  ];

  const market = marketData[symbol as keyof typeof marketData] || marketData["APPLE/HODL"];

  return (
    <div className="grid grid-cols-12 gap-4 h-screen p-4">
      {/* Order Entry Panel */}
      <div className="col-span-3 space-y-4">
        <div className="border rounded-lg p-4">
          <h3 className="font-semibold mb-4 flex items-center gap-2">
            📊 Professional Order Entry
          </h3>
          
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium mb-1">Symbol</label>
              <div className="text-lg font-bold">{symbol}</div>
              <div className="text-sm text-gray-600">${market.lastPrice}</div>
            </div>

            <div className="grid grid-cols-2 gap-2">
              <button 
                onClick={() => setSide("buy")}
                className={`p-2 rounded font-semibold ${
                  side === "buy" 
                    ? "bg-green-500 text-white" 
                    : "border border-green-500 text-green-500"
                }`}
              >
                BUY
              </button>
              <button 
                onClick={() => setSide("sell")}
                className={`p-2 rounded font-semibold ${
                  side === "sell" 
                    ? "bg-red-500 text-white" 
                    : "border border-red-500 text-red-500"
                }`}
              >
                SELL
              </button>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Order Type</label>
              <select 
                value={orderType} 
                onChange={(e) => setOrderType(e.target.value)}
                className="w-full border rounded p-2"
              >
                <option value="market">Market</option>
                <option value="limit">Limit</option>
                <option value="stop">Stop</option>
                <option value="stop_limit">Stop Limit</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Time in Force</label>
              <select 
                value={timeInForce} 
                onChange={(e) => setTimeInForce(e.target.value)}
                className="w-full border rounded p-2"
              >
                <option value="gtc">Good Till Cancelled (GTC)</option>
                <option value="ioc">Immediate or Cancel (IOC)</option>
                <option value="fok">Fill or Kill (FOK)</option>
                <option value="gtd">Good Till Date (GTD)</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Quantity</label>
              <input 
                type="number" 
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
                placeholder="0"
                className="w-full border rounded p-2"
              />
            </div>

            {(orderType === "limit" || orderType === "stop_limit") && (
              <div>
                <label className="block text-sm font-medium mb-1">Price</label>
                <input 
                  type="number" 
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  placeholder="0.00"
                  step="0.01"
                  className="w-full border rounded p-2"
                />
              </div>
            )}

            {(orderType === "stop" || orderType === "stop_limit") && (
              <div>
                <label className="block text-sm font-medium mb-1">Stop Price</label>
                <input 
                  type="number" 
                  value={stopPrice}
                  onChange={(e) => setStopPrice(e.target.value)}
                  placeholder="0.00"
                  step="0.01"
                  className="w-full border rounded p-2"
                />
              </div>
            )}

            {timeInForce === "fok" && (
              <div className="bg-blue-50 border border-blue-200 rounded p-3">
                <div className="text-sm text-blue-800">
                  <strong>Fill or Kill (FOK):</strong> Order must be filled completely at the specified price or better, or it will be immediately cancelled.
                </div>
              </div>
            )}

            {timeInForce === "ioc" && (
              <div className="bg-orange-50 border border-orange-200 rounded p-3">
                <div className="text-sm text-orange-800">
                  <strong>Immediate or Cancel (IOC):</strong> Fill whatever quantity is immediately available, cancel the rest.
                </div>
              </div>
            )}

            <button className={`w-full py-3 rounded font-semibold text-white ${
              side === "buy" ? "bg-green-500" : "bg-red-500"
            }`}>
              {side === "buy" ? "BUY" : "SELL"} {symbol}
            </button>
          </div>
        </div>

        {/* Market Status */}
        <div className="border rounded-lg p-4">
          <h4 className="font-semibold mb-3">🔄 Market Status</h4>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span>Status:</span>
              <span className="text-green-600 font-semibold">OPEN</span>
            </div>
            <div className="flex justify-between">
              <span>Circuit Breaker:</span>
              <span className="text-green-600">NORMAL</span>
            </div>
            <div className="flex justify-between">
              <span>24h Volume:</span>
              <span>${(market.volume / 1000000).toFixed(1)}M</span>
            </div>
            <div className="flex justify-between">
              <span>Settlement:</span>
              <span className="text-blue-600 font-semibold">T+0 (6sec)</span>
            </div>
          </div>
        </div>
      </div>

      {/* Order Book */}
      <div className="col-span-3">
        <div className="border rounded-lg p-4 h-full">
          <h3 className="font-semibold mb-4">📋 Order Book</h3>
          
          <div className="space-y-1 text-xs">
            {/* Asks (Sell Orders) */}
            <div className="text-center text-gray-500 font-medium mb-2">ASKS</div>
            {orderBook.asks.reverse().map((ask, i) => (
              <div key={i} className="grid grid-cols-3 gap-1 py-1 hover:bg-red-50">
                <div className="text-right text-red-600">{ask.price.toFixed(2)}</div>
                <div className="text-right">{ask.size}</div>
                <div className="text-right text-gray-500">{ask.total}</div>
              </div>
            ))}
            
            {/* Spread */}
            <div className="text-center py-2 bg-gray-50 font-medium">
              Spread: ${(market.ask - market.bid).toFixed(2)}
            </div>
            
            {/* Bids (Buy Orders) */}
            <div className="text-center text-gray-500 font-medium mb-2">BIDS</div>
            {orderBook.bids.map((bid, i) => (
              <div key={i} className="grid grid-cols-3 gap-1 py-1 hover:bg-green-50">
                <div className="text-right text-green-600">{bid.price.toFixed(2)}</div>
                <div className="text-right">{bid.size}</div>
                <div className="text-right text-gray-500">{bid.total}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Price Chart Area */}
      <div className="col-span-4">
        <div className="border rounded-lg p-4 h-full">
          <h3 className="font-semibold mb-4">📈 Real-Time Price Chart</h3>
          <div className="h-64 bg-gray-50 rounded flex items-center justify-center">
            <div className="text-center text-gray-500">
              <div className="text-4xl mb-2">📊</div>
              <div>Live trading chart</div>
              <div className="text-sm">TradingView integration ready</div>
            </div>
          </div>
          
          <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
            <div>
              <div className="text-gray-600">Last Price</div>
              <div className="text-xl font-bold">${market.lastPrice}</div>
            </div>
            <div>
              <div className="text-gray-600">24h Change</div>
              <div className="text-green-600 font-semibold">+2.45%</div>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Trades & My Orders */}
      <div className="col-span-2 space-y-4">
        <div className="border rounded-lg p-4">
          <h4 className="font-semibold mb-3">⚡ Recent Trades</h4>
          <div className="space-y-1 text-xs">
            <div className="grid grid-cols-3 gap-1 text-gray-500 font-medium">
              <div>Time</div>
              <div className="text-right">Price</div>
              <div className="text-right">Size</div>
            </div>
            {recentTrades.map((trade, i) => (
              <div key={i} className="grid grid-cols-3 gap-1 py-1">
                <div className="text-gray-600">{trade.time}</div>
                <div className={`text-right font-medium ${
                  trade.side === "buy" ? "text-green-600" : "text-red-600"
                }`}>
                  {trade.price.toFixed(2)}
                </div>
                <div className="text-right">{trade.size}</div>
              </div>
            ))}
          </div>
        </div>

        <div className="border rounded-lg p-4">
          <h4 className="font-semibold mb-3">📋 My Orders</h4>
          <div className="space-y-2 text-xs">
            {myOrders.map((order) => (
              <div key={order.id} className="border rounded p-2">
                <div className="flex justify-between items-center mb-1">
                  <span className={`font-semibold ${
                    order.side === "buy" ? "text-green-600" : "text-red-600"
                  }`}>
                    {order.side.toUpperCase()} {order.quantity}
                  </span>
                  <span className={`text-xs px-2 py-1 rounded ${
                    order.status === "open" 
                      ? "bg-blue-100 text-blue-800"
                      : order.status === "rejected"
                      ? "bg-red-100 text-red-800" 
                      : "bg-green-100 text-green-800"
                  }`}>
                    {order.status.toUpperCase()}
                  </span>
                </div>
                <div className="text-gray-600">
                  {order.type.toUpperCase()} @ ${order.price} ({order.tif.toUpperCase()})
                </div>
                {order.status === "rejected" && order.type === "fok" && (
                  <div className="text-red-600 text-xs mt-1">
                    FOK order rejected - insufficient liquidity
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Trading Safeguards Alert */}
      <div className="col-span-12 mt-4">
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <h4 className="font-semibold text-blue-800 mb-2">🛡️ ShareHODL Trading Safeguards Active</h4>
          <div className="grid grid-cols-4 gap-4 text-sm text-blue-700">
            <div>
              <strong>Circuit Breakers:</strong> Automatic halt at 15% price movement
            </div>
            <div>
              <strong>Ownership Limits:</strong> 49.9% maximum individual ownership
            </div>
            <div>
              <strong>Professional Orders:</strong> FOK, IOC, GTC, GTD supported
            </div>
            <div>
              <strong>Instant Settlement:</strong> T+0 with 6-second finality
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}