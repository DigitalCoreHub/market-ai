package models

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Model          string    `json:"model" db:"model"`
	Status         string    `json:"status" db:"status"`
	InitialBalance float64   `json:"initial_balance" db:"initial_balance"`
	CurrentBalance float64   `json:"current_balance" db:"current_balance"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type AgentMetrics struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	AgentID             uuid.UUID `json:"agent_id" db:"agent_id"`
	TotalTrades         int       `json:"total_trades" db:"total_trades"`
	WinningTrades       int       `json:"winning_trades" db:"winning_trades"`
	LosingTrades        int       `json:"losing_trades" db:"losing_trades"`
	TotalProfitLoss     float64   `json:"total_profit_loss" db:"total_profit_loss"`
	TotalPortfolioValue float64   `json:"total_portfolio_value" db:"total_portfolio_value"`
	WinRate             float64   `json:"win_rate" db:"win_rate"`
	ROI                 float64   `json:"roi" db:"roi"`
	SharpeRatio         float64   `json:"sharpe_ratio" db:"sharpe_ratio"`
	MaxDrawdown         float64   `json:"max_drawdown" db:"max_drawdown"`
	CalculatedAt        time.Time `json:"calculated_at" db:"calculated_at"`
}

type AgentWithMetrics struct {
	Agent   Agent        `json:"agent"`
	Metrics AgentMetrics `json:"metrics"`
}
