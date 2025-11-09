package fusion

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/1batu/market-ai/internal/datasources/scraper"
	tw "github.com/1batu/market-ai/internal/datasources/twitter"
	"github.com/1batu/market-ai/internal/datasources/yahoo"
	"github.com/1batu/market-ai/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// sourceStats her kaynak için hafif bellek içi güvenilirlik istatistiklerini tutar
type sourceStats struct {
	total         int
	success       int
	totalDuration time.Duration // başarılı getirme sürelerinin toplamı
}

// Service Yahoo, Web kazıyıcılar ve Twitter + duygu verilerini birleştirir
// Ayrıca fiyat füzyonu ve güven skorlaması için hafif güvenilirlik metrikleri tutar.
type Service struct {
	db       *pgxpool.Pool
	yahoo    *yahoo.YahooFinanceClient
	scraper  *scraper.WebScraper
	twitter  *tw.Client
	analyzer *tw.Analyzer

	// API sık çağrıldığında harici çağrıları azaltmak için basit bellek içi önbellek
	cacheTTL       time.Duration
	lastSymbolsKey string
	lastAt         time.Time
	lastCtx        *models.MarketContext

	// güvenilirlik / metrikler
	stats map[string]*sourceStats // anahtar: kaynak adı ("yahoo", sonra diğerleri)
}

func New(db *pgxpool.Pool, y *yahoo.YahooFinanceClient, s *scraper.WebScraper, t *tw.Client, a *tw.Analyzer) *Service {
	return &Service{
		db:       db,
		yahoo:    y,
		scraper:  s,
		twitter:  t,
		analyzer: a,
		cacheTTL: 30 * time.Second,
		stats: map[string]*sourceStats{
			"Yahoo Finance API":    {},
			"Bloomberg HT Scraper": {},
			"Twitter API Search":   {},
		},
	}
}

// recordFetch güvenilirlik istatistiklerini günceller (bellekte ve veritabanında)
func (svc *Service) recordFetch(source string, dur time.Duration, success bool) {
	st, ok := svc.stats[source]
	if !ok {
		st = &sourceStats{}
		svc.stats[source] = st
	}
	st.total++
	if success {
		st.success++
		st.totalDuration += dur
	}
	// data_sources tablosunu güncelle (bloklamadan)
	go svc.updateDataSource(source, dur, success)
}

// updateDataSource getirme istatistiklerini data_sources tablosuna kaydeder
func (svc *Service) updateDataSource(source string, dur time.Duration, success bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	avgMs := int(dur.Milliseconds())
	var lastError *string
	if !success {
		errMsg := "fetch failed"
		lastError = &errMsg
	}
	status := "active"
	if !success {
		status = "error"
	}
	_, err := svc.db.Exec(ctx, `
		UPDATE data_sources
		SET total_fetches = total_fetches + 1,
		    success_count = success_count + CASE WHEN $2 THEN 1 ELSE 0 END,
		    error_count = error_count + CASE WHEN $2 THEN 0 ELSE 1 END,
		    avg_response_time_ms = (avg_response_time_ms * GREATEST(total_fetches - 1, 0) + $3) / GREATEST(total_fetches, 1),
		    last_fetch_at = NOW(),
		    last_error = $4,
		    status = $5,
		    updated_at = NOW()
		WHERE source_name = $1
	`, source, success, avgMs, lastError, status)
	if err != nil {
		log.Error().Err(err).Str("source", source).Msg("updateDataSource failed")
	}
}

