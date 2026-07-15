package main

import (
	"log"
	"net/http"
	"os"

	"github.com/varad/exchange-backend/internal/db"
	"github.com/varad/exchange-backend/internal/engine"
	"github.com/varad/exchange-backend/internal/ws"
)

func main() {
	// Initialize WebSocket manager
	wsManager := ws.NewManager()
	go wsManager.Run()

	// Initialize DB (Postgres)
	dbConnString := os.Getenv("DATABASE_URL")
	var database *db.DB
	var err error
	if dbConnString != "" {
		database, err = db.NewDB(dbConnString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		log.Println("Connected to PostgreSQL")
	} else {
		log.Println("No DATABASE_URL provided, running without DB persistence")
	}

	// Initialize Engine
	exchangeEngine := engine.NewEngine(database, wsManager)

	// API Routes
	http.HandleFunc("/order", corsMiddleware(exchangeEngine.HandleOrder))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsManager.ServeWS(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// Simple CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
