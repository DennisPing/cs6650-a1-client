package main

import (
	"math/rand"
	"time"

	api "github.com/DennisPing/twinder-sdk-go"
)

// An api client that has a random number generator
type RngClient struct {
	apiClient *api.APIClient
	rng       *rand.Rand
}

func NewRngClient(serverURL string) *RngClient {
	cfg := api.NewConfiguration()
	apiClient := api.NewAPIClient(cfg)
	apiClient.ChangeBasePath(serverURL)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &RngClient{
		apiClient: apiClient,
		rng:       rng,
	}
}
