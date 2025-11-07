package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/1batu/market-ai/internal/ai"
	"github.com/1batu/market-ai/internal/api"
	"github.com/1batu/market-ai/internal/api/handlers"
	"github.com/1batu/market-ai/internal/config"
	"github.com/1batu/market-ai/internal/database"
	"github.com/1batu/market-ai/internal/services"
	"github.com/1batu/market-ai/internal/websocket"
	"github.com/1batu/market-ai/pkg/logger"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	logger.Init(cfg.Log.Level)
	log.Info().Msg("Logger initialized")

	db, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()
	log.Info().Msg("Connected to PostgreSQL")

	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()
	log.Info().Msg("Connected to Redis")

	hub := websocket.NewHub()
	go hub.Run()
	log.Info().Msg("WebSocket hub started")

	// Initialize v0.3 services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// === NEWS AGGREGATOR (30 min cycle) ===
	rssFeeds := strings.Split(cfg.News.Feeds, ",")
	newsAggregator := services.NewNewsAggregator(
		db,
		redisClient,
		hub,
		cfg.News.APIKey,
		rssFeeds,
		time.Duration(cfg.News.UpdateInterval)*time.Minute,
		time.Duration(cfg.News.CacheTTL)*time.Minute,
	)
	go newsAggregator.Start(ctx)
	log.Info().Msg("News aggregator started (30 min cycle)")

	// === TRADING ENGINE & RISK MANAGER ===
	tradingEngine := services.NewTradingEngine(db)
	riskManager := services.NewRiskManager(db, 5.0, 20.0, 70.0)

	// === AGENT ENGINE (30-60 sec decisions) ===
	agentEngine := services.NewAgentEngine(
		db,
		redisClient,
		hub,
		tradingEngine,
		riskManager,
		newsAggregator,
		30*time.Second,
		60*time.Second,
	)

	// === AI CLIENTS ===
	openaiClient := ai.NewOpenAIClient(cfg.AI.OpenAIKey, cfg.AI.GPTModel)
	claudeClient := ai.NewAnthropicClient(cfg.AI.AnthropicKey, cfg.AI.ClaudeModel)

	// Get agent IDs from database and register AI clients
	var gptAgentID, claudeAgentID uuid.UUID
	err = db.QueryRow(ctx, "SELECT id FROM agents WHERE name ILIKE '%GPT%' LIMIT 1").Scan(&gptAgentID)
	if err == nil {
		agentEngine.RegisterAgent(gptAgentID, openaiClient)
		log.Info().Str("agent", "GPT-4 Turbo").Msg("AI client registered")
	}

	err = db.QueryRow(ctx, "SELECT id FROM agents WHERE name ILIKE '%Claude%' LIMIT 1").Scan(&claudeAgentID)
	if err == nil {
		agentEngine.RegisterAgent(claudeAgentID, claudeClient)
		log.Info().Str("agent", "Claude 3").Msg("AI client registered")
	}

	// Start agent engine
	go agentEngine.Start(ctx)
	log.Info().Msg("Agent engine started (30-60 sec decision cycle)")

	app := api.NewServer(cfg)

	healthHandler := handlers.NewHealthHandler(db, redisClient)
	agentHandler := handlers.NewAgentHandler(db)
	stockHandler := handlers.NewStockHandler(db)
	tradeHandler := handlers.NewTradeHandler(db, tradingEngine)

	api.SetupRoutes(app, healthHandler, agentHandler, stockHandler, tradeHandler, hub)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Info().Str("port", cfg.Server.Port).Msg("Server starting")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	cancel() // Stop all services
	if err := app.ShutdownWithContext(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
