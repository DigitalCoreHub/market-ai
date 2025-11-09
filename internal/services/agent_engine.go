package services

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/1batu/market-ai/internal/ai"
	"github.com/1batu/market-ai/internal/datasources/fusion"
	"github.com/1batu/market-ai/internal/models"
	"github.com/1batu/market-ai/internal/websocket"
)

// AgentEngine ajanlar için otonom ticareti düzenler
type AgentEngine struct {
	db             *pgxpool.Pool
	redis          *redis.Client
	hub            *websocket.Hub
	tradingEngine  *TradingEngine
	riskManager    *RiskManager
	newsAggregator *NewsAggregator
	aiClients      map[uuid.UUID]ai.Client
	minInterval    time.Duration
	maxInterval    time.Duration

	// v0.5 bağlam
	fusionService  *fusion.Service
	contextSymbols []string
}

// NewAgentEngine yeni bir ajan motoru oluşturur
func NewAgentEngine(
	db *pgxpool.Pool,
	redis *redis.Client,
	hub *websocket.Hub,
	tradingEngine *TradingEngine,
	riskManager *RiskManager,
	newsAggregator *NewsAggregator,
	minInterval, maxInterval time.Duration,
) *AgentEngine {
	return &AgentEngine{
		db:             db,
		redis:          redis,
		hub:            hub,
		tradingEngine:  tradingEngine,
		riskManager:    riskManager,
		newsAggregator: newsAggregator,
		aiClients:      make(map[uuid.UUID]ai.Client),
		minInterval:    minInterval,
		maxInterval:    maxInterval,
	}
}

// SetFusionService piyasa bağlamı füzyon servisini enjekte eder
func (ae *AgentEngine) SetFusionService(fs *fusion.Service) { ae.fusionService = fs }

// SetContextSymbols piyasa bağlamı için kullanılacak sembolleri yapılandırır
func (ae *AgentEngine) SetContextSymbols(symbols []string) { ae.contextSymbols = symbols }

// RegisterAgent bir ajan için YZ istemcisi kaydeder
func (ae *AgentEngine) RegisterAgent(agentID uuid.UUID, client ai.Client) {
	ae.aiClients[agentID] = client
	log.Info().
		Str("agent_id", agentID.String()).
		Str("model", client.GetModelName()).
		Msg("AI agent registered")
}

// Start otonom ajan karar döngüsünü başlatır
func (ae *AgentEngine) Start(ctx context.Context) {
	log.Info().
		Dur("min_interval", ae.minInterval).
		Dur("max_interval", ae.maxInterval).
		Msg("Agent engine started with random decision intervals")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Agent engine stopped")
			return
		default:
			// Rastgele uyku süresini hesapla
			randomDuration := time.Duration(rand.Int63n(int64(ae.maxInterval-ae.minInterval))) + ae.minInterval

			log.Debug().Dur("next_decision_in", randomDuration).Msg("Waiting for next decision cycle")
			time.Sleep(randomDuration)

			// Tüm ajanları işle
			ae.processAllAgents(ctx)
		}
	}
}

// processAllAgents tüm aktif ajanlar için kararlar verir
func (ae *AgentEngine) processAllAgents(ctx context.Context) {
	// Tüm aktif ajanları al
	query := `SELECT id, name, model, current_balance FROM agents WHERE status = 'active'`
	rows, err := ae.db.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch active agents")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var agentID uuid.UUID
		var agentName, model string
		var balance float64

		if err := rows.Scan(&agentID, &agentName, &model, &balance); err != nil {
			log.Error().Err(err).Msg("Failed to scan agent")
			continue
		}

		// YZ istemcisinin var olup olmadığını kontrol et
		aiClient, exists := ae.aiClients[agentID]
		if !exists {
			log.Warn().Str("agent_id", agentID.String()).Msg("No AI client registered")
			continue
		}

		// Goroutine içinde işle
		go ae.processAgentDecision(ctx, agentID, agentName, balance, aiClient)
	}
}

