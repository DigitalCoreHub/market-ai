package news

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/rs/zerolog/log"
)

// NewsAPIClient interacts with NewsAPI.org
type NewsAPIClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// newsAPIResponse represents the News API response structure
type newsAPIResponse struct {
	Status       string           `json:"status"`
	TotalResults int              `json:"totalResults"`
	Articles     []newsAPIArticle `json:"articles"`
}

// newsAPIArticle represents a single article from News API
type newsAPIArticle struct {
	Source      newsAPISource `json:"source"`
	Author      string        `json:"author"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	URL         string        `json:"url"`
	PublishedAt string        `json:"publishedAt"`
	Content     string        `json:"content"`
}

// newsAPISource represents the source in News API
type newsAPISource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewNewsAPIClient creates a new News API client
func NewNewsAPIClient(apiKey string) *NewsAPIClient {
	return &NewsAPIClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: "https://newsapi.org/v2",
	}
}

// GetTurkeyFinanceNews fetches latest Turkish finance news
func (c *NewsAPIClient) GetTurkeyFinanceNews(ctx context.Context) ([]models.NewsArticle, error) {
	if c.apiKey == "" {
		log.Warn().Msg("NewsAPI key not configured, skipping News API fetch")
		return []models.NewsArticle{}, nil
	}

	// Build query: Turkey economy OR BIST OR Borsa Istanbul
	query := url.QueryEscape("(Turkey OR Türkiye) AND (economy OR BIST OR \"Borsa Istanbul\" OR finans OR hisse)")

	// API endpoint
	endpoint := fmt.Sprintf("%s/everything?q=%s&language=tr&sortBy=publishedAt&pageSize=50&apiKey=%s",
		c.baseURL, query, c.apiKey)

	log.Debug().Msg("Fetching from News API")

	// Make request
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("news API returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp newsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Status != "ok" {
		return nil, fmt.Errorf("news API returned status: %s", apiResp.Status)
	}

	// Convert to our model
	articles := make([]models.NewsArticle, 0, len(apiResp.Articles))
	for _, apiArticle := range apiResp.Articles {
		// Skip if no URL
		if apiArticle.URL == "" {
			continue
		}

		// Parse published time
		publishedAt, err := time.Parse(time.RFC3339, apiArticle.PublishedAt)
		if err != nil {
			log.Warn().Str("time", apiArticle.PublishedAt).Msg("Failed to parse publish time")
			publishedAt = time.Now()
		}

		// Extract stock symbols from title/description (simple heuristic)
		relatedStocks := extractStockSymbols(apiArticle.Title + " " + apiArticle.Description)

		article := models.NewsArticle{
			Title:         apiArticle.Title,
			Description:   apiArticle.Description,
			Content:       apiArticle.Content,
			Source:        apiArticle.Source.Name,
			URL:           apiArticle.URL,
			EventType:     "news",
			RelatedStocks: relatedStocks,
			PublishedAt:   publishedAt,
			FetchedAt:     time.Now(),
		}

		articles = append(articles, article)
	}

	log.Info().Int("count", len(articles)).Msg("News API fetch completed")

	return articles, nil
}

// extractStockSymbols extracts BIST stock symbols from text
func extractStockSymbols(text string) []string {
	// Known BIST stocks
	knownStocks := []string{
		"THYAO", "AKBNK", "GARAN", "ASELS", "TCELL", "SISE",
		"ARCLK", "KCHOL", "TOASO", "PETKM", "EREGL", "VESTL",
		"AKSEN", "DOAS", "KOZAL", "KOZAA", "HALKB", "ISCTR",
		"VFBNK", "SAHOL", "AEFES", "TATGD", "KARTN", "OTKAR",
		"BIMAS", "ENERY", "TTRAK", "VARBS", "AGRIW", "TRAZT",
	}

	text = strings.ToUpper(text)
	found := make(map[string]bool)

	for _, stock := range knownStocks {
		if strings.Contains(text, stock) {
			found[stock] = true
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(found))
	for stock := range found {
		result = append(result, stock)
	}

	return result
}
