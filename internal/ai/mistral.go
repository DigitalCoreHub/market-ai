package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/1batu/market-ai/internal/models"
)

// MistralClient uses Mistral's OpenAI-compatible endpoint
type MistralClient struct {
	client *openai.Client
	model  string
}

func NewMistralClient(apiKey, model string) *MistralClient {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = "https://api.mistral.ai/v1"
	return &MistralClient{client: openai.NewClientWithConfig(cfg), model: model}
}

func (c *MistralClient) GetTradingDecision(ctx context.Context, prompt string) (*models.AIDecision, error) {
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
		return nil, fmt.Errorf("mistral request failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from mistral")
	}
	var decision models.AIDecision
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse mistral response: %w", err)
	}
	return &decision, nil
}

func (c *MistralClient) GetModelName() string { return c.model }
