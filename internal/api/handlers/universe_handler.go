package handlers

import (
	"context"
	"time"

	"github.com/1batu/market-ai/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UniverseHandler struct {
	db      *pgxpool.Pool
	service *services.StockUniverseService
}

func NewUniverseHandler(db *pgxpool.Pool, svc *services.StockUniverseService) *UniverseHandler {
	return &UniverseHandler{db: db, service: svc}
}

// GetActiveStocks returns current active stock universe
func (h *UniverseHandler) GetActiveStocks(c *fiber.Ctx) error {
	query := `
		SELECT symbol, name, current_price, market_cap,
		       discovery_source, mention_count_7d, twitter_count_24h
		FROM stocks
		WHERE is_active = TRUE
		ORDER BY COALESCE(mention_count_7d,0) DESC, COALESCE(market_cap,0) DESC`
	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	defer rows.Close()

	var stocks []map[string]interface{}
	for rows.Next() {
		var (
			symbol       string
			name         string
			price        float64
			marketCap    int64
			source       string
			mentionCount int
			twitterCount int
		)
		if err := rows.Scan(&symbol, &name, &price, &marketCap, &source, &mentionCount, &twitterCount); err != nil {
			continue
		}
		stocks = append(stocks, map[string]interface{}{
			"symbol":            symbol,
			"name":              name,
			"current_price":     price,
			"market_cap":        marketCap,
			"discovery_source":  source,
			"mention_count_7d":  mentionCount,
			"twitter_count_24h": twitterCount,
		})
	}

	// Stats
	var stats struct {
		Total       int
		FromNews    int
		FromTwitter int
		FromTrades  int
	}
	if err := h.db.QueryRow(context.Background(), `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE mention_count_7d > 0) as from_news,
			COUNT(*) FILTER (WHERE twitter_count_24h > 0) as from_twitter,
			COUNT(*) FILTER (WHERE last_trade_date > NOW() - INTERVAL '7 days') as from_trades
		FROM stocks
		WHERE is_active = TRUE
	`).Scan(&stats.Total, &stats.FromNews, &stats.FromTwitter, &stats.FromTrades); err != nil {
		// Use zero values if query fails
		stats.Total = 0
		stats.FromNews = 0
		stats.FromTwitter = 0
		stats.FromTrades = 0
	}

	return c.JSON(fiber.Map{
		"success": true,
		"stocks":  stocks,
		"stats": fiber.Map{
			"total":        stats.Total,
			"from_news":    stats.FromNews,
			"from_twitter": stats.FromTwitter,
			"from_trades":  stats.FromTrades,
		},
	})
}

// TriggerUniverseUpdate manually triggers update
func (h *UniverseHandler) TriggerUniverseUpdate(c *fiber.Ctx) error {
	go func() {
		if err := h.service.UpdateUniverse(context.Background()); err != nil {
			// Log error but don't block response - error is handled by service
			_ = err
		}
	}()
	return c.JSON(fiber.Map{"success": true, "message": "Universe update triggered"})
}

// GetUniverseHistory returns last universe updates
func (h *UniverseHandler) GetUniverseHistory(c *fiber.Ctx) error {
	query := `
		SELECT total_stocks, from_trades, from_news, from_twitter,
		       added_stocks, removed_stocks, update_reason, created_at
		FROM stock_universe_log
		ORDER BY created_at DESC
		LIMIT 20`
	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var (
			totalStocks int
			fromTrades  int
			fromNews    int
			fromTwitter int
			added       []string
			removed     []string
			reason      string
			createdAt   time.Time
		)
		if err := rows.Scan(&totalStocks, &fromTrades, &fromNews, &fromTwitter, &added, &removed, &reason, &createdAt); err != nil {
			continue
		}
		history = append(history, map[string]interface{}{
			"total_stocks": totalStocks,
			"from_trades":  fromTrades,
			"from_news":    fromNews,
			"from_twitter": fromTwitter,
			"added":        added,
			"removed":      removed,
			"reason":       reason,
			"timestamp":    createdAt,
		})
	}

	return c.JSON(fiber.Map{"success": true, "history": history})
}
