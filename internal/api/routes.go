package api

import (
	"github.com/1batu/market-ai/internal/api/handlers"
	"github.com/1batu/market-ai/internal/middleware"
	"github.com/1batu/market-ai/internal/websocket"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	healthHandler *handlers.HealthHandler,
	agentHandler *handlers.AgentHandler,
	stockHandler *handlers.StockHandler,
	tradeHandler *handlers.TradeHandler,
	leaderboardHandler *handlers.LeaderboardHandler,
	roiHistoryHandler *handlers.ROIHistoryHandler,
	marketCtxHandler *handlers.MarketContextHandler,
	debugHandler *handlers.DebugDataHandler,
	metricsHandler *handlers.MetricsHandler,
	universeHandler *handlers.UniverseHandler,
	newsHandler *handlers.NewsHandler,
	authHandler *handlers.AuthHandler,
	hub *websocket.Hub,
) {
	app.Get("/health", healthHandler.Check)
	app.Get("/ws", websocket.HandleWebSocket(hub))

	v1 := app.Group("/api/v1")
	v1.Get("/ping", healthHandler.Ping)

	// Auth routes (public)
	auth := v1.Group("/auth")
	auth.Post("/login", authHandler.Login)

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

	// Leaderboard
	v1.Get("/leaderboard", leaderboardHandler.GetLeaderboard)
	v1.Get("/leaderboard/roi-history", roiHistoryHandler.GetAllAgentsROIHistory)

	// Market context (v0.5)
	v1.Get("/market/context", marketCtxHandler.GetContext)

	// Metrics endpoint (observability)
	v1.Get("/metrics", metricsHandler.Get)
	v1.Get("/metrics/prometheus", metricsHandler.GetPrometheus)

	// Dynamic stock universe endpoints
	universe := v1.Group("/universe")
	universe.Get("/active", universeHandler.GetActiveStocks)
	universe.Post("/update", middleware.APIKeyOrJWTProtected(), universeHandler.TriggerUniverseUpdate) // Protected (API key or JWT)
	universe.Get("/history", universeHandler.GetUniverseHistory)

	// News endpoints
	news := v1.Group("/news")
	news.Post("/fetch", middleware.APIKeyOrJWTProtected(), newsHandler.TriggerNewsFetch) // Protected (API key or JWT)
	news.Get("/latest", newsHandler.GetLatestNews)                                       // Public

	// Debug endpoints (per-source)
	dbg := v1.Group("/debug")
	dbg.Get("/yahoo", debugHandler.GetYahoo)
	dbg.Get("/scraper", debugHandler.GetScraper)
	dbg.Get("/tweets", debugHandler.GetTweets)
}
