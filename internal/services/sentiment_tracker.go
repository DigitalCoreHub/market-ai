package services

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// SentimentTracker periodically updates aggregated sentiment windows using DB function update_sentiment_aggregate
type SentimentTracker struct {
	db        *pgxpool.Pool
	interval  time.Duration
	symbols   []string
	windowMin int
}

func NewSentimentTracker(db *pgxpool.Pool, symbols []string, intervalSec, windowMinutes int) *SentimentTracker {
	if intervalSec <= 0 {
		intervalSec = 300
	}
	if windowMinutes <= 0 {
		windowMinutes = 60
	}
	return &SentimentTracker{db: db, interval: time.Duration(intervalSec) * time.Second, symbols: symbols, windowMin: windowMinutes}
}

func (s *SentimentTracker) Start(ctx context.Context) {
	log.Info().Msg("SentimentTracker started")
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("SentimentTracker stopped (context)")
			return
		case <-ticker.C:
			s.updateAll(ctx)
		}
	}
}

func (s *SentimentTracker) updateAll(ctx context.Context) {
	for _, sym := range s.symbols {
		_, err := s.db.Exec(ctx, "SELECT update_sentiment_aggregate($1, $2)", sym, s.windowMin)
		if err != nil {
			log.Error().Err(err).Str("symbol", sym).Msg("update_sentiment_aggregate failed")
		}
	}
	log.Debug().Int("count", len(s.symbols)).Msg("Sentiment aggregates updated")
}
