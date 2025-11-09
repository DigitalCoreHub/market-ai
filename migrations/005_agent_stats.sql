-- ============================================
-- Market AI v0.4 - Agent Statistics & Leaderboard
-- ============================================

-- Enable UUID extension (idempotent)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Agent performance snapshots (hourly)
CREATE TABLE IF NOT EXISTS agent_performance_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,

    -- Balance metrics
    balance DECIMAL(15,2) NOT NULL,
    portfolio_value DECIMAL(15,2) NOT NULL,
    total_value DECIMAL(15,2) NOT NULL,

    -- Performance metrics
    total_profit_loss DECIMAL(15,2) NOT NULL,
    roi_percent DECIMAL(10,2) NOT NULL,

    -- Trading stats
    total_trades INTEGER NOT NULL,
    winning_trades INTEGER NOT NULL,
    losing_trades INTEGER NOT NULL,
    win_rate DECIMAL(5,2) NOT NULL,

    -- Risk metrics
    max_drawdown DECIMAL(10,2),
    sharpe_ratio DECIMAL(10,4),

    snapshot_time TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_snapshots_agent ON agent_performance_snapshots(agent_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_time ON agent_performance_snapshots(snapshot_time DESC);

-- Leaderboard rankings (periodically updated)
CREATE TABLE IF NOT EXISTS leaderboard_rankings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,

    -- Rankings
    rank_overall INTEGER NOT NULL,
    rank_by_roi INTEGER NOT NULL,
    rank_by_winrate INTEGER NOT NULL,
    rank_by_profit INTEGER NOT NULL,

    -- Current metrics
    current_roi DECIMAL(10,2) NOT NULL,
    current_profit_loss DECIMAL(15,2) NOT NULL,
    current_win_rate DECIMAL(5,2) NOT NULL,
    total_trades INTEGER NOT NULL,

    -- Badges/achievements
    badges TEXT[],

    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rankings_agent ON leaderboard_rankings(agent_id);
CREATE INDEX IF NOT EXISTS idx_rankings_overall ON leaderboard_rankings(rank_overall);
CREATE UNIQUE INDEX IF NOT EXISTS idx_rankings_agent_unique ON leaderboard_rankings(agent_id);

-- Head-to-head matchups
CREATE TABLE IF NOT EXISTS agent_matchups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent1_id UUID NOT NULL REFERENCES agents(id),
    agent2_id UUID NOT NULL REFERENCES agents(id),

    agent1_wins INTEGER DEFAULT 0,
    agent2_wins INTEGER DEFAULT 0,
    draws INTEGER DEFAULT 0,

    agent1_total_profit DECIMAL(15,2) DEFAULT 0,
    agent2_total_profit DECIMAL(15,2) DEFAULT 0,

    last_winner UUID REFERENCES agents(id),
    last_matchup_time TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(agent1_id, agent2_id)
);

CREATE INDEX IF NOT EXISTS idx_matchups_agent1 ON agent_matchups(agent1_id);
CREATE INDEX IF NOT EXISTS idx_matchups_agent2 ON agent_matchups(agent2_id);

-- Agent daily statistics
CREATE TABLE IF NOT EXISTS agent_daily_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    stat_date DATE NOT NULL,

    trades_count INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,

    profit_loss DECIMAL(15,2) DEFAULT 0,
    volume_traded DECIMAL(15,2) DEFAULT 0,

    best_trade_profit DECIMAL(15,2),
    worst_trade_loss DECIMAL(15,2),

    avg_confidence DECIMAL(5,2),
    avg_execution_time_ms INTEGER,

    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(agent_id, stat_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_stats_agent ON agent_daily_stats(agent_id);
CREATE INDEX IF NOT EXISTS idx_daily_stats_date ON agent_daily_stats(stat_date DESC);

-- Function to update leaderboard rankings
CREATE OR REPLACE FUNCTION update_leaderboard_rankings()
RETURNS void AS $$
BEGIN
    -- Clear old rankings
    TRUNCATE leaderboard_rankings;

    -- Calculate new rankings from agent_metrics
    WITH agent_stats AS (
        SELECT
            a.id AS agent_id,
            a.name,
            am.roi,
            am.total_profit_loss,
            am.win_rate,
            am.total_trades,
            RANK() OVER (ORDER BY am.roi DESC) AS rank_roi,
            RANK() OVER (ORDER BY am.total_profit_loss DESC) AS rank_profit,
            RANK() OVER (ORDER BY am.win_rate DESC) AS rank_winrate,
            RANK() OVER (ORDER BY (am.roi * 0.4 + am.win_rate * 0.3 + (am.total_profit_loss / 1000) * 0.3) DESC) AS rank_overall
        FROM agents a
        JOIN agent_metrics am ON a.id = am.agent_id
        WHERE a.status = 'active'
    )
    INSERT INTO leaderboard_rankings (
        agent_id, rank_overall, rank_by_roi, rank_by_winrate, rank_by_profit,
        current_roi, current_profit_loss, current_win_rate, total_trades, updated_at
    )
    SELECT
        agent_id, rank_overall, rank_roi, rank_winrate, rank_profit,
        roi, total_profit_loss, win_rate, total_trades, NOW()
    FROM agent_stats;

    RAISE NOTICE 'Leaderboard rankings updated';
END;
$$ LANGUAGE plpgsql;
