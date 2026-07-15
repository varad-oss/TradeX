package orderbook

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type Trade struct {
	ID           string `json:"id"`
	MakerOrderID string `json:"makerOrderId"`
	TakerOrderID string `json:"takerOrderId"`
	Price        int64  `json:"price"`
	Quantity     int64  `json:"quantity"`
	Timestamp    int64  `json:"timestamp"`
}

type OrderBook struct {
	mu   sync.RWMutex
	Bids *BidQueue
	Asks *AskQueue
}

func NewOrderBook() *OrderBook {
	bids := &BidQueue{OrderQueue: make([]*Order, 0)}
	asks := &AskQueue{OrderQueue: make([]*Order, 0)}
	heap.Init(bids)
	heap.Init(asks)
	return &OrderBook{
		Bids: bids,
		Asks: asks,
	}
}

// ProcessOrder processes a new order, matching it against the book if possible,
// and returning a slice of generated trades. If the order is not fully filled,
// the remainder is added to the order book.
func (ob *OrderBook) ProcessOrder(order *Order) []Trade {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	var trades []Trade

	if order.Side == Buy {
		trades = ob.processBuyOrder(order)
	} else {
		trades = ob.processSellOrder(order)
	}

	return trades
}

func (ob *OrderBook) processBuyOrder(order *Order) []Trade {
	var trades []Trade

	for order.Quantity > 0 && ob.Asks.Len() > 0 {
		bestAsk := (*ob.Asks).OrderQueue[0]
		if order.Price < bestAsk.Price {
			break // No match possible
		}

		// Match found!
		tradeQty := order.Quantity
		if bestAsk.Quantity < tradeQty {
			tradeQty = bestAsk.Quantity
		}

		trade := Trade{
			ID:           fmt.Sprintf("tr_%d", time.Now().UnixNano()), // simple ID gen
			MakerOrderID: bestAsk.ID,
			TakerOrderID: order.ID,
			Price:        bestAsk.Price, // Maker price determines trade price
			Quantity:     tradeQty,
			Timestamp:    time.Now().UnixNano(),
		}
		trades = append(trades, trade)

		order.Quantity -= tradeQty
		bestAsk.Quantity -= tradeQty

		if bestAsk.Quantity == 0 {
			heap.Pop(ob.Asks) // Remove fully filled ask
		} else {
			// Best ask is partially filled, no need to Pop and Push since price/timestamp 
			// don't change, but in a strict heap, we might need heap.Fix(ob.Asks, 0) if keys changed.
			// Since Quantity is not a heap key, it's fine.
		}
	}

	if order.Quantity > 0 {
		heap.Push(ob.Bids, order)
	}

	return trades
}

func (ob *OrderBook) processSellOrder(order *Order) []Trade {
	var trades []Trade

	for order.Quantity > 0 && ob.Bids.Len() > 0 {
		bestBid := (*ob.Bids).OrderQueue[0]
		if order.Price > bestBid.Price {
			break // No match possible
		}

		// Match found!
		tradeQty := order.Quantity
		if bestBid.Quantity < tradeQty {
			tradeQty = bestBid.Quantity
		}

		trade := Trade{
			ID:           fmt.Sprintf("tr_%d", time.Now().UnixNano()),
			MakerOrderID: bestBid.ID,
			TakerOrderID: order.ID,
			Price:        bestBid.Price, // Maker price determines trade price
			Quantity:     tradeQty,
			Timestamp:    time.Now().UnixNano(),
		}
		trades = append(trades, trade)

		order.Quantity -= tradeQty
		bestBid.Quantity -= tradeQty

		if bestBid.Quantity == 0 {
			heap.Pop(ob.Bids) // Remove fully filled bid
		}
	}

	if order.Quantity > 0 {
		heap.Push(ob.Asks, order)
	}

	return trades
}

type Level struct {
	Price    int64 `json:"price"`
	Quantity int64 `json:"quantity"`
}

type Snapshot struct {
	Bids []Level `json:"bids"`
	Asks []Level `json:"asks"`
}

// GetSnapshot returns a quick copy of the top levels of the order book for streaming
func (ob *OrderBook) GetSnapshot(depth int) Snapshot {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bidAgg := make(map[int64]int64)
	for _, o := range ob.Bids.OrderQueue {
		bidAgg[o.Price] += o.Quantity
	}
	
	askAgg := make(map[int64]int64)
	for _, o := range ob.Asks.OrderQueue {
		askAgg[o.Price] += o.Quantity
	}
	
	var snap Snapshot
	for p, q := range bidAgg {
		snap.Bids = append(snap.Bids, Level{Price: p, Quantity: q})
	}
	for p, q := range askAgg {
		snap.Asks = append(snap.Asks, Level{Price: p, Quantity: q})
	}
	
	return snap
}
