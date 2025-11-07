package models

import (
	"time"

	"github.com/google/uuid"
)

// AIDecision represents a decision made by AI
type AIDecision struct {
	Action           string         `json:"action"`
	StockSymbol      string         `json:"stock_symbol"`
	Quantity         int            `json:"quantity"`
	TargetPrice      float64        `json:"target_price"`
	StopLoss         float64        `json:"stop_loss"`
	ReasoningSummary string         `json:"reasoning_summary"`
	ReasoningFull    string         `json:"reasoning_full"`
	Confidence       float64        `json:"confidence"`
	RiskLevel        string         `json:"risk_level"`
	ThinkingSteps    []ThinkingStep `json:"thinking_steps"`
}

// ThinkingStep represents a step in the AI's reasoning process
type ThinkingStep struct {
	Step        string `json:"step"`
	Observation string `json:"observation"`
}

// AgentDecision represents a stored agent decision
type AgentDecision struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	AgentID          uuid.UUID  `json:"agent_id" db:"agent_id"`
	StockSymbol      *string    `json:"stock_symbol" db:"stock_symbol"`
	Decision         string     `json:"decision" db:"decision"`
	Quantity         *int       `json:"quantity" db:"quantity"`
	TargetPrice      *float64   `json:"target_price" db:"target_price"`
	StopLoss         *float64   `json:"stop_loss" db:"stop_loss"`
	ReasoningFull    string     `json:"reasoning_full" db:"reasoning_full"`
	ReasoningSummary string     `json:"reasoning_summary" db:"reasoning_summary"`
	ConfidenceScore  float64    `json:"confidence_score" db:"confidence_score"`
	RiskScore        float64    `json:"risk_score" db:"risk_score"`
	RiskLevel        string     `json:"risk_level" db:"risk_level"`
	MarketContext    string     `json:"market_context" db:"market_context"`
	Executed         bool       `json:"executed" db:"executed"`
	TradeID          *uuid.UUID `json:"trade_id" db:"trade_id"`
	Outcome          string     `json:"outcome" db:"outcome"`
	ActualProfitLoss *float64   `json:"actual_profit_loss" db:"actual_profit_loss"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// AgentThought represents a thinking step stored in database
type AgentThought struct {
	ID         uuid.UUID `json:"id" db:"id"`
	AgentID    uuid.UUID `json:"agent_id" db:"agent_id"`
	DecisionID uuid.UUID `json:"decision_id" db:"decision_id"`
	StepNumber int       `json:"step_number" db:"step_number"`
	StepName   string    `json:"step_name" db:"step_name"`
	Thought    string    `json:"thought" db:"thought"`
	Data       string    `json:"data" db:"data"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AgentStrategy represents an agent's trading strategy
type AgentStrategy struct {
	ID           uuid.UUID `json:"id" db:"id"`
	AgentID      uuid.UUID `json:"agent_id" db:"agent_id"`
	StrategyType string    `json:"strategy_type" db:"strategy_type"`
	Description  string    `json:"description" db:"description"`
	Parameters   string    `json:"parameters" db:"parameters"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
