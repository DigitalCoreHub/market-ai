package fusion

import (
	"context"
	"testing"

	"github.com/1batu/market-ai/internal/models"
)

func TestAggregateSentiment(t *testing.T) {
	symbols := []string{"THYAO", "AKBNK"}
	tweets := []models.Tweet{
		{Text: "THYAO çok iyi gidiyor", StockSymbols: []string{"THYAO"}, SentimentLabel: "positive", SentimentScore: 0.8, ImpactScore: 1.0},
		{Text: "AKBNK kötü haber", StockSymbols: []string{"AKBNK"}, SentimentLabel: "negative", SentimentScore: -0.6, ImpactScore: 2.0},
		{Text: "THYAO nötr", StockSymbols: []string{"THYAO"}, SentimentLabel: "neutral", SentimentScore: 0.0, ImpactScore: 0.5},
	}

	agg := aggregateSentiment(tweets, symbols)

	if got := agg["THYAO"].TweetCount; got != 2 {
		t.Fatalf("THYAO TweetCount = %d, want 2", got)
	}
	if got := agg["THYAO"].PositiveCount; got != 1 {
		t.Fatalf("THYAO PositiveCount = %d, want 1", got)
	}
	if got := agg["THYAO"].NeutralCount; got != 1 {
		t.Fatalf("THYAO NeutralCount = %d, want 1", got)
	}
	if got := agg["AKBNK"].NegativeCount; got != 1 {
		t.Fatalf("AKBNK NegativeCount = %d, want 1", got)
	}
	if got := agg["AKBNK"].AvgSentiment; got >= 0.0 {
		t.Fatalf("AKBNK AvgSentiment = %f, want < 0", got)
	}

	// sanity: context function should not crash on empty fetch
	svc := &Service{}
	if _, err := svc.MarketContext(context.Background(), []string{"THYAO"}); err == nil {
		// svc has nil clients; should not panic but will likely log errors and still return context.
	}
}
