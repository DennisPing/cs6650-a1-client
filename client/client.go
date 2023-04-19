package client

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// An api client that has a random number generator
type ApiClient struct {
	ServerUrl  string
	HttpClient *http.Client
	Rng        *rand.Rand
}

func NewApiClient(serverUrl string) *ApiClient {
	return &ApiClient{
		ServerUrl:  serverUrl,
		HttpClient: &http.Client{},
		Rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (client *ApiClient) CreateRequest(ctx context.Context, method, url string, data interface{}) (*http.Request, error) {
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

func (client *ApiClient) SendRequestWithTimeout(req *http.Request, maxRetries int) (*http.Response, error) {
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
