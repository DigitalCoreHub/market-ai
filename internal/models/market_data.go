package models

import "time"

// StockPrice represents a stock price from a specific source
type StockPrice struct {
	Symbol          string    `json:"symbol"`
	Price           float64   `json:"price"`
	Open            float64   `json:"open"`
	High            float64   `json:"high"`
	Low             float64   `json:"low"`
	Volume          int64     `json:"volume"`
	Source          string    `json:"source"`
	Timestamp       time.Time `json:"timestamp"`
	DelayMinutes    int       `json:"delay_minutes"`
	ConfidenceScore float64   `json:"confidence_score"`
}

// ScrapedArticle represents a scraped news article
type ScrapedArticle struct {
	Source        string    `json:"source"`
	Title         string    `json:"title"`
	Content       string    `json:"content,omitempty"`
	URL           string    `json:"url"`
	RelatedStocks []string  `json:"related_stocks,omitempty"`
	PublishedAt   time.Time `json:"published_at,omitempty"`
	ScrapedAt     time.Time `json:"scraped_at"`
}

// MarketContext is the fused multi-source snapshot
type MarketContext struct {
	Prices          []*StockPrice              `json:"prices"`
	News            []ScrapedArticle           `json:"news"`
	Tweets          []Tweet                    `json:"tweets"`
	StockSentiments map[string]*StockSentiment `json:"stock_sentiments"`
	UpdatedAt       time.Time                  `json:"updated_at"`
	FetchDurations  map[string]time.Duration   `json:"fetch_durations"`
}
