package handlers

import (
	"context"

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
	services := make(map[string]string)

	if err := h.db.Ping(context.Background()); err != nil {
		services["postgres"] = "unhealthy"
	} else {
		services["postgres"] = "healthy"
	}

	if err := h.redis.Ping(context.Background()).Err(); err != nil {
		services["redis"] = "unhealthy"
	} else {
		services["redis"] = "healthy"
	}

	status := "healthy"
	for _, v := range services {
		if v == "unhealthy" {
			status = "unhealthy"
			break
		}
	}

	response := models.Response{
		Success: status == "healthy",
		Data: models.HealthResponse{
			Status:   status,
			Services: services,
		},
	}

	if status == "unhealthy" {
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
