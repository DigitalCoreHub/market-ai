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

	// === AGENT ENGINE (decision intervals) ===
	minDec := 30 * time.Second
	maxDec := 60 * time.Second
	if cfg.AI.BudgetMode {
		// Stretch intervals to cut daily call volume roughly in half/third
		minDec = 60 * time.Second
		maxDec = 120 * time.Second
	}
	agentEngine := services.NewAgentEngine(
		db,
		redisClient,
		hub,
		tradingEngine,
		riskManager,
		newsAggregator,
		minDec,
		maxDec,
	)

	// === AI CLIENTS ===
	openaiClient := ai.NewOpenAIClient(cfg.AI.OpenAIKey, cfg.AI.GPTModel)
	gpt4MiniClient := ai.NewOpenAIClient(cfg.AI.OpenAIKey, cfg.AI.GPT4MiniModel)
	claudeClient := ai.NewAnthropicClient(cfg.AI.AnthropicKey, cfg.AI.ClaudeModel)
	var geminiClient *ai.GoogleClient
	if c, err := ai.NewGoogleClient(cfg.AI.GoogleKey, cfg.AI.GoogleModel); err == nil {
		geminiClient = c
	} else {
		log.Warn().Err(err).Msg("Gemini client disabled")
	}
	deepseekClient := ai.NewDeepSeekClient(cfg.AI.DeepSeekKey, cfg.AI.DeepSeekModel)
	groqClient := ai.NewGroqClient(cfg.AI.GroqKey, cfg.AI.GroqModel)
	mistralClient := ai.NewMistralClient(cfg.AI.MistralKey, cfg.AI.MistralModel)
	xaiClient := ai.NewXAIClient(cfg.AI.XAIKey, cfg.AI.XAIModel)

	// Get agent IDs from database and register AI clients (conditional by cost flags)
	// Premium detection: GPT-4, Claude Sonnet/Opus, Grok
	premiumModels := map[string]bool{
		cfg.AI.GPTModel:    true,
		cfg.AI.ClaudeModel: true,
		cfg.AI.XAIModel:    true,
	}

	// Register all known agents by name substrings
	rows, qerr := db.Query(ctx, "SELECT id, name FROM agents WHERE status = 'active'")
	if qerr == nil {
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				continue
			}
			switch {
			case strings.Contains(strings.ToLower(name), "gpt-4o mini") || strings.Contains(strings.ToLower(name), "gpt-4o-mini"):
				if gpt4MiniClient != nil {
					agentEngine.RegisterAgent(id, gpt4MiniClient)
				}
			case strings.Contains(strings.ToLower(name), "gpt"):
				if openaiClient != nil && (cfg.AI.EnablePremiumModels || !premiumModels[openaiClient.GetModelName()]) {
					agentEngine.RegisterAgent(id, openaiClient)
				}
			case strings.Contains(strings.ToLower(name), "claude"):
				if claudeClient != nil && (cfg.AI.EnablePremiumModels || !premiumModels[claudeClient.GetModelName()]) {
					agentEngine.RegisterAgent(id, claudeClient)
				}
			case strings.Contains(strings.ToLower(name), "gemini") && geminiClient != nil:
				// Gemini treated as mid-tier: always allowed unless explicitly disabled by empty key
				agentEngine.RegisterAgent(id, geminiClient)
			case strings.Contains(strings.ToLower(name), "deepseek"):
				if deepseekClient != nil {
					agentEngine.RegisterAgent(id, deepseekClient)
				}
			case strings.Contains(strings.ToLower(name), "llama"):
				if groqClient != nil {
					agentEngine.RegisterAgent(id, groqClient)
				}
			case strings.Contains(strings.ToLower(name), "mixtral"):
				if mistralClient != nil {
					agentEngine.RegisterAgent(id, mistralClient)
				}
			case strings.Contains(strings.ToLower(name), "grok"):
				if xaiClient != nil && (cfg.AI.EnablePremiumModels || !premiumModels[xaiClient.GetModelName()]) {
					agentEngine.RegisterAgent(id, xaiClient)
				}
			}
		}
	} else {
		log.Warn().Err(qerr).Msg("Failed to query agents for registration")
	}

	// Start agent engine
	go agentEngine.Start(ctx)
	log.Info().Msg("Agent engine started (30-60 sec decision cycle)")

	app := api.NewServer(cfg)

	healthHandler := handlers.NewHealthHandler(db, redisClient)
	agentHandler := handlers.NewAgentHandler(db)
	stockHandler := handlers.NewStockHandler(db)
	tradeHandler := handlers.NewTradeHandler(db, tradingEngine)
	leaderboardHandler := handlers.NewLeaderboardHandler(db)
	roiHistoryHandler := handlers.NewROIHistoryHandler(db)

	api.SetupRoutes(app, healthHandler, agentHandler, stockHandler, tradeHandler, leaderboardHandler, roiHistoryHandler, hub)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Info().Str("port", cfg.Server.Port).Msg("Server starting")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// === LEADERBOARD SERVICE (interval from env, default 60s) ===
	lbInterval := time.Duration(cfg.Leaderboard.UpdateInterval)
	if lbInterval <= 0 {
		lbInterval = 60
	}
	leaderboardSvc := services.NewLeaderboardService(db, hub, lbInterval*time.Second)
	go leaderboardSvc.Start(ctx)

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
