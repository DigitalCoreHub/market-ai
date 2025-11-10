package handlers

import (
	"github.com/1batu/market-ai/internal/services"
	"github.com/gofiber/fiber/v2"
)

// NewsHandler handles news-related API requests
type NewsHandler struct {
	newsAggregator *services.NewsAggregator
}

// NewNewsHandler creates a new news handler
func NewNewsHandler(newsAggregator *services.NewsAggregator) *NewsHandler {
	return &NewsHandler{newsAggregator: newsAggregator}
}

// TriggerNewsFetch manually triggers news fetching
// POST /api/v1/news/fetch
func (h *NewsHandler) TriggerNewsFetch(c *fiber.Ctx) error {
	go func() {
		// Call FetchAndStore in a goroutine to avoid blocking the response
		h.newsAggregator.FetchAndStore(c.Context())
	}()
	return c.JSON(fiber.Map{
		"success": true,
		"message": "News fetch triggered",
	})
}

// GetLatestNews returns the latest news articles
// GET /api/v1/news/latest
func (h *NewsHandler) GetLatestNews(c *fiber.Ctx) error {
	articles, err := h.newsAggregator.GetLatestNews(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch latest news",
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    articles,
		"count":   len(articles),
	})
}

