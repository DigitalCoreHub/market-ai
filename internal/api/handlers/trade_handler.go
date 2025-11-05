package handlers

import (
	"github.com/1batu/market-ai/internal/models"
	"github.com/1batu/market-ai/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TradeHandler struct {
	db     *pgxpool.Pool
	engine *services.TradingEngine
}

func NewTradeHandler(db *pgxpool.Pool, engine *services.TradingEngine) *TradeHandler {
	return &TradeHandler{
		db:     db,
		engine: engine,
	}
}

func (h *TradeHandler) Execute(c *fiber.Ctx) error {
	var req models.TradeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Success: false,
			Message: "Invalid request body",
		})
	}

	trade, err := h.engine.ExecuteTrade(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.Response{
		Success: true,
		Message: "Trade executed successfully",
		Data:    trade,
	})
}

func (h *TradeHandler) GetHistory(c *fiber.Ctx) error {
	agentID := c.Query("agent_id")
	limit := c.QueryInt("limit", 50)

	query := `
		SELECT id, agent_id, stock_symbol, trade_type, quantity, price,
		       total_amount, commission, reasoning, created_at
		FROM trades
		WHERE ($1 = '' OR agent_id::text = $1)
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := h.db.Query(c.Context(), query, agentID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Success: false,
			Message: "Failed to fetch trades",
		})
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var trade models.Trade
		if err := rows.Scan(
			&trade.ID, &trade.AgentID, &trade.StockSymbol, &trade.TradeType,
			&trade.Quantity, &trade.Price, &trade.TotalAmount,
			&trade.Commission, &trade.Reasoning, &trade.CreatedAt,
		); err != nil {
			continue
		}
		trades = append(trades, trade)
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    trades,
	})
}
