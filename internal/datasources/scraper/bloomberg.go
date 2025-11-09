package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/gocolly/colly/v2"
	"github.com/rs/zerolog/log"
)

// BloombergHTScraper scrapes news from Bloomberg HT
type BloombergHTScraper struct {
	collector *colly.Collector
	baseURL   string
}

func NewBloombergHTScraper() *BloombergHTScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("www.bloomberght.com"),
		colly.UserAgent("Mozilla/5.0 (compatible; MarketAI/1.0)"),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*bloomberght.com*", Delay: 2 * time.Second, RandomDelay: 1 * time.Second})
	return &BloombergHTScraper{collector: c, baseURL: "https://www.bloomberght.com"}
}

// ScrapeLatestNews fetches latest financial news articles
func (b *BloombergHTScraper) ScrapeLatestNews(ctx context.Context) ([]models.ScrapedArticle, error) {
	var articles []models.ScrapedArticle
	log.Debug().Msg("Scraping Bloomberg HT...")

	b.collector.OnHTML(".news-list .news-item", func(e *colly.HTMLElement) {
		article := models.ScrapedArticle{
			Source:    "Bloomberg HT",
			Title:     e.ChildText(".news-title"),
			URL:       b.baseURL + e.ChildAttr("a", "href"),
			ScrapedAt: time.Now(),
		}
		article.RelatedStocks = extractStockSymbols(article.Title)
		if article.Title != "" && article.URL != "" {
			articles = append(articles, article)
		}
	})
	b.collector.OnError(func(r *colly.Response, err error) {
		log.Error().Err(err).Str("url", r.Request.URL.String()).Msg("Bloomberg scrape error")
	})

	if err := b.collector.Visit(b.baseURL + "/borsa"); err != nil {
		return nil, fmt.Errorf("visit bloomberg: %w", err)
	}
	log.Info().Int("count", len(articles)).Msg("Bloomberg HT articles scraped")
	return articles, nil
}

// extractStockSymbols finds stock symbols in text (basic match)
func extractStockSymbols(text string) []string {
	known := []string{"THYAO", "AKBNK", "ASELS", "TUPRS", "EREGL", "GARAN", "ISCTR", "KCHOL", "PETKM", "SAHOL"}
	var out []string
	upper := strings.ToUpper(text)
	for _, s := range known {
		if strings.Contains(upper, s) {
			out = append(out, s)
		}
	}
	return out
}
