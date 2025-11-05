package models

import (
	"time"

	"github.com/google/uuid"
)

type Portfolio struct {
	ID                uuid.UUID `json:"id" db:"id"`
	AgentID           uuid.UUID `json:"agent_id" db:"agent_id"`
	StockSymbol       string    `json:"stock_symbol" db:"stock_symbol"`
	Quantity          int       `json:"quantity" db:"quantity"`
	AvgBuyPrice       float64   `json:"avg_buy_price" db:"avg_buy_price"`
	TotalInvested     float64   `json:"total_invested" db:"total_invested"`
	CurrentValue      float64   `json:"current_value" db:"current_value"`
	ProfitLoss        float64   `json:"profit_loss" db:"profit_loss"`
	ProfitLossPercent float64   `json:"profit_loss_percent" db:"profit_loss_percent"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type PortfolioSummary struct {
	AgentID           uuid.UUID   `json:"agent_id"`
	AgentName         string      `json:"agent_name"`
	CurrentBalance    float64     `json:"current_balance"`
	PortfolioValue    float64     `json:"portfolio_value"`
	TotalValue        float64     `json:"total_value"`
	TotalProfitLoss   float64     `json:"total_profit_loss"`
	ProfitLossPercent float64     `json:"profit_loss_percent"`
	Holdings          []Portfolio `json:"holdings"`
}
