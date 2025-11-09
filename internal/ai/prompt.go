package ai

import (
	"fmt"
	"strings"
	"time"
)

// GetSystemPrompt returns the system prompt for AI trading
func GetSystemPrompt() string {
	return `You are a professional BIST (Borsa Istanbul) trader with 10+ years of experience.
Your goal is to maximize portfolio returns while managing risk effectively.

CRITICAL: You must respond ONLY with valid JSON in this exact format:
{
  "action": "BUY|SELL|HOLD",
  "stock_symbol": "THYAO",
  "quantity": 50,
  "target_price": 250.00,
  "stop_loss": 240.00,
  "reasoning_summary": "Brief one-line explanation",
  "reasoning_full": "Detailed multi-step analysis",
  "confidence": 85,
  "risk_level": "low|medium|high",
  "thinking_steps": [
    {
      "step": "Market Analysis",
      "observation": "Price trending up with high volume"
    },
    {
      "step": "Technical Indicators",
      "observation": "RSI at 45, MACD positive crossover"
    },
    {
      "step": "Decision",
      "observation": "Strong buy signal with good risk/reward"
    }
  ]
}

Rules:
- NEVER invest more than 5% of balance in a single trade
- Always set stop loss (max 3% loss per trade)
- Only trade when confidence > 70%
- Consider portfolio diversification
- If uncertain, choose HOLD
- News context is CRITICAL - major news can cause 10%+ moves
- Always factor in news sentiment when making decisions`
}

