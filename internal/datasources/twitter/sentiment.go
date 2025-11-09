package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	m "github.com/1batu/market-ai/internal/models"
	openai "github.com/sashabaranov/go-openai"
)

// Analyzer uses OpenAI to classify tweet sentiment
type Analyzer struct{ client *openai.Client }

func NewAnalyzer(apiKey string) *Analyzer {
	if apiKey == "" {
		return nil
	}
	return &Analyzer{client: openai.NewClient(apiKey)}
}

func (a *Analyzer) AnalyzeTweet(ctx context.Context, t *m.Tweet) error {
	if a == nil || a.client == nil {
		return nil
	}
	prompt := fmt.Sprintf(`Analyze the sentiment of this Turkish stock market tweet.
Tweet: %q

Respond ONLY with valid JSON:
{
  "sentiment": "positive|negative|neutral",
  "score": 0.75,
  "confidence": 0.85
}

Score range: -1.0 (very negative) to +1.0 (very positive)`, t.Text)

	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       "gpt-3.5-turbo",
		Temperature: 0.2,
		MaxTokens:   120,
		Messages: []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		}},
	})
	if err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("no openai choices")
	}
	var result struct {
		Sentiment  string  `json:"sentiment"`
		Score      float64 `json:"score"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return err
	}
	t.SentimentLabel = result.Sentiment
	t.SentimentScore = result.Score
	t.SentimentConfidence = result.Confidence

	// Impact score = engagement × score × author influence
	influence := float64(t.AuthorFollowers) / 10000.0
	if influence > 10 {
		influence = 10
	}
	engagement := float64(t.Likes + t.Retweets*2)
	t.ImpactScore = engagement * t.SentimentScore * influence
	return nil
}

func (a *Analyzer) AnalyzeBatch(ctx context.Context, tweets []m.Tweet) ([]m.Tweet, error) {
	for i := range tweets {
		_ = a.AnalyzeTweet(ctx, &tweets[i])
		time.Sleep(200 * time.Millisecond)
	}
	return tweets, nil
}
