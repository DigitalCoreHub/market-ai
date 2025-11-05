package handlers

import (
	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockHandler struct {
	db *pgxpool.Pool
}

func NewStockHandler(db *pgxpool.Pool) *StockHandler {
	return &StockHandler{db: db}
}

func (h *StockHandler) GetAll(c *fiber.Ctx) error {
	query := `
		SELECT id, symbol, name, current_price, previous_close,
		       change_percent, volume, last_updated, created_at
		FROM stocks
		ORDER BY symbol ASC
	`

	rows, err := h.db.Query(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Success: false,
			Message: "Failed to fetch stocks",
		})
	}
	defer rows.Close()

	var stocks []models.Stock
	for rows.Next() {
		var stock models.Stock
		if err := rows.Scan(
			&stock.ID, &stock.Symbol, &stock.Name, &stock.CurrentPrice,
			&stock.PreviousClose, &stock.ChangePercent, &stock.Volume,
			&stock.LastUpdated, &stock.CreatedAt,
		); err != nil {
			continue
		}
		stocks = append(stocks, stock)
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    stocks,
	})
}

func (h *StockHandler) GetBySymbol(c *fiber.Ctx) error {
	symbol := c.Params("symbol")

	var stock models.Stock
	query := `
		SELECT id, symbol, name, current_price, previous_close,
		       change_percent, volume, last_updated, created_at
		FROM stocks WHERE symbol = $1
	`

	err := h.db.QueryRow(c.Context(), query, symbol).Scan(
		&stock.ID, &stock.Symbol, &stock.Name, &stock.CurrentPrice,
		&stock.PreviousClose, &stock.ChangePercent, &stock.Volume,
		&stock.LastUpdated, &stock.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Success: false,
			Message: "Stock not found",
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    stock,
	})
}

func (h *StockHandler) GetHistory(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	timeframe := c.Query("timeframe", "1m")

	query := `
		SELECT id, stock_symbol, open_price, close_price, high_price, low_price,
		       volume, timestamp, timeframe
		FROM market_data
		WHERE stock_symbol = $1 AND timeframe = $2
		ORDER BY timestamp DESC
		LIMIT 100
	`

	rows, err := h.db.Query(c.Context(), query, symbol, timeframe)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Success: false,
			Message: "Failed to fetch market data",
		})
	}
	defer rows.Close()

	var data []models.MarketData
	for rows.Next() {
		var md models.MarketData
		if err := rows.Scan(
			&md.ID, &md.Symbol, &md.OpenPrice, &md.ClosePrice,
			&md.HighPrice, &md.LowPrice, &md.Volume,
			&md.Timestamp, &md.Timeframe,
		); err != nil {
			continue
		}
		data = append(data, md)
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    data,
	})
}
