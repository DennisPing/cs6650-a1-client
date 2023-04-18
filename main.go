package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DennisPing/cs6650-a1-client/models"
	"github.com/DennisPing/cs6650-distributed-systems/assignment1/client-single/log"
)

const (
	charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxWorkers  = 10
	numRequests = 100_000
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
	log.Logger.Info().Msgf("Client PORT: %s", port)

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

	log.Logger.Info().Msgf("Starting %d requests...", numRequests)
	startTime := time.Now()

	// Spawn numWorkers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rngClient := NewRngClient(serverURL)

			// Do tasks until taskQueue is empty
			for range taskQueue {
				direction := randDirection(rngClient.rng)
				swipeLeftOrRight(ctx, rngClient, direction)
			}
		}()
	}
	wg.Wait()
	duration := time.Since(startTime)
	log.Logger.Info().Msgf("Total run time: %v\n", duration)
	log.Logger.Info().Msgf("Success count: %d\n", atomic.LoadUint64(&successCount))
	log.Logger.Info().Msgf("Error count: %d\n", atomic.LoadUint64(&errorCount))
}

func swipeLeftOrRight(ctx context.Context, client *RngClient, direction string) {
	swipeRequest := models.SwipeRequest{
		Swiper:  strconv.Itoa(randInt(client.rng, 1, 5000)),
		Swipee:  strconv.Itoa(randInt(client.rng, 1, 1_000_000)),
		Comment: randComment(client.rng, 256),
	}
	swipeEndpoint := fmt.Sprintf("%s/swipe/%s/", client.serverUrl, direction)

	body, err := json.Marshal(swipeRequest)
	if err != nil {
		log.Logger.Error().Msg(err.Error())
	}
	reader := bytes.NewReader(body)

	ctxWithTimeout, cancelCtx := context.WithTimeout(ctx, 32*time.Second)
	defer cancelCtx()
	req, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodPost, swipeEndpoint, reader)
	if err != nil {
		log.Logger.Error().Msg(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	maxRetries := 5
	baseBackoff := 100 * time.Millisecond

	var resp *http.Response
	for i := 1; i <= maxRetries; i++ {
		// Send POST request
		resp, err = client.httpClient.Do(req)
		if err == nil {
			break // Successful API call
		}
		// Exponential backoff with jitter
		backoffDuration := time.Duration(math.Pow(2, float64(i))) * baseBackoff
		sleepDuration := backoffDuration + time.Duration(client.rng.Int63n(1000))*time.Millisecond
		log.Logger.Info().Msgf("sleeping for %v", sleepDuration)
		time.Sleep(sleepDuration)
	}
	if err != nil {
		atomic.AddUint64(&errorCount, 1)
		log.Logger.Error().Msgf("max retries hit: %s", err.Error())
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

// Each goroutine client should pass in their own RNG
func randInt(rng *rand.Rand, start, stop int) int {
	return rng.Intn(stop) + start
}

// Each goroutine client should pass in their own RNG
func randComment(rng *rand.Rand, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}

// Each goroutine client should pass in their own RNG
func randDirection(rng *rand.Rand) string {
	if rng.Intn(2) == 1 {
		return "right"
	}
	return "left"
}
