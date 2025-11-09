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

// NewsAggregator birden fazla kaynaktan haber getirir ve önbelleğe alır
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

// NewNewsAggregator yeni bir haber toplayıcı servisi oluşturur
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

// Start haber toplama döngüsünü başlatır
func (na *NewsAggregator) Start(ctx context.Context) {
	log.Info().
		Dur("interval", na.interval).
		Msg("News aggregator started")

	// Önce hemen getir
	na.fetchAndStore(ctx)

	// Sonra aralıklarla getir
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

// fetchAndStore tüm kaynaklardan haber getirir ve veritabanına/önbelleğe kaydeder
func (na *NewsAggregator) fetchAndStore(ctx context.Context) {
	start := time.Now()
	log.Debug().Msg("Fetching news from all sources")

	var allArticles []models.NewsArticle

	// News API'den getir
	if newsAPIArticles, err := na.newsAPI.GetTurkeyFinanceNews(ctx); err == nil {
		allArticles = append(allArticles, newsAPIArticles...)
		log.Debug().Int("count", len(newsAPIArticles)).Msg("Fetched from News API")
	} else {
		log.Warn().Err(err).Msg("Failed to fetch from News API")
	}

	// RSS beslemelerinden getir
	if rssArticles, err := na.rssParser.GetEconomyNews(ctx); err == nil {
		allArticles = append(allArticles, rssArticles...)
		log.Debug().Int("count", len(rssArticles)).Msg("Fetched from RSS feeds")
	} else {
		log.Warn().Err(err).Msg("Failed to fetch from RSS feeds")
	}

	// URL'ye göre tekrarları kaldır
	uniqueArticles := na.deduplicateByURL(allArticles)
	log.Info().Int("total", len(allArticles)).Int("unique", len(uniqueArticles)).Msg("News fetched")

	// Veritabanına kaydet
	storedCount := 0
	for _, article := range uniqueArticles {
		if err := na.storeArticle(ctx, article); err != nil {
			log.Warn().Err(err).Str("title", article.Title).Msg("Failed to store article")
			continue
		}
		storedCount++
	}

	log.Info().Int("stored", storedCount).Msg("Articles stored in database")

	// Redis'e önbelleğe al
	if err := na.newsCache.SetLatestNews(ctx, uniqueArticles); err != nil {
		log.Error().Err(err).Msg("Failed to cache news in Redis")
	}

	// WebSocket istemcilerine yayınla
	na.hub.BroadcastMessage("news_update", map[string]interface{}{
		"count":     len(uniqueArticles),
		"articles":  uniqueArticles[:min(len(uniqueArticles), 5)], // İlk 5
		"timestamp": time.Now().Unix(),
	})

	// Veritabanına kaydet
	executionTime := time.Since(start).Milliseconds()
	na.logFetch("News API + RSS", len(allArticles), storedCount, true, "", int(executionTime))

	log.Info().
		Int("fetched", len(allArticles)).
		Int("stored", storedCount).
		Dur("duration", time.Since(start)).
		Msg("News aggregation cycle completed")
}

// storeArticle tek bir makaleyi veritabanına kaydeder
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

// deduplicateByURL URL'ye göre tekrar eden makaleleri kaldırır
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

// logFetch bir haber getirme işlemini kaydeder
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

// GetLatestNews en son önbelleğe alınmış haberleri getirir
func (na *NewsAggregator) GetLatestNews(ctx context.Context) ([]models.NewsArticle, error) {
	// Önce önbelleği dene
	cached, err := na.newsCache.GetLatestNews(ctx)
	if err == nil {
		return cached, nil
	}

	// Veritabanına geri dön
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

// min iki tam sayının en küçüğünü almak için yardımcı fonksiyon
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
