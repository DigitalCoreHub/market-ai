package handlers

import (
	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LeaderboardHandler struct {
	db *pgxpool.Pool
}

func NewLeaderboardHandler(db *pgxpool.Pool) *LeaderboardHandler { return &LeaderboardHandler{db: db} }

// GetLeaderboard returns current leaderboard entries
func (h *LeaderboardHandler) GetLeaderboard(c *fiber.Ctx) error {
	const q = `
        SELECT
            lr.rank_overall,
            a.id,
            a.name,
            a.model,
            lr.current_roi,
            lr.current_profit_loss,
            lr.current_win_rate,
            lr.total_trades,
            a.current_balance,
            am.total_portfolio_value,
            lr.badges,
            lr.updated_at
        FROM leaderboard_rankings lr
        JOIN agents a ON lr.agent_id = a.id
        JOIN agent_metrics am ON a.id = am.agent_id
        ORDER BY lr.rank_overall ASC`

	rows, err := h.db.Query(c.Context(), q)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{Success: false, Message: "Failed to fetch leaderboard"})
	}
	defer rows.Close()

	entries := []models.LeaderboardEntry{}
	for rows.Next() {
		var e models.LeaderboardEntry
		var badges []string
		if err := rows.Scan(&e.Rank, &e.AgentID, &e.AgentName, &e.Model, &e.ROI, &e.ProfitLoss, &e.WinRate, &e.TotalTrades, &e.Balance, &e.PortfolioValue, &badges, &e.UpdatedAt); err != nil {
			continue
		}
		e.Badges = badges
		e.TotalValue = e.Balance + e.PortfolioValue
		entries = append(entries, e)
	}

	return c.JSON(models.Response{Success: true, Data: entries})
}
