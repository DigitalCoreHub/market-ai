# 🤖 Market AI

### _Türkiye'nin ilk yapay zekâ destekli finans simülasyon arenası_

> **"AI'lar Borsa İstanbul'da yarışsaydı kim kazanırdı?"**

---

## 📖 Proje Hakkında

Market AI, finansal piyasalarda yapay zekâ ajanlarının (AI agents) farklı stratejilerle nasıl kararlar aldığını gözlemlemeyi amaçlayan, deneysel bir simülasyon ve test projesidir.

## 🎯 v0.3 - Autonomous AI Agent System with News Integration

### ✨ Yeni Özellikler

- **Otonom AI Ajanları**: 30-60 saniye aralıklarla haber bağlamında kendi kendine karar veren AI ajanları
- **Haber Entegrasyonu**: News API + RSS feeds ile Türkiye finans haberlerinin gerçek zamanlı toplanması
- **AI Model Desteği**: OpenAI (GPT-3.5/GPT-4) ve Anthropic (Claude 3 Haiku/Opus)
- **Risk Yönetimi**: Trade'leri gerçekleştirmeden önce otomatik risk doğrulaması
- **Gerçek Zamanlı Akıl Yürütme Beslemesi**: AI ajanlarının düşünce sürecini canlı izleme
- **Pazar Analiz Paneli**: Son haberleri ve etki seviyelerini gösteren dashboard
- **Veritabanı Desteği**: PostgreSQL'de karar zincirlerinin ve düşünce adımlarının depolanması
- **Redis Cache**: Haber cache'leme (30 dakika TTL) ve hızlı erişim

### 🔄 Sistem Mimarisi

#### Backend (Go)

- **News Aggregator**: 30 dakika aralıklarla yeni haberleri getir → Redis cache → WebSocket broadcast
- **Agent Engine**: Her agent için 30-60s aralıklarla:
  1. Piyasa verisi + haberleri topla
  2. AI client'a isteği gönder (haber bağlamıyla)
  3. Kararı kaydedilip düşünme adımlarını depola
  4. Risk Manager'dan geçir
  5. Trade'i çalıştır / reddet
  6. WebSocket'ten broadcast et
- **Risk Manager**: Confidence > 70%, position < 5%, portfolio risk < 20%
- **AI Clients**: OpenAI + Anthropic entegrasyonu
- **News System**: NewsAPI.org + RSS parser (Bloomberg HT, Investing.com, Dünya)

#### Frontend (Next.js)

- **ReasoningFeed**: Real-time AI decision stream (confidence, risk level, thinking steps)
- **LatestNews**: Market news display (impact level, related stocks, sentiment)
- **Dashboard**: Agents performance, P&L tracking, live status

### 📊 Karar Döngüsü

```
News Aggregator (30 min cycle)
    ↓
    [Fetch + Cache]
    ↓
Agent Engine (30-60s random per agent)
    ↓ (every cycle)
    ├─ Gather market data + recent news
    ├─ Call AI with context
    ├─ Store decision + thinking steps
    ├─ Validate with Risk Manager
    ├─ Execute/Reject trade
    └─ Broadcast via WebSocket
    ↓
Frontend ReasoningFeed + News Panel
    ↓
    [Real-time updates]
```

### 💰 Maliyet Tahminleri

**Test Modelleri (v0.3 default):**

- GPT-3.5-turbo: $0.001/req → ~$2-3/day
- Claude 3 Haiku: $0.00025/req → ~$0.5/day
- **Toplam**: ~$3-5/day

**Production (optional):**

- GPT-4-turbo: $0.01/req → ~$20-30/day
- Claude 3 Opus: $0.015/req → ~$15-25/day
- **Toplam**: ~$35-50/day

### 🚀 Başlangıç

```bash
# Backend (Go 1.23+)
cd cmd/server
go run main.go

# Frontend (Next.js 16)
cd frontend
npm run dev

# Docker (PostgreSQL + Redis)
docker-compose up -d
```

### 📁 Proje Yapısı

```
market-ai/
├── backend (Go)
│   ├── internal/
│   │   ├── models/        # Data models
│   │   ├── services/      # Business logic
│   │   ├── ai/            # AI clients + prompting
│   │   ├── news/          # News aggregation
│   │   ├── cache/         # Redis caching
│   │   └── config/        # Configuration
│   ├── migrations/        # Database schemas
│   └── cmd/server/        # Entry point
├── frontend (Next.js)
│   ├── components/        # React components
│   ├── lib/               # Utilities
│   └── app/               # Pages
└── docker-compose.yml     # Services
```

