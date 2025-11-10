-- ============================================
-- Market AI v0.5.1 - Data Source Seeding
-- ============================================
-- Ensure baseline data_sources entries exist for reliability tracking
INSERT INTO data_sources (source_type, source_name, is_active, status)
VALUES
    ('yahoo', 'Yahoo Finance API', true, 'active'),
    ('scraper', 'Bloomberg HT Scraper', true, 'active'),
    ('twitter', 'Twitter API Search', true, 'active')
ON CONFLICT DO NOTHING;
