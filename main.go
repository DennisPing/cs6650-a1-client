package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DennisPing/cs6650-distributed-systems/assignment1/client-single/log"
	api "github.com/DennisPing/twinder-sdk-go"
)

const (
	charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxWorkers  = 10
	numRequests = 1000
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
	// Needed for local testing since the client and the server can't both listen on port 8080
	port := os.Getenv("PORT") // Set the PORT to 8081 for local testing
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

	// Initialize RNG client pool
	RngClientPool := make(chan *RngClient, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		RngClientPool <- NewRngClient(serverURL)
	}

	log.Logger.Info().Msgf("Starting %d requests...", numRequests)
	startTime := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			rngClient := <-RngClientPool // Get an RNG client from the pool
			defer func() {
				RngClientPool <- rngClient // Release the slot so that other goroutines can acquire client
			}()
			direction := randDirection(rngClient.rng)
			swipeLeftOrRight(ctx, rngClient, direction)
		}()
	}
	wg.Wait()
	duration := time.Since(startTime)
	log.Logger.Info().Msgf("Total run time: %v\n", duration)
	log.Logger.Info().Msgf("Success count: %d\n", atomic.LoadUint64(&successCount))
	log.Logger.Info().Msgf("Error count: %d\n", atomic.LoadUint64(&errorCount))
}

func swipeLeftOrRight(ctx context.Context, client *RngClient, direction string) {
	reqBody := api.SwipeDetails{
		Swiper:  strconv.Itoa(randInt(client.rng, 1, 5000)),
		Swipee:  strconv.Itoa(randInt(client.rng, 1, 1_000_000)),
		Comment: randComment(client.rng, 256),
	}

	var err error
	var resp *http.Response
	maxRetries := 5
	baseBackoff := 100 * time.Millisecond
	maxBackoff := 32 * time.Second

	// https://cloud.google.com/iot/docs/how-tos/exponential-backoff
	for i := 0; i < maxRetries; i++ {
		// Send POST request
		resp, err = client.apiClient.SwipeApi.Swipe(ctx, reqBody, direction)
		if err == nil {
			break // Successful API call
		}
		// Exponential backoff with jitter
		backoffDuration := time.Duration(math.Pow(2, float64(i))) * baseBackoff
		if backoffDuration > maxBackoff {
			backoffDuration = maxBackoff
		}
		sleepDuration := backoffDuration + time.Duration(client.rng.Int63n(1000))*time.Millisecond
		time.Sleep(sleepDuration)
	}
	if err != nil {
		atomic.AddUint64(&errorCount, 1)
		log.Logger.Error().Msg(err.Error())
		return
	}

	// StatusCode should be 200 or 201, else log warn
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		atomic.AddUint64(&successCount, 1)
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
