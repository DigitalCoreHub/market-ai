package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/1batu/market-ai/internal/websocket"
)

// StockUniverseService manages the dynamic stock universe
type StockUniverseService struct {
	db              *pgxpool.Pool
	hub             *websocket.Hub
	updateInterval  time.Duration
	minMentions     int // Minimum news mentions in 7d
	minTwitterCount int // Minimum tweets in 24h
}

// NewStockUniverseService creates a new stock universe service
func NewStockUniverseService(db *pgxpool.Pool, hub *websocket.Hub, updateInterval time.Duration) *StockUniverseService {
	return &StockUniverseService{
		db:              db,
		hub:             hub,
		updateInterval:  updateInterval,
		minMentions:     3,
		minTwitterCount: 5,
	}
}

// Start begins the stock universe update loop
func (sus *StockUniverseService) Start(ctx context.Context) {
	// Update immediately on start
	if err := sus.UpdateUniverse(ctx); err != nil {
		log.Warn().Err(err).Msg("initial universe update failed")
	}

	ticker := time.NewTicker(sus.updateInterval)
	defer ticker.Stop()
	log.Info().Dur("interval", sus.updateInterval).Msg("Stock universe service started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stock universe service stopped")
			return
		case <-ticker.C:
			if err := sus.UpdateUniverse(ctx); err != nil {
				log.Error().Err(err).Msg("universe update failed")
			}
		}
	}
}

// UpdateUniverse updates the active stock universe from multiple sources
func (sus *StockUniverseService) UpdateUniverse(ctx context.Context) error {
	log.Info().Msg("Updating stock universe (autonomous)")
	start := time.Now()

	before, err := sus.getActiveStocks(ctx)
	if err != nil {
		return err
	}

	news, err := sus.discoverFromNews(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("discoverFromNews failed")
	}
	twitter, err := sus.discoverFromTwitter(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("discoverFromTwitter failed")
	}
	traded, err := sus.getRecentlyTradedStocks(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("getRecentlyTradedStocks failed")
	}

	merged := sus.mergeStockLists(news, twitter, traded)
	added, err := sus.updateStockActivity(ctx, merged)
	if err != nil {
		return err
	}

	if _, err := sus.db.Exec(ctx, "SELECT update_stock_activity()"); err != nil {
		log.Warn().Err(err).Msg("update_stock_activity db func error")
	}

	after, _ := sus.getActiveStocks(ctx)
	sus.logUniverseUpdate(ctx, before, after, news, twitter, traded)

	sus.hub.BroadcastMessage("universe_updated", map[string]interface{}{
		"total_active":    len(after),
		"from_news":       len(news),
		"from_twitter":    len(twitter),
		"from_trades":     len(traded),
		"newly_added":     added,
		"update_duration": time.Since(start).Seconds(),
		"autonomous":      true,
	})

	log.Info().Int("active", len(after)).Int("added", added).Dur("dur", time.Since(start)).Msg("universe updated")
	return nil
}

func (sus *StockUniverseService) getActiveStocks(ctx context.Context) ([]string, error) {
	rows, err := sus.db.Query(ctx, `SELECT symbol FROM stocks WHERE is_active = TRUE ORDER BY symbol`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil {
			out = append(out, s)
		}
	}
	return out, nil
}

