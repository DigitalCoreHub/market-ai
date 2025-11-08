package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/1batu/market-ai/internal/models"
)

// GroqClient uses Groq's OpenAI-compatible endpoint for Llama models
type GroqClient struct {
	client *openai.Client
	model  string
}

func NewGroqClient(apiKey, model string) *GroqClient {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = "https://api.groq.com/openai/v1"
	return &GroqClient{client: openai.NewClientWithConfig(cfg), model: model}
}

func (c *GroqClient) GetTradingDecision(ctx context.Context, prompt string) (*models.AIDecision, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: GetSystemPrompt()},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature:    0.7,
		MaxTokens:      1500,
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
	})
	if err != nil {
		return nil, fmt.Errorf("groq request failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from groq")
	}
	var decision models.AIDecision
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse groq response: %w", err)
	}
	return &decision, nil
}

func (c *GroqClient) GetModelName() string { return c.model }
