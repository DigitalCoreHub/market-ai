# Güvenilirlik Skorlaması & Gözlemlenebilirlik

## Genel Bakış

Market AI v0.5, çoklu kaynaklı piyasa verisi füzyonu için **güvenilirlik skorlaması** ve **gözlemlenebilirlik metrikleri** sunar. Yahoo Finance'dan (ve gelecekteki kaynaklardan) gelen her fiyat teklifi, geçmiş başarı oranları, yanıt süreleri ve kaynak varyansına göre **güven skoru** (0-100) alır.

## Bileşenler

### 1. Güvenilirlik Skorlaması (`internal/datasources/fusion/reliability.go`)

Formül:

```
base = 50 + 40 * successRate
responsePenalty = max(0, (avgMs - 1500) / 150) maksimum 10
variancePenalty = min(10, sqrt(priceVariance))
confidence = clamp(base - responsePenalty - variancePenalty, 5, 99.9)
```

- **successRate**: başarılı sorguların oranı (0.0–1.0)
- **avgMs**: ortalama yanıt süresi (ms); başlangıç değeri 1500 ms (Yahoo 15 dakika gecikme)
- **priceVariance**: kaynaklararası fiyat varyansı (tek kaynak için 0; çoklu kaynak için gelecekte)

### 2. Veri Kaynakları Tablosu (`migrations/006_data_sources.sql`)

Takip edilenler:

- `total_fetches`, `success_count`, `error_count`
- `avg_response_time_ms`
- `last_fetch_at`, `status`, `last_error`

`migrations/007_data_sources_seed.sql` ile başlatılan kaynaklar:

- Yahoo Finance API
- Bloomberg HT Scraper
- Twitter API Search

### 3. Füzyon Servisi Güncellemeleri (`internal/datasources/fusion/service.go`)

- **Bellek içi istatistikler** (`sourceStats`): başarı/hata ve süre birikimi için hafif sayaçlar
- **`recordFetch`**: bellek içi istatistikleri günceller + `data_sources` tablosunu asenkron günceller
- **`ComputeConfidence`**: `price_sources`'a kaydedilmeden önce fiyat başına uygulanır

### 4. Metrik Endpoint'i (`/api/v1/metrics`)

Handler: `internal/api/handlers/metrics_handler.go`

Tüm veri kaynaklarının canlı istatistiklerini içeren JSON dizisi döner:

```json
{
  "success": true,
  "data": {
    "data_sources": [
      {
        "source_type": "yahoo",
        "source_name": "Yahoo Finance API",
        "is_active": true,
        "total_fetches": 120,
        "success_count": 118,
        "error_count": 2,
        "avg_response_time_ms": 850,
        "status": "active",
        "last_fetch_at": "2025-11-08T12:34:56Z"
      }
    ]
  }
}
```

## Kullanım

### Migration Çalıştırma

```bash
docker exec -i marketai-postgres psql -U marketai -d marketai_dev < migrations/007_data_sources_seed.sql
```

### Metrikleri Sorgulama

```bash
curl http://localhost:8080/api/v1/metrics | jq
```

### Güvenilirlik Skorlaması Testi

```bash
go test ./internal/datasources/fusion -v -run TestComputeConfidence
```

## Gelecek Geliştirmeler

- Çoklu kaynak fiyat füzyonu (güvene göre ortalama veya ağırlıklı)
- Çakışma tespiti ve çözümü (kaynaklar %5'ten fazla farklıysa uyarı)
- Grafana dashboardları için Prometheus exporter
- Güvenilir olmayan kaynakları dinamik kapatma (status -> 'inactive')

## Referanslar

- Güvenilirlik mantığı: `internal/datasources/fusion/reliability.go`
- Testler: `internal/datasources/fusion/reliability_test.go`
- Metrik handler: `internal/api/handlers/metrics_handler.go`
- Veri modeli: `migrations/006_data_sources.sql` + `007_data_sources_seed.sql`
