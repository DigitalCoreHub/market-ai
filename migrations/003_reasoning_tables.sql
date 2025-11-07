-- ============================================
-- Market AI v0.3 - Reasoning Tables
-- ============================================

-- Agent decisions and reasoning
CREATE TABLE IF NOT EXISTS agent_decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    stock_symbol VARCHAR(10) REFERENCES stocks(symbol),
    decision VARCHAR(10) NOT NULL CHECK (decision IN ('BUY', 'SELL', 'HOLD')),
    quantity INTEGER,
    target_price DECIMAL(10,2),
    stop_loss DECIMAL(10,2),

    -- Reasoning
    reasoning_full TEXT NOT NULL,
    reasoning_summary TEXT NOT NULL,

    -- Confidence & Risk
    confidence_score DECIMAL(5,2) CHECK (confidence_score >= 0 AND confidence_score <= 100),
    risk_score DECIMAL(5,2) CHECK (risk_score >= 0 AND risk_score <= 100),
    risk_level VARCHAR(20) CHECK (risk_level IN ('low', 'medium', 'high')),

    -- Market context at decision time
    market_context JSONB,

    -- Result tracking
    executed BOOLEAN DEFAULT FALSE,
    trade_id UUID REFERENCES trades(id),
    outcome VARCHAR(20) CHECK (outcome IN ('pending', 'success', 'failed', 'profit', 'loss')),
    actual_profit_loss DECIMAL(15,2),

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_decisions_agent ON agent_decisions(agent_id);
CREATE INDEX idx_decisions_stock ON agent_decisions(stock_symbol);
CREATE INDEX idx_decisions_created ON agent_decisions(created_at DESC);
CREATE INDEX idx_decisions_decision ON agent_decisions(decision);

-- Agent thinking logs (detailed step-by-step)
CREATE TABLE IF NOT EXISTS agent_thoughts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    decision_id UUID REFERENCES agent_decisions(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    step_name VARCHAR(100) NOT NULL,
    thought TEXT NOT NULL,
    data JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_thoughts_decision ON agent_thoughts(decision_id);
CREATE INDEX idx_thoughts_agent ON agent_thoughts(agent_id);

-- Agent strategies/personas
CREATE TABLE IF NOT EXISTS agent_strategies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    strategy_type VARCHAR(50) NOT NULL,
    description TEXT,
    parameters JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(agent_id)
);

CREATE INDEX idx_strategies_agent ON agent_strategies(agent_id);

-- Update agent_decisions outcome after trade
CREATE OR REPLACE FUNCTION update_decision_outcome()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the decision that led to this trade
    UPDATE agent_decisions
    SET executed = TRUE,
        trade_id = NEW.id,
        outcome = 'success'
    WHERE id = (
        SELECT id FROM agent_decisions
        WHERE agent_id = NEW.agent_id
          AND stock_symbol = NEW.stock_symbol
          AND decision = NEW.trade_type
          AND executed = FALSE
          AND created_at >= NOW() - INTERVAL '5 minutes'
        ORDER BY created_at DESC
        LIMIT 1
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS after_trade_insert ON trades;
CREATE TRIGGER after_trade_insert
    AFTER INSERT ON trades
    FOR EACH ROW
    EXECUTE FUNCTION update_decision_outcome();
