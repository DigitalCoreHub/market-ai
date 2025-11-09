package services

import (
	"context"
	"fmt"

	"github.com/1batu/market-ai/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RiskManager risk kurallarına göre işlemleri doğrular
type RiskManager struct {
	db                 *pgxpool.Pool
	maxRiskPerTrade    float64 // yüzde
	maxPortfolioRisk   float64 // yüzde
	minConfidenceScore float64
}

// NewRiskManager yeni bir risk yöneticisi oluşturur
func NewRiskManager(db *pgxpool.Pool, maxRiskPerTrade, maxPortfolioRisk, minConfidence float64) *RiskManager {
	return &RiskManager{
		db:                 db,
		maxRiskPerTrade:    maxRiskPerTrade,  // 5.0 %5 için
		maxPortfolioRisk:   maxPortfolioRisk, // 20.0 %20 için
		minConfidenceScore: minConfidence,    // 70.0 %70 için
	}
}

// ValidateTrade bir alım-satım kararını risk kurallarına göre doğrular
func (rm *RiskManager) ValidateTrade(ctx context.Context, agentID uuid.UUID, decision *models.AIDecision) error {
	// Miktar kontrolü
	if decision.Quantity <= 0 {
		return fmt.Errorf("invalid quantity: %d (must be > 0)", decision.Quantity)
	}
	// Güven kontrolü
	if decision.Confidence < rm.minConfidenceScore {
		return fmt.Errorf("confidence too low: %.1f%% < %.1f%%", decision.Confidence, rm.minConfidenceScore)
	}

	// Ajan bakiyesini al
	var balance float64
	err := rm.db.QueryRow(ctx, "SELECT current_balance FROM agents WHERE id = $1", agentID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("failed to get agent balance: %w", err)
	}

	// Hisse fiyatını al
	var stockPrice float64
	err = rm.db.QueryRow(ctx, "SELECT current_price FROM stocks WHERE symbol = $1", decision.StockSymbol).Scan(&stockPrice)
	if err != nil {
		return fmt.Errorf("stock not found: %w", err)
	}

	tradeAmount := float64(decision.Quantity) * stockPrice

	// İşlem büyüklüğünü kontrol et
	if decision.Action == "BUY" {
		maxTradeAmount := balance * (rm.maxRiskPerTrade / 100)
		if tradeAmount > maxTradeAmount {
			return fmt.Errorf("trade amount %.2f TL exceeds max %.2f TL (%.1f%% of balance)",
				tradeAmount, maxTradeAmount, rm.maxRiskPerTrade)
		}

		// Yeterli bakiye olup olmadığını kontrol et (komisyon dahil)
		commission := tradeAmount * 0.001 // %0.1 komisyon varsayımı
		totalCost := tradeAmount + commission
		if totalCost > balance {
			return fmt.Errorf("insufficient balance: %.2f TL < %.2f TL (trade: %.2f + commission: %.2f) - reduce to %d lots",
				balance, totalCost, tradeAmount, commission, int(balance/stockPrice))
		}
	}
	// SELL action validation can be added here if needed
	// For now, SELL actions are allowed without additional validation

	// Portföy yoğunluğunu kontrol et
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
	// SELL action portfolio risk validation can be added here if needed
	// For now, SELL actions are allowed without additional portfolio risk validation

	return nil
}

// getPortfolioValue bir ajan için toplam portföy değerini hesaplar
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