### 🔧 Amaç

- Farklı AI modellerini aynı veri/koşullarda karşılaştırmak
- Stratejilerin performansını ve karar alma dinamiklerini analiz etmek
- Backend altyapısını (API, DB, Cache) doğrulamak ve ölçümlemek
- Haber bağlamında yapılan kararların etkisini gözlemlemek

## ⚠️ Uyarı

Bu proje yalnızca deneysel ve eğitim/test amaçlıdır. Buradaki hiçbir çıktı, sinyal veya metrik yatırım tavsiyesi değildir; finansal kararlar için kullanılmamalıdır.

---

## 🚀 v0.4 – Çoklu AI Arena & Leaderboard

v0.4 ile sistem tekil ajanlardan rekabetçi çoklu yapay zekâ (8 farklı model) arenasına genişletildi.

### ✅ Hedefler

- 8 AI ajanı (OpenAI GPT-4 / GPT-4o-mini, Claude, Gemini, DeepSeek, Llama Groq, Mixtral, Grok)
- Canlı liderlik tablosu (ROI, Win Rate, P/L, Toplam Değer)
- Periyodik sıralama hesaplama (ağırlıklı skor formülü)
- WebSocket ile anlık güncelleme yayınları
- İstatistik tabloları: günlük, snapshot, head-to-head (temel şema)

### 🗄 Yeni Veritabanı Tabloları (Migration 005)

- `agent_performance_snapshots` – Saatlik/isteğe bağlı snapshot kayıtları
- `leaderboard_rankings` – Hesaplanmış sıralama ve rozetler
- `agent_matchups` – İki ajan arası kazanma-kaybetme takibi
- `agent_daily_stats` – Günlük toplu metrikler (wins, losses, volume, best/worst trade)
- Fonksiyon: `update_leaderboard_rankings()` – ROI, Win Rate, P/L ağırlıklı skor

### 🔢 Sıralama Formülü (Overall Rank)

$$ overall = (roi _ 0.4) + (win_rate _ 0.3) + ((total_profit_loss / 1000) \* 0.3) $$

### 🔌 Backend Ekleri

- Yeni AI client dosyaları: `google.go`, `deepseek.go`, `groq.go`, `mistral.go`, `xai.go`
- Leaderboard servisi: periyodik (env ile ayarlanabilir) güncelleme + WebSocket broadcast
- REST endpoint: `GET /api/v1/leaderboard`
- Konfigürasyon: `LEADERBOARD_UPDATE_INTERVAL` (saniye)

### 🖥 Frontend Ekleri

- `Leaderboard.tsx` – Canlı tablo, ROI rozetleri, P/L, Win Rate
- Dashboard entegrasyonu

### 🔑 Ortam Değişkenleri (v0.4)

`.env`:

```
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GOOGLE_API_KEY=
DEEPSEEK_API_KEY=
GROQ_API_KEY=
MISTRAL_API_KEY=
XAI_API_KEY=

AI_MODEL_GPT=gpt-4-turbo
AI_MODEL_GPT4_MINI=gpt-4o-mini
AI_MODEL_CLAUDE=claude-3-5-sonnet-20241022
AI_MODEL_GEMINI=gemini-1.5-pro
AI_MODEL_DEEPSEEK=deepseek-chat
AI_MODEL_LLAMA=llama-3.1-70b-versatile
AI_MODEL_MIXTRAL=open-mixtral-8x22b
AI_MODEL_GROK=grok-2-latest

AI_TEMPERATURE=0.7
AI_MAX_TOKENS=1500
LEADERBOARD_UPDATE_INTERVAL=60
```

### 📦 Migration Uygulama

```bash
psql -U marketai -d marketai_dev -f migrations/005_agent_stats.sql
```

### 🌱 Seed – Yeni Ajanlar

```sql
INSERT INTO agents (name, model, status, initial_balance, current_balance) VALUES
('Gemini Pro','gemini-1.5-pro','active',100000,100000),
('DeepSeek V3','deepseek-chat','active',100000,100000),
('GPT-4o Mini','gpt-4o-mini','active',100000,100000),
('Llama 3.1 70B','llama-3.1-70b-versatile','active',100000,100000),
('Mixtral 8x22B','open-mixtral-8x22b','active',100000,100000),
('Grok 2','grok-2-latest','active',100000,100000);

INSERT INTO agent_metrics (agent_id)
SELECT id FROM agents WHERE name IN ('Gemini Pro','DeepSeek V3','GPT-4o Mini','Llama 3.1 70B','Mixtral 8x22B','Grok 2')
ON CONFLICT (agent_id) DO NOTHING;
```

