package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	newsLatestKey = "news:latest"
	newsHashKey   = "news:hash"
)

// NewsCache handles Redis caching for news
type NewsCache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewNewsCache creates a new news cache
func NewNewsCache(redis *redis.Client, ttl time.Duration) *NewsCache {
	return &NewsCache{
		redis: redis,
		ttl:   ttl,
	}
}

// GetLatestNews retrieves cached news articles
func (nc *NewsCache) GetLatestNews(ctx context.Context) ([]models.NewsArticle, error) {
	val, err := nc.redis.Get(ctx, newsLatestKey).Result()
	if err == redis.Nil {
		return nil, ErrCacheNotFound
	}
	if err != nil {
		return nil, err
	}

	var articles []models.NewsArticle
	if err := json.Unmarshal([]byte(val), &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

// SetLatestNews caches news articles
func (nc *NewsCache) SetLatestNews(ctx context.Context, articles []models.NewsArticle) error {
	data, err := json.Marshal(articles)
	if err != nil {
		return err
	}

	return nc.redis.Set(ctx, newsLatestKey, data, nc.ttl).Err()
}

// GetNewsHash retrieves the current news hash (for change detection)
func (nc *NewsCache) GetNewsHash(ctx context.Context) (string, error) {
	return nc.redis.Get(ctx, newsHashKey).Result()
}

// SetNewsHash stores the news hash
func (nc *NewsCache) SetNewsHash(ctx context.Context, hash string) error {
	return nc.redis.Set(ctx, newsHashKey, hash, nc.ttl).Err()
}

// Clear removes all news cache
func (nc *NewsCache) Clear(ctx context.Context) error {
	return nc.redis.Del(ctx, newsLatestKey, newsHashKey).Err()
}

var ErrCacheNotFound = redis.Nil
