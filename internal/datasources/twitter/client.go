package twitter

import (
	"context"
	"fmt"
	"strings"
	"time"

	m "github.com/1batu/market-ai/internal/models"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/rs/zerolog/log"
)

// Client wraps twitter API usage (search + stream)
type Client struct {
	client   *twitter.Client
	hashtags []string
	keywords []string
}

func NewClient(apiKey, apiSecret, accessToken, accessSecret string) *Client {
	if apiKey == "" || apiSecret == "" || accessToken == "" || accessSecret == "" {
		log.Warn().Msg("Twitter credentials missing; client will be disabled")
		return nil
	}
	cfg := oauth1.NewConfig(apiKey, apiSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := cfg.Client(oauth1.NoContext, token)
	tc := twitter.NewClient(httpClient)
	return &Client{
		client:   tc,
		hashtags: []string{"#BIST", "#BorsaIstanbul", "#THYAO", "#AKBNK", "#ASELS", "#Borsa"},
		keywords: []string{"BIST", "Borsa Istanbul", "hisse", "yükseliş", "düşüş"},
	}
}

// SearchRecent fetches recent tweets matching finance query
func (c *Client) SearchRecent(ctx context.Context, maxResults int) ([]m.Tweet, error) {
	if c == nil {
		return nil, fmt.Errorf("twitter client disabled")
	}
	query := strings.Join(c.hashtags, " OR ")
	params := &twitter.SearchTweetParams{Query: query, Count: maxResults, ResultType: "recent", Lang: "tr"}
	res, _, err := c.client.Search.Tweets(params)
	if err != nil {
		return nil, fmt.Errorf("twitter search: %w", err)
	}
	out := make([]m.Tweet, 0, len(res.Statuses))
	for _, st := range res.Statuses {
		tw := m.Tweet{
			ID:              st.IDStr,
			Text:            st.Text,
			Author:          st.User.ScreenName,
			AuthorFollowers: st.User.FollowersCount,
			Likes:           st.FavoriteCount,
			Retweets:        st.RetweetCount,
			CreatedAt:       parseTime(st.CreatedAt),
			URL:             fmt.Sprintf("https://twitter.com/%s/status/%s", st.User.ScreenName, st.IDStr),
		}
		tw.StockSymbols = extractStockMentions(tw.Text)
		out = append(out, tw)
	}
	log.Debug().Int("tweets", len(out)).Msg("Twitter search completed")
	return out, nil
}

// StreamFinancial starts streaming tweets; returns channel; caller must cancel context.
func (c *Client) StreamFinancial(ctx context.Context) (<-chan m.Tweet, error) {
	if c == nil {
		return nil, fmt.Errorf("twitter client disabled")
	}
	params := &twitter.StreamFilterParams{Track: c.hashtags, StallWarnings: twitter.Bool(true), Language: []string{"tr"}}
	stream, err := c.client.Streams.Filter(params)
	if err != nil {
		return nil, fmt.Errorf("start stream: %w", err)
	}
	ch := make(chan m.Tweet, 100)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				stream.Stop()
				return
			case msg := <-stream.Messages:
				if tw, ok := msg.(*twitter.Tweet); ok {
					t := m.Tweet{
						ID:              tw.IDStr,
						Text:            tw.Text,
						Author:          tw.User.ScreenName,
						AuthorFollowers: tw.User.FollowersCount,
						Likes:           tw.FavoriteCount,
						Retweets:        tw.RetweetCount,
						CreatedAt:       parseTime(tw.CreatedAt),
						StockSymbols:    extractStockMentions(tw.Text),
					}
					ch <- t
				}
			}
		}
	}()
	log.Info().Msg("Twitter stream started")
	return ch, nil
}

func extractStockMentions(text string) []string {
	known := []string{"THYAO", "AKBNK", "ASELS", "TUPRS", "EREGL", "GARAN", "ISCTR"}
	var out []string
	upper := strings.ToUpper(text)
	for _, s := range known {
		if strings.Contains(upper, s) {
			out = append(out, s)
		}
	}
	return out
}

func parseTime(ts string) time.Time {
	// Twitter format example: "Mon Jan 02 15:04:05 -0700 2006"
	t, err := time.Parse(time.RubyDate, ts)
	if err != nil {
		return time.Now()
	}
	return t
}
