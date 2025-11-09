# 🚀 Market AI Deployment Guide

Bu doküman Market AI uygulamasının Fly.io üzerinde nasıl deploy edileceğini açıklar.

---

## 📋 Önkoşullar

1. **Fly.io hesabı** oluşturun: https://fly.io
2. **Fly CLI** kurulumu:
   ```bash
   curl -L https://fly.io/install.sh | sh
   ```
3. **GitHub Actions** için `FLY_API_TOKEN` secret'ı ekleyin

---

## 🐳 Local Docker Setup

### 1. Servisleri Başlat

```bash
make up
# veya
docker-compose up -d
```

### 2. Migrasyonları Uygula

```bash
docker exec -i marketai-postgres psql -U marketai -d marketai_dev < migrations/001_initial_schema.sql
docker exec -i marketai-postgres psql -U marketai -d marketai_dev < migrations/002_trading_tables.sql
# ... diğer migration'lar
```

### 3. Logları İzle

```bash
make logs
# veya
docker-compose logs -f
```

---

## ☁️ Fly.io Deployment

### 1. Fly.io'ya Giriş Yap

```bash
flyctl auth login
```

### 2. Uygulamayı Oluştur

```bash
flyctl apps create market-ai
```

### 3. PostgreSQL ve Redis Servislerini Bağla

```bash
# PostgreSQL
flyctl postgres create --name marketai-db
flyctl postgres attach marketai-db --app market-ai

# Redis (Fly.io Redis eklentisi varsa)
flyctl redis create --name marketai-redis
flyctl redis attach marketai-redis --app market-ai
```

### 4. Environment Variables Ayarla

```bash
flyctl secrets set \
  OPENAI_API_KEY=your_key \
  ANTHROPIC_API_KEY=your_key \
  DB_HOST=your_db_host \
  REDIS_HOST=your_redis_host
```

### 5. Deploy Et

```bash
flyctl deploy
```

### 6. Health Check

```bash
flyctl status
curl https://market-ai.fly.dev/health
```

---

## 🔄 CI/CD Pipeline

GitHub Actions otomatik olarak:
1. **Lint** kontrolü yapar
2. **Test** çalıştırır
3. **Docker image** build eder
4. **Fly.io'ya deploy** eder (sadece `main` branch için)

### GitHub Secrets

Repository Settings → Secrets → Actions:
- `FLY_API_TOKEN`: Fly.io API token'ı

---

## 📊 Monitoring

### Prometheus
- URL: `http://localhost:9090` (local)
- Metrics endpoint: `/api/v1/metrics/prometheus`

### Grafana
- URL: `http://localhost:3000` (local)
- Default credentials: `admin/admin`

---

## 🔧 Troubleshooting

### Uygulama başlamıyor
```bash
flyctl logs
flyctl status
```

### Database bağlantı hatası
```bash
flyctl postgres connect -a marketai-db
```

### Redis bağlantı hatası
```bash
flyctl redis connect -a marketai-redis
```

---

## 📝 Notlar

- Production'da mutlaka SSL kullanın (Fly.io otomatik sağlar)
- Environment variables'ı `.env` dosyasından değil, Fly.io secrets'tan yönetin
- Health check endpoint'i (`/health`) Fly.io tarafından otomatik kullanılır

---

**Son Güncelleme:** 2025-01-XX
**Versiyon:** v1.0

