package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/1batu/market-ai/internal/models"
)

// AnthropicClient implements the Client interface for Anthropic models
type AnthropicClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// AnthropicMessageRequest is the request format for Anthropic API
type AnthropicMessageRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system"`
	Messages  []interface{} `json:"messages"`
}

// AnthropicMessage is a single message in the conversation
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicMessageResponse is the response from Anthropic API
type AnthropicMessageResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30,
		},
	}
}

// GetTradingDecision gets a trading decision from Anthropic
func (c *AnthropicClient) GetTradingDecision(ctx context.Context, prompt string) (*models.AIDecision, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("anthropic API key not configured")
	}

	// Build request
	reqBody := AnthropicMessageRequest{
		Model:     c.model,
		MaxTokens: 1500,
		System:    GetSystemPrompt(),
		Messages: []interface{}{
			AnthropicMessage{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp AnthropicMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("no response from anthropic")
	}

	var decision models.AIDecision
	if err := json.Unmarshal([]byte(apiResp.Content[0].Text), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse decision response: %w", err)
	}

	return &decision, nil
}

// GetModelName returns the model name
func (c *AnthropicClient) GetModelName() string {
	return c.model
}
