package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/1batu/market-ai/internal/models"
)

// GoogleClient implements Gemini model access
type GoogleClient struct {
	client *genai.Client
	model  string
}

// NewGoogleClient creates a Gemini client
func NewGoogleClient(apiKey, model string) (*GoogleClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("google gemini API key not configured")
	}
	ctx := context.Background()
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}
	return &GoogleClient{client: c, model: model}, nil
}

// GetTradingDecision queries Gemini for a decision
func (gc *GoogleClient) GetTradingDecision(ctx context.Context, prompt string) (*models.AIDecision, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	model := gc.client.GenerativeModel(gc.model)
	model.SetTemperature(0.7)
	model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(GetSystemPrompt())}}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty gemini response")
	}
	raw := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	var decision models.AIDecision
	if err := json.Unmarshal([]byte(raw), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse gemini JSON: %w", err)
	}
	return &decision, nil
}

// GetModelName returns model id
func (gc *GoogleClient) GetModelName() string { return gc.model }

// Close shuts down client
func (gc *GoogleClient) Close() error { return gc.client.Close() }
