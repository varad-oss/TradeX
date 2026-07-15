package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	NumTraders  = 20
	OrdersPer   = 2500 // Total = 50,000 orders
	APIEndpoint = "http://localhost:8080/order"
)

type OrderRequest struct {
	TraderID string `json:"traderId"`
	Side     string `json:"side"`
	Price    int64  `json:"price"`
	Quantity int64  `json:"quantity"`
}

func main() {
	fmt.Printf("Starting simulation: %d traders, %d orders each (Total: %d)\n", NumTraders, OrdersPer, NumTraders*OrdersPer)
	
	var wg sync.WaitGroup
	start := time.Now()

	var totalLatency time.Duration
	var latenciesMu sync.Mutex
	var successCount int

	for i := 0; i < NumTraders; i++ {
		wg.Add(1)
		go func(traderIdx int) {
			defer wg.Done()
			traderID := fmt.Sprintf("trader_%d", traderIdx)
			client := &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
				},
			}

			var localLatency time.Duration
			var localSuccess int

			for j := 0; j < OrdersPer; j++ {
				side := "BUY"
				if rand.Intn(2) == 0 {
					side = "SELL"
				}
				
				// Generate prices around 10000 (e.g. $100.00)
				price := int64(10000 + (rand.Intn(200) - 100))
				qty := int64(rand.Intn(100) + 1)

				reqBody := OrderRequest{
					TraderID: traderID,
					Side:     side,
					Price:    price,
					Quantity: qty,
				}
				
				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest("POST", APIEndpoint, bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				reqStart := time.Now()
				resp, err := client.Do(req)
				reqDuration := time.Since(reqStart)

				if err == nil {
					resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						localSuccess++
						localLatency += reqDuration
					}
				}
			}

			latenciesMu.Lock()
			totalLatency += localLatency
			successCount += localSuccess
			latenciesMu.Unlock()
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Println("--- Simulation Complete ---")
	fmt.Printf("Total time: %v\n", duration)
	fmt.Printf("Successful orders: %d\n", successCount)
	
	if successCount > 0 {
		avgLatency := totalLatency / time.Duration(successCount)
		throughput := float64(successCount) / duration.Seconds()
		fmt.Printf("Average Latency: %v\n", avgLatency)
		fmt.Printf("Throughput: %.2f orders/sec\n", throughput)
		
		if throughput >= 1500 {
			fmt.Println("✅ PASS: Sustained >1.5K orders/sec")
		} else {
			fmt.Println("❌ FAIL: Did not meet throughput metric")
		}
		if avgLatency < time.Millisecond {
			fmt.Println("✅ PASS: Sub-ms matching latency")
		} else {
			fmt.Println("❌ FAIL: Latency too high")
		}
	}
}