// MarketContext çok kaynaklı anlık görüntü getirir ve önemli kısımları kaydeder
func (svc *Service) MarketContext(ctx context.Context, symbols []string) (*models.MarketContext, error) {
	// önbellek kontrolü
	key := symbolsKey(symbols)
	if svc.cacheTTL > 0 && svc.lastCtx != nil && key == svc.lastSymbolsKey && time.Since(svc.lastAt) < svc.cacheTTL {
		return svc.lastCtx, nil
	}

	// 1) Yahoo fiyatları
	var prices []*models.StockPrice
	var durYahoo time.Duration
	var yahooErr error
	if svc.yahoo != nil {
		startYahoo := time.Now()
		prices, yahooErr = svc.yahoo.GetMultipleStocks(ctx, symbols)
		durYahoo = time.Since(startYahoo)
		if yahooErr != nil {
			log.Error().Err(yahooErr).Msg("Yahoo fetch error")
		}
		// güvenilirlik istatistikleri (tek toplu getirme olarak değerlendir)
		svc.recordFetch("Yahoo Finance API", durYahoo, yahooErr == nil && len(prices) > 0)
	}

	// Fiyat başına güven skoru türet (şimdilik tek kaynak)
	if st, ok := svc.stats["Yahoo Finance API"]; ok && st.total > 0 {
		successRate := float64(st.success) / float64(st.total)
		avgMs := 0
		if st.success > 0 {
			avgMs = int((st.totalDuration / time.Duration(st.success)).Milliseconds())
		}
		for _, p := range prices {
			p.ConfidenceScore = ComputeConfidence(successRate, avgMs, 0) // çok kaynaklı olana kadar varyans=0
		}
	}

	// 2) Kazınmış haberler
	startScrape := time.Now()
	var news []models.ScrapedArticle
	if svc.scraper != nil {
		if n, err := svc.scraper.ScrapeAll(ctx); err == nil {
			news = n
		} else {
			log.Error().Err(err).Msg("Scrape error")
		}
	}
	durScrape := time.Since(startScrape)

	// 3) Twitter tweetleri
	startTweets := time.Now()
	var tweets []models.Tweet
	if svc.twitter != nil {
		if ts, err := svc.twitter.SearchRecent(ctx, 50); err == nil {
			tweets = ts
		} else {
			log.Error().Err(err).Msg("Twitter search error")
		}
	}
	durTweets := time.Since(startTweets)

	// 4) Duygu analizi yap
	if svc.analyzer != nil && len(tweets) > 0 {
		if analyzed, err := svc.analyzer.AnalyzeBatch(ctx, tweets); err == nil {
			tweets = analyzed
		}
	}

	// 5) Hisse başına duyguları topla
	sentiments := aggregateSentiment(tweets, symbols)

	// 6) Güvenilirlik takibi için fiyatları kaydet
	svc.storePriceSources(ctx, prices)

	// 7) Tweet duygularını kaydet
	svc.storeTweets(ctx, tweets)

	ctxOut := &models.MarketContext{
		Prices:          prices,
		News:            news,
		Tweets:          tweets,
		StockSentiments: sentiments,
		UpdatedAt:       time.Now(),
		FetchDurations: map[string]time.Duration{
			"yahoo":   durYahoo,
			"scraper": durScrape,
			"twitter": durTweets,
		},
	}
	svc.lastCtx = ctxOut
	svc.lastSymbolsKey = key
	svc.lastAt = time.Now()
	return ctxOut, nil
}

func aggregateSentiment(tweets []models.Tweet, symbols []string) map[string]*models.StockSentiment {
	out := make(map[string]*models.StockSentiment)
	for _, s := range symbols {
		out[s] = &models.StockSentiment{Symbol: s}
	}
	for _, tw := range tweets {
		for _, sym := range tw.StockSymbols {
			if agg, ok := out[sym]; ok {
				agg.TweetCount++
				agg.AvgSentiment += tw.SentimentScore
				switch tw.SentimentLabel {
				case "positive":
					agg.PositiveCount++
				case "negative":
					agg.NegativeCount++
				default:
					agg.NeutralCount++
				}
				if agg.TopTweet == nil || tw.ImpactScore > agg.TopTweet.ImpactScore {
					t := tw
					agg.TopTweet = &t
				}
			}
		}
	}
	for _, agg := range out {
		if agg.TweetCount > 0 {
			agg.AvgSentiment /= float64(agg.TweetCount)
		}
	}
	return out
}

func (svc *Service) storePriceSources(ctx context.Context, prices []*models.StockPrice) {
	for _, p := range prices {
		_, err := svc.db.Exec(ctx, `
            INSERT INTO price_sources (stock_symbol, yahoo_price, bloomberg_price, investing_price, final_price, max_diff, price_variance, confidence_score, timestamp)
            VALUES ($1,$2,NULL,NULL,$3,NULL,$4,$5,$6)
        `, p.Symbol, p.Price, p.Price, 0.0, p.ConfidenceScore, p.Timestamp)
		if err != nil {
			log.Error().Err(err).Str("symbol", p.Symbol).Msg("storePriceSources")
		}
	}
}

func (svc *Service) storeTweets(ctx context.Context, tweets []models.Tweet) {
	for _, t := range tweets {
		primary := ""
		if len(t.StockSymbols) > 0 {
			primary = t.StockSymbols[0]
		}
		_, err := svc.db.Exec(ctx, `
            INSERT INTO twitter_sentiment (
                tweet_id, tweet_text, tweet_url, author_username, author_followers,
                stock_symbols, primary_symbol, sentiment_score, sentiment_label,
                confidence, likes, retweets, replies, impact_score
            ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
            ON CONFLICT (tweet_id) DO NOTHING
        `,
			t.ID, t.Text, t.URL, t.Author, t.AuthorFollowers,
			t.StockSymbols, primary, t.SentimentScore, t.SentimentLabel,
			t.SentimentConfidence, t.Likes, t.Retweets, 0, t.ImpactScore,
		)
		if err != nil {
			log.Error().Err(err).Str("tweet_id", t.ID).Msg("storeTweets")
		}
	}
}

func symbolsKey(symbols []string) string {
	if len(symbols) == 0 {
		return ""
	}
	s := make([]string, len(symbols))
	copy(s, symbols)
	// kararlı önbellek anahtarı için sırayı normalize et
	sort.Strings(s)
	return strings.Join(s, ",")
}