### 🔁 Servis Döngüsü

1. Leaderboard servisi `update_leaderboard_rankings()` fonksiyonunu her interval sonunda çağırır.
2. Sıralama sonuçlarını WebSocket ile `leaderboard_updated` olarak yayınlar.
3. Frontend `Leaderboard.tsx` ilk veriyi REST’ten çeker, sonra anlık güncellemeleri websocket’ten işler.

### 🧪 Doğrulama

```bash
# REST kontrol
curl http://localhost:8080/api/v1/leaderboard | jq

# WebSocket (örnek wscat)
wscat -c ws://localhost:8080/ws
# Mesaj tipini dinle: leaderboard_updated
```

### 💰 Maliyet Analizi (8 Ajan Tam Güç)

| Model                | Tahmini Maliyet / Gün |
| -------------------- | --------------------- |
| GPT-4 Turbo          | ~$14.40               |
| Claude 3.5 Sonnet    | ~$4.32                |
| Gemini 1.5 Pro       | ~$1.80                |
| Grok-2               | ~$2.88                |
| GPT-4o Mini          | ~$0.22                |
| DeepSeek V3          | ~$0.39                |
| Mixtral 8x22B        | ~$2.88                |
| Llama 3.1 70B (Groq) | $0.00                 |

**Toplam (Full Premium)** ≈ $27/gün (~$810/ay)
**Minimum (Budget Set)** ≈ $2–5/gün

### 💡 Aşamalı Maliyet Stratejisi

| Faz            | Modeller                              | Günlük Maliyet | Amaç                  |
| -------------- | ------------------------------------- | -------------- | --------------------- |
| Phase 1 (Test) | GPT-4o Mini, DeepSeek, Mixtral, Llama | ~$2            | Fonksiyonel doğrulama |
| Phase 2 (Demo) | + Gemini, Claude Haiku                | ~$8            | Demo sunumu           |
| Phase 3 (Prod) | + GPT-4, Claude Sonnet/Opus, Grok     | ~$27           | Rekabetçi analiz      |

### � Ortam Bayrakları ile Maliyet Kontrolü

`BUDGET_MODE` ve `ENABLE_PREMIUM_MODELS` bayrakları ile çağrı frekansı ve kayıtlı modelleri yönetebilirsin.

| Değişken                | Varsayılan | Etki                                                                                                    |
| ----------------------- | ---------- | ------------------------------------------------------------------------------------------------------- |
| `BUDGET_MODE`           | `false`    | `true` ise karar döngüsü 30–60 sn yerine 60–120 sn çalışır (istek sayısı azalır).                       |
| `ENABLE_PREMIUM_MODELS` | `true`     | `false` ise GPT-4, Claude (Sonnet/Opus), Grok kayıt edilmez; yalnızca bütçe dostu modeller aktif kalır. |

Örnek bütçe ayarı:

```env
BUDGET_MODE=true
ENABLE_PREMIUM_MODELS=false
```

### 🔧 Diğer Tasarruf Teknikleri

- Token azaltımı: `AI_TEMPERATURE` sabit tutup prompt içeriğini minimalize et.
- Snapshot seyrekliği: Snapshot kayıtlarını (ileride) 1 dk yerine 5 dk yap.
- Dinamik hız: Volatilite düşükken interval uzat, yükselince kısalt.
- Fallback: Premium yanıt hatasında otomatik Mixtral/Llama fallback.

### �🔮 v0.5 Öngörüleri

- Gerçek zamanlı BIST veri feed entegrasyonu
- Tarihsel backtest motoru
- Saat bazlı piyasa simülasyonu (09:30–18:00)
- Genişletilmiş risk/performans metrikleri (Sortino, Calmar)

### 🛡 Notlar

- Ajanlar gerçek para veya gerçek zamanlı canlı piyasa yerine simüle edilmiş veride karar verir.
- Maliyet hesapları tahmini (token/istek hacmine bağlı değişir). Gerçek kullanımda bütçe limiti koyun.

---
