package handlers

import (
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ROIHistoryHandler struct{ db *pgxpool.Pool }

func NewROIHistoryHandler(db *pgxpool.Pool) *ROIHistoryHandler { return &ROIHistoryHandler{db: db} }

// GetAllAgentsROIHistory returns ROI time series for all active agents (recent N snapshots)
func (h *ROIHistoryHandler) GetAllAgentsROIHistory(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	if limit < 1 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	query := `
        SELECT aps.agent_id, aps.snapshot_time, aps.roi_percent
        FROM agent_performance_snapshots aps
        JOIN agents a ON aps.agent_id = a.id
        WHERE a.status = 'active'
        ORDER BY aps.snapshot_time DESC
        LIMIT $1`

	rows, err := h.db.Query(c.Context(), query, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{Success: false, Message: "Failed to fetch ROI history"})
	}
	defer rows.Close()

	type Point struct {
		Time time.Time `json:"time"`
		ROI  float64   `json:"roi"`
	}
	data := map[uuid.UUID][]Point{}
	for rows.Next() {
		var agentID uuid.UUID
		var ts time.Time
		var roi float64
		if err := rows.Scan(&agentID, &ts, &roi); err != nil {
			continue
		}
		data[agentID] = append(data[agentID], Point{Time: ts, ROI: roi})
	}

	return c.JSON(models.Response{Success: true, Data: data})
}
