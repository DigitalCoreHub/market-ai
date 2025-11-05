# Market AI - Function Documentation

Bu doküman projedeki tüm Go fonksiyonlarının ne işe yaradığını açıklar.

---

## 📁 internal/config/config.go

### `Load() (*Config, error)`
**Amaç:** .env dosyasından tüm konfigürasyonu okur ve Config struct'ını döndürür

**Ne yapar:**
- Viper kütüphanesini kullanarak .env dosyasını parse eder
- Server, Database, Redis ve Log ayarlarını yükler
- Environment değişkenlerini otomatik olarak override eder

**Kullanım:**
```go
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Port:", cfg.Server.Port)
```

**Dönüş:**
- Başarılı: Dolu Config struct'ı
- Hatalı: nil ve error

---

## 📁 pkg/logger/logger.go

### `Init(level string)`
**Amaç:** Zerolog logger'ını başlatır ve log seviyesini ayarlar

**Parametreler:**
- `level`: "debug", "info", "warn", "error"

**Ne yapar:**
- Log seviyesini global olarak ayarlar
- Development ortamında renkli console output kullanır
- Production ortamında JSON format kullanır

**Kullanım:**
```go
logger.Init("debug")
log.Info().Msg("Uygulama başladı")
```

### `Get() *zerolog.Logger`
**Amaç:** Global logger instance'ını döndürür

**Ne yapar:**
- Zerolog'un global logger'ını return eder
- Tüm dosyalardan aynı logger instance'ı kullanılır

**Kullanım:**
```go
logger := logger.Get()
logger.Error().Err(err).Msg("Hata oluştu")
```

---

## 📁 internal/database/postgres.go

### `NewPostgresPool(cfg DatabaseConfig) (*pgxpool.Pool, error)`
**Amaç:** PostgreSQL connection pool oluşturur ve bağlantıyı test eder

**Ne yapar:**
- DSN (Data Source Name) string'i oluşturur
- pgx/v5 ile connection pool başlatır
- Ping komutuyla database'e erişimi doğrular
- Pool sayesinde concurrent istekleri verimli yönetir

**Kullanım:**
```go
pool, err := database.NewPostgresPool(cfg.Database)
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

// Query örneği
row := pool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", userId)
```

**Avantajları:**
- Connection pooling (performans)
- Otomatik reconnection
- Concurrent safe

---

## 📁 internal/database/redis.go

### `NewRedisClient(cfg RedisConfig) (*redis.Client, error)`
**Amaç:** Redis client oluşturur ve bağlantıyı test eder

**Ne yapar:**
- Redis connection parametrelerini ayarlar
- PING komutuyla bağlantıyı doğrular
- Cache operasyonları için hazır client döner

**Kullanım:**
```go
client, err := database.NewRedisClient(cfg.Redis)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Cache işlemleri
client.Set(ctx, "key", "value", time.Hour)
val, err := client.Get(ctx, "key").Result()
```

**Use Case'ler:**
- Session storage
- Rate limiting
- Caching
- Real-time data

---

## 📁 internal/models/response.go

### `Response` Struct
**Amaç:** Tüm API endpoint'lerinde kullanılan standart yanıt formatı

**Alanlar:**
- `Success` (bool): İşlem başarılı mı?
- `Message` (string): Kullanıcıya gösterilecek mesaj
- `Data` (interface{}): Döndürülecek veri (herhangi bir tip olabilir)
- `Error` (string): Hata mesajı

**Örnek Kullanım:**
```go
// Başarılı yanıt
return c.JSON(models.Response{
    Success: true,
    Data: userData,
})

// Hata yanıtı
return c.Status(400).JSON(models.Response{
    Success: false,
    Error: "Geçersiz input",
})
```

### `HealthResponse` Struct
**Amaç:** /health endpoint'i için özel yanıt modeli

**Alanlar:**
- `Status` (string): "healthy" veya "unhealthy"
- `Services` (map[string]string): Her servisin ayrı durumu

**Kullanım:**
```go
response := models.HealthResponse{
    Status: "healthy",
    Services: map[string]string{
        "postgres": "healthy",
        "redis": "healthy",
    },
}
```

---

## 📁 internal/api/handlers/health.go

### `NewHealthHandler(db, redis) *HealthHandler`
**Amaç:** HealthHandler instance'ı oluşturur (Constructor)

**Dependency Injection:**
- PostgreSQL pool
- Redis client

**Kullanım:**
```go
handler := handlers.NewHealthHandler(db, redisClient)
```

### `Check(c *fiber.Ctx) error`
**Amaç:** Tüm servislerin sağlık durumunu kontrol eder

**Endpoint:** `GET /health`

**Ne kontrol eder:**
- PostgreSQL bağlantısı (Ping)
- Redis bağlantısı (Ping)

**Yanıt Kodları:**
- 200: Tüm servisler healthy
- 503: En az bir servis unhealthy

**Kullanım:**
```bash
curl http://localhost:8080/health
```

**Kullanım Alanları:**
- Kubernetes liveness/readiness probe
- Load balancer health check
- Monitoring sistemleri (Prometheus, Datadog)

### `Ping(c *fiber.Ctx) error`
**Amaç:** Basit connectivity test endpoint'i

**Endpoint:** `GET /api/v1/ping`

