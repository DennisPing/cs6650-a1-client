package main

import (
	"math/rand"
	"net/http"
	"time"
)

// An api client that has a random number generator
type RngClient struct {
	serverUrl  string
	httpClient *http.Client
	rng        *rand.Rand
}

func NewRngClient(serverUrl string) *RngClient {
	return &RngClient{
		serverUrl:  serverUrl,
		httpClient: &http.Client{},
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
