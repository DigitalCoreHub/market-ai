package scraper

import (
	"context"

	"github.com/1batu/market-ai/internal/models"
)

// WebScraper orchestrates multiple site scrapers
type WebScraper struct {
	bloomberg *BloombergHTScraper
	// investing, kap could be added here
}

func NewWebScraper() *WebScraper {
	return &WebScraper{
		bloomberg: NewBloombergHTScraper(),
	}
}

// ScrapeAll collects latest articles across sources
func (w *WebScraper) ScrapeAll(ctx context.Context) ([]models.ScrapedArticle, error) {
	var out []models.ScrapedArticle
	if w.bloomberg != nil {
		if items, err := w.bloomberg.ScrapeLatestNews(ctx); err == nil {
			out = append(out, items...)
		}
	}
	return out, nil
}
