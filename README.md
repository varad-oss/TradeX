# Real-Time Equity Trading Exchange

A highly-performant real-time equity trading exchange matching engine and visualizer built to handle high-throughput concurrent order matching.

## Architecture

This project is a microservices stack containing:
1. **High-Performance Go Matching Engine (`/backend`)**: An $O(\log n)$ dual-heap matching structure guaranteeing strict price-time priority execution. 
2. **Next.js Real-time Terminal (`/frontend`)**: A visually engaging web terminal built with Next.js, TailwindCSS, and Framer Motion that consumes real-time market data over WebSockets.
3. **PostgreSQL Database**: A relational database to asynchronously persist executed trades and ensure ACID compliance without blocking the main matching thread.

## Quick Start (Docker)

To run the entire exchange locally:

```bash
docker-compose up --build -d
```

- The **Next.js Frontend** will be available at `http://localhost:3000`
- The **Go Backend** WebSocket & HTTP API will run on `http://localhost:8080`
- The **PostgreSQL Database** is exposed locally on port `5433`

## Running the High-Throughput Simulation (Proving 1.5K+ Orders/sec)

To test the system's performance limits and mimic heavy load, you can run the built-in Golang simulation script. This script will spin up concurrent simulated traders bombarding the exchange with buy/sell orders.

1. Ensure your backend is running (`docker-compose up -d`)
2. Go into the backend directory:
   ```bash
   cd backend
   ```
3. Run the simulator:
   ```bash
   go run cmd/simulator/main.go
   ```
4. Check your frontend terminal (`http://localhost:3000`) and watch the orders matching in real time!
