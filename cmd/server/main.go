package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/1batu/market-ai/internal/api"
	"github.com/1batu/market-ai/internal/api/handlers"
	"github.com/1batu/market-ai/internal/config"
	"github.com/1batu/market-ai/internal/database"
	"github.com/1batu/market-ai/internal/services"
	"github.com/1batu/market-ai/internal/websocket"
	"github.com/1batu/market-ai/pkg/logger"
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

	// marketSimulator := services.NewMarketSimulator(db, hub)
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// go marketSimulator.Start(ctx)
	// log.Info().Msg("Market simulator started")

	app := api.NewServer(cfg)

	healthHandler := handlers.NewHealthHandler(db, redisClient)
	agentHandler := handlers.NewAgentHandler(db)
	stockHandler := handlers.NewStockHandler(db)
	tradingEngine := services.NewTradingEngine(db)
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
	if err := app.ShutdownWithContext(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
