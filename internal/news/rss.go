package news

import (
	"context"
	"strings"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
)

// RSSParser fetches news from RSS feeds
type RSSParser struct {
	parser *gofeed.Parser
	feeds  []string
}

// NewRSSParser creates a new RSS parser
func NewRSSParser(feedURLs []string) *RSSParser {
	return &RSSParser{
		parser: gofeed.NewParser(),
		feeds:  feedURLs,
	}
}

// GetEconomyNews fetches news from all configured RSS feeds
func (rp *RSSParser) GetEconomyNews(ctx context.Context) ([]models.NewsArticle, error) {
	var allArticles []models.NewsArticle

	// Fetch from each feed
	for _, feedURL := range rp.feeds {
		articles, err := rp.fetchFromFeed(ctx, feedURL)
		if err != nil {
			log.Error().Err(err).Str("feed", feedURL).Msg("Failed to fetch RSS feed")
			continue // Skip failed feeds, don't stop entire process
		}
		allArticles = append(allArticles, articles...)
	}

	log.Info().Int("count", len(allArticles)).Int("feeds", len(rp.feeds)).Msg("RSS fetch completed")

	return allArticles, nil
}

// fetchFromFeed fetches and parses a single RSS feed
func (rp *RSSParser) fetchFromFeed(ctx context.Context, feedURL string) ([]models.NewsArticle, error) {
	feed, err := rp.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return nil, err
	}

	// Determine source from feed
	source := feed.Title
	if source == "" {
		// Extract from URL (e.g., "bloomberght.com")
		parts := strings.Split(feedURL, "/")
		if len(parts) > 2 {
			source = parts[2]
		}
	}

	// Convert items to articles
	articles := make([]models.NewsArticle, 0, len(feed.Items))
	for _, item := range feed.Items {
		// Parse published time
		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else {
			publishedAt = time.Now()
		}

		// Skip old articles (older than 24 hours)
		if time.Since(publishedAt) > 24*time.Hour {
			continue
		}

		// Skip if no link
		if item.Link == "" {
			continue
		}

		// Extract stock symbols
		relatedStocks := extractStockSymbols(item.Title + " " + item.Description)

		article := models.NewsArticle{
			Title:         item.Title,
			Description:   item.Description,
			Content:       item.Content,
			Source:        source,
			URL:           item.Link,
			EventType:     "news",
			RelatedStocks: relatedStocks,
			PublishedAt:   publishedAt,
			FetchedAt:     time.Now(),
		}

		articles = append(articles, article)
	}

	return articles, nil
}
