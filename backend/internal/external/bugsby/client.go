package bugsby

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/omnikam04/release-notes-generator/internal/logger"
)

const (
	defaultBaseURL    = "https://bugs-service.infra.corp.arista.io"
	defaultAPIVersion = "v3"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 3
	maxResponseSize   = 5 * 1024 * 1024 // 5MB
)

// Retryable HTTP status codes
var retryableStatusCodes = map[int]bool{
	http.StatusTooManyRequests:     true, // 429
	http.StatusInternalServerError: true, // 500
	http.StatusBadGateway:          true, // 502
	http.StatusServiceUnavailable:  true, // 503
	http.StatusGatewayTimeout:      true, // 504
}

// Client defines the interface for Bugsby API operations
type Client interface {
	// Generic HTTP methods - support ALL Bugsby operations
	Get(ctx context.Context, endpoint string, params map[string]string) (*http.Response, error)
	Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
	Put(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
	Patch(ctx context.Context, endpoint string, body interface{}) (*http.Response, error)
	Delete(ctx context.Context, endpoint string) (*http.Response, error)

	// Convenience methods for common operations
	Query(ctx context.Context, query string, limit int) (*BugsbyResponse, error)
	GetBugByID(ctx context.Context, bugID int) (*BugsbyBug, error)
	GetBugsByRelease(ctx context.Context, release string, filters *BugFilters) (*BugsbyResponse, error)
}

// client is the concrete implementation of Client
type client struct {
	baseURL       string
	apiVersion    string
	tokenProvider *TokenProvider
	httpClient    *http.Client
	maxRetries    int
}

// Config holds configuration for creating a Bugsby client
type Config struct {
	BaseURL    string
	APIVersion string
	TokenFile  string
	Timeout    time.Duration
	MaxRetries int
}

// NewClient creates a new Bugsby API client
func NewClient(cfg *Config) (Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	// Set defaults
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = defaultAPIVersion
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = defaultMaxRetries
	}

	tokenProvider := NewTokenProvider(cfg.TokenFile)

	// Validate token is available
	token, err := tokenProvider.GetToken()
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to load Bugsby auth token - client will operate without authentication")
	} else {
		logger.Info().Int("token_length", len(token)).Msg("Bugsby auth token loaded successfully")
	}

	return &client{
		baseURL:       cfg.BaseURL,
		apiVersion:    cfg.APIVersion,
		tokenProvider: tokenProvider,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		maxRetries: cfg.MaxRetries,
	}, nil
}

// buildURL constructs the full URL for an endpoint
func (c *client) buildURL(endpoint string) string {
	// If endpoint already contains the base URL, return as-is
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}

	// Remove leading slash from endpoint if present
	endpoint = strings.TrimPrefix(endpoint, "/")

	// If endpoint doesn't start with version, add it
	if !strings.HasPrefix(endpoint, c.apiVersion) {
		endpoint = c.apiVersion + "/" + endpoint
	}

	return c.baseURL + "/" + endpoint
}

// buildHeaders constructs HTTP headers including authentication
func (c *client) buildHeaders(customHeaders map[string]string) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	// Add authentication token
	token, err := c.tokenProvider.GetToken()
	if err == nil && token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	// Merge custom headers
	for key, value := range customHeaders {
		headers[key] = value
	}

	return headers
}

// addQueryParams adds query parameters to a URL
func addQueryParams(baseURL string, params map[string]string) string {
	if len(params) == 0 {
		return baseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// doRequestWithRetry performs an HTTP request with retry logic
func (c *client) doRequestWithRetry(ctx context.Context, method, url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	var lastErr error
	backoffs := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		// Create request
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// Execute request
		logger.Debug().
			Str("method", method).
			Str("url", url).
			Int("attempt", attempt+1).
			Msg("Making Bugsby API request")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			logger.Warn().
				Err(err).
				Int("attempt", attempt+1).
				Dur("retry_after", backoffs[attempt]).
				Msg("Request failed - will retry")

			if attempt < c.maxRetries-1 {
				time.Sleep(backoffs[attempt])
			}
			continue
		}

		// Check if status code is retryable
		if retryableStatusCodes[resp.StatusCode] {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			_ = resp.Body.Close()

			lastErr = fmt.Errorf("received retryable status %d: %s", resp.StatusCode, string(bodyBytes))
			logger.Warn().
				Int("status_code", resp.StatusCode).
				Int("attempt", attempt+1).
				Dur("retry_after", backoffs[attempt]).
				Msg("Received retryable status - will retry")

			if attempt < c.maxRetries-1 {
				time.Sleep(backoffs[attempt])
			}
			continue
		}

		// Success or non-retryable error
		logger.Debug().
			Int("status_code", resp.StatusCode).
			Str("method", method).
			Str("url", url).
			Msg("Bugsby API request completed")

		return resp, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("request failed after %d retries", c.maxRetries)
	}

	return nil, lastErr
}

// Get performs a GET request
func (c *client) Get(ctx context.Context, endpoint string, params map[string]string) (*http.Response, error) {
	url := c.buildURL(endpoint)
	url = addQueryParams(url, params)
	headers := c.buildHeaders(nil)

	return c.doRequestWithRetry(ctx, "GET", url, headers, nil)
}

// Post performs a POST request
func (c *client) Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	url := c.buildURL(endpoint)
	headers := c.buildHeaders(nil)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	return c.doRequestWithRetry(ctx, "POST", url, headers, bodyReader)
}

// Put performs a PUT request
func (c *client) Put(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	url := c.buildURL(endpoint)
	headers := c.buildHeaders(nil)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	return c.doRequestWithRetry(ctx, "PUT", url, headers, bodyReader)
}

// Patch performs a PATCH request
func (c *client) Patch(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	url := c.buildURL(endpoint)
	headers := c.buildHeaders(nil)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	return c.doRequestWithRetry(ctx, "PATCH", url, headers, bodyReader)
}

// Delete performs a DELETE request
func (c *client) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	url := c.buildURL(endpoint)
	headers := c.buildHeaders(nil)

	return c.doRequestWithRetry(ctx, "DELETE", url, headers, nil)
}

// parseResponse parses the HTTP response into the target structure
func parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}

// Query performs a generic Bugsby query and returns the response
func (c *client) Query(ctx context.Context, query string, limit int) (*BugsbyResponse, error) {
	if limit <= 0 {
		limit = 100
	}

	params := map[string]string{
		"q":     query,
		"limit": fmt.Sprintf("%d", limit),
	}

	resp, err := c.Get(ctx, "bugs", params)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	var result BugsbyResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBugByID retrieves a single bug by its ID
func (c *client) GetBugByID(ctx context.Context, bugID int) (*BugsbyBug, error) {
	query := fmt.Sprintf("id==%d", bugID)

	resp, err := c.Query(ctx, query, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get bug %d: %w", bugID, err)
	}

	if len(resp.Bugs) == 0 {
		return nil, fmt.Errorf("bug %d not found", bugID)
	}

	return &resp.Bugs[0], nil
}

// GetBugsByRelease retrieves bugs for a specific release with optional filters
func (c *client) GetBugsByRelease(ctx context.Context, release string, filters *BugFilters) (*BugsbyResponse, error) {
	if filters == nil {
		filters = &BugFilters{}
	}
	filters.Release = release

	query := filters.BuildQuery()
	if query == "" {
		return nil, fmt.Errorf("no valid filters provided")
	}

	logger.Info().
		Str("release", release).
		Str("query", query).
		Msg("Fetching bugs from Bugsby")

	return c.Query(ctx, query, 1000) // Default limit for release queries
}
