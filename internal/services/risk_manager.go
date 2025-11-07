package services

import (
	"context"
	"fmt"

	"github.com/1batu/market-ai/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RiskManager validates trades based on risk rules
type RiskManager struct {
	db                 *pgxpool.Pool
	maxRiskPerTrade    float64 // percentage
	maxPortfolioRisk   float64 // percentage
	minConfidenceScore float64
}

// NewRiskManager creates a new risk manager
func NewRiskManager(db *pgxpool.Pool, maxRiskPerTrade, maxPortfolioRisk, minConfidence float64) *RiskManager {
	return &RiskManager{
		db:                 db,
		maxRiskPerTrade:    maxRiskPerTrade,  // 5.0 for 5%
		maxPortfolioRisk:   maxPortfolioRisk, // 20.0 for 20%
		minConfidenceScore: minConfidence,    // 70.0 for 70%
	}
}

// ValidateTrade validates a trading decision against risk rules
func (rm *RiskManager) ValidateTrade(ctx context.Context, agentID uuid.UUID, decision *models.AIDecision) error {
	// Check confidence
	if decision.Confidence < rm.minConfidenceScore {
		return fmt.Errorf("confidence too low: %.1f%% < %.1f%%", decision.Confidence, rm.minConfidenceScore)
	}

	// Get agent balance
	var balance float64
	err := rm.db.QueryRow(ctx, "SELECT current_balance FROM agents WHERE id = $1", agentID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("failed to get agent balance: %w", err)
	}

	// Get stock price
	var stockPrice float64
	err = rm.db.QueryRow(ctx, "SELECT current_price FROM stocks WHERE symbol = $1", decision.StockSymbol).Scan(&stockPrice)
	if err != nil {
		return fmt.Errorf("stock not found: %w", err)
	}

	tradeAmount := float64(decision.Quantity) * stockPrice

	// Check trade size
	if decision.Action == "BUY" {
		maxTradeAmount := balance * (rm.maxRiskPerTrade / 100)
		if tradeAmount > maxTradeAmount {
			return fmt.Errorf("trade amount %.2f TL exceeds max %.2f TL (%.1f%% of balance)",
				tradeAmount, maxTradeAmount, rm.maxRiskPerTrade)
		}

		// Check if sufficient balance
		if tradeAmount > balance {
			return fmt.Errorf("insufficient balance: %.2f TL < %.2f TL", balance, tradeAmount)
		}
	}

	// Check portfolio concentration
	portfolioValue, err := rm.getPortfolioValue(ctx, agentID)
	if err != nil {
		return err
	}

	totalValue := balance + portfolioValue
	if decision.Action == "BUY" {
		newPortfolioValue := portfolioValue + tradeAmount
		portfolioRisk := (newPortfolioValue / totalValue) * 100
		if portfolioRisk > rm.maxPortfolioRisk {
			return fmt.Errorf("portfolio risk %.1f%% exceeds max %.1f%%", portfolioRisk, rm.maxPortfolioRisk)
		}
	}

	return nil
}

// getPortfolioValue calculates total portfolio value for an agent
func (rm *RiskManager) getPortfolioValue(ctx context.Context, agentID uuid.UUID) (float64, error) {
	var value float64
	query := `
		SELECT COALESCE(SUM(p.quantity * s.current_price), 0)
		FROM portfolio p
		JOIN stocks s ON p.stock_symbol = s.symbol
		WHERE p.agent_id = $1
	`
	err := rm.db.QueryRow(ctx, query, agentID).Scan(&value)
	return value, err
}
