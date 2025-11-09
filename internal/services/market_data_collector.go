package services

import (
	"context"
	"time"

	"github.com/1batu/market-ai/internal/datasources/fusion"
	"github.com/rs/zerolog/log"
)

// MarketDataCollector periodically pulls multi-source market context
type MarketDataCollector struct {
	fusion    *fusion.Service
	symbols   []string
	yahooInt  time.Duration
	scrapeInt time.Duration
	twInt     time.Duration
	stopChan  chan struct{}
}

func NewMarketDataCollector(f *fusion.Service, symbols []string, yahooSec, scrapeSec, twitterSec int) *MarketDataCollector {
	if yahooSec <= 0 {
		yahooSec = 300
	}
	if scrapeSec <= 0 {
		scrapeSec = 120
	}
	if twitterSec <= 0 {
		twitterSec = 60
	}
	return &MarketDataCollector{
		fusion:    f,
		symbols:   symbols,
		yahooInt:  time.Duration(yahooSec) * time.Second,
		scrapeInt: time.Duration(scrapeSec) * time.Second,
		twInt:     time.Duration(twitterSec) * time.Second,
		stopChan:  make(chan struct{}),
	}
}

func (m *MarketDataCollector) Start(ctx context.Context) {
	log.Info().Msg("MarketDataCollector started")
	yahooTicker := time.NewTicker(m.yahooInt)
	scrapeTicker := time.NewTicker(m.scrapeInt)
	twitterTicker := time.NewTicker(m.twInt)
	defer yahooTicker.Stop()
	defer scrapeTicker.Stop()
	defer twitterTicker.Stop()

	// To avoid triple fetch overload, track last fetch results; full context can be built each cycle union
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("MarketDataCollector stopped (context)")
			return
		case <-m.stopChan:
			log.Info().Msg("MarketDataCollector stopped (signal)")
			return
		case <-yahooTicker.C:
			m.fetch(ctx, "yahoo")
		case <-scrapeTicker.C:
			m.fetch(ctx, "scraper")
		case <-twitterTicker.C:
			m.fetch(ctx, "twitter")
		}
	}
}

func (m *MarketDataCollector) fetch(ctx context.Context, cause string) {
	// We call fusion service for unified context each time; it internally fetches all sources
	// For cost control, could optimize later to partial refresh. Simplicity first.
	if m.fusion == nil {
		return
	}
	ctxOut, err := m.fusion.MarketContext(ctx, m.symbols)
	if err != nil {
		log.Error().Err(err).Str("source", cause).Msg("fusion context error")
		return
	}
	log.Debug().Str("trigger", cause).Int("prices", len(ctxOut.Prices)).Int("tweets", len(ctxOut.Tweets)).Int("news", len(ctxOut.News)).Msg("Market context updated")
}

func (m *MarketDataCollector) Stop() { close(m.stopChan) }
