-- ============================================
-- Market AI v0.2 - Trading Tables Migration
-- ============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- 1. AGENTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'inactive' CHECK (status IN ('active', 'inactive', 'paused')),
    initial_balance DECIMAL(15,2) DEFAULT 100000.00,
    current_balance DECIMAL(15,2) DEFAULT 100000.00,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_name ON agents(name);

-- ============================================
-- 2. STOCKS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS stocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    current_price DECIMAL(10,2) NOT NULL,
    previous_close DECIMAL(10,2),
    change_percent DECIMAL(5,2),
    volume BIGINT DEFAULT 0,
    last_updated TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_stocks_symbol ON stocks(symbol);

-- ============================================
-- 3. TRADES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    stock_symbol VARCHAR(10) NOT NULL REFERENCES stocks(symbol),
    trade_type VARCHAR(10) NOT NULL CHECK (trade_type IN ('BUY', 'SELL')),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    total_amount DECIMAL(15,2) NOT NULL,
    commission DECIMAL(10,2) DEFAULT 0,
    reasoning TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_trades_agent ON trades(agent_id);
CREATE INDEX idx_trades_stock ON trades(stock_symbol);
CREATE INDEX idx_trades_created ON trades(created_at DESC);
CREATE INDEX idx_trades_type ON trades(trade_type);

-- ============================================
-- 4. PORTFOLIO TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS portfolio (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    stock_symbol VARCHAR(10) NOT NULL REFERENCES stocks(symbol),
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    avg_buy_price DECIMAL(10,2) NOT NULL CHECK (avg_buy_price > 0),
    total_invested DECIMAL(15,2) NOT NULL,
    current_value DECIMAL(15,2) DEFAULT 0,
    profit_loss DECIMAL(15,2) DEFAULT 0,
    profit_loss_percent DECIMAL(10,2) DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(agent_id, stock_symbol)
);

CREATE INDEX idx_portfolio_agent ON portfolio(agent_id);
CREATE INDEX idx_portfolio_stock ON portfolio(stock_symbol);

-- ============================================
-- 5. MARKET DATA TABLE (Historical Prices)
-- ============================================
CREATE TABLE IF NOT EXISTS market_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stock_symbol VARCHAR(10) NOT NULL REFERENCES stocks(symbol),
    open_price DECIMAL(10,2) NOT NULL,
    close_price DECIMAL(10,2) NOT NULL,
    high_price DECIMAL(10,2) NOT NULL,
    low_price DECIMAL(10,2) NOT NULL,
    volume BIGINT DEFAULT 0,
    timestamp TIMESTAMP NOT NULL,
    timeframe VARCHAR(10) DEFAULT '1m' CHECK (timeframe IN ('1m', '5m', '15m', '1h', '1d'))
);

CREATE INDEX idx_market_data_symbol ON market_data(stock_symbol);
CREATE INDEX idx_market_data_timestamp ON market_data(timestamp DESC);
CREATE INDEX idx_market_data_symbol_time ON market_data(stock_symbol, timestamp DESC);

-- ============================================
-- 6. AGENT METRICS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS agent_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    total_trades INTEGER DEFAULT 0,
    winning_trades INTEGER DEFAULT 0,
    losing_trades INTEGER DEFAULT 0,
    total_profit_loss DECIMAL(15,2) DEFAULT 0,
    total_portfolio_value DECIMAL(15,2) DEFAULT 0,
    win_rate DECIMAL(5,2) DEFAULT 0,
    roi DECIMAL(10,2) DEFAULT 0,
    sharpe_ratio DECIMAL(10,4) DEFAULT 0,
    max_drawdown DECIMAL(10,2) DEFAULT 0,
    calculated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(agent_id)
);

CREATE INDEX idx_metrics_agent ON agent_metrics(agent_id);
CREATE INDEX idx_metrics_roi ON agent_metrics(roi DESC);

-- ============================================
-- 7. TRIGGER: Update agent updated_at
-- ============================================
CREATE OR REPLACE FUNCTION update_agent_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER agent_update_timestamp
    BEFORE UPDATE ON agents
    FOR EACH ROW
    EXECUTE FUNCTION update_agent_timestamp();