func (sus *StockUniverseService) discoverFromNews(ctx context.Context) ([]string, error) {
	q := `
		SELECT DISTINCT UNNEST(related_stocks) AS symbol, COUNT(*) as mention_count
		FROM market_events
		WHERE published_at > NOW() - INTERVAL '7 days'
		  AND related_stocks IS NOT NULL
		GROUP BY symbol
		HAVING COUNT(*) >= $1
		ORDER BY mention_count DESC`
	rows, err := sus.db.Query(ctx, q, sus.minMentions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []string
	for rows.Next() {
		var sym string
		var cnt int
		if err := rows.Scan(&sym, &cnt); err == nil {
			stocks = append(stocks, sym)
			log.Debug().Str("symbol", sym).Int("mentions", cnt).Msg("news discovered")
		}
	}
	return stocks, nil
}

func (sus *StockUniverseService) discoverFromTwitter(ctx context.Context) ([]string, error) {
	q := `
		SELECT primary_symbol, COUNT(*) as tweet_count
		FROM twitter_sentiment
		WHERE created_at > NOW() - INTERVAL '24 hours'
		  AND primary_symbol IS NOT NULL
		GROUP BY primary_symbol
		HAVING COUNT(*) >= $1
		ORDER BY tweet_count DESC
		LIMIT 30`
	rows, err := sus.db.Query(ctx, q, sus.minTwitterCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []string
	for rows.Next() {
		var sym string
		var cnt int
		if err := rows.Scan(&sym, &cnt); err == nil {
			stocks = append(stocks, sym)
			log.Debug().Str("symbol", sym).Int("tweets", cnt).Msg("twitter discovered")
		}
	}
	return stocks, nil
}

func (sus *StockUniverseService) getRecentlyTradedStocks(ctx context.Context) ([]string, error) {
	q := `
		SELECT DISTINCT stock_symbol, COUNT(*) as trade_count
		FROM trades
		WHERE created_at > NOW() - INTERVAL '7 days'
		GROUP BY stock_symbol
		HAVING COUNT(*) >= 5
		ORDER BY trade_count DESC`
	rows, err := sus.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []string
	for rows.Next() {
		var sym string
		var cnt int
		if err := rows.Scan(&sym, &cnt); err == nil {
			stocks = append(stocks, sym)
			log.Debug().Str("symbol", sym).Int("trades", cnt).Msg("recently traded")
		}
	}
	return stocks, nil
}

func (sus *StockUniverseService) mergeStockLists(lists ...[]string) []string {
	seen := make(map[string]bool)
	var merged []string
	for _, l := range lists {
		for _, s := range l {
			if !seen[s] {
				seen[s] = true
				merged = append(merged, s)
			}
		}
	}
	return merged
}

func (sus *StockUniverseService) updateStockActivity(ctx context.Context, symbols []string) (int, error) {
	if len(symbols) == 0 {
		return 0, nil
	}
	res, err := sus.db.Exec(ctx, `
		UPDATE stocks
		SET is_active = TRUE, discovery_source = COALESCE(discovery_source, 'news')
		WHERE symbol = ANY($1) AND is_active = FALSE
	`, symbols)
	if err != nil {
		return 0, err
	}
	return int(res.RowsAffected()), nil
}

func (sus *StockUniverseService) logUniverseUpdate(
	ctx context.Context,
	before []string,
	after []string,
	news []string,
	twitter []string,
	traded []string,
) {
	beforeMap := map[string]bool{}
	for _, s := range before {
		beforeMap[s] = true
	}
	afterMap := map[string]bool{}
	for _, s := range after {
		afterMap[s] = true
	}

	var added, removed []string
	for _, s := range after {
		if !beforeMap[s] {
			added = append(added, s)
		}
	}
	for _, s := range before {
		if !afterMap[s] {
			removed = append(removed, s)
		}
	}

	reason := "Autonomous data-driven update"
	if len(added) > 0 {
		reason += fmt.Sprintf(" - Added: %d stocks", len(added))
	}
	if len(removed) > 0 {
		reason += fmt.Sprintf(" - Removed: %d stocks", len(removed))
	}

	_, err := sus.db.Exec(ctx, `
		INSERT INTO stock_universe_log (
			total_stocks, from_trades, from_news, from_twitter,
			added_stocks, removed_stocks, update_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, len(after), len(traded), len(news), len(twitter), added, removed, reason)
	if err != nil {
		log.Error().Err(err).Msg("Failed to log universe update")
	}
}
