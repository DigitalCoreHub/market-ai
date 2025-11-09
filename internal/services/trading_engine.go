package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/1batu/market-ai/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TradingEngine struct {
	db *pgxpool.Pool
}

func NewTradingEngine(db *pgxpool.Pool) *TradingEngine {
	return &TradingEngine{db: db}
}

const CommissionRate = 0.001

func (te *TradingEngine) ExecuteTrade(ctx context.Context, req models.TradeRequest) (*models.Trade, error) {
	tx, err := te.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			// Ignore rollback error if transaction was committed
		}
	}()

	var stockPrice float64
	err = tx.QueryRow(ctx, "SELECT current_price FROM stocks WHERE symbol = $1", req.StockSymbol).Scan(&stockPrice)
	if err != nil {
		return nil, fmt.Errorf("stock not found: %w", err)
	}

	var agentBalance float64
	err = tx.QueryRow(ctx, "SELECT current_balance FROM agents WHERE id = $1", req.AgentID).Scan(&agentBalance)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	totalAmount := float64(req.Quantity) * stockPrice
	commission := totalAmount * CommissionRate

	if req.TradeType == "BUY" {
		if agentBalance < totalAmount+commission {
			return nil, errors.New("insufficient balance")
		}

		_, err = tx.Exec(ctx,
			"UPDATE agents SET current_balance = current_balance - $1 WHERE id = $2",
			totalAmount+commission, req.AgentID)
		if err != nil {
			return nil, fmt.Errorf("failed to update balance: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO portfolio (agent_id, stock_symbol, quantity, avg_buy_price, total_invested)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (agent_id, stock_symbol)
			DO UPDATE SET
				quantity = portfolio.quantity + EXCLUDED.quantity,
				avg_buy_price = (portfolio.total_invested + EXCLUDED.total_invested) / (portfolio.quantity + EXCLUDED.quantity),
				total_invested = portfolio.total_invested + EXCLUDED.total_invested,
				updated_at = NOW()
		`, req.AgentID, req.StockSymbol, req.Quantity, stockPrice, totalAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to update portfolio: %w", err)
		}
	} else if req.TradeType == "SELL" {
		var currentQuantity int
		err = tx.QueryRow(ctx,
			"SELECT quantity FROM portfolio WHERE agent_id = $1 AND stock_symbol = $2",
			req.AgentID, req.StockSymbol).Scan(&currentQuantity)
		if err != nil || currentQuantity < req.Quantity {
			return nil, errors.New("insufficient stocks")
		}

		_, err = tx.Exec(ctx,
			"UPDATE agents SET current_balance = current_balance + $1 WHERE id = $2",
			totalAmount-commission, req.AgentID)
		if err != nil {
			return nil, fmt.Errorf("failed to update balance: %w", err)
		}

		if currentQuantity == req.Quantity {
			_, err = tx.Exec(ctx,
				"DELETE FROM portfolio WHERE agent_id = $1 AND stock_symbol = $2",
				req.AgentID, req.StockSymbol)
		} else {
			_, err = tx.Exec(ctx, `
				UPDATE portfolio
				SET quantity = quantity - $1,
				    total_invested = total_invested * (quantity - $1) / quantity,
				    updated_at = NOW()
				WHERE agent_id = $2 AND stock_symbol = $3
			`, req.Quantity, req.AgentID, req.StockSymbol)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update portfolio: %w", err)
		}
	}

	trade := &models.Trade{
		ID:          uuid.New(),
		AgentID:     req.AgentID,
		StockSymbol: req.StockSymbol,
		TradeType:   req.TradeType,
		Quantity:    req.Quantity,
		Price:       stockPrice,
		TotalAmount: totalAmount,
		Commission:  commission,
		Reasoning:   req.Reasoning,
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO trades (id, agent_id, stock_symbol, trade_type, quantity, price, total_amount, commission, reasoning)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, trade.ID, trade.AgentID, trade.StockSymbol, trade.TradeType, trade.Quantity,
		trade.Price, trade.TotalAmount, trade.Commission, trade.Reasoning)
	if err != nil {
		return nil, fmt.Errorf("failed to insert trade: %w", err)
	}

	_, err = tx.Exec(ctx, "SELECT update_agent_metrics($1)", req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to update metrics: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return trade, nil
}
