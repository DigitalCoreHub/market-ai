-- ============================================
-- Market AI v0.5 - Multi-Source Data Tables
-- ============================================

-- Data source tracking
CREATE TABLE IF NOT EXISTS data_sources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('yahoo', 'scraper', 'twitter')),
    source_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,

    -- Statistics
    last_fetch_at TIMESTAMP,
    total_fetches INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    avg_response_time_ms INTEGER,

    -- Status
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'error')),
    last_error TEXT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sources_type ON data_sources(source_type);
CREATE INDEX IF NOT EXISTS idx_sources_status ON data_sources(status);

-- Price data from multiple sources (for comparison & reliability)
CREATE TABLE IF NOT EXISTS price_sources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stock_symbol VARCHAR(10) NOT NULL,

    -- Prices from different sources
    yahoo_price DECIMAL(10,2),
    bloomberg_price DECIMAL(10,2),
    investing_price DECIMAL(10,2),

    -- Final price (after fusion)
    final_price DECIMAL(10,2) NOT NULL,

    -- Differences
    max_diff DECIMAL(10,2),
    price_variance DECIMAL(10,4),

    -- Reliability score
    confidence_score DECIMAL(5,2), -- 0-100

    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_price_sources_symbol ON price_sources(stock_symbol);
CREATE INDEX IF NOT EXISTS idx_price_sources_timestamp ON price_sources(timestamp DESC);

-- Twitter sentiment data
CREATE TABLE IF NOT EXISTS twitter_sentiment (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Tweet info
    tweet_id VARCHAR(100) UNIQUE,
    tweet_text TEXT NOT NULL,
    tweet_url TEXT,
    author_username VARCHAR(100),
    author_followers INTEGER,

    -- Stock mentions
    stock_symbols TEXT[],
    primary_symbol VARCHAR(10),

    -- Sentiment analysis
    sentiment_score DECIMAL(5,2) NOT NULL, -- -1.0 to +1.0
    sentiment_label VARCHAR(20), -- 'positive', 'negative', 'neutral'
    confidence DECIMAL(5,2),

    -- Engagement
    likes INTEGER DEFAULT 0,
    retweets INTEGER DEFAULT 0,
    replies INTEGER DEFAULT 0,

    -- Impact score (engagement × sentiment × author influence)
    impact_score DECIMAL(10,2),

    created_at TIMESTAMP DEFAULT NOW(),
    processed_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sentiment_symbols ON twitter_sentiment USING GIN(stock_symbols);
CREATE INDEX IF NOT EXISTS idx_sentiment_primary ON twitter_sentiment(primary_symbol);
CREATE INDEX IF NOT EXISTS idx_sentiment_created ON twitter_sentiment(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sentiment_impact ON twitter_sentiment(impact_score DESC);

-- Aggregated sentiment per stock (updated every 5 minutes)
CREATE TABLE IF NOT EXISTS stock_sentiment_aggregates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stock_symbol VARCHAR(10) NOT NULL,

    -- Time window
    window_start TIMESTAMP NOT NULL,
    window_end TIMESTAMP NOT NULL,

    -- Aggregate metrics
    total_tweets INTEGER DEFAULT 0,
    positive_tweets INTEGER DEFAULT 0,
    negative_tweets INTEGER DEFAULT 0,
    neutral_tweets INTEGER DEFAULT 0,

    avg_sentiment DECIMAL(5,2), -- Average sentiment score
    weighted_sentiment DECIMAL(5,2), -- Weighted by engagement

    total_engagement INTEGER, -- Total likes + retweets
    top_tweet_id VARCHAR(100),

    -- Trend
    sentiment_trend VARCHAR(20), -- 'rising', 'falling', 'stable'

    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(stock_symbol, window_start)
);

CREATE INDEX IF NOT EXISTS idx_agg_sentiment_symbol ON stock_sentiment_aggregates(stock_symbol);
CREATE INDEX IF NOT EXISTS idx_agg_sentiment_window ON stock_sentiment_aggregates(window_start DESC);

-- Scraped articles/news
CREATE TABLE IF NOT EXISTS scraped_articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    source VARCHAR(100) NOT NULL, -- 'Bloomberg HT', 'Investing.com', 'KAP'
    title TEXT NOT NULL,
    content TEXT,
    url TEXT UNIQUE,

    related_stocks TEXT[],

    published_at TIMESTAMP,
    scraped_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scraped_source ON scraped_articles(source);
CREATE INDEX IF NOT EXISTS idx_scraped_stocks ON scraped_articles USING GIN(related_stocks);
CREATE INDEX IF NOT EXISTS idx_scraped_published ON scraped_articles(published_at DESC);

-- Function to calculate stock sentiment aggregate
CREATE OR REPLACE FUNCTION update_sentiment_aggregate(p_stock_symbol VARCHAR, p_window_minutes INTEGER)
RETURNS void AS $$
DECLARE
    v_window_start TIMESTAMP;
    v_window_end TIMESTAMP;
    v_total INTEGER;
    v_positive INTEGER;
    v_negative INTEGER;
    v_neutral INTEGER;
    v_avg_sentiment DECIMAL(5,2);
    v_weighted_sentiment DECIMAL(5,2);
    v_total_engagement INTEGER;
    v_top_tweet VARCHAR(100);
BEGIN
    -- Define time window
    v_window_end := NOW();
    v_window_start := v_window_end - (p_window_minutes || ' minutes')::INTERVAL;

    -- Calculate metrics
    SELECT
        COUNT(*),
        COUNT(*) FILTER (WHERE sentiment_label = 'positive'),
        COUNT(*) FILTER (WHERE sentiment_label = 'negative'),
        COUNT(*) FILTER (WHERE sentiment_label = 'neutral'),
        AVG(sentiment_score),
        SUM(sentiment_score * (likes + retweets)) / NULLIF(SUM(likes + retweets), 0),
        SUM(likes + retweets),
        (SELECT tweet_id FROM twitter_sentiment
         WHERE p_stock_symbol = ANY(stock_symbols)
         AND created_at >= v_window_start
         ORDER BY impact_score DESC LIMIT 1)
    INTO v_total, v_positive, v_negative, v_neutral,
         v_avg_sentiment, v_weighted_sentiment, v_total_engagement, v_top_tweet
    FROM twitter_sentiment
    WHERE p_stock_symbol = ANY(stock_symbols)
      AND created_at >= v_window_start
      AND created_at <= v_window_end;

    -- Insert or update aggregate
    INSERT INTO stock_sentiment_aggregates (
        stock_symbol, window_start, window_end,
        total_tweets, positive_tweets, negative_tweets, neutral_tweets,
        avg_sentiment, weighted_sentiment, total_engagement, top_tweet_id
    ) VALUES (
        p_stock_symbol, v_window_start, v_window_end,
        v_total, v_positive, v_negative, v_neutral,
        v_avg_sentiment, v_weighted_sentiment, v_total_engagement, v_top_tweet
    )
    ON CONFLICT (stock_symbol, window_start)
    DO UPDATE SET
        total_tweets = EXCLUDED.total_tweets,
        positive_tweets = EXCLUDED.positive_tweets,
        negative_tweets = EXCLUDED.negative_tweets,
        neutral_tweets = EXCLUDED.neutral_tweets,
        avg_sentiment = EXCLUDED.avg_sentiment,
        weighted_sentiment = EXCLUDED.weighted_sentiment,
        total_engagement = EXCLUDED.total_engagement,
        top_tweet_id = EXCLUDED.top_tweet_id;
END;
$$ LANGUAGE plpgsql;
