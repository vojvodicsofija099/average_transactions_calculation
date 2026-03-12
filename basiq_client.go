package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	baseURL    = "https://au-api.basiq.io"
	tokenPath  = "/token"
	usersPath  = "/users"
	apiVersion = "2.1"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type TokenCache struct {
	AccessToken string
	ExpiresAt   time.Time
}

var tokenCache TokenCache

type BasiqClient struct {
	BaseURL string
	Client  *http.Client
	Token   string
}

func NewBasiqClient(token string) *BasiqClient {
	return &BasiqClient{
		BaseURL: baseURL,
		Client:  httpClient,
		Token:   token,
	}
}

func getAccessToken(apiKey string) (*TokenResponse, error) {
	if tokenCache.AccessToken != "" && time.Now().Before(tokenCache.ExpiresAt) {
		return &TokenResponse{
			AccessToken: tokenCache.AccessToken,
			TokenType:   "Bearer",
			ExpiresIn:   int(time.Until(tokenCache.ExpiresAt).Seconds()),
		}, nil
	}

	form := url.Values{}
	form.Set("scope", "SERVER_ACCESS")

	tokenURL := baseURL + tokenPath

	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, fmt.Errorf("request creation unsuccessful: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("basiq-version", apiVersion)

	var tokenResp TokenResponse
	if err := doJSONRequest(httpClient, req, &tokenResp); err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}

	// safety buffer of 30 seconds
	tokenCache = TokenCache{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn-30) * time.Second),
	}

	return &tokenResp, nil
}

func (c *BasiqClient) createUser(request CreateUserRequest) (*UserResponse, error) {
	usersURL := c.BaseURL + usersPath

	req, err := c.newJSONRequest(http.MethodPost, usersURL, request)
	if err != nil {
		return nil, fmt.Errorf("cannot create user request: %w", err)
	}

	var userResp UserResponse
	if err := doJSONRequest(c.Client, req, &userResp); err != nil {
		return nil, fmt.Errorf("create user request failed: %w", err)
	}

	return &userResp, nil
}

func (c *BasiqClient) createConnection(userID string, request CreateConnectionRequest) (*ConnectionResponse, error) {
	connectionURL := fmt.Sprintf("%s/users/%s/connections", c.BaseURL, userID)

	req, err := c.newJSONRequest(http.MethodPost, connectionURL, request)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection request: %w", err)
	}

	var connectionResp ConnectionResponse
	if err := doJSONRequest(c.Client, req, &connectionResp); err != nil {
		return nil, fmt.Errorf("connection request failed: %w", err)
	}

	return &connectionResp, nil
}

func (c *BasiqClient) getJob(jobID string) (*JobResponse, error) {
	jobURL := fmt.Sprintf("%s/jobs/%s", c.BaseURL, jobID)

	req, err := c.newGETRequest(jobURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create job request: %w", err)
	}

	var jobResponse JobResponse
	if err := doJSONRequest(c.Client, req, &jobResponse); err != nil {
		return nil, fmt.Errorf("get job request failed: %w", err)
	}

	return &jobResponse, nil
}

func (c *BasiqClient) waitForJobSuccess(jobID string) (*JobResponse, error) {
	for {
		job, err := c.getJob(jobID)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Job steps: %+v\n", job.Steps)

		if allStepsSuccessful(job.Steps) {
			return job, nil
		}

		if anyStepFailing(job.Steps) {
			return nil, fmt.Errorf("job %s failed", jobID)
		}

		time.Sleep(2 * time.Second)
	}
}

func (c *BasiqClient) getTransactions(userID string) ([]Transaction, error) {
	transactionsURL := fmt.Sprintf("%s/users/%s/transactions", c.BaseURL, userID)

	reqURL, err := url.Parse(transactionsURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse transactions url: %w", err)
	}

	query := reqURL.Query()
	query.Set("filter", "transaction.direction.eq('debit')")
	reqURL.RawQuery = query.Encode()

	var allTransactions []Transaction
	nextURL := reqURL.String()

	for nextURL != "" {
		req, err := c.newGETRequest(nextURL)
		if err != nil {
			return nil, fmt.Errorf("cannot create transactions request: %w", err)
		}

		var transactionsResponse TransactionsResponse
		if err := doJSONRequest(c.Client, req, &transactionsResponse); err != nil {
			return nil, fmt.Errorf("get transactions request failed: %w", err)
		}

		allTransactions = append(allTransactions, transactionsResponse.Data...)
		nextURL = transactionsResponse.Links.Next
	}

	fmt.Println("---------------------------------------------")
	fmt.Println("Total transactions:", len(allTransactions))

	return allTransactions, nil
}

func calculateAveragesPerCategory(transactions []Transaction) map[string]float64 {
	sum := make(map[string]float64)
	count := make(map[string]int)

	for _, t := range transactions {
		if t.SubClass == nil {
			continue
		}

		amount, err := strconv.ParseFloat(t.Amount, 64)
		if err != nil {
			continue
		}

		// defensive check; API is already filtered to debit transactions
		if amount >= 0 {
			continue
		}

		category := t.SubClass.Code
		sum[category] += amount
		count[category]++
	}

	avg := make(map[string]float64)

	for category, total := range sum {
		avg[category] = total / float64(count[category])
	}

	return avg
}

func allStepsSuccessful(steps []JobStep) bool {
	if len(steps) == 0 {
		return false
	}

	for _, step := range steps {
		if step.Status != "success" {
			return false
		}
	}

	return true
}

func anyStepFailing(steps []JobStep) bool {
	for _, step := range steps {
		if step.Status == "failed" {
			return true
		}
	}
	return false
}

func (c *BasiqClient) newJSONRequest(method string, url string, body interface{}) (*http.Request, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cannot serialize request body: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("basiq-version", apiVersion)

	return req, nil
}

func (c *BasiqClient) newGETRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create GET request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("basiq-version", apiVersion)

	return req, nil
}

func doJSONRequest(client *http.Client, req *http.Request, result interface{}) error {
	const maxAttempts = 3

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		currentReq := req

		if attempt > 1 && req.GetBody != nil {
			bodyCopy, err := req.GetBody()
			if err != nil {
				return fmt.Errorf("cannot recreate request body: %w", err)
			}

			clonedReq := req.Clone(req.Context())
			clonedReq.Body = bodyCopy
			currentReq = clonedReq
		}

		resp, err := client.Do(currentReq)
		if err != nil {
			lastErr = fmt.Errorf("request sending unsuccessful: %w", err)

			if attempt < maxAttempts && isRetryableMethod(req) && shouldRetry(0, err) {
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				continue
			}

			return lastErr
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("response body read unsuccessful: %w", readErr)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("unexpected response status code: %d: %s", resp.StatusCode, string(body))

			if attempt < maxAttempts && isRetryableMethod(req) && shouldRetry(resp.StatusCode, nil) {
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				continue
			}

			return lastErr
		}

		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("response parsing unsuccessful: %w", err)
		}

		return nil
	}

	return lastErr
}

func shouldRetry(statusCode int, err error) bool {
	if err != nil {
		return true
	}

	if statusCode == http.StatusTooManyRequests {
		return true
	}

	if statusCode >= 500 && statusCode <= 599 {
		return true
	}

	return false
}

func isRetryableMethod(req *http.Request) bool {
	return req.Method == http.MethodGet
}
