# 🗄️ Fly.io Database Migrations

Bu dokümantasyon Fly.io'da PostgreSQL migrations'larını uygulamak için gerekli adımları içerir.

---

## 📋 Migrations Listesi

1. `001_initial_schema.sql` - Initial schema (system_info)
2. `002_trading_tables.sql` - Trading tables (agents, trades, portfolios)
3. `003_reasoning_tables.sql` - Reasoning tables (decisions, thinking_steps)
4. `004_news_tables.sql` - News tables (articles, sources)
5. `005_agent_stats.sql` - Agent statistics tables
6. `006_data_sources.sql` - Data sources tables (price_sources, twitter_sentiment, scraped_articles)
7. `007_data_sources_seed.sql` - Data sources seed data
8. `008_dynamic_universe.sql` - Dynamic universe tables (stocks, universe_history)

---

## 🔍 Tabloları Kontrol Et

### PostgreSQL'e Bağlan

```bash
fly postgres connect -a marketai-db
```

### Tabloları Listele

PostgreSQL console'da:

```sql
\dt
```

veya

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

---

## 🚀 Migrations Uygulama

### Yöntem 1: Fly.io PostgreSQL Console (Önerilen)

1. **PostgreSQL'e bağlan:**
```bash
fly postgres connect -a marketai-db
```

2. **Her migration'ı sırayla uygula:**

```sql
-- Migration 001
\i /path/to/migrations/001_initial_schema.sql

-- Migration 002
\i /path/to/migrations/002_trading_tables.sql

-- ... diğer migrations
```

**Not:** Bu yöntem için migration dosyalarını local'den kopyalamak gerekir.

### Yöntem 2: psql ile Connection String Kullan (Önerilen)

**Connection String:**
```
postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable
```

**Migrations uygula:**

```bash
# Her migration'ı sırayla uygula
psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < migrations/001_initial_schema.sql
psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < migrations/002_trading_tables.sql
psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < migrations/003_reasoning_tables.sql
psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < migrations/004_news_tables.sql
psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < migrations/005_agent_stats.sql
psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < migrations/006_data_sources.sql
psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < migrations/007_data_sources_seed.sql
psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < migrations/008_dynamic_universe.sql
```

**Veya tek komutla:**

```bash
for file in migrations/*.sql; do
  psql "postgres://marketai_backend:WEiflJW0ithygW2@marketai-db.flycast:5432/marketai_backend?sslmode=disable" < "$file"
done
```

### Yöntem 3: Fly.io SSH ile

1. **App'e SSH ile bağlan:**
```bash
fly ssh console -a marketai-backend
```

2. **Migration dosyalarını kopyala:**
```bash
# Local'den migration dosyalarını kopyala (gerekirse)
```

3. **psql ile uygula:**
```bash
psql $DATABASE_URL < /path/to/migration.sql
```

---

## ✅ Tabloları Doğrula

### Tüm Tabloları Listele

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

### Beklenen Tablolar

- `system_info`
- `agents`
- `trades`
- `portfolios`
- `decisions`
- `thinking_steps`
- `articles`
- `sources`
- `agent_stats`
- `price_sources`
- `twitter_sentiment`
- `scraped_articles`
- `stocks`
- `universe_history`

---

## 🔧 Sorun Giderme

### Connection String Çalışmıyor

1. **Connection string'i kontrol et:**
```bash
fly secrets list -a marketai-backend | grep DATABASE_URL
```

2. **Flycast network kullan:**
   - `marketai-db.flycast` (internal network)
   - `marketai-db.internal` (alternatif)

### Migration Hataları

1. **"relation already exists" hatası:**
   - Tablo zaten var, migration'ı atla
   - Veya `DROP TABLE IF EXISTS` kullan

2. **"permission denied" hatası:**
   - App kullanıcısı yerine admin kullanıcı kullan
   - `postgres://postgres:I2YBj37fGQJRrFk@marketai-db.flycast:5432`

### Tablolar Görünmüyor

1. **Schema kontrol et:**
```sql
SELECT current_schema();
```

2. **Tüm schemas'ları listele:**
```sql
SELECT schema_name FROM information_schema.schemata;
```

---

## 📝 Notlar

1. **Migrations sırası önemli:** Migration'ları sırayla uygula
2. **Backup al:** Production'da migration öncesi backup al
3. **Test et:** Migration'ları test ortamında önce test et
4. **Connection string güvenli:** Connection string'i güvenli tut

---

**Son Güncelleme:** 2025-01-XX

