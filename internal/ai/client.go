package ai

import (
	"context"

	"github.com/1batu/market-ai/internal/models"
)

// Client defines the interface for AI trading decision makers
type Client interface {
	// GetTradingDecision asks AI to make a trading decision
	GetTradingDecision(ctx context.Context, prompt string) (*models.AIDecision, error)

	// GetModelName returns the model name
	GetModelName() string
}

// DecisionRequest contains all data needed for an AI trading decision
type DecisionRequest struct {
	AgentID        string
	AgentName      string
	CurrentBalance float64
	Portfolio      []models.Portfolio
	Stocks         []models.Stock
	MarketData     []models.MarketData
	RecentTrades   []models.Trade
	Strategy       string
	News           []models.NewsArticle // Latest news articles
	NewsCount      int                  // Number of news articles
}
