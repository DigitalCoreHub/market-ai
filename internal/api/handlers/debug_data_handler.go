package handlers

import (
	"strings"

	"github.com/1batu/market-ai/internal/datasources/scraper"
	tw "github.com/1batu/market-ai/internal/datasources/twitter"
	"github.com/1batu/market-ai/internal/datasources/yahoo"
	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
)

// DebugDataHandler exposes per-source debug endpoints (Yahoo, Scraper, Twitter)
type DebugDataHandler struct {
	yahoo    *yahoo.YahooFinanceClient
	scraper  *scraper.WebScraper
	twitter  *tw.Client
	analyzer *tw.Analyzer
}

func NewDebugDataHandler(y *yahoo.YahooFinanceClient, s *scraper.WebScraper, t *tw.Client, a *tw.Analyzer) *DebugDataHandler {
	return &DebugDataHandler{yahoo: y, scraper: s, twitter: t, analyzer: a}
}

// GET /api/v1/debug/yahoo?symbols=THYAO,AKBNK
func (h *DebugDataHandler) GetYahoo(c *fiber.Ctx) error {
	symbolsParam := c.Query("symbols", "")
	if symbolsParam == "" || h.yahoo == nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{Success: false, Message: "symbols param required or yahoo disabled"})
	}
	raw := strings.Split(symbolsParam, ",")
	var symbols []string
	for _, s := range raw {
		if s = strings.TrimSpace(s); s != "" {
			symbols = append(symbols, strings.ToUpper(s))
		}
	}
	prices, err := h.yahoo.GetMultipleStocks(c.Context(), symbols)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{Success: false, Message: "yahoo fetch error"})
	}
	return c.JSON(models.Response{Success: true, Data: prices})
}

// GET /api/v1/debug/scraper
func (h *DebugDataHandler) GetScraper(c *fiber.Ctx) error {
	if h.scraper == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Response{Success: false, Message: "scraper disabled"})
	}
	items, err := h.scraper.ScrapeAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{Success: false, Message: "scrape error"})
	}
	return c.JSON(models.Response{Success: true, Data: items})
}

// GET /api/v1/debug/tweets?max=50&analyze=true
func (h *DebugDataHandler) GetTweets(c *fiber.Ctx) error {
	if h.twitter == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Response{Success: false, Message: "twitter disabled"})
	}
	max := c.QueryInt("max", 30)
	analyze := c.QueryBool("analyze", false)
	tweets, err := h.twitter.SearchRecent(c.Context(), max)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{Success: false, Message: "twitter search error"})
	}
	if analyze && h.analyzer != nil {
		analyzed, _ := h.analyzer.AnalyzeBatch(c.Context(), tweets)
		tweets = analyzed
	}
	return c.JSON(models.Response{Success: true, Data: tweets})
}
