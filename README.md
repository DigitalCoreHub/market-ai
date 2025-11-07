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
