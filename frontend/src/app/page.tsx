"use client";

import React, { useEffect, useState, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";

type Level = { price: number; quantity: number };
type Trade = {
  id: string;
  price: number;
  quantity: number;
  timestamp: number;
};

export default function Home() {
  const [bids, setBids] = useState<Level[]>([]);
  const [asks, setAsks] = useState<Level[]>([]);
  const [trades, setTrades] = useState<Trade[]>([]);
  const ws = useRef<WebSocket | null>(null);

  // Form State
  const [side, setSide] = useState<"BUY" | "SELL">("BUY");
  const [price, setPrice] = useState("");
  const [quantity, setQuantity] = useState("");

  useEffect(() => {
    // In docker-compose, frontend container needs to know backend URL.
    // For local dev, we default to localhost:8080.
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";
    ws.current = new WebSocket(wsUrl);

    ws.current.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === "SNAPSHOT") {
        const sortedBids = (msg.data.bids || []).sort((a: Level, b: Level) => b.price - a.price);
        const sortedAsks = (msg.data.asks || []).sort((a: Level, b: Level) => a.price - b.price);
        setBids(sortedBids.slice(0, 15));
        setAsks(sortedAsks.slice(0, 15).reverse());
      } else if (msg.type === "TRADES") {
        setTrades((prev) => {
          const newTrades = [...msg.data, ...prev];
          return newTrades.slice(0, 50); // Keep last 50
        });
      }
    };

    return () => {
      ws.current?.close();
    };
  }, []);

  const submitOrder = async (e: React.FormEvent) => {
    e.preventDefault();
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    await fetch(`${apiUrl}/order`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        traderId: "ui_client",
        side,
        price: Number(price),
        quantity: Number(quantity),
      }),
    });
    setPrice("");
    setQuantity("");
  };

  return (
    <div className="min-h-screen bg-[#09090b] text-gray-200 p-8">
      <header className="mb-8 flex items-center justify-between border-b border-gray-800 pb-4">
        <div>
          <h1 className="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-emerald-400">
            Real-Time Equity Exchange
          </h1>
          <p className="text-sm text-gray-500">TEST/USD • Sub-ms Matching • WebSocket Live</p>
        </div>
        <div className="flex items-center gap-2">
          <span className="relative flex h-3 w-3">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
          </span>
          <span className="text-sm text-emerald-500 font-medium">Market Open</span>
        </div>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Order Book Panel */}
        <div className="col-span-1 bg-black/40 backdrop-blur-md rounded-xl border border-gray-800 p-4 shadow-xl">
          <h2 className="text-sm font-semibold text-gray-400 mb-4 uppercase tracking-wider">Order Book</h2>
          
          <div className="flex justify-between text-xs text-gray-500 mb-2 px-2">
            <span>Price</span>
            <span>Quantity</span>
          </div>

          <div className="space-y-[2px]">
            {/* Asks (Sell Orders - Red) */}
            <div className="flex flex-col gap-[2px]">
              {asks.length === 0 && <div className="text-center text-gray-600 text-xs py-4">No Asks</div>}
              {asks.map((ask, i) => (
                <div key={`ask-${ask.price}-${i}`} className="flex justify-between text-sm px-2 py-1 rounded bg-red-500/10 hover:bg-red-500/20 transition-colors">
                  <span className="text-red-400 font-mono">${(ask.price / 100).toFixed(2)}</span>
                  <span className="text-gray-300 font-mono">{ask.quantity}</span>
                </div>
              ))}
            </div>

            {/* Spread / Mid Market */}
            <div className="py-2 flex items-center justify-center gap-2 border-y border-gray-800 my-2">
              <span className="text-gray-500 text-xs">Spread</span>
              {bids.length > 0 && asks.length > 0 ? (
                <span className="text-gray-300 font-mono text-sm">
                  ${((asks[asks.length - 1].price - bids[0].price) / 100).toFixed(2)}
                </span>
              ) : (
                <span className="text-gray-600">-</span>
              )}
            </div>

            {/* Bids (Buy Orders - Green) */}
            <div className="flex flex-col gap-[2px]">
              {bids.length === 0 && <div className="text-center text-gray-600 text-xs py-4">No Bids</div>}
              {bids.map((bid, i) => (
                <div key={`bid-${bid.price}-${i}`} className="flex justify-between text-sm px-2 py-1 rounded bg-emerald-500/10 hover:bg-emerald-500/20 transition-colors">
                  <span className="text-emerald-400 font-mono">${(bid.price / 100).toFixed(2)}</span>
                  <span className="text-gray-300 font-mono">{bid.quantity}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Trade Entry & Chart Panel */}
        <div className="col-span-1 lg:col-span-1 flex flex-col gap-6">
          <div className="bg-black/40 backdrop-blur-md rounded-xl border border-gray-800 p-4 shadow-xl">
            <h2 className="text-sm font-semibold text-gray-400 mb-4 uppercase tracking-wider">Place Order</h2>
            <form onSubmit={submitOrder} className="space-y-4">
              <div className="flex bg-gray-900 p-1 rounded-lg">
                <button
                  type="button"
                  onClick={() => setSide("BUY")}
                  className={`flex-1 py-2 text-sm font-medium rounded-md transition-all ${
                    side === "BUY" ? "bg-emerald-500 text-white shadow-lg" : "text-gray-400 hover:text-white"
                  }`}
                >
                  Buy
                </button>
                <button
                  type="button"
                  onClick={() => setSide("SELL")}
                  className={`flex-1 py-2 text-sm font-medium rounded-md transition-all ${
                    side === "SELL" ? "bg-red-500 text-white shadow-lg" : "text-gray-400 hover:text-white"
                  }`}
                >
                  Sell
                </button>
              </div>

              <div>
                <label className="text-xs text-gray-500">Price (in cents)</label>
                <input
                  type="number"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  className="w-full bg-gray-900 border border-gray-800 rounded-lg px-4 py-2 mt-1 text-white focus:outline-none focus:border-blue-500 transition-colors"
                  placeholder="e.g. 10000"
                  required
                />
              </div>

              <div>
                <label className="text-xs text-gray-500">Quantity</label>
                <input
                  type="number"
                  value={quantity}
                  onChange={(e) => setQuantity(e.target.value)}
                  className="w-full bg-gray-900 border border-gray-800 rounded-lg px-4 py-2 mt-1 text-white focus:outline-none focus:border-blue-500 transition-colors"
                  placeholder="e.g. 50"
                  required
                />
              </div>

              <button
                type="submit"
                className={`w-full py-3 rounded-lg font-semibold transition-all hover:scale-[1.02] active:scale-95 ${
                  side === "BUY" ? "bg-emerald-500 hover:bg-emerald-400" : "bg-red-500 hover:bg-red-400"
                } text-white`}
              >
                Place {side} Order
              </button>
            </form>
          </div>
        </div>

        {/* Live Trades Panel */}
        <div className="col-span-1 bg-black/40 backdrop-blur-md rounded-xl border border-gray-800 p-4 shadow-xl overflow-hidden flex flex-col h-[600px]">
          <h2 className="text-sm font-semibold text-gray-400 mb-4 uppercase tracking-wider">Live Trades</h2>
          
          <div className="flex justify-between text-xs text-gray-500 mb-2 px-2">
            <span>Price</span>
            <span>Quantity</span>
            <span>Time</span>
          </div>

          <div className="flex-1 overflow-y-auto space-y-[2px] pr-2">
            <AnimatePresence initial={false}>
              {trades.map((trade) => (
                <motion.div
                  key={trade.id}
                  initial={{ opacity: 0, x: -20, backgroundColor: "#3b82f6" }}
                  animate={{ opacity: 1, x: 0, backgroundColor: "transparent" }}
                  transition={{ duration: 0.5 }}
                  className="flex justify-between text-sm px-2 py-2 rounded border-b border-gray-800/50"
                >
                  <span className="text-white font-mono">${(trade.price / 100).toFixed(2)}</span>
                  <span className="text-gray-300 font-mono">{trade.quantity}</span>
                  <span className="text-gray-500 text-xs">
                    {new Date(trade.timestamp / 1e6).toLocaleTimeString([], { hour12: false, fractionalSecondDigits: 3 })}
                  </span>
                </motion.div>
              ))}
            </AnimatePresence>
            {trades.length === 0 && (
              <div className="text-center text-gray-600 text-sm mt-10">Waiting for trades...</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
