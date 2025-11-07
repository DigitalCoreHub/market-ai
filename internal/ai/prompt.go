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
			sb.WriteString(fmt.Sprintf("- %s %s %d lots @ %.2f TL (%s)\n",
				t.CreatedAt.Format("15:04"), t.TradeType, t.Quantity, t.Price, t.Reasoning))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("=== QUESTION ===\n")
	sb.WriteString("Based on the above information, what trading decision should you make RIGHT NOW?\n")
	sb.WriteString("Consider: technical analysis, risk management, portfolio balance, market trends, AND NEWS IMPACT.\n")
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
