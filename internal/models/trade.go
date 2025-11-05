package models

import (
	"time"

	"github.com/google/uuid"
)

type Trade struct {
	ID          uuid.UUID `json:"id" db:"id"`
	AgentID     uuid.UUID `json:"agent_id" db:"agent_id"`
	StockSymbol string    `json:"stock_symbol" db:"stock_symbol"`
	TradeType   string    `json:"trade_type" db:"trade_type"`
	Quantity    int       `json:"quantity" db:"quantity"`
	Price       float64   `json:"price" db:"price"`
	TotalAmount float64   `json:"total_amount" db:"total_amount"`
	Commission  float64   `json:"commission" db:"commission"`
	Reasoning   string    `json:"reasoning" db:"reasoning"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type TradeRequest struct {
	AgentID     uuid.UUID `json:"agent_id" validate:"required"`
	StockSymbol string    `json:"stock_symbol" validate:"required"`
	TradeType   string    `json:"trade_type" validate:"required,oneof=BUY SELL"`
	Quantity    int       `json:"quantity" validate:"required,min=1"`
	Reasoning   string    `json:"reasoning"`
}
