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
	"github.com/1batu/market-ai/internal/datasources/fusion"
	"github.com/1batu/market-ai/internal/datasources/scraper"
	tw "github.com/1batu/market-ai/internal/datasources/twitter"
	"github.com/1batu/market-ai/internal/datasources/yahoo"
	"github.com/1batu/market-ai/internal/middleware"
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

	// Initialize authentication
	middleware.InitAuth(cfg.Auth.JWTSecret, cfg.Auth.APIKey)
	log.Info().Msg("Authentication initialized")

	db, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()
	log.Info().Msg("Connected to PostgreSQL")

	// Run database migrations (checks if tables exist before applying)
	if err := database.RunMigrations(context.Background(), db); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}
	log.Info().Msg("Database migrations completed")

	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()
	log.Info().Msg("Connected to Redis")

	hub := websocket.NewHub()
	go hub.Run()
	log.Info().Msg("WebSocket hub started")

	// v0.3 servislerini başlat
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// === HABER TOPLAYICI (30 dk döngü) ===
	// RSS feeds'i parse et (boş string ise boş array)
	var rssFeeds []string
	if cfg.News.Feeds != "" {
		rssFeeds = strings.Split(cfg.News.Feeds, ",")
		// Trim whitespace
		for i, feed := range rssFeeds {
			rssFeeds[i] = strings.TrimSpace(feed)
		}
	}

	// Update interval kontrolü (minimum 1 dakika)
	updateInterval := time.Duration(cfg.News.UpdateInterval) * time.Minute
	if updateInterval <= 0 {
		updateInterval = 30 * time.Minute // Default: 30 minutes
	}

	// Cache TTL kontrolü (minimum 1 dakika)
	cacheTTL := time.Duration(cfg.News.CacheTTL) * time.Minute
	if cacheTTL <= 0 {
		cacheTTL = 60 * time.Minute // Default: 60 minutes
	}

	newsAggregator := services.NewNewsAggregator(
		db,
		redisClient,
		hub,
		cfg.News.APIKey,
		rssFeeds,
		updateInterval,
		cacheTTL,
	)
	go newsAggregator.Start(ctx)
	log.Info().Dur("interval", updateInterval).Msg("News aggregator started")

	// === TİCARET MOTORU & RİSK YÖNETİCİSİ ===
	tradingEngine := services.NewTradingEngine(db)
	riskManager := services.NewRiskManager(db, 5.0, 20.0, 70.0)

	// === AJAN MOTORU (karar aralıkları) ===
	minDec := 30 * time.Second
	maxDec := 60 * time.Second
	if cfg.AI.BudgetMode {
		// Günlük çağrı hacmini kabaca yarıya/üçte bire indirmek için aralıkları uzat
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

	// === YZ İSTEMCİLERİ ===
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

	// Veritabanından ajan kimliklerini al ve YZ istemcilerini kaydet (maliyet bayraklarına göre koşullu)
	// Premium tespit: GPT-4, Claude Sonnet/Opus, Grok
	premiumModels := map[string]bool{
		cfg.AI.GPTModel:    true,
		cfg.AI.ClaudeModel: true,
		cfg.AI.XAIModel:    true,
	}

	// Tüm bilinen ajanları isim alt dizelerine göre kaydet
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
				// Gemini orta seviye olarak kabul ediliyor: boş anahtar ile açıkça devre dışı bırakılmadıkça her zaman izin verilir
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

	// Ajan motorunu başlat
	go agentEngine.Start(ctx)
	log.Info().Msg("Agent engine started (30-60 sec decision cycle)")

	app := api.NewServer(cfg)

	// === v0.5 VERİ KAYNAĞI İSTEMCİLERİ & FÜZYON SERVİSİ ===
	// Çevre yapılandırmasından sembol evreni (boşsa varsayılanlara geri dön)
	symbols := []string{"THYAO", "AKBNK", "ASELS", "GARAN", "BIMAS", "KCHOL", "SISE"}
	if su := strings.TrimSpace(cfg.DataSources.SymbolUniverse); su != "" {
		parts := strings.Split(su, ",")
		var parsed []string
		for _, p := range parts {
			p = strings.ToUpper(strings.TrimSpace(p))
			if p != "" {
				parsed = append(parsed, p)
			}
		}
		if len(parsed) > 0 {
			symbols = parsed
		}
	}

	yahooClient := yahoo.NewYahooFinanceClient()
	webScraper := scraper.NewWebScraper()
	// Çalışma zamanı anahtarları doğrulama ve nazik devre dışı bırakma
	var twitterClient *tw.Client
	if cfg.DataSources.TwitterAPIKey != "" && cfg.DataSources.TwitterAPISecret != "" &&
		cfg.DataSources.TwitterAccessToken != "" && cfg.DataSources.TwitterAccessSecret != "" {
		twitterClient = tw.NewClient(
			cfg.DataSources.TwitterAPIKey,
			cfg.DataSources.TwitterAPISecret,
			cfg.DataSources.TwitterAccessToken,
			cfg.DataSources.TwitterAccessSecret,
		)
	} else {
		log.Warn().Msg("Twitter credentials missing; Twitter features disabled")
	}
	var tweetAnalyzer *tw.Analyzer
	if cfg.AI.OpenAIKey != "" {
		tweetAnalyzer = tw.NewAnalyzer(cfg.AI.OpenAIKey) // duygu analizi için OpenAI anahtarını yeniden kullan
	} else {
		log.Warn().Msg("OPENAI_API_KEY missing; tweet sentiment analysis disabled")
	}

	fusionService := fusion.New(db, yahooClient, webScraper, twitterClient, tweetAnalyzer)
	marketCtxHandler := handlers.NewMarketContextHandler(fusionService)
	debugHandler := handlers.NewDebugDataHandler(yahooClient, webScraper, twitterClient, tweetAnalyzer)
	metricsHandler := handlers.NewMetricsHandler(db)
	// Dynamic stock universe service (6h interval)
	stockUniverseSvc := services.NewStockUniverseService(db, hub, 6*time.Hour)
	go stockUniverseSvc.Start(ctx)
	universeHandler := handlers.NewUniverseHandler(db, stockUniverseSvc)
	// Prompt bağlamı için füzyon + sembolleri ajan motoruna enjekte et
	agentEngine.SetFusionService(fusionService)
	agentEngine.SetContextSymbols(symbols)

	// === HTTP İŞLEYİCİLERİ ===
	healthHandler := handlers.NewHealthHandler(db, redisClient)
	agentHandler := handlers.NewAgentHandler(db)
	stockHandler := handlers.NewStockHandler(db)
	tradeHandler := handlers.NewTradeHandler(db, tradingEngine)
	leaderboardHandler := handlers.NewLeaderboardHandler(db)
	roiHistoryHandler := handlers.NewROIHistoryHandler(db)
	newsHandler := handlers.NewNewsHandler(newsAggregator)
	authHandler := handlers.NewAuthHandler(cfg)

	api.SetupRoutes(app, healthHandler, agentHandler, stockHandler, tradeHandler, leaderboardHandler, roiHistoryHandler, marketCtxHandler, debugHandler, metricsHandler, universeHandler, newsHandler, authHandler, hub)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Info().Str("port", cfg.Server.Port).Msg("Server starting")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// === LİDER TABLOSU SERVİSİ (çevreden aralık, varsayılan 60s) ===
	lbInterval := time.Duration(cfg.Leaderboard.UpdateInterval)
	if lbInterval <= 0 {
		lbInterval = 60
	}
	leaderboardSvc := services.NewLeaderboardService(db, hub, lbInterval*time.Second)
	go leaderboardSvc.Start(ctx)

	// === PİYASA VERİSİ TOPLAYICI & DUYGU TAKİPCİSİ (v0.5) ===
	mdc := services.NewMarketDataCollector(
		fusionService,
		symbols,
		cfg.DataSources.YahooFetchInterval,
		cfg.DataSources.ScraperFetchInterval,
		cfg.DataSources.TwitterFetchInterval,
	)
	go mdc.Start(ctx)

	sentimentTracker := services.NewSentimentTracker(
		db,
		symbols,
		cfg.DataSources.SentimentUpdateInterval,
		60, // duygu toplama için pencere dakikaları (daha sonra çevre tarafından yönlendirilebilir)
	)
	go sentimentTracker.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	cancel() // Tüm servisleri durdur
	if err := app.ShutdownWithContext(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
