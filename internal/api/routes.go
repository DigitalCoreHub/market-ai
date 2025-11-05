package api

import (
	"github.com/1batu/market-ai/internal/api/handlers"
	"github.com/1batu/market-ai/internal/websocket"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	healthHandler *handlers.HealthHandler,
	agentHandler *handlers.AgentHandler,
	stockHandler *handlers.StockHandler,
	tradeHandler *handlers.TradeHandler,
	hub *websocket.Hub,
) {
	app.Get("/health", healthHandler.Check)
	app.Get("/ws", websocket.HandleWebSocket(hub))

	v1 := app.Group("/api/v1")
	v1.Get("/ping", healthHandler.Ping)

	agents := v1.Group("/agents")
	agents.Get("/", agentHandler.GetAll)
	agents.Get("/:id", agentHandler.GetByID)
	agents.Get("/:id/metrics", agentHandler.GetMetrics)
	agents.Get("/:id/portfolio", agentHandler.GetPortfolio)

	stocks := v1.Group("/stocks")
	stocks.Get("/", stockHandler.GetAll)
	stocks.Get("/:symbol", stockHandler.GetBySymbol)
	stocks.Get("/:symbol/history", stockHandler.GetHistory)

	trades := v1.Group("/trades")
	trades.Post("/", tradeHandler.Execute)
	trades.Get("/", tradeHandler.GetHistory)
}
