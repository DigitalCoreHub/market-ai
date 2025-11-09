package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricsHandler struct {
	db *pgxpool.Pool
}

func NewMetricsHandler(db *pgxpool.Pool) *MetricsHandler {
	return &MetricsHandler{db: db}
}

type DataSourceMetrics struct {
	SourceType        string `json:"source_type"`
	SourceName        string `json:"source_name"`
	IsActive          bool   `json:"is_active"`
	TotalFetches      int    `json:"total_fetches"`
	SuccessCount      int    `json:"success_count"`
	ErrorCount        int    `json:"error_count"`
	AvgResponseTimeMs int    `json:"avg_response_time_ms"`
	Status            string `json:"status"`
	LastError         string `json:"last_error,omitempty"`
	LastFetchAt       string `json:"last_fetch_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

// Get /api/v1/metrics
func (h *MetricsHandler) Get(c *fiber.Ctx) error {
	rows, err := h.db.Query(context.Background(), `
		SELECT source_type, source_name, COALESCE(is_active,true),
		       COALESCE(total_fetches,0), COALESCE(success_count,0), COALESCE(error_count,0),
		       COALESCE(avg_response_time_ms,0), COALESCE(status,'active'),
		       COALESCE(last_error,''), COALESCE(last_fetch_at::text,''), COALESCE(updated_at::text,'')
		FROM data_sources
		ORDER BY source_type, source_name
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{Success: false, Message: "metrics query error"})
	}
	defer rows.Close()
	var metrics []DataSourceMetrics
	for rows.Next() {
		var m DataSourceMetrics
		if err := rows.Scan(&m.SourceType, &m.SourceName, &m.IsActive, &m.TotalFetches, &m.SuccessCount, &m.ErrorCount, &m.AvgResponseTimeMs, &m.Status, &m.LastError, &m.LastFetchAt, &m.UpdatedAt); err != nil {
			continue
		}
		metrics = append(metrics, m)
	}
	return c.JSON(models.Response{Success: true, Data: fiber.Map{
		"data_sources": metrics,
	}})
}

// GetPrometheus returns metrics in Prometheus format
// GET /api/v1/metrics/prometheus
func (h *MetricsHandler) GetPrometheus(c *fiber.Ctx) error {
	var output []string

	// Active agents count
	var activeAgents int
	if err := h.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM agents WHERE status = 'active'").Scan(&activeAgents); err != nil {
		activeAgents = 0
	}
	output = append(output, fmt.Sprintf("marketai_active_agents %d", activeAgents))

	// Total trades count
	var totalTrades int
	if err := h.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM trades").Scan(&totalTrades); err != nil {
		totalTrades = 0
	}
	output = append(output, fmt.Sprintf("marketai_total_trades %d", totalTrades))

	// Average trade latency (simplified - using created_at)
	var avgLatency float64
	if err := h.db.QueryRow(context.Background(), `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at))), 0)
		FROM trades
		WHERE created_at > NOW() - INTERVAL '1 hour'
	`).Scan(&avgLatency); err != nil {
		avgLatency = 0
	}
	output = append(output, fmt.Sprintf("marketai_avg_trade_latency_seconds %.2f", avgLatency))

	// Reasoning logs per minute (simplified - using agent_decisions)
	var decisionsPerMin float64
	if err := h.db.QueryRow(context.Background(), `
		SELECT COALESCE(COUNT(*)::float / 60.0, 0)
		FROM agent_decisions
		WHERE created_at > NOW() - INTERVAL '1 minute'
	`).Scan(&decisionsPerMin); err != nil {
		decisionsPerMin = 0
	}
	output = append(output, fmt.Sprintf("marketai_reasoning_logs_per_minute %.2f", decisionsPerMin))

	// Data source metrics
	rows, err := h.db.Query(context.Background(), `
		SELECT source_name, success_count, error_count, avg_response_time_ms
		FROM data_sources
		WHERE is_active = true
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sourceName string
			var successCount, errorCount, avgResponseTime int
			if err := rows.Scan(&sourceName, &successCount, &errorCount, &avgResponseTime); err != nil {
				continue
			}

			// Sanitize source name for Prometheus (replace spaces with underscores)
			sanitizedName := fmt.Sprintf("marketai_datasource_%s", sanitizeMetricName(sourceName))
			output = append(output, fmt.Sprintf("%s_success_count %d", sanitizedName, successCount))
			output = append(output, fmt.Sprintf("%s_error_count %d", sanitizedName, errorCount))
			output = append(output, fmt.Sprintf("%s_avg_response_time_ms %d", sanitizedName, avgResponseTime))
		}
	}

	// Add timestamp
	output = append(output, fmt.Sprintf("# Timestamp: %d", time.Now().Unix()))

	c.Set("Content-Type", "text/plain; version=0.0.4")
	result := ""
	for _, line := range output {
		result += line + "\n"
	}
	return c.SendString(result)
}

func sanitizeMetricName(name string) string {
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result += string(r)
		} else if r == ' ' {
			result += "_"
		}
	}
	return result
}
