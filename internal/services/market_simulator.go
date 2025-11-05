package services

import (
	"context"
	"math/rand"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/1batu/market-ai/internal/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type MarketSimulator struct {
	db  *pgxpool.Pool
	hub *websocket.Hub
}

func NewMarketSimulator(db *pgxpool.Pool, hub *websocket.Hub) *MarketSimulator {
	return &MarketSimulator{
		db:  db,
		hub: hub,
	}
}

func (ms *MarketSimulator) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Info().Msg("Market simulator started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Market simulator stopped")
			return
		case <-ticker.C:
			ms.updatePrices()
		}
	}
}

func (ms *MarketSimulator) updatePrices() {
	query := `SELECT symbol, current_price FROM stocks`
	rows, err := ms.db.Query(context.Background(), query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch stocks")
		return
	}
	defer rows.Close()

	var updates []models.Stock

	for rows.Next() {
		var symbol string
		var currentPrice float64

		if err := rows.Scan(&symbol, &currentPrice); err != nil {
			log.Error().Err(err).Msg("Failed to scan stock")
			continue
		}

		changePercent := (rand.Float64() - 0.5) * 4
		newPrice := currentPrice * (1 + changePercent/100)

		updateQuery := `
			UPDATE stocks
			SET current_price = $1,
			    change_percent = $2,
			    last_updated = NOW()
			WHERE symbol = $3
		`
		_, err := ms.db.Exec(context.Background(), updateQuery, newPrice, changePercent, symbol)
		if err != nil {
			log.Error().Err(err).Str("symbol", symbol).Msg("Failed to update price")
			continue
		}

		updates = append(updates, models.Stock{
			Symbol:        symbol,
			CurrentPrice:  newPrice,
			ChangePercent: changePercent,
		})
	}

	if len(updates) > 0 {
		ms.hub.BroadcastMessage("price_update", updates)
		log.Debug().Int("count", len(updates)).Msg("Prices updated")
	}
}
