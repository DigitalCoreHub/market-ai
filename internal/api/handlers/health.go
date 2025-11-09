package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Check tüm servislerin sağlık durumunu kontrol eder (GET /health)
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	services := make(map[string]interface{})
	ctx := context.Background()

	// PostgreSQL health check
	dbStatus := "healthy"
	var dbLatency int64
	start := time.Now()
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "unhealthy"
	} else {
		dbLatency = time.Since(start).Milliseconds()
	}
	services["postgres"] = map[string]interface{}{
		"status":  dbStatus,
		"latency": fmt.Sprintf("%dms", dbLatency),
	}

	// Redis health check
	redisStatus := "healthy"
	var redisLatency int64
	start = time.Now()
	if err := h.redis.Ping(ctx).Err(); err != nil {
		redisStatus = "unhealthy"
	} else {
		redisLatency = time.Since(start).Milliseconds()
	}
	services["redis"] = map[string]interface{}{
		"status":  redisStatus,
		"latency": fmt.Sprintf("%dms", redisLatency),
	}

	// WebSocket hub status (always healthy if server is running)
	services["websocket"] = map[string]interface{}{
		"status": "healthy",
		"note":   "Hub is running",
	}

	overallStatus := "healthy"
	if dbStatus == "unhealthy" || redisStatus == "unhealthy" {
		overallStatus = "unhealthy"
	}

	response := models.Response{
		Success: overallStatus == "healthy",
		Data: models.HealthResponse{
			Status:   overallStatus,
			Services: services,
		},
	}

	if overallStatus == "unhealthy" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	return c.JSON(response)
}

// Ping basit connectivity test endpoint'i (GET /api/v1/ping)
func (h *HealthHandler) Ping(c *fiber.Ctx) error {
	return c.JSON(models.Response{
		Success: true,
		Message: "pong",
	})
}
