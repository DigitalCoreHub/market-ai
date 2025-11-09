-- 008_dynamic_universe.sql
-- Dynamic Stock Universe & Activity Tracking
-- Created: 2025-11-09

-- Add new columns to stocks if missing
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS market_cap BIGINT DEFAULT 0;
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS last_trade_date DATE;
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS last_news_mention TIMESTAMP;
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS last_twitter_mention TIMESTAMP;
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS discovery_source VARCHAR(50) DEFAULT 'manual';
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS mention_count_7d INTEGER DEFAULT 0;
ALTER TABLE stocks ADD COLUMN IF NOT EXISTS twitter_count_24h INTEGER DEFAULT 0;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_stocks_active ON stocks(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_stocks_market_cap ON stocks(market_cap DESC);
CREATE INDEX IF NOT EXISTS idx_stocks_mentions ON stocks(mention_count_7d DESC);
CREATE INDEX IF NOT EXISTS idx_stocks_last_news ON stocks(last_news_mention DESC NULLS LAST);

-- Universe update log
CREATE TABLE IF NOT EXISTS stock_universe_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    total_stocks INTEGER NOT NULL,
    from_trades INTEGER,
    from_news INTEGER,
    from_twitter INTEGER,
    added_stocks TEXT[],
    removed_stocks TEXT[],
    update_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Helper function: update stock activity from news & twitter
CREATE OR REPLACE FUNCTION update_stock_activity()
RETURNS void AS $$
BEGIN
    -- News based activation
    UPDATE stocks s
    SET is_active = TRUE,
        last_news_mention = (
            SELECT MAX(published_at)
            FROM market_events
            WHERE s.symbol = ANY(related_stocks)
        ),
        mention_count_7d = (
            SELECT COUNT(*)
            FROM market_events
            WHERE s.symbol = ANY(related_stocks)
              AND published_at > NOW() - INTERVAL '7 days'
        )
    WHERE EXISTS (
        SELECT 1 FROM market_events
        WHERE s.symbol = ANY(related_stocks)
          AND published_at > NOW() - INTERVAL '7 days'
    );

    -- Twitter based activation
    UPDATE stocks s
    SET is_active = TRUE,
        last_twitter_mention = (
            SELECT MAX(created_at)
            FROM twitter_sentiment
            WHERE s.symbol = ANY(stock_symbols)
        ),
        twitter_count_24h = (
            SELECT COUNT(*)
            FROM twitter_sentiment
            WHERE s.symbol = ANY(stock_symbols)
              AND created_at > NOW() - INTERVAL '24 hours'
        )
    WHERE EXISTS (
        SELECT 1 FROM twitter_sentiment
        WHERE s.symbol = ANY(stock_symbols)
          AND created_at > NOW() - INTERVAL '24 hours'
    );

    -- Deactivate if inactive 30+ days everywhere
    UPDATE stocks
    SET is_active = FALSE
    WHERE (last_news_mention IS NULL OR last_news_mention < NOW() - INTERVAL '30 days')
      AND (last_twitter_mention IS NULL OR last_twitter_mention < NOW() - INTERVAL '30 days')
      AND (last_trade_date IS NULL OR last_trade_date < NOW() - INTERVAL '30 days');

    RAISE NOTICE 'Stock activity updated';
END;
$$ LANGUAGE plpgsql;

-- Seed top BIST names (inactive bootstrap, system will activate when data appears)
INSERT INTO stocks (symbol, name, current_price, previous_close, market_cap, is_active, discovery_source) VALUES
('THYAO', 'Türk Hava Yolları', 1.00, 1.00, 50000000000, FALSE, 'seed'),
('AKBNK', 'Akbank', 1.00, 1.00, 45000000000, FALSE, 'seed'),
('GARAN', 'Garanti Bankası', 1.00, 1.00, 42000000000, FALSE, 'seed'),
('ISCTR', 'İş Bankası', 1.00, 1.00, 40000000000, FALSE, 'seed'),
('KCHOL', 'Koç Holding', 1.00, 1.00, 38000000000, FALSE, 'seed'),
('TUPRS', 'Tüpraş', 1.00, 1.00, 35000000000, FALSE, 'seed'),
('ASELS', 'Aselsan', 1.00, 1.00, 32000000000, FALSE, 'seed'),
('EREGL', 'Ereğli Demir Çelik', 1.00, 1.00, 30000000000, FALSE, 'seed'),
('PETKM', 'Petkim', 1.00, 1.00, 28000000000, FALSE, 'seed'),
('SAHOL', 'Sabancı Holding', 1.00, 1.00, 27000000000, FALSE, 'seed'),
('BIMAS', 'BİM', 1.00, 1.00, 25000000000, FALSE, 'seed'),
('SISE', 'Şişe Cam', 1.00, 1.00, 24000000000, FALSE, 'seed'),
('TCELL', 'Turkcell', 1.00, 1.00, 23000000000, FALSE, 'seed'),
('HALKB', 'Halk Bankası', 1.00, 1.00, 22000000000, FALSE, 'seed'),
('YKBNK', 'Yapı Kredi', 1.00, 1.00, 21000000000, FALSE, 'seed'),
('VAKBN', 'Vakıfbank', 1.00, 1.00, 20000000000, FALSE, 'seed'),
('TTKOM', 'Türk Telekom', 1.00, 1.00, 19000000000, FALSE, 'seed'),
('SODA', 'Soda Sanayii', 1.00, 1.00, 18000000000, FALSE, 'seed'),
('TOASO', 'Tofaş', 1.00, 1.00, 17000000000, FALSE, 'seed'),
('ARCLK', 'Arçelik', 1.00, 1.00, 16000000000, FALSE, 'seed')
ON CONFLICT (symbol) DO UPDATE SET
    market_cap = EXCLUDED.market_cap,
    discovery_source = EXCLUDED.discovery_source;

-- Universe log table comment
COMMENT ON TABLE stock_universe_log IS 'Tracks autonomous updates to the dynamic stock universe';
