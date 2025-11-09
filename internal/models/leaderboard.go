package models

import (
	"time"

	"github.com/google/uuid"
)

// LeaderboardEntry represents a row in the live leaderboard
type LeaderboardEntry struct {
	Rank           int       `json:"rank"`
	AgentID        uuid.UUID `json:"agent_id" db:"agent_id"`
	AgentName      string    `json:"agent_name" db:"agent_name"`
	Model          string    `json:"model" db:"model"`
	ROI            float64   `json:"roi" db:"roi"`
	ProfitLoss     float64   `json:"profit_loss" db:"profit_loss"`
	WinRate        float64   `json:"win_rate" db:"win_rate"`
	TotalTrades    int       `json:"total_trades" db:"total_trades"`
	Balance        float64   `json:"balance" db:"balance"`
	PortfolioValue float64   `json:"portfolio_value" db:"portfolio_value"`
	TotalValue     float64   `json:"total_value"`
	Badges         []string  `json:"badges" db:"badges"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
