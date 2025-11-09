package models

import (
	"time"

	"github.com/google/uuid"
)

type Stock struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	Symbol             string     `json:"symbol" db:"symbol"`
	Name               string     `json:"name" db:"name"`
	CurrentPrice       float64    `json:"current_price" db:"current_price"`
	PreviousClose      float64    `json:"previous_close" db:"previous_close"`
	ChangePercent      float64    `json:"change_percent" db:"change_percent"`
	Volume             int64      `json:"volume" db:"volume"`
	LastUpdated        time.Time  `json:"last_updated" db:"last_updated"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	MarketCap          int64      `json:"market_cap" db:"market_cap"`
	IsActive           bool       `json:"is_active" db:"is_active"`
	LastTradeDate      *time.Time `json:"last_trade_date" db:"last_trade_date"`
	LastNewsMention    *time.Time `json:"last_news_mention" db:"last_news_mention"`
	LastTwitterMention *time.Time `json:"last_twitter_mention" db:"last_twitter_mention"`
	DiscoverySource    string     `json:"discovery_source" db:"discovery_source"`
	MentionCount7d     int        `json:"mention_count_7d" db:"mention_count_7d"`
	TwitterCount24h    int        `json:"twitter_count_24h" db:"twitter_count_24h"`
}

type MarketData struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Symbol     string    `json:"symbol" db:"stock_symbol"`
	OpenPrice  float64   `json:"open_price" db:"open_price"`
	ClosePrice float64   `json:"close_price" db:"close_price"`
	HighPrice  float64   `json:"high_price" db:"high_price"`
	LowPrice   float64   `json:"low_price" db:"low_price"`
	Volume     int64     `json:"volume" db:"volume"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	Timeframe  string    `json:"timeframe" db:"timeframe"`
}
