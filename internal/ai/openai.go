package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/1batu/market-ai/internal/models"
	"github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the Client interface for OpenAI models
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

// GetTradingDecision gets a trading decision from OpenAI
func (c *OpenAIClient) GetTradingDecision(ctx context.Context, prompt string) (*models.AIDecision, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: GetSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1500,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	var decision models.AIDecision
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse openai response: %w", err)
	}

	return &decision, nil
}

// GetModelName returns the model name
func (c *OpenAIClient) GetModelName() string {
	return c.model
}