-- ============================================
-- 8. TRIGGER: Update portfolio timestamp
-- ============================================
CREATE TRIGGER portfolio_update_timestamp
    BEFORE UPDATE ON portfolio
    FOR EACH ROW
    EXECUTE FUNCTION update_agent_timestamp();

-- ============================================
-- 9. FUNCTION: Calculate portfolio value
-- ============================================
CREATE OR REPLACE FUNCTION calculate_portfolio_value(p_agent_id UUID)
RETURNS DECIMAL(15,2) AS $$
DECLARE
    total_value DECIMAL(15,2);
BEGIN
    SELECT COALESCE(SUM(p.quantity * s.current_price), 0)
    INTO total_value
    FROM portfolio p
    JOIN stocks s ON p.stock_symbol = s.symbol
    WHERE p.agent_id = p_agent_id;

    RETURN total_value;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 10. FUNCTION: Update agent metrics
-- ============================================
CREATE OR REPLACE FUNCTION update_agent_metrics(p_agent_id UUID)
RETURNS VOID AS $$
DECLARE
    v_total_trades INTEGER;
    v_winning_trades INTEGER;
    v_losing_trades INTEGER;
    v_total_profit_loss DECIMAL(15,2);
    v_portfolio_value DECIMAL(15,2);
    v_win_rate DECIMAL(5,2);
    v_roi DECIMAL(10,2);
    v_initial_balance DECIMAL(15,2);
BEGIN
    -- Get agent's initial balance
    SELECT initial_balance INTO v_initial_balance
    FROM agents WHERE id = p_agent_id;

    -- Calculate total trades
    SELECT COUNT(*) INTO v_total_trades
    FROM trades WHERE agent_id = p_agent_id;

    -- Calculate portfolio value
    v_portfolio_value := calculate_portfolio_value(p_agent_id);

    -- Calculate profit/loss
    SELECT COALESCE(SUM(profit_loss), 0) INTO v_total_profit_loss
    FROM portfolio WHERE agent_id = p_agent_id;

    -- Calculate winning/losing trades (simplified)
    SELECT
        COUNT(*) FILTER (WHERE trade_type = 'SELL' AND price > avg_buy_price),
        COUNT(*) FILTER (WHERE trade_type = 'SELL' AND price <= avg_buy_price)
    INTO v_winning_trades, v_losing_trades
    FROM trades t
    LEFT JOIN portfolio p ON t.agent_id = p.agent_id AND t.stock_symbol = p.stock_symbol
    WHERE t.agent_id = p_agent_id;

    -- Calculate win rate
    IF v_total_trades > 0 THEN
        v_win_rate := (v_winning_trades::DECIMAL / v_total_trades::DECIMAL) * 100;
    ELSE
        v_win_rate := 0;
    END IF;

    -- Calculate ROI
    IF v_initial_balance > 0 THEN
        v_roi := ((v_total_profit_loss / v_initial_balance) * 100);
    ELSE
        v_roi := 0;
    END IF;

    -- Upsert metrics
    INSERT INTO agent_metrics (
        agent_id, total_trades, winning_trades, losing_trades,
        total_profit_loss, total_portfolio_value, win_rate, roi, calculated_at
    )
    VALUES (
        p_agent_id, v_total_trades, v_winning_trades, v_losing_trades,
        v_total_profit_loss, v_portfolio_value, v_win_rate, v_roi, NOW()
    )
    ON CONFLICT (agent_id)
    DO UPDATE SET
        total_trades = EXCLUDED.total_trades,
        winning_trades = EXCLUDED.winning_trades,
        losing_trades = EXCLUDED.losing_trades,
        total_profit_loss = EXCLUDED.total_profit_loss,
        total_portfolio_value = EXCLUDED.total_portfolio_value,
        win_rate = EXCLUDED.win_rate,
        roi = EXCLUDED.roi,
        calculated_at = NOW();
END;
$$ LANGUAGE plpgsql;
