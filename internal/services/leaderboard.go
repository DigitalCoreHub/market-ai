package services

import (
	"context"
	"time"

	"github.com/1batu/market-ai/internal/models"
	"github.com/1batu/market-ai/internal/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// LeaderboardService recalculates and broadcasts leaderboard updates
type LeaderboardService struct {
	db       *pgxpool.Pool
	hub      *websocket.Hub
	interval time.Duration
}

func NewLeaderboardService(db *pgxpool.Pool, hub *websocket.Hub, interval time.Duration) *LeaderboardService {
	return &LeaderboardService{db: db, hub: hub, interval: interval}
}

// Start begins periodic updates
func (ls *LeaderboardService) Start(ctx context.Context) {
	ls.update(ctx)
	ticker := time.NewTicker(ls.interval)
	defer ticker.Stop()
	log.Info().Dur("interval", ls.interval).Msg("Leaderboard service started")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Leaderboard service stopped")
			return
		case <-ticker.C:
			ls.update(ctx)
		}
	}
}

func (ls *LeaderboardService) update(ctx context.Context) {
	if _, err := ls.db.Exec(ctx, "SELECT update_leaderboard_rankings()"); err != nil {
		log.Error().Err(err).Msg("Failed to update leaderboard rankings")
		return
	}

	// Insert performance snapshots (lightweight ROI history) per agent from agent_metrics
	snapshotInsert := `
		INSERT INTO agent_performance_snapshots (
			agent_id, balance, portfolio_value, total_value,
			total_profit_loss, roi_percent, total_trades,
			winning_trades, losing_trades, win_rate, max_drawdown, sharpe_ratio, snapshot_time
		)
		SELECT a.id,
			   a.current_balance,
			   am.total_portfolio_value,
			   a.current_balance + am.total_portfolio_value,
			   am.total_profit_loss,
			   am.roi,
			   am.total_trades,
			   am.winning_trades,
			   am.losing_trades,
			   am.win_rate,
			   am.max_drawdown,
			   am.sharpe_ratio,
			   NOW()
		FROM agents a
		JOIN agent_metrics am ON a.id = am.agent_id
		WHERE a.status = 'active';`
	if _, err := ls.db.Exec(ctx, snapshotInsert); err != nil {
		log.Warn().Err(err).Msg("Failed to insert performance snapshots")
	}

	entries, err := ls.getCurrent(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch leaderboard")
		return
	}
	ls.hub.BroadcastMessage("leaderboard_updated", entries)
	log.Info().Int("agents", len(entries)).Msg("Leaderboard updated")
}

func (ls *LeaderboardService) getCurrent(ctx context.Context) ([]models.LeaderboardEntry, error) {
	const q = `
        SELECT
            lr.rank_overall,
            a.id,
            a.name,
            a.model,
            lr.current_roi,
            lr.current_profit_loss,
            lr.current_win_rate,
            lr.total_trades,
            a.current_balance,
            am.total_portfolio_value,
            lr.badges,
            lr.updated_at
        FROM leaderboard_rankings lr
        JOIN agents a ON lr.agent_id = a.id
        JOIN agent_metrics am ON a.id = am.agent_id
        ORDER BY lr.rank_overall ASC`

	rows, err := ls.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.LeaderboardEntry{}
	for rows.Next() {
		var e models.LeaderboardEntry
		var badges []string
		if err := rows.Scan(&e.Rank, &e.AgentID, &e.AgentName, &e.Model, &e.ROI, &e.ProfitLoss, &e.WinRate, &e.TotalTrades, &e.Balance, &e.PortfolioValue, &badges, &e.UpdatedAt); err != nil {
			log.Error().Err(err).Msg("scan leaderboard row")
			continue
		}
		e.Badges = badges
		e.TotalValue = e.Balance + e.PortfolioValue
		out = append(out, e)
	}
	return out, nil
}
