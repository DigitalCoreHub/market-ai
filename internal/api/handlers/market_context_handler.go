package handlers

import (
	"strings"

	"github.com/1batu/market-ai/internal/datasources/fusion"
	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
)

type MarketContextHandler struct {
	fusion *fusion.Service
}

func NewMarketContextHandler(f *fusion.Service) *MarketContextHandler {
	return &MarketContextHandler{fusion: f}
}

// GetContext istenen semboller için taze piyasa bağlamı oluşturur (virgülle ayrılmış ?symbols=THYAO,AKBNK)
func (h *MarketContextHandler) GetContext(c *fiber.Ctx) error {
	symbolsParam := c.Query("symbols", "")
	if symbolsParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{Success: false, Message: "symbols param required"})
	}
	raw := strings.Split(symbolsParam, ",")
	var symbols []string
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s != "" {
			symbols = append(symbols, strings.ToUpper(s))
		}
	}
	if len(symbols) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{Success: false, Message: "no valid symbols"})
	}
	ctxOut, err := h.fusion.MarketContext(c.Context(), symbols)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{Success: false, Message: "failed to collect context"})
	}
	return c.JSON(models.Response{Success: true, Data: ctxOut})
}
