package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/rs/zerolog/log"
)

// YahooFinanceClient fetches stock data from Yahoo Finance (15 min delay for BIST)
type YahooFinanceClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewYahooFinanceClient() *YahooFinanceClient {
	return &YahooFinanceClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    "https://query1.finance.yahoo.com/v8/finance",
	}
}

// GetStockPrice fetches current price for a BIST stock (symbol without .IS)
func (y *YahooFinanceClient) GetStockPrice(ctx context.Context, symbol string) (*models.StockPrice, error) {
	yahooSymbol := fmt.Sprintf("%s.IS", symbol)
	url := fmt.Sprintf("%s/chart/%s?interval=1d&range=1d", y.baseURL, yahooSymbol)
	log.Debug().Str("symbol", symbol).Str("url", url).Msg("Fetching Yahoo Finance price")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MarketAI/1.0)")

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo status %d", resp.StatusCode)
	}

	var result yahooResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(result.Chart.Result) == 0 || len(result.Chart.Result[0].Indicators.Quote) == 0 {
		return nil, fmt.Errorf("empty quote data for %s", symbol)
	}
	quote := result.Chart.Result[0].Indicators.Quote[0]
	lastIdx := len(quote.Close) - 1
	if lastIdx < 0 {
		return nil, fmt.Errorf("no close data")
	}
	price := &models.StockPrice{
		Symbol:       symbol,
		Price:        quote.Close[lastIdx],
		Open:         pick(quote.Open, lastIdx),
		High:         pick(quote.High, lastIdx),
		Low:          pick(quote.Low, lastIdx),
		Volume:       pickInt(quote.Volume, lastIdx),
		Source:       "yahoo",
		Timestamp:    time.Now(),
		DelayMinutes: 15,
	}
	log.Info().Str("symbol", symbol).Float64("price", price.Price).Msg("Yahoo price fetched")
	return price, nil
}

// GetMultipleStocks batch fetch (sequential with small sleep)
func (y *YahooFinanceClient) GetMultipleStocks(ctx context.Context, symbols []string) ([]*models.StockPrice, error) {
	out := make([]*models.StockPrice, 0, len(symbols))
	for _, s := range symbols {
		p, err := y.GetStockPrice(ctx, s)
		if err != nil {
			log.Error().Err(err).Str("symbol", s).Msg("Yahoo fetch failed")
			continue
		}
		out = append(out, p)
		time.Sleep(1 * time.Second)
	}
	return out, nil
}

// Helper to safely pick float slices
func pick(arr []float64, idx int) float64 {
	if idx >= 0 && idx < len(arr) {
		return arr[idx]
	}
	return 0
}

func pickInt(arr []int64, idx int) int64 {
	if idx >= 0 && idx < len(arr) {
		return arr[idx]
	}
	return 0
}

// Response structs
type yahooResponse struct {
	Chart struct {
		Result []struct {
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}
