package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DennisPing/cs6650-a1-client/client"
	"github.com/DennisPing/cs6650-a1-client/data"
	"github.com/DennisPing/cs6650-a1-client/log"
	"github.com/DennisPing/cs6650-a1-client/models"
	"github.com/montanaflynn/stats"
)

const (
	maxWorkers  = 25
	numRequests = 500_000
)

var (
	successCount uint64
	errorCount   uint64
)

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		log.Logger.Fatal().Msg("SERVER_URL env variable not set")
	}

	port := os.Getenv("CLIENT_PORT") // Set the PORT to 8081 for local testing
	if port == "" {
		port = "8080" // For Cloud Run
	}

	// Set up dummy HTTP server to satisfy Cloud Run requirements
	go func() {
		http.HandleFunc("/dummy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		addr := fmt.Sprintf(":%s", port)
		log.Logger.Fatal().Msg(http.ListenAndServe(addr, nil).Error())
	}()

	ctx := context.Background()

	// Populate the task queue with tasks (token)
	taskQueue := make(chan struct{}, numRequests)
	for i := 0; i < numRequests; i++ {
		taskQueue <- struct{}{}
	}
	close(taskQueue) // Close the queue. Nothing is ever being put into the queue.

	var wg sync.WaitGroup

	log.Logger.Info().Msgf("Using %d goroutines", maxWorkers)
	log.Logger.Info().Msgf("Starting %d requests...", numRequests)
	startTime := time.Now()

	responseTimes := make([][]time.Duration, maxWorkers)

	// Spawn numWorkers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			apiClient := client.NewApiClient(serverURL)

			// Do tasks until taskQueue is empty. Then all workers will move on.
			for range taskQueue {
				ctxWithTimeout, cancelCtx := context.WithTimeout(ctx, 32*time.Second)
				direction := data.RandDirection(apiClient.Rng)
				t0 := time.Now()
				swipeLeftOrRight(ctxWithTimeout, apiClient, direction) // The actual HTTP request
				t1 := time.Since(t0)
				cancelCtx()
				// Each worker only appends to their own slice. Thread safe.
				responseTimes[workerId] = append(responseTimes[workerId], t1)
			}
		}(i)
	}
	wg.Wait()

	// Calculate metrics
	duration := time.Since(startTime)
	success := atomic.LoadUint64(&successCount)
	errors := atomic.LoadUint64(&errorCount)
	throughput := float64(success) / duration.Seconds()

	log.Logger.Info().Msgf("Success count: %d", success)
	log.Logger.Info().Msgf("Error count: %d", errors)
	log.Logger.Info().Msgf("Total run time: %v", duration)
	log.Logger.Info().Msgf("Throughput: %.2f req/sec", throughput)

	allResponseTimes := make([]float64, 0, numRequests)
	for _, slice := range responseTimes { // Convert all time.Duration to float64
		for _, rt := range slice {
			rtFloat := float64(rt.Milliseconds())
			allResponseTimes = append(allResponseTimes, rtFloat)
		}
	}
	mean, _ := stats.Mean(allResponseTimes)
	median, _ := stats.Median(allResponseTimes)
	p99, _ := stats.Percentile(allResponseTimes, 99)
	min, _ := stats.Min(allResponseTimes)
	max, _ := stats.Max(allResponseTimes)

	log.Logger.Info().Msgf("Mean response time: %.2f ms", mean)
	log.Logger.Info().Msgf("Median response time: %.2f ms", median)
	log.Logger.Info().Msgf("P99 response time: %.2f ms", p99)
	log.Logger.Info().Msgf("Min response time: %.2f ms", min)
	log.Logger.Info().Msgf("Max response time: %.2f ms", max)
}

func swipeLeftOrRight(ctx context.Context, client *client.ApiClient, direction string) {
	swipeRequest := models.SwipeRequest{
		Swiper:  strconv.Itoa(data.RandInt(client.Rng, 1, 5000)),
		Swipee:  strconv.Itoa(data.RandInt(client.Rng, 1, 1_000_000)),
		Comment: data.RandComment(client.Rng, 256),
	}
	swipeEndpoint := fmt.Sprintf("%s/swipe/%s/", client.ServerUrl, direction)

	req, err := client.CreateRequest(ctx, http.MethodPost, swipeEndpoint, swipeRequest)
	if err != nil {
		log.Logger.Error().Msg(err.Error())
		return
	}

	resp, err := client.SendRequestWithTimeout(req, 5)
	if err != nil {
		atomic.AddUint64(&errorCount, 1)
		log.Logger.Error().Msgf("max retries hit: %v", err)
		return
	}
	defer resp.Body.Close()

	// StatusCode should be 200 or 201, else log warn
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		atomic.AddUint64(&successCount, 1)
		log.Logger.Debug().Msg(resp.Status)
	} else {
		atomic.AddUint64(&errorCount, 1)
		log.Logger.Warn().Msg(resp.Status)
	}
}
