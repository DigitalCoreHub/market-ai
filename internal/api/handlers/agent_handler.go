package handlers

import (
	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AgentHandler struct {
	db *pgxpool.Pool
}

func NewAgentHandler(db *pgxpool.Pool) *AgentHandler {
	return &AgentHandler{db: db}
}

func (h *AgentHandler) GetAll(c *fiber.Ctx) error {
	query := `
		SELECT a.id, a.name, a.model, a.status, a.initial_balance, a.current_balance,
		       a.created_at, a.updated_at,
		       COALESCE(m.total_profit_loss, 0) as profit_loss,
		       COALESCE(m.roi, 0) as roi
		FROM agents a
		LEFT JOIN agent_metrics m ON a.id = m.agent_id
		ORDER BY a.created_at DESC
	`

	rows, err := h.db.Query(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Success: false,
			Message: "Failed to fetch agents",
		})
	}
	defer rows.Close()

	type AgentWithMetrics struct {
		models.Agent
		ProfitLoss float64 `json:"profit_loss"`
		ROI        float64 `json:"roi"`
	}

	var agents []AgentWithMetrics
	for rows.Next() {
		var agent AgentWithMetrics
		if err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Model, &agent.Status,
			&agent.InitialBalance, &agent.CurrentBalance,
			&agent.CreatedAt, &agent.UpdatedAt,
			&agent.ProfitLoss, &agent.ROI,
		); err != nil {
			continue
		}
		agents = append(agents, agent)
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    agents,
	})
}

func (h *AgentHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Success: false,
			Message: "Invalid agent ID",
		})
	}

	var agent models.Agent
	query := `
		SELECT id, name, model, status, initial_balance, current_balance, created_at, updated_at
		FROM agents WHERE id = $1
	`

	err = h.db.QueryRow(c.Context(), query, id).Scan(
		&agent.ID, &agent.Name, &agent.Model, &agent.Status,
		&agent.InitialBalance, &agent.CurrentBalance,
		&agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Success: false,
			Message: "Agent not found",
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    agent,
	})
}

func (h *AgentHandler) GetMetrics(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Success: false,
			Message: "Invalid agent ID",
		})
	}

	var metrics models.AgentMetrics
	query := `
		SELECT id, agent_id, total_trades, winning_trades, losing_trades,
		       total_profit_loss, total_portfolio_value, win_rate, roi,
		       sharpe_ratio, max_drawdown, calculated_at
		FROM agent_metrics WHERE agent_id = $1
	`

	err = h.db.QueryRow(c.Context(), query, id).Scan(
		&metrics.ID, &metrics.AgentID, &metrics.TotalTrades,
		&metrics.WinningTrades, &metrics.LosingTrades,
		&metrics.TotalProfitLoss, &metrics.TotalPortfolioValue,
		&metrics.WinRate, &metrics.ROI, &metrics.SharpeRatio,
		&metrics.MaxDrawdown, &metrics.CalculatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.Response{
			Success: false,
			Message: "Metrics not found",
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    metrics,
	})
}

func (h *AgentHandler) GetPortfolio(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{
			Success: false,
			Message: "Invalid agent ID",
		})
	}

	query := `
		SELECT p.id, p.agent_id, p.stock_symbol, p.quantity, p.avg_buy_price,
		       p.total_invested, p.current_value, p.profit_loss,
		       p.profit_loss_percent, p.updated_at
		FROM portfolio p
		WHERE p.agent_id = $1
		ORDER BY p.total_invested DESC
	`

	rows, err := h.db.Query(c.Context(), query, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Success: false,
			Message: "Failed to fetch portfolio",
		})
	}
	defer rows.Close()

	var holdings []models.Portfolio
	for rows.Next() {
		var p models.Portfolio
		if err := rows.Scan(
			&p.ID, &p.AgentID, &p.StockSymbol, &p.Quantity,
			&p.AvgBuyPrice, &p.TotalInvested, &p.CurrentValue,
			&p.ProfitLoss, &p.ProfitLossPercent, &p.UpdatedAt,
		); err != nil {
			continue
		}
		holdings = append(holdings, p)
	}

	return c.JSON(models.Response{
		Success: true,
		Data:    holdings,
	})
}
