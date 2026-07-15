package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/varad/exchange-backend/internal/db"
	"github.com/varad/exchange-backend/internal/orderbook"
	"github.com/varad/exchange-backend/internal/ws"
)

type Engine struct {
	ob *orderbook.OrderBook
	db *db.DB
	ws *ws.Manager
}

func NewEngine(db *db.DB, ws *ws.Manager) *Engine {
	return &Engine{
		ob: orderbook.NewOrderBook(),
		db: db,
		ws: ws,
	}
}

type OrderRequest struct {
	TraderID string `json:"traderId"`
	Side     string `json:"side"` // "BUY" or "SELL"
	Price    int64  `json:"price"`
	Quantity int64  `json:"quantity"`
}

func (e *Engine) HandleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	side := orderbook.Buy
	if req.Side == "SELL" {
		side = orderbook.Sell
	}

	order := &orderbook.Order{
		ID:        fmt.Sprintf("o_%d", time.Now().UnixNano()),
		TraderID:  req.TraderID,
		Side:      side,
		Price:     req.Price,
		Quantity:  req.Quantity,
		Timestamp: time.Now().UnixNano(),
	}

	// 1. Process Order (Sub-ms matching)
	trades := e.ob.ProcessOrder(order)

	// 2. Async Persist
	if e.db != nil {
		e.db.SaveOrderAsync(order)
		for _, t := range trades {
			e.db.SaveTradeAsync(t)
		}
	}

	// 3. Broadcast Market Data (Websocket Fan-out)
	// We send the order book snapshot. In a real app we might only send diffs.
	snapshot := e.ob.GetSnapshot(50) // Top 50 levels
	
	msg := map[string]interface{}{
		"type": "SNAPSHOT",
		"data": snapshot,
	}
	e.ws.Broadcast(msg)

	// If there were trades, broadcast them too
	if len(trades) > 0 {
		tradeMsg := map[string]interface{}{
			"type": "TRADES",
			"data": trades,
		}
		e.ws.Broadcast(tradeMsg)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"orderId": order.ID,
		"trades": trades,
	})
}