// processAgentDecision tek bir ajan için ticaret kararı verir
func (ae *AgentEngine) processAgentDecision(
	ctx context.Context,
	agentID uuid.UUID,
	agentName string,
	balance float64,
	aiClient ai.Client,
) {
	log.Debug().Str("agent", agentName).Msg("Processing agent decision")

	// "Düşünüyor" durumunu yayınla
	ae.hub.BroadcastMessage("agent_thinking", map[string]interface{}{
		"agent_id":   agentID,
		"agent_name": agentName,
		"timestamp":  time.Now().Unix(),
	})

	// Karar için veri topla
	decisionReq, err := ae.gatherDecisionData(ctx, agentID, agentName, balance)
	if err != nil {
		log.Error().Err(err).Str("agent", agentName).Msg("Failed to gather decision data")
		return
	}

	// Prompt oluştur
	prompt := ai.BuildDecisionPrompt(decisionReq)

	// YZ kararını al
	aiDecision, err := aiClient.GetTradingDecision(ctx, prompt)
	if err != nil {
		log.Error().Err(err).Str("agent", agentName).Msg("Failed to get AI decision")
		return
	}

	log.Info().
		Str("agent", agentName).
		Str("action", aiDecision.Action).
		Str("stock", aiDecision.StockSymbol).
		Float64("confidence", aiDecision.Confidence).
		Msg("AI decision received")

	// Kararı kaydet
	decisionID, err := ae.storeDecision(ctx, agentID, aiDecision)
	if err != nil {
		log.Error().Err(err).Msg("Failed to store decision")
		return
	}

	// Kararı yayınla
	ae.hub.BroadcastMessage("agent_decision", map[string]interface{}{
		"agent_id":          agentID,
		"agent_name":        agentName,
		"decision_id":       decisionID,
		"action":            aiDecision.Action,
		"stock_symbol":      aiDecision.StockSymbol,
		"quantity":          aiDecision.Quantity,
		"reasoning_summary": aiDecision.ReasoningSummary,
		"confidence":        aiDecision.Confidence,
		"risk_level":        aiDecision.RiskLevel,
		"thinking_steps":    aiDecision.ThinkingSteps,
		"timestamp":         time.Now().Unix(),
	})

	// HOLD değilse işlemi gerçekleştir
	if aiDecision.Action != "HOLD" {
		// Risk yöneticisi ile doğrula
		if err := ae.riskManager.ValidateTrade(ctx, agentID, aiDecision); err != nil {
			log.Warn().Err(err).Str("agent", agentName).Msg("Trade rejected by risk manager")
			ae.hub.BroadcastMessage("trade_rejected", map[string]interface{}{
				"agent_id":   agentID,
				"agent_name": agentName,
				"reason":     err.Error(),
				"timestamp":  time.Now().Unix(),
			})
			return
		}

		// İşlemi gerçekleştir
		tradeReq := models.TradeRequest{
			AgentID:     agentID,
			StockSymbol: aiDecision.StockSymbol,
			TradeType:   aiDecision.Action,
			Quantity:    aiDecision.Quantity,
			Reasoning:   aiDecision.ReasoningSummary,
		}

		trade, err := ae.tradingEngine.ExecuteTrade(ctx, tradeReq)
		if err != nil {
			log.Error().Err(err).Str("agent", agentName).Msg("Failed to execute trade")
			return
		}

		log.Info().
			Str("agent", agentName).
			Str("trade_id", trade.ID.String()).
			Msg("Trade executed successfully")

		// İşlemi yayınla
		ae.hub.BroadcastMessage("trade_executed", trade)
	} else {
		// HOLD action - no trade executed
		log.Debug().Str("agent", agentName).Msg("Agent decided to HOLD - no trade executed")
	}
}

