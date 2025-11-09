package models

import (
	"time"

	"github.com/google/uuid"
)

type AgentPerformanceSnapshot struct {
	ID              uuid.UUID `json:"id" db:"id"`
	AgentID         uuid.UUID `json:"agent_id" db:"agent_id"`
	Balance         float64   `json:"balance" db:"balance"`
	PortfolioValue  float64   `json:"portfolio_value" db:"portfolio_value"`
	TotalValue      float64   `json:"total_value" db:"total_value"`
	TotalProfitLoss float64   `json:"total_profit_loss" db:"total_profit_loss"`
	ROIPercent      float64   `json:"roi_percent" db:"roi_percent"`
	TotalTrades     int       `json:"total_trades" db:"total_trades"`
	WinningTrades   int       `json:"winning_trades" db:"winning_trades"`
	LosingTrades    int       `json:"losing_trades" db:"losing_trades"`
	WinRate         float64   `json:"win_rate" db:"win_rate"`
	MaxDrawdown     float64   `json:"max_drawdown" db:"max_drawdown"`
	SharpeRatio     float64   `json:"sharpe_ratio" db:"sharpe_ratio"`
	SnapshotTime    time.Time `json:"snapshot_time" db:"snapshot_time"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type AgentDailyStats struct {
	ID              uuid.UUID `json:"id" db:"id"`
	AgentID         uuid.UUID `json:"agent_id" db:"agent_id"`
	StatDate        time.Time `json:"stat_date" db:"stat_date"`
	TradesCount     int       `json:"trades_count" db:"trades_count"`
	Wins            int       `json:"wins" db:"wins"`
	Losses          int       `json:"losses" db:"losses"`
	ProfitLoss      float64   `json:"profit_loss" db:"profit_loss"`
	VolumeTraded    float64   `json:"volume_traded" db:"volume_traded"`
	BestTradeProfit float64   `json:"best_trade_profit" db:"best_trade_profit"`
	WorstTradeLoss  float64   `json:"worst_trade_loss" db:"worst_trade_loss"`
	AvgConfidence   float64   `json:"avg_confidence" db:"avg_confidence"`
	AvgExecTimeMS   int       `json:"avg_execution_time_ms" db:"avg_execution_time_ms"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type AgentMatchup struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Agent1ID          uuid.UUID  `json:"agent1_id" db:"agent1_id"`
	Agent2ID          uuid.UUID  `json:"agent2_id" db:"agent2_id"`
	Agent1Wins        int        `json:"agent1_wins" db:"agent1_wins"`
	Agent2Wins        int        `json:"agent2_wins" db:"agent2_wins"`
	Draws             int        `json:"draws" db:"draws"`
	Agent1TotalProfit float64    `json:"agent1_total_profit" db:"agent1_total_profit"`
	Agent2TotalProfit float64    `json:"agent2_total_profit" db:"agent2_total_profit"`
	LastWinner        *uuid.UUID `json:"last_winner" db:"last_winner"`
	LastMatchupTime   *time.Time `json:"last_matchup_time" db:"last_matchup_time"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}
