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
	"github.com/1batu/market-ai/internal/models"
	"github.com/1batu/market-ai/internal/websocket"
)

// AgentEngine orchestrates autonomous trading for agents
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
}

// NewAgentEngine creates a new agent engine
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

// RegisterAgent registers an AI client for an agent
func (ae *AgentEngine) RegisterAgent(agentID uuid.UUID, client ai.Client) {
	ae.aiClients[agentID] = client
	log.Info().
		Str("agent_id", agentID.String()).
		Str("model", client.GetModelName()).
		Msg("AI agent registered")
}

// Start begins the autonomous agent decision loop
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
			// Calculate random sleep duration
			randomDuration := time.Duration(rand.Int63n(int64(ae.maxInterval-ae.minInterval))) + ae.minInterval

			log.Debug().Dur("next_decision_in", randomDuration).Msg("Waiting for next decision cycle")
			time.Sleep(randomDuration)

			// Process all agents
			ae.processAllAgents(ctx)
		}
	}
}

// processAllAgents makes decisions for all active agents
func (ae *AgentEngine) processAllAgents(ctx context.Context) {
	// Get all active agents
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

		// Check if AI client exists
		aiClient, exists := ae.aiClients[agentID]
		if !exists {
			log.Warn().Str("agent_id", agentID.String()).Msg("No AI client registered")
			continue
		}

		// Process in goroutine
		go ae.processAgentDecision(ctx, agentID, agentName, balance, aiClient)
	}
}

// processAgentDecision makes a trading decision for a single agent
func (ae *AgentEngine) processAgentDecision(
	ctx context.Context,
	agentID uuid.UUID,
	agentName string,
	balance float64,
	aiClient ai.Client,
) {
	log.Debug().Str("agent", agentName).Msg("Processing agent decision")

	// Broadcast "thinking" status
	ae.hub.BroadcastMessage("agent_thinking", map[string]interface{}{
		"agent_id":   agentID,
		"agent_name": agentName,
		"timestamp":  time.Now().Unix(),
	})

	// Gather data for decision
	decisionReq, err := ae.gatherDecisionData(ctx, agentID, agentName, balance)
	if err != nil {
		log.Error().Err(err).Str("agent", agentName).Msg("Failed to gather decision data")
		return
	}

	// Build prompt
	prompt := ai.BuildDecisionPrompt(decisionReq)

	// Get AI decision
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

	// Store decision
	decisionID, err := ae.storeDecision(ctx, agentID, aiDecision)
	if err != nil {
		log.Error().Err(err).Msg("Failed to store decision")
		return
	}

	// Broadcast decision
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

	// Execute trade if not HOLD
	if aiDecision.Action != "HOLD" {
		// Validate with risk manager
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

		// Execute trade
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

		// Broadcast trade
		ae.hub.BroadcastMessage("trade_executed", trade)
	}
}

// gatherDecisionData collects all data needed for a decision
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

	// Get portfolio
	portfolioQuery := `
		SELECT stock_symbol, quantity, avg_buy_price, COALESCE(quantity * $1, 0) as current_value,
		       COALESCE(quantity * $1 - total_invested, 0) as profit_loss
		FROM portfolio WHERE agent_id = $2
	`
	// Note: This is simplified - in production you'd join with stocks table
	rows, err := ae.db.Query(ctx, portfolioQuery, 100, agentID) // 100 is placeholder price
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

	// Get stocks
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

	// Get recent trades
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

	// Get latest news from aggregator
	if latestNews, err := ae.newsAggregator.GetLatestNews(ctx); err == nil {
		req.News = latestNews
		req.NewsCount = len(latestNews)
	} else {
		req.News = []models.NewsArticle{}
		req.NewsCount = 0
	}

	return req, nil
}

// storeDecision stores an AI decision in the database
func (ae *AgentEngine) storeDecision(ctx context.Context, agentID uuid.UUID, decision *models.AIDecision) (uuid.UUID, error) {
	decisionID := uuid.New()

	// Marshal market context
	marketContext, _ := json.Marshal(map[string]interface{}{
		"timestamp": time.Now(),
	})

	// Calculate risk score
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

	// Store thinking steps
	for i, step := range decision.ThinkingSteps {
		thoughtQuery := `
			INSERT INTO agent_thoughts (agent_id, decision_id, step_number, step_name, thought)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, _ = ae.db.Exec(ctx, thoughtQuery, agentID, decisionID, i+1, step.Step, step.Observation)
	}

	return decisionID, nil
}
