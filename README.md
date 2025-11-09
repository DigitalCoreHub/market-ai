# 🤖 Market AI

Türkiye odaklı, çoklu veri kaynağı ve çoklu model desteği ile AI tabanlı otonom al-sat simülasyon altyapısı.

—

## İçindekiler

- Özellikler
- Mimarinin Özeti
- Kurulum ve Çalıştırma
- Ortam Değişkenleri (.env)
- API Uç Noktaları
- Veri Tabanı ve Migrasyonlar
- Test ve Kalite
- Notlar ve Sorumluluk Reddi

—

## Özellikler

- Otonom AI ajanları (rastgele karar aralığı, bütçe modu ile uzatılabilir)
- Haber + scraper + Twitter verisi ile piyasa bağlamı (Füzyon Servisi)
- Dinamik Hisse Evreni (v0.6)
  - Haber/Twitter/İşlem aktivitesine göre otomatik aktif/pasif hisse yönetimi
  - WebSocket ile “universe_updated” yayını, geçmiş log kaydı
- AI Kararlarında Lot/Miktar Kontrolü (v0.6)
  - Prompt’ta maksimum işlem tutarı rehberi
  - Risk Yöneticisi ile miktar > 0, bakiye + komisyon kontrolü
- Güvenilirlik skorlaması ve metrikler (v0.5)
- Çoklu model desteği: OpenAI, Anthropic, Google, DeepSeek, Groq/Llama, Mistral, XAI
- PostgreSQL + Redis altyapısı, WebSocket yayınları
- **v1.0: Production Ready**
  - JWT + API Key authentication
  - Rate limiting (60 req/min)
  - Input validation
  - Prometheus metrics & Grafana dashboards
  - Dockerized deployment (Docker Compose)
  - CI/CD pipeline (GitHub Actions)

—

## Mimarinin Özeti

- News Aggregator: Haberleri (NewsAPI + RSS) periyodik toplar, önbellekler ve yayınlar.
- Fusion Service: Yahoo fiyatları, scraper haberleri ve Twitter duygu verisini birleştirir.
- Agent Engine: Piyasa bağlamıyla AI kararını üretir, veritabanına kaydeder, riskten geçirir ve işlemi uygular.
- StockUniverseService: 6 saatte bir (otonom) evren günceller; manuel tetiklenebilir.
- Leaderboard Service: Belirli aralıkta sıralama hesaplar ve yayınlar.

—

## Kurulum ve Çalıştırma

Önkoşullar: Go 1.24+, Docker, Docker Compose, Node 20+ (frontend için - opsiyonel).

1. Servisleri başlatın

```bash
make up
# veya
docker-compose up -d  # PostgreSQL + Redis + App
```

**Not:** Local development için Prometheus ve Grafana varsayılan olarak kapalıdır. Gerekirse:

```bash
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d
```

2. Migrasyonları uygulayın (v0.5 veri kaynakları + v0.6 dinamik evren)

```bash
docker exec -i marketai-postgres psql -U marketai -d marketai_dev < migrations/006_data_sources.sql
docker exec -i marketai-postgres psql -U marketai -d marketai_dev < migrations/007_data_sources_seed.sql
docker exec -i marketai-postgres psql -U marketai -d marketai_dev < migrations/008_dynamic_universe.sql
```

3. Sunucuyu başlatın

```bash
go build -o bin/market-ai ./cmd/server
./bin/market-ai
```

Opsiyonel (Frontend):

```bash
cd frontend && npm install && npm run dev
```

—

## Ortam Değişkenleri (.env)

Temel

- PORT, ENV
- LOG_LEVEL (debug|info|warn|error)

Veritabanı

- DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE

Redis

- REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB

Haber (v0.3)

- NEWS_API_KEY, NEWS_UPDATE_INTERVAL (dakika), NEWS_CACHE_TTL (dakika), RSS_FEEDS (virgüllü)

AI Sağlayıcıları (v0.4+)

- OPENAI_API_KEY, ANTHROPIC_API_KEY, GOOGLE_API_KEY, DEEPSEEK_API_KEY, GROQ_API_KEY, MISTRAL_API_KEY, XAI_API_KEY

Model İsimleri (v0.4+)

- AI_MODEL_GPT, AI_MODEL_GPT4_MINI, AI_MODEL_CLAUDE, AI_MODEL_GEMINI, AI_MODEL_DEEPSEEK, AI_MODEL_LLAMA, AI_MODEL_MIXTRAL, AI_MODEL_GROK
- AI_TEMPERATURE, AI_MAX_TOKENS

Maliyet Bayrakları

- BUDGET_MODE (true|false), ENABLE_PREMIUM_MODELS (true|false)

Veri Kaynakları ve Aralıklar (v0.5)

- YAHOO_FETCH_INTERVAL, SCRAPER_FETCH_INTERVAL, TWITTER_FETCH_INTERVAL (saniye)
- SENTIMENT_UPDATE_INTERVAL (saniye)
- TWITTER_API_KEY, TWITTER_API_SECRET, TWITTER_ACCESS_TOKEN, TWITTER_ACCESS_SECRET

Opsiyonel

- SYMBOL_UNIVERSE: Başlangıç/bağlam sembolleri (Dinamik evren açıkken opsiyoneldir)
- LEADERBOARD_UPDATE_INTERVAL (varsayılan 60s)

Authentication (v1.0)

- JWT_SECRET: JWT token imzalama secret'ı (production'da mutlaka değiştir!)
- API_KEY: Master API key (API key ile login yapıp JWT token almak için)

Kaldırılan/Artık Kullanılmayan

- AGENT_DECISION_INTERVAL_MIN/MAX, AGENT_MAX_RISK_PER_TRADE, AGENT_MAX_PORTFOLIO_RISK, AGENT_MIN_CONFIDENCE, AGENT_INITIAL_BALANCE → KULLANILMIYOR

—

## API Uç Noktaları (seçmece)

Authentication (v1.0)

- POST /api/v1/auth/login → API key ile login, JWT token al

Public Endpoints

- GET /health → Health check
- GET /api/v1/ping → Ping test
- GET /api/v1/market/context?symbols=THYAO,AKBNK
- GET /api/v1/metrics, GET /api/v1/metrics/prometheus
- GET /api/v1/debug/yahoo | /debug/scraper | /debug/tweets
- GET /api/v1/leaderboard, GET /api/v1/leaderboard/roi-history
- GET /api/v1/universe/active, GET /api/v1/universe/history

Protected Endpoints (API Key veya JWT Token gerekli)

- POST /api/v1/universe/update → Hisse evrenini güncelle

—

## Veri Tabanı ve Migrasyonlar

- 002: Temel trading tabloları (agents, stocks, trades, portfolio, ...)
- 006–007: Veri kaynakları ve seed
- 008: Dinamik hisse evreni, log ve aktivite fonksiyonu

—

## Test ve Kalite

```bash
go test ./...
```

Kalite kapıları: Build PASS, mevcut testler PASS. Yeni özellikler için ek testler önerilir (özellikle StockUniverseService).

—

## Notlar ve Sorumluluk Reddi

- Bu proje eğitim ve deneysel amaçlıdır; yatırım tavsiyesi değildir.
- API maliyetleri modele ve çağrı sıklığına göre değişir; bütçe bayraklarını kullanın.

—

Teşekkürler: Katkılar, geri bildirimler ve PR’lar memnuniyetle karşılanır.
