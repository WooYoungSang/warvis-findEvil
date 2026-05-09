package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for Ollama API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	model      string
}

// NewClient creates a new Ollama client.
// baseURL should be like "http://localhost:11434"
// model should be like "gemma4:26b-a4b-it-q4_K_M"
func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			// 26B-class quantized models can take >2min for first-token on
			// cold load even on a 24 GB VRAM GPU. 600s leaves room for
			// model load + reasoning; outer hunt timeout (state/hunt budgets)
			// still bounds total wall time.
			Timeout: 600 * time.Second,
		},
	}
}

// Chat sends a chat request to the Ollama API and returns the response.
// Implements 3-retry logic for network timeouts.
func (c *Client) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.chatOnce(ctx, messages)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		// Only retry on network/timeout errors, not on malformed responses
		if isRetryableError(err) && attempt < maxRetries-1 {
			// Exponential backoff: 100ms, 200ms, 400ms
			backoff := time.Duration(100*(1<<uint(attempt))) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}
		break
	}

	return nil, fmt.Errorf("ollama chat failed after %d retries: %w", maxRetries, lastErr)
}

// chatOnce makes a single chat request (no retries).
func (c *Client) chatOnce(ctx context.Context, messages []Message) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
		Format:   "json", // Request JSON output from Ollama
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d: %s", httpResp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// ListModels returns the list of available models from Ollama.
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d: %s", httpResp.StatusCode, string(respBody))
	}

	var modelsResp ModelsResponse
	if err := json.Unmarshal(respBody, &modelsResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return modelsResp.Models, nil
}

// CheckModelAvailable checks if a specific model is available on the Ollama server.
func (c *Client) CheckModelAvailable(ctx context.Context) (bool, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return false, err
	}

	for _, m := range models {
		if m.Name == c.model {
			return true, nil
		}
	}
	return false, nil
}

// isRetryableError returns true if the error is a network/timeout error that should be retried.
func isRetryableError(err error) bool {
	// Check for context cancellation/deadline
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false // Don't retry on explicit cancellation
	}

	// Check for timeout errors using net.Error interface
	type timeoutError interface {
		Timeout() bool
	}
	if te, ok := err.(timeoutError); ok && te.Timeout() {
		return true
	}

	// Check for connection errors
	errMsg := err.Error()
	return bytes.Contains([]byte(errMsg), []byte("connection refused")) ||
		bytes.Contains([]byte(errMsg), []byte("connection reset")) ||
		bytes.Contains([]byte(errMsg), []byte("no such host"))
}
