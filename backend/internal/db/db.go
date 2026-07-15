package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/varad/exchange-backend/internal/orderbook"
)

type DB struct {
	pool       *pgxpool.Pool
	tradeQueue chan orderbook.Trade
	orderQueue chan *orderbook.Order
}

func NewDB(connString string) (*DB, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(50) PRIMARY KEY,
			trader_id VARCHAR(50),
			side VARCHAR(10),
			price BIGINT,
			quantity BIGINT,
			timestamp BIGINT
		);
		CREATE TABLE IF NOT EXISTS trades (
			id VARCHAR(50) PRIMARY KEY,
			maker_order_id VARCHAR(50),
			taker_order_id VARCHAR(50),
			price BIGINT,
			quantity BIGINT,
			timestamp BIGINT
		);
	`)
	if err != nil {
		return nil, err
	}

	db := &DB{
		pool:       pool,
		tradeQueue: make(chan orderbook.Trade, 100000), // Large buffer to avoid blocking
		orderQueue: make(chan *orderbook.Order, 100000),
	}

	// Start async workers
	go db.tradeWorker()
	go db.orderWorker()

	return db, nil
}

func (db *DB) SaveOrderAsync(order *orderbook.Order) {
	db.orderQueue <- order
}

func (db *DB) SaveTradeAsync(trade orderbook.Trade) {
	db.tradeQueue <- trade
}

func (db *DB) tradeWorker() {
	ctx := context.Background()
	for trade := range db.tradeQueue {
		_, err := db.pool.Exec(ctx,
			"INSERT INTO trades (id, maker_order_id, taker_order_id, price, quantity, timestamp) VALUES ($1, $2, $3, $4, $5, $6)",
			trade.ID, trade.MakerOrderID, trade.TakerOrderID, trade.Price, trade.Quantity, trade.Timestamp,
		)
		if err != nil {
			log.Println("Error inserting trade:", err)
		}
	}
}

func (db *DB) orderWorker() {
	ctx := context.Background()
	for order := range db.orderQueue {
		_, err := db.pool.Exec(ctx,
			"INSERT INTO orders (id, trader_id, side, price, quantity, timestamp) VALUES ($1, $2, $3, $4, $5, $6)",
			order.ID, order.TraderID, order.Side, order.Price, order.Quantity, order.Timestamp,
		)
		if err != nil {
			log.Println("Error inserting order:", err)
		}
	}
}
