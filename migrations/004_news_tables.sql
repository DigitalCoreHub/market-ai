-- ============================================
-- Market AI v0.3 - News System Tables
-- ============================================

-- News/Events storage
CREATE TABLE IF NOT EXISTS market_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Content
    title TEXT NOT NULL,
    description TEXT,
    content TEXT,                    -- Full article content (optional)

    -- Meta information
    source VARCHAR(100) NOT NULL,    -- 'Bloomberg HT', 'Investing.com', 'News API'
    url TEXT,                        -- Article URL
    content_hash VARCHAR(64),        -- MD5 hash of title+description for dedup

    -- Categorization
    event_type VARCHAR(50) DEFAULT 'news' CHECK (event_type IN ('news', 'economic_data', 'earnings', 'policy')),
    category VARCHAR(50),            -- 'banking', 'aviation', 'energy', 'general'

    -- Stock relations
    related_stocks TEXT[],           -- Array of stock symbols: ['THYAO', 'AKBNK']

    -- Sentiment analysis (for v1.0)
    sentiment VARCHAR(20) CHECK (sentiment IN ('positive', 'negative', 'neutral')),
    sentiment_score DECIMAL(5,2),    -- Score from -1.0 to +1.0

    -- Impact assessment
    impact_level VARCHAR(20) CHECK (impact_level IN ('low', 'medium', 'high')),

    -- Timestamps
    published_at TIMESTAMP NOT NULL,
    fetched_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_events_published ON market_events(published_at DESC);
CREATE INDEX idx_events_stocks ON market_events USING GIN(related_stocks);
CREATE INDEX idx_events_source ON market_events(source);
CREATE INDEX idx_events_type ON market_events(event_type);
CREATE INDEX idx_events_category ON market_events(category);

-- Unique constraint on non-null URL to prevent duplicates
CREATE UNIQUE INDEX idx_events_url ON market_events(url) WHERE url IS NOT NULL;

-- Simple: just use content_hash with no WHERE clause for ON CONFLICT
ALTER TABLE market_events ADD CONSTRAINT unique_content_hash UNIQUE NULLS NOT DISTINCT (content_hash);

-- Function to cleanup old news (keep only last 7 days)
CREATE OR REPLACE FUNCTION cleanup_old_news()
RETURNS void AS $$
BEGIN
    DELETE FROM market_events
    WHERE published_at < NOW() - INTERVAL '7 days';

    RAISE NOTICE 'Old news cleaned up';
END;
$$ LANGUAGE plpgsql;

-- News fetch log (track aggregator runs)
CREATE TABLE IF NOT EXISTS news_fetch_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source VARCHAR(100) NOT NULL,
    articles_fetched INTEGER DEFAULT 0,
    articles_stored INTEGER DEFAULT 0,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    execution_time_ms INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_fetch_log_created ON news_fetch_log(created_at DESC);

-- View: Recent news with stock info
CREATE OR REPLACE VIEW v_recent_news_with_stocks AS
SELECT
    me.id,
    me.title,
    me.description,
    me.source,
    me.published_at,
    me.related_stocks,
    me.impact_level,
    COALESCE(
        ARRAY_AGG(
            DISTINCT s.name
        ) FILTER (WHERE s.name IS NOT NULL),
        ARRAY[]::TEXT[]
    ) AS stock_names,
    me.created_at
FROM market_events me
LEFT JOIN LATERAL UNNEST(me.related_stocks) AS stock_symbol ON TRUE
LEFT JOIN stocks s ON s.symbol = stock_symbol
WHERE me.published_at > NOW() - INTERVAL '24 hours'
GROUP BY me.id, me.title, me.description, me.source, me.published_at,
         me.related_stocks, me.impact_level, me.created_at
ORDER BY me.published_at DESC;
