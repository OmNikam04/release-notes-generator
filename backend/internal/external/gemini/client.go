package gemini

import (
	"context"
	"fmt"
	"strings"
	"time"

	genai "google.golang.org/genai"
)

// Client wraps the Google Gemini API client
type Client struct {
	client    *genai.Client
	config    *Config
	projectID string
	location  string
	model     string
}

// NewClient creates a new Gemini client
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("GCP_PROJECT_ID is required")
	}
	if cfg.Location == "" {
		return nil, fmt.Errorf("GCP_LOCATION is required")
	}
	if cfg.Model == "" {
		cfg.Model = "gemini-2.5-pro" // Default model
	}

	// Initialize Gemini client with Vertex AI
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  cfg.ProjectID,
		Location: cfg.Location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client:    client,
		config:    cfg,
		projectID: cfg.ProjectID,
		location:  cfg.Location,
		model:     cfg.Model,
	}, nil
}

// Close closes the Gemini client
func (c *Client) Close() error {
	// Gemini client doesn't have a Close method in the new SDK
	return nil
}

// GenerateContent generates content using Gemini
func (c *Client) GenerateContent(ctx context.Context, prompt string) (string, error) {
	// Set timeout for API call
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Create content parts
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: prompt},
			},
			Role: "user",
		},
	}

	// Configure generation parameters
	config := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.7)), // Balanced creativity
		MaxOutputTokens: 4096,                    // Increased to allow complete JSON response with all fields
		TopP:            genai.Ptr(float32(0.95)),
		TopK:            genai.Ptr(float32(40)),
	}

	// Generate content with retry logic
	var response *genai.GenerateContentResponse
	var err error

	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		response, err = c.client.Models.GenerateContent(ctx, c.model, contents, config)
		if err == nil {
			break
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			return "", fmt.Errorf("non-retryable error from Gemini API: %w", err)
		}

		// Exponential backoff
		if attempt < maxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate content after %d attempts: %w", maxRetries, err)
	}

	// Extract text from response
	text := response.Text()
	if text == "" {
		return "", fmt.Errorf("empty response from Gemini API")
	}

	return text, nil
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retry on rate limits, timeouts, and temporary failures
	retryableErrors := []string{
		"rate limit",
		"quota",
		"timeout",
		"deadline exceeded",
		"unavailable",
		"internal error",
		"503",
		"429",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return false
}
