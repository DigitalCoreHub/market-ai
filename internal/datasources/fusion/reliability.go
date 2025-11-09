package fusion

import "math"

// ComputeConfidence veri kaynağı başarı oranı, ortalama yanıt süresi (ms) ve çapraz kaynak
// varyansına (tek kaynak ise 0) göre fiyat anlık görüntüsü için 0-100 arası güven skoru hesaplar.
// Formül basit ve yorumlanabilir:
//
//	base = 50 + 40*successRate
//	responsePenalty = max(0, (avgMs-1500)/150) maksimum 10
//	variancePenalty = min(10, priceVarianceScaled)
//	confidence = clamp(base - responsePenalty - variancePenalty, 5, 99.9)
func ComputeConfidence(successRate float64, avgMs int, priceVariance float64) float64 {
	if successRate < 0 {
		successRate = 0
	}
	if successRate > 1 {
		successRate = 1
	}
	base := 50.0 + 40.0*successRate

	responsePenalty := 0.0
	if avgMs > 1500 {
		responsePenalty = float64(avgMs-1500) / 150.0
		if responsePenalty > 10 {
			responsePenalty = 10
		}
	}

	// Varyans cezasını ölçeklendir: varyansı TL^2 olarak değerlendir; ölçeği ılımlı tutmak için karekök kullan
	variancePenalty := math.Min(10.0, math.Sqrt(math.Abs(priceVariance)))

	conf := base - responsePenalty - variancePenalty
	if conf < 5 {
		conf = 5
	}
	if conf > 99.9 {
		conf = 99.9
	}
	return conf
}
