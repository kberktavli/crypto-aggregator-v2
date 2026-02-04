package services

import "v2-trading-bot/internal/core/domain"

// CalculateSMA: Basit Hareketli Ortalama (Simple Moving Average) hesaplar.
// period: Son kaç mumu baz alacağız? (Örn: 14)
func CalculateSMA(candles []domain.Candle, period int) float64 {
	if len(candles) < period {
		return 0 // Yeterli veri yoksa hesaplama yapma
	}

	var sum float64
	// Son 'period' kadar mumu topla
	// Go'da slice işlemleri: candles[len-period : len]
	start := len(candles) - period
	subset := candles[start:]

	for _, c := range subset {
		sum += c.Close
	}

	return sum / float64(period)
}

// CalculateRSI: Göreceli Güç Endeksi (Relative Strength Index) hesaplar.
// Klasik RSI formülü: 100 - (100 / (1 + RS))
func CalculateRSI(candles []domain.Candle, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	var gains, losses float64

	// Son 'period' mumluk değişimlere bak
	start := len(candles) - period - 1
	subset := candles[start:]

	for i := 1; i < len(subset); i++ {
		change := subset[i].Close - subset[i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses -= change // Negatif değeri pozitife çevir
		}
	}

	if losses == 0 {
		return 100 // Hiç düşüş yoksa RSI 100'dür
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	rs := avgGain / avgLoss

	return 100 - (100 / (1 + rs))
}
