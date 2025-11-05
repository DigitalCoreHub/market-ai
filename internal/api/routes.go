package api

import (
	"github.com/1batu/market-ai/internal/api/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes tüm HTTP route'ları tanımlar
func SetupRoutes(app *fiber.App, healthHandler *handlers.HealthHandler) {
	app.Get("/health", healthHandler.Check)

	v1 := app.Group("/api/v1")
	v1.Get("/ping", healthHandler.Ping)
}
