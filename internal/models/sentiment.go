package models

import "time"

// Tweet represents a financial tweet mentioning stocks
type Tweet struct {
	ID                  string    `json:"id"`
	Text                string    `json:"text"`
	URL                 string    `json:"url,omitempty"`
	Author              string    `json:"author"`
	AuthorFollowers     int       `json:"author_followers"`
	Likes               int       `json:"likes"`
	Retweets            int       `json:"retweets"`
	CreatedAt           time.Time `json:"created_at"`
	StockSymbols        []string  `json:"stock_symbols"`
	SentimentScore      float64   `json:"sentiment_score"`
	SentimentLabel      string    `json:"sentiment_label"`
	SentimentConfidence float64   `json:"sentiment_confidence"`
	ImpactScore         float64   `json:"impact_score"`
}

// StockSentiment aggregates sentiment for a stock
type StockSentiment struct {
	Symbol        string  `json:"symbol"`
	TweetCount    int     `json:"tweet_count"`
	AvgSentiment  float64 `json:"avg_sentiment"`
	PositiveCount int     `json:"positive_count"`
	NegativeCount int     `json:"negative_count"`
	NeutralCount  int     `json:"neutral_count"`
	TopTweet      *Tweet  `json:"top_tweet,omitempty"`
}