**Ne yapar:**
- Sadece "pong" yanıtı döner
- Hiçbir dependency kontrolü yapmaz
- API'nin çalıştığını hızlıca test etmek için

**Kullanım:**
```bash
curl http://localhost:8080/api/v1/ping
# {"success": true, "message": "pong"}
```

---

## 📁 internal/api/routes.go

### `SetupRoutes(app, handlers)`
**Amaç:** Tüm HTTP route'ları tanımlar ve handler'lara bağlar

**Ne yapar:**
- `/health` endpoint'ini tanımlar
- `/api/v1/*` route grubunu oluşturur
- Yeni endpoint'ler buraya eklenir

**Route Yapısı:**
```
/health              -> healthHandler.Check
/api/v1/ping        -> healthHandler.Ping
/api/v1/...         -> (gelecek endpoint'ler)
```

**Yeni Endpoint Ekleme:**
```go
v1.Post("/users", userHandler.Create)
v1.Get("/users/:id", userHandler.GetByID)
```

---

## 📁 internal/api/server.go

### `NewServer(cfg) *fiber.App`
**Amaç:** Fiber HTTP server oluşturur ve middleware'leri ekler

**Eklenen Middleware'ler:**
1. **Recover:** Panic'leri yakalar, server crash olmaz
2. **Logger:** HTTP isteklerini loglar
3. **CORS:** Cross-origin isteklere izin verir

**Kullanım:**
```go
app := api.NewServer(cfg)
```

**Middleware Sırası Neden Önemli:**
```
İstek → Recover → Logger → CORS → Route Handler → Yanıt
```

### `errorHandler(c, err) error`
**Amaç:** Global error handler - tüm hatalar buradan geçer

**Ne yapar:**
- Fiber.Error ise status code'unu kullanır
- Diğer hatalar için 500 döner
- Tutarlı JSON error response oluşturur

**Örnek Yanıt:**
```json
{
  "success": false,
  "error": "Database connection failed"
}
```

---

## 📁 cmd/server/main.go

### `main()`
**Amaç:** Uygulamanın giriş noktası - tüm başlatma işlemlerini yapar

**Başlatma Sırası:**
1. **Konfigürasyon yükleme** (.env)
2. **Logger başlatma** (zerolog)
3. **PostgreSQL bağlantısı** (connection pool)
4. **Redis bağlantısı** (client)
5. **HTTP server kurulumu** (Fiber)
6. **Graceful shutdown** bekleme

**Graceful Shutdown Nedir:**
- SIGINT (Ctrl+C) veya SIGTERM sinyali geldiğinde
- Yeni istekleri reddet
- Aktif istekleri tamamla
- Database connection'ları kapat
- Server'ı düzgünce kapat

**Neden Goroutine Kullanılır:**
```go
go func() {
    app.Listen(":8080") // Blocking operation
}()
// Ana thread'de shutdown sinyali bekle
```

**Defer Kullanımı:**
```go
defer db.Close()        // main bitince çalışır
defer redisClient.Close()
```

---

## 🎯 Fonksiyon Kullanım Akışı

### Uygulama Başlatma:
```
main()
  ├─> config.Load()
  ├─> logger.Init()
  ├─> database.NewPostgresPool()
  ├─> database.NewRedisClient()
  ├─> api.NewServer()
  │    └─> errorHandler()
  ├─> handlers.NewHealthHandler()
  └─> api.SetupRoutes()
```

### HTTP İstek Akışı:
```
GET /health
  ├─> Recover Middleware
  ├─> Logger Middleware
  ├─> CORS Middleware
  ├─> healthHandler.Check()
  │    ├─> db.Ping()
  │    └─> redis.Ping()
  └─> JSON Response
```

---

## 📊 Fonksiyon İstatistikleri

- **Toplam Fonksiyon:** 12
- **Constructor Fonksiyon:** 4 (New*)
- **HTTP Handler:** 2 (Check, Ping)
- **Setup Fonksiyon:** 3 (Init, Load, SetupRoutes)
- **Error Handler:** 1 (errorHandler)

---

## 🔥 En Sık Kullanılacak Fonksiyonlar

1. **config.Load()** - Her başlatmada
2. **logger.Init()** - Her başlatmada
3. **NewPostgresPool()** - Database operasyonları için
4. **NewRedisClient()** - Cache operasyonları için
5. **healthHandler.Check()** - Monitoring için

---

## 💡 Best Practices

### 1. Error Handling
```go
if err != nil {
    log.Error().Err(err).Msg("Açıklayıcı mesaj")
    return c.Status(500).JSON(models.Response{
        Success: false,
        Error: err.Error(),
    })
}
```

### 2. Context Kullanımı
```go
ctx := context.Background()
// veya
ctx := c.Context() // Fiber context
```

### 3. Defer ile Cleanup
```go
pool, err := NewPostgresPool(cfg)
if err != nil {
    return err
}
defer pool.Close() // Fonksiyon bitince otomatik kapat
```

### 4. Logging
```go
log.Info().
    Str("user_id", userId).
    Int("count", count).
    Msg("İşlem tamamlandı")
```

---

**Son Güncelleme:** 5 Kasım 2025
**Versiyon:** 0.1
**Hazırlayan:** Market AI Development Team