// gatherDecisionData bir karar için gereken tüm verileri toplar
func (ae *AgentEngine) gatherDecisionData(
	ctx context.Context,
	agentID uuid.UUID,
	agentName string,
	balance float64,
) (*ai.DecisionRequest, error) {
	req := &ai.DecisionRequest{
		AgentID:        agentID.String(),
		AgentName:      agentName,
		CurrentBalance: balance,
		Strategy:       "balanced",
	}

	// Portföyü al (güncel fiyat ile hesapla)
	portfolioQuery := `
		SELECT p.stock_symbol, p.quantity, p.avg_buy_price,
			   COALESCE(p.quantity * s.current_price, 0) as current_value,
			   COALESCE(p.quantity * s.current_price - p.total_invested, 0) as profit_loss
		FROM portfolio p
		JOIN stocks s ON s.symbol = p.stock_symbol
		WHERE p.agent_id = $1
	`
	rows, err := ae.db.Query(ctx, portfolioQuery, agentID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p models.Portfolio
			if err := rows.Scan(&p.StockSymbol, &p.Quantity, &p.AvgBuyPrice, &p.CurrentValue, &p.ProfitLoss); err != nil {
				continue
			}
			req.Portfolio = append(req.Portfolio, p)
		}
	}

	// Hisseleri al
	stocksQuery := `SELECT symbol, name, current_price, change_percent, volume FROM stocks ORDER BY symbol LIMIT 20`
	rows, err = ae.db.Query(ctx, stocksQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s models.Stock
			if err := rows.Scan(&s.Symbol, &s.Name, &s.CurrentPrice, &s.ChangePercent, &s.Volume); err != nil {
				continue
			}
			req.Stocks = append(req.Stocks, s)
		}
	}

	// Son işlemleri al
	tradesQuery := `
		SELECT stock_symbol, trade_type, quantity, price, reasoning, created_at
		FROM trades WHERE agent_id = $1 ORDER BY created_at DESC LIMIT 5
	`
	rows, err = ae.db.Query(ctx, tradesQuery, agentID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t models.Trade
			if err := rows.Scan(&t.StockSymbol, &t.TradeType, &t.Quantity, &t.Price, &t.Reasoning, &t.CreatedAt); err != nil {
				continue
			}
			req.RecentTrades = append(req.RecentTrades, t)
		}
	}

	// Toplayıcıdan en son haberleri al
	if latestNews, err := ae.newsAggregator.GetLatestNews(ctx); err == nil {
		req.News = latestNews
		req.NewsCount = len(latestNews)
	} else {
		req.News = []models.NewsArticle{}
		req.NewsCount = 0
	}

	// Piyasa Bağlamı (füzyonlanmış) - füzyon servisi mevcutsa isteğe bağlı
	if ae.fusionService != nil && len(ae.contextSymbols) > 0 {
		if mc, err := ae.fusionService.MarketContext(ctx, ae.contextSymbols); err == nil && mc != nil {
			req.MCPrices = mc.Prices
			req.MCSentiments = mc.StockSentiments
			if len(mc.Tweets) > 0 {
				req.MCTopTweets = mc.Tweets
			}
			req.MCNotes = "Context aggregated from multi-source feed; consider sentiment extremes and sudden volume spikes."
		}
	}

	return req, nil
}

// storeDecision bir YZ kararını veritabanına kaydeder
func (ae *AgentEngine) storeDecision(ctx context.Context, agentID uuid.UUID, decision *models.AIDecision) (uuid.UUID, error) {
	decisionID := uuid.New()

	// Piyasa bağlamını marshal et
	marketContext, _ := json.Marshal(map[string]interface{}{
		"timestamp": time.Now(),
	})

	// Risk skorunu hesapla
	riskScore := 100.0 - decision.Confidence

	query := `
		INSERT INTO agent_decisions (
			id, agent_id, stock_symbol, decision, quantity, target_price, stop_loss,
			reasoning_full, reasoning_summary, confidence_score, risk_score, risk_level,
			market_context, outcome
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := ae.db.Exec(ctx, query,
		decisionID, agentID, decision.StockSymbol, decision.Action, decision.Quantity,
		decision.TargetPrice, decision.StopLoss, decision.ReasoningFull, decision.ReasoningSummary,
		decision.Confidence, riskScore, decision.RiskLevel, string(marketContext), "pending",
	)
	if err != nil {
		return uuid.Nil, err
	}

	// Düşünme adımlarını kaydet
	for i, step := range decision.ThinkingSteps {
		thoughtQuery := `
			INSERT INTO agent_thoughts (agent_id, decision_id, step_number, step_name, thought)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, _ = ae.db.Exec(ctx, thoughtQuery, agentID, decisionID, i+1, step.Step, step.Observation)
	}

	return decisionID, nil
}