// BuildDecisionPrompt builds the prompt for AI decision making
func BuildDecisionPrompt(req *DecisionRequest) string {
	var sb strings.Builder

	// Agent Info
	sb.WriteString("=== AGENT STATUS ===\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", req.AgentName))
	sb.WriteString(fmt.Sprintf("Available Balance: %.2f TL\n", req.CurrentBalance))
	sb.WriteString(fmt.Sprintf("Strategy: %s\n\n", req.Strategy))

	// Portfolio
	sb.WriteString("=== CURRENT PORTFOLIO ===\n")
	if len(req.Portfolio) == 0 {
		sb.WriteString("No positions\n")
	} else {
		for _, p := range req.Portfolio {
			sb.WriteString(fmt.Sprintf("- %s: %d lots @ %.2f TL avg (Current Value: %.2f TL, P/L: %.2f TL)\n",
				p.StockSymbol, p.Quantity, p.AvgBuyPrice, p.CurrentValue, p.ProfitLoss))
		}
	}
	sb.WriteString("\n")

	// Available Stocks
	sb.WriteString("=== AVAILABLE STOCKS ===\n")
	for _, s := range req.Stocks {
		direction := "↑"
		if s.ChangePercent < 0 {
			direction = "↓"
		}
		sb.WriteString(fmt.Sprintf("- %s (%s): %.2f TL (%s%.2f%%) | Volume: %d\n",
			s.Symbol, s.Name, s.CurrentPrice, direction, s.ChangePercent, s.Volume))
	}
	sb.WriteString("\n")

	// === AI QUANTITY CONTROL (Position Sizing Guidance) ===
	maxTradeAmount := req.CurrentBalance * 0.05
	sb.WriteString("=== 💰 YOUR TRADING AUTHORITY ===\n")
	sb.WriteString(fmt.Sprintf("Current Balance: %.2f TL\n", req.CurrentBalance))
	sb.WriteString(fmt.Sprintf("Max Per Trade: %.2f TL (5%% rule)\n\n", maxTradeAmount))
	sb.WriteString("YOU have FULL CONTROL over:\n")
	sb.WriteString("1. Which stock to trade\n")
	sb.WriteString("2. How many LOTS to buy/sell (MUST be an exact integer)\n")
	sb.WriteString("3. When to trade (timing)\n\n")

	sb.WriteString("=== 📊 DYNAMIC UNIVERSE SNAPSHOT ===\n")
	if len(req.Stocks) == 0 {
		sb.WriteString("No active stocks available.\n\n")
	} else {
		for _, stock := range req.Stocks {
			if stock.CurrentPrice <= 0 {
				continue
			}
			maxLots := int(maxTradeAmount / stock.CurrentPrice)
			sb.WriteString(fmt.Sprintf("- %s (%s): Price: %.2f TL | Max lots: %d (≈ %.2f TL)\n",
				stock.Symbol, stock.Name, stock.CurrentPrice, maxLots, float64(maxLots)*stock.CurrentPrice))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("=== ⚠️ QUANTITY CALCULATION RULES ===\n")
	sb.WriteString("When deciding quantity, consider:\n")
	sb.WriteString("1. Stock price (higher price = fewer lots)\n")
	sb.WriteString("2. Your conviction level (higher confidence = more lots)\n")
	sb.WriteString("3. Risk management (don't exceed 5% per trade)\n")
	sb.WriteString("4. Portfolio diversification (avoid over-concentration)\n")
	sb.WriteString("5. Market volatility (more volatile = smaller position)\n\n")

	// Market Data (Recent candles)
	if len(req.MarketData) > 0 {
		sb.WriteString("=== RECENT MARKET DATA (Last 5 candles) ===\n")
		for i, md := range req.MarketData {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("%s - O:%.2f H:%.2f L:%.2f C:%.2f V:%d\n",
				md.Timestamp.Format("15:04"), md.OpenPrice, md.HighPrice, md.LowPrice, md.ClosePrice, md.Volume))
		}
		sb.WriteString("\n")
	}

	// *** NEWS SECTION - Most important for decision making ***
	sb.WriteString("=== 📰 LATEST ECONOMIC NEWS (Last 3 Hours) ===\n")
	if len(req.News) == 0 {
		sb.WriteString("No significant news in the last 3 hours.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Total: %d news articles (showing top 10)\n\n", req.NewsCount))

		// Show max 10 news
		newsCount := len(req.News)
		if newsCount > 10 {
			newsCount = 10
		}

		for i := 0; i < newsCount; i++ {
			article := req.News[i]
			timeAgo := time.Since(article.PublishedAt)

			// Format time ago
			timeStr := formatDuration(timeAgo)

			sb.WriteString(fmt.Sprintf("%d. [%s] [%s] %s\n",
				i+1,
				timeStr,
				article.Source,
				article.Title,
			))

			// Add description if available
			if article.Description != "" && len(article.Description) > 0 {
				desc := article.Description
				if len(desc) > 150 {
					desc = desc[:150] + "..."
				}
				sb.WriteString(fmt.Sprintf("   📝 %s\n", desc))
			}

			// Show related stocks if available
			if len(article.RelatedStocks) > 0 {
				sb.WriteString(fmt.Sprintf("   🎯 Related: %s\n", strings.Join(article.RelatedStocks, ", ")))
			}

			sb.WriteString("\n")
		}
	}

	sb.WriteString("=== ⚠️ IMPORTANT - Consider News Impact ===\n")
	sb.WriteString("- How might these news affect BIST stocks?\n")
	sb.WriteString("- Are there any company-specific announcements?\n")
	sb.WriteString("- What's the overall market sentiment from news?\n")
	sb.WriteString("- Any economic indicators or policy changes?\n")
	sb.WriteString("- Breaking news might cause significant price movements!\n\n")

	// Recent Trades
	if len(req.RecentTrades) > 0 {
		sb.WriteString("=== YOUR RECENT TRADES ===\n")
		for i, t := range req.RecentTrades {
			if i >= 3 {
				break
			}

			// === MARKET CONTEXT (v0.5) ===
			if len(req.MCPrices) > 0 || len(req.MCSentiments) > 0 || len(req.MCTopTweets) > 0 {
				sb.WriteString("\n=== 📊 MARKET CONTEXT (Fused: Yahoo + Scraper + Twitter) ===\n")
				if len(req.MCPrices) > 0 {
					sb.WriteString("Prices (delayed ~15m):\n")
					max := 5
					if len(req.MCPrices) < max {
						max = len(req.MCPrices)
					}
					for i := 0; i < max; i++ {
						p := req.MCPrices[i]
						sb.WriteString(fmt.Sprintf("- %s: %.2f TL (O:%.2f H:%.2f L:%.2f V:%d)\n", p.Symbol, p.Price, p.Open, p.High, p.Low, p.Volume))
					}
				}
				if len(req.MCSentiments) > 0 {
					sb.WriteString("Sentiment summary (tweets):\n")
					shown := 0
					for sym, agg := range req.MCSentiments {
						sb.WriteString(fmt.Sprintf("- %s: avg=%.2f pos=%d neu=%d neg=%d\n", sym, agg.AvgSentiment, agg.PositiveCount, agg.NeutralCount, agg.NegativeCount))
						shown++
						if shown >= 5 {
							break
						}
					}
				}
				if len(req.MCTopTweets) > 0 {
					sb.WriteString("Top tweets:\n")
					max := 3
					if len(req.MCTopTweets) < max {
						max = len(req.MCTopTweets)
					}
					for i := 0; i < max; i++ {
						t := req.MCTopTweets[i]
						sb.WriteString(fmt.Sprintf("- @%s (impact %.2f): %s\n", t.Author, t.ImpactScore, truncate(t.Text, 140)))
					}
				}
				if req.MCNotes != "" {
					sb.WriteString(req.MCNotes + "\n")
				}
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("- %s %s %d lots @ %.2f TL (%s)\n",
				t.CreatedAt.Format("15:04"), t.TradeType, t.Quantity, t.Price, t.Reasoning))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("=== QUESTION ===\n")
	sb.WriteString("Based on the above information, what trading decision should you make RIGHT NOW?\n")
	sb.WriteString("Consider: technical analysis, risk management, portfolio balance, market trends, AND NEWS IMPACT.\n")
	sb.WriteString("Also consider market context (multi-source sentiment and recent signals).\n")
	sb.WriteString("Remember to respond ONLY with valid JSON in the specified format.\n")

	return sb.String()
}

// formatDuration formats time duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd ago", days)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
