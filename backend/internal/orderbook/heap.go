package orderbook

// Side represents the side of an order

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

// Order represents a single order in the order book
type Order struct {
	ID        string
	TraderID  string
	Side      Side
	Price     int64 // Storing as integer (e.g. cents) for precision
	Quantity  int64
	Timestamp int64 // Nanoseconds for strict price-time priority
}

// OrderQueue is a slice of Orders that implements heap.Interface
type OrderQueue []*Order

func (oq OrderQueue) Len() int      { return len(oq) }
func (oq OrderQueue) Swap(i, j int) { oq[i], oq[j] = oq[j], oq[i] }

func (oq *OrderQueue) Push(x interface{}) {
	item := x.(*Order)
	*oq = append(*oq, item)
}

func (oq *OrderQueue) Pop() interface{} {
	old := *oq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*oq = old[0 : n-1]
	return item
}

// BidQueue implements a Max-Heap for Buy orders (highest price first).
// If prices are equal, earlier timestamp wins.
type BidQueue struct {
	OrderQueue
}

func (bq BidQueue) Less(i, j int) bool {
	if bq.OrderQueue[i].Price == bq.OrderQueue[j].Price {
		return bq.OrderQueue[i].Timestamp < bq.OrderQueue[j].Timestamp
	}
	return bq.OrderQueue[i].Price > bq.OrderQueue[j].Price
}

// AskQueue implements a Min-Heap for Sell orders (lowest price first).
// If prices are equal, earlier timestamp wins.
type AskQueue struct {
	OrderQueue
}

func (aq AskQueue) Less(i, j int) bool {
	if aq.OrderQueue[i].Price == aq.OrderQueue[j].Price {
		return aq.OrderQueue[i].Timestamp < aq.OrderQueue[j].Timestamp
	}
	return aq.OrderQueue[i].Price < aq.OrderQueue[j].Price
}
