package services

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/1batu/market-ai/internal/cache"
	"github.com/1batu/market-ai/internal/models"
	"github.com/1batu/market-ai/internal/news"
	"github.com/1batu/market-ai/internal/websocket"
)

// NewsAggregator fetches and caches news from multiple sources
type NewsAggregator struct {
	db        *pgxpool.Pool
	redis     *redis.Client
	hub       *websocket.Hub
	newsCache *cache.NewsCache
	newsAPI   *news.NewsAPIClient
	rssParser *news.RSSParser
	interval  time.Duration
	cacheTTL  time.Duration
}

// NewNewsAggregator creates a new news aggregator service
func NewNewsAggregator(
	db *pgxpool.Pool,
	redis *redis.Client,
	hub *websocket.Hub,
	newsAPIKey string,
	rssFeeds []string,
	interval time.Duration,
	cacheTTL time.Duration,
) *NewsAggregator {
	newsCache := cache.NewNewsCache(redis, cacheTTL)
	newsAPIClient := news.NewNewsAPIClient(newsAPIKey)
	rssParser := news.NewRSSParser(rssFeeds)

	return &NewsAggregator{
		db:        db,
		redis:     redis,
		hub:       hub,
		newsCache: newsCache,
		newsAPI:   newsAPIClient,
		rssParser: rssParser,
		interval:  interval,
		cacheTTL:  cacheTTL,
	}
}

// Start begins the news aggregation loop
func (na *NewsAggregator) Start(ctx context.Context) {
	log.Info().
		Dur("interval", na.interval).
		Msg("News aggregator started")

	// First fetch immediately
	na.fetchAndStore(ctx)

	// Then fetch at intervals
	ticker := time.NewTicker(na.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("News aggregator stopped")
			return
		case <-ticker.C:
			na.fetchAndStore(ctx)
		}
	}
}

// fetchAndStore fetches news from all sources and stores in DB/cache
func (na *NewsAggregator) fetchAndStore(ctx context.Context) {
	start := time.Now()
	log.Debug().Msg("Fetching news from all sources")

	var allArticles []models.NewsArticle

	// Fetch from News API
	if newsAPIArticles, err := na.newsAPI.GetTurkeyFinanceNews(ctx); err == nil {
		allArticles = append(allArticles, newsAPIArticles...)
		log.Debug().Int("count", len(newsAPIArticles)).Msg("Fetched from News API")
	} else {
		log.Warn().Err(err).Msg("Failed to fetch from News API")
	}

	// Fetch from RSS feeds
	if rssArticles, err := na.rssParser.GetEconomyNews(ctx); err == nil {
		allArticles = append(allArticles, rssArticles...)
		log.Debug().Int("count", len(rssArticles)).Msg("Fetched from RSS feeds")
	} else {
		log.Warn().Err(err).Msg("Failed to fetch from RSS feeds")
	}

	// Deduplicate by URL
	uniqueArticles := na.deduplicateByURL(allArticles)
	log.Info().Int("total", len(allArticles)).Int("unique", len(uniqueArticles)).Msg("News fetched")

	// Store in database
	storedCount := 0
	for _, article := range uniqueArticles {
		if err := na.storeArticle(ctx, article); err != nil {
			log.Warn().Err(err).Str("title", article.Title).Msg("Failed to store article")
			continue
		}
		storedCount++
	}

	log.Info().Int("stored", storedCount).Msg("Articles stored in database")

	// Cache in Redis
	if err := na.newsCache.SetLatestNews(ctx, uniqueArticles); err != nil {
		log.Error().Err(err).Msg("Failed to cache news in Redis")
	}

	// Broadcast to WebSocket clients
	na.hub.BroadcastMessage("news_update", map[string]interface{}{
		"count":     len(uniqueArticles),
		"articles":  uniqueArticles[:min(len(uniqueArticles), 5)], // Top 5
		"timestamp": time.Now().Unix(),
	})

	// Log to database
	executionTime := time.Since(start).Milliseconds()
	na.logFetch("News API + RSS", len(allArticles), storedCount, true, "", int(executionTime))

	log.Info().
		Int("fetched", len(allArticles)).
		Int("stored", storedCount).
		Dur("duration", time.Since(start)).
		Msg("News aggregation cycle completed")
}

// storeArticle stores a single article in the database
func (na *NewsAggregator) storeArticle(ctx context.Context, article models.NewsArticle) error {
	query := `
		INSERT INTO market_events (
			title, description, content, source, url,
			event_type, related_stocks, published_at, fetched_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT DO NOTHING
	`

	_, err := na.db.Exec(ctx, query,
		article.Title,
		article.Description,
		article.Content,
		article.Source,
		article.URL,
		article.EventType,
		article.RelatedStocks,
		article.PublishedAt,
	)

	return err
}

// deduplicateByURL removes duplicate articles by URL
func (na *NewsAggregator) deduplicateByURL(articles []models.NewsArticle) []models.NewsArticle {
	seen := make(map[string]bool)
	unique := make([]models.NewsArticle, 0, len(articles))

	for _, article := range articles {
		if article.URL != "" && !seen[article.URL] {
			seen[article.URL] = true
			unique = append(unique, article)
		}
	}

	return unique
}

// logFetch logs a news fetch operation
func (na *NewsAggregator) logFetch(source string, fetched, stored int, success bool, errMsg string, execTimeMs int) {
	query := `
		INSERT INTO news_fetch_log (source, articles_fetched, articles_stored, success, error_message, execution_time_ms)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := na.db.Exec(context.Background(), query, source, fetched, stored, success, errMsg, execTimeMs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to log fetch operation")
	}
}

// GetLatestNews retrieves the latest cached news
func (na *NewsAggregator) GetLatestNews(ctx context.Context) ([]models.NewsArticle, error) {
	// Try cache first
	cached, err := na.newsCache.GetLatestNews(ctx)
	if err == nil {
		return cached, nil
	}

	// Fallback to database
	query := `
		SELECT id, title, description, source, url, related_stocks, published_at, fetched_at, created_at
		FROM market_events
		WHERE published_at > NOW() - INTERVAL '3 hours'
		ORDER BY published_at DESC
		LIMIT 20
	`

	rows, err := na.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.NewsArticle
	for rows.Next() {
		var article models.MarketEvent
		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Description,
			&article.Source,
			&article.URL,
			&article.RelatedStocks,
			&article.PublishedAt,
			&article.FetchedAt,
			&article.CreatedAt,
		); err != nil {
			continue
		}

		articles = append(articles, models.NewsArticle{
			Title:         article.Title,
			Description:   article.Description,
			Source:        article.Source,
			URL:           article.URL,
			RelatedStocks: article.RelatedStocks,
			PublishedAt:   article.PublishedAt,
			FetchedAt:     article.FetchedAt,
		})
	}

	return articles, nil
}

// min is a helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
