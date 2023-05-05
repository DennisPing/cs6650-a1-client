package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/DennisPing/cs6650-a1-client/data"
	"github.com/DennisPing/cs6650-a1-client/log"
	"github.com/DennisPing/cs6650-a1-client/models"
)

// An api client that has a random number generator
type ApiClient struct {
	ServerUrl    string
	HttpClient   *http.Client
	Rng          *rand.Rand
	SuccessCount uint64
	ErrorCount   uint64
}

func NewApiClient(transport *http.Transport, serverUrl string) *ApiClient {
	return &ApiClient{
		ServerUrl:  serverUrl,
		HttpClient: &http.Client{Transport: transport},
		Rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// POST /swipe/{leftorright}/
func (client *ApiClient) SwipeLeftOrRight(ctx context.Context, direction string) {
	swipeRequest := models.SwipeRequest{
		Swiper:  strconv.Itoa(data.RandInt(client.Rng, 1, 5000)),
		Swipee:  strconv.Itoa(data.RandInt(client.Rng, 1, 1_000_000)),
		Comment: data.RandComment(client.Rng, 256),
	}
	swipeEndpoint := fmt.Sprintf("%s/swipe/%s/", client.ServerUrl, direction)

	req, err := client.createRequest(ctx, http.MethodPost, swipeEndpoint, swipeRequest)
	if err != nil {
		log.Logger.Error().Msg(err.Error())
		return
	}

	resp, err := client.sendRequest(req, 5)
	if err != nil {
		client.ErrorCount += 1
		log.Logger.Error().Msgf("max retries hit: %v", err)
		return
	}
	defer resp.Body.Close()

	// StatusCode should be 200 or 201, else log warn
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		client.SuccessCount += 1
		log.Logger.Debug().Msg(resp.Status)
	} else {
		client.ErrorCount += 1
		log.Logger.Warn().Msg(resp.Status)
	}
}

// Create HTTP request with a timeout context
func (client *ApiClient) createRequest(ctx context.Context, method, url string, data interface{}) (*http.Request, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(body)

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	return req, nil
}

// Send HTTP request with retry limit
func (client *ApiClient) sendRequest(req *http.Request, maxRetries int) (*http.Response, error) {
	baseBackoff := 100 * time.Millisecond

	var resp *http.Response
	var err error
	for i := 1; i <= maxRetries; i++ {
		resp, err = client.HttpClient.Do(req)
		if err == nil {
			break // Successful API call
		}
		// Exponential backoff with jitter
		backoffDuration := time.Duration(math.Pow(2, float64(i))) * baseBackoff
		sleepDuration := backoffDuration + time.Duration(client.Rng.Int63n(1000))*time.Millisecond
		time.Sleep(sleepDuration)
	}
	return resp, err
}
