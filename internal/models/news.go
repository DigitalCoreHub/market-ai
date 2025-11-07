package models

import (
	"time"

	"github.com/google/uuid"
)

// NewsArticle represents a news article from any source
type NewsArticle struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Title          string    `json:"title" db:"title"`
	Description    string    `json:"description" db:"description"`
	Content        string    `json:"content" db:"content"`
	Source         string    `json:"source" db:"source"`
	URL            string    `json:"url" db:"url"`
	EventType      string    `json:"event_type" db:"event_type"`
	Category       string    `json:"category" db:"category"`
	RelatedStocks  []string  `json:"related_stocks" db:"related_stocks"`
	Sentiment      string    `json:"sentiment,omitempty" db:"sentiment"`
	SentimentScore float64   `json:"sentiment_score,omitempty" db:"sentiment_score"`
	ImpactLevel    string    `json:"impact_level,omitempty" db:"impact_level"`
	PublishedAt    time.Time `json:"published_at" db:"published_at"`
	FetchedAt      time.Time `json:"fetched_at" db:"fetched_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// NewsFetchResult tracks news fetch operations
type NewsFetchResult struct {
	Source          string
	ArticlesFetched int
	ArticlesStored  int
	Success         bool
	ErrorMessage    string
	ExecutionTimeMs int
}

// MarketEvent represents a market-related event/news
type MarketEvent struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Title          string    `json:"title" db:"title"`
	Description    string    `json:"description" db:"description"`
	Content        string    `json:"content" db:"content"`
	Source         string    `json:"source" db:"source"`
	URL            string    `json:"url" db:"url"`
	EventType      string    `json:"event_type" db:"event_type"`
	Category       string    `json:"category" db:"category"`
	RelatedStocks  []string  `json:"related_stocks" db:"related_stocks"`
	Sentiment      string    `json:"sentiment" db:"sentiment"`
	SentimentScore float64   `json:"sentiment_score" db:"sentiment_score"`
	ImpactLevel    string    `json:"impact_level" db:"impact_level"`
	PublishedAt    time.Time `json:"published_at" db:"published_at"`
	FetchedAt      time.Time `json:"fetched_at" db:"fetched_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
