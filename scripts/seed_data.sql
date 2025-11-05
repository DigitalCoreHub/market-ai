-- ============================================
-- Seed Data for Market AI v0.2
-- ============================================

-- Insert initial stocks (BIST favorites)
INSERT INTO stocks (symbol, name, current_price, previous_close, volume) VALUES
('THYAO', 'Türk Hava Yolları', 245.50, 244.00, 15000000),
('AKBNK', 'Akbank', 58.25, 57.80, 45000000),
('ASELS', 'Aselsan', 125.75, 124.50, 8000000),
('TUPRS', 'Tüpraş', 178.50, 177.00, 12000000),
('EREGL', 'Ereğli Demir Çelik', 42.30, 42.00, 25000000)
ON CONFLICT (symbol) DO NOTHING;

-- Update change percentages
UPDATE stocks SET change_percent = ROUND(((current_price - previous_close) / previous_close * 100)::numeric, 2);

-- Insert demo agents (will be replaced with real AI agents later)
INSERT INTO agents (name, model, status, initial_balance, current_balance) VALUES
('GPT-4 Turbo', 'gpt-4-turbo', 'active', 100000.00, 100000.00),
('Claude Opus', 'claude-3-opus', 'active', 100000.00, 100000.00),
('Gemini Pro', 'gemini-pro', 'active', 100000.00, 100000.00)
ON CONFLICT DO NOTHING;

-- Initialize agent metrics
INSERT INTO agent_metrics (agent_id)
SELECT id FROM agents
ON CONFLICT (agent_id) DO NOTHING;

-- Insert some historical market data (last 30 minutes, 1-minute candles)
DO $$
DECLARE
    stock_rec RECORD;
    base_time TIMESTAMP := NOW() - INTERVAL '30 minutes';
    i INTEGER;
    base_price DECIMAL(10,2);
    rand_change DECIMAL(10,2);
BEGIN
    FOR stock_rec IN SELECT symbol, current_price FROM stocks LOOP
        base_price := stock_rec.current_price;

        FOR i IN 0..29 LOOP
            rand_change := (RANDOM() * 2 - 1) * 2; -- Random change between -2 and +2

            INSERT INTO market_data (
                stock_symbol,
                open_price,
                close_price,
                high_price,
                low_price,
                volume,
                timestamp,
                timeframe
            ) VALUES (
                stock_rec.symbol,
                base_price + rand_change,
                base_price + rand_change + (RANDOM() * 2 - 1),
                base_price + rand_change + ABS(RANDOM() * 2),
                base_price + rand_change - ABS(RANDOM() * 2),
                (RANDOM() * 1000000)::BIGINT,
                base_time + (i || ' minutes')::INTERVAL,
                '1m'
            );
        END LOOP;
    END LOOP;
END $$;
