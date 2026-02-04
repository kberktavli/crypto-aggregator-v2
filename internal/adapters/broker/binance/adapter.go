package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"v2-trading-bot/internal/core/domain"
	"v2-trading-bot/internal/core/ports"

	"github.com/gorilla/websocket"
)

// BinanceAdapter: Dış dünyadan (Binance) veri akısını yöneten yapıdır.
type BinanceAdapter struct {
	service ports.TradingService
}

// NewBinanceAdapter: Adaptörü oluşturur.
func NewBinanceAdapter(service ports.TradingService) *BinanceAdapter {
	return &BinanceAdapter{
		service: service,
	}
}

// Connect: Belirtilen sembol için WebSocket bağlantısını başlatır.
// Bu fonksiyon "blocking" (kilitleyici) olmamalıdır. o yüzden goroutine içinde çağrılacak.
func (b *BinanceAdapter) Connect(symbol string) {
	// 1. Binance WebSocket URL'ini hazırla (küçük harf zorunlu: btcusdt)
	// format: wss://stream.binance.com:9443/ws/<symbol>@kline_<interval>
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@kline_1m", symbol)
	fmt.Printf("Binance'e bağlanılıyor: %s\n", url)

	// 2. Bağlantıyı kur (dial)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("WebSocket bağlantı hatası: %v", err)
	}
	// Fonksiyon bitince bağlantıyı kapatmayı unutma.
	defer conn.Close()

	// 3. Sonsuz döngü: Mesajları Dinle
	for {
		// Mesajı oku (readmessage programı durdurur, mesaj gelene kadar bekler)
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Okuma hatası: %v", err)
			// hata varsa 2 saniye bekletiyoruz, döngüye devam et
			time.Sleep(2 * time.Second)
			continue
		}

		// 4. Gelen JSON verisini go struct'ına çeviriyoruz
		var event BinanceKlineEvent
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("JSON parse hatası: %v", err)
			continue
		}

		// 5. Binance formatını -> Bizim Domain formatına (Candle) çevir (Mapping)
		// Sadece mum kapanmışsa (IsClosed = true) işleme alacağız
		if event.Kline.IsClosed {
			candle, err := event.ToDomain()
			if err != nil {
				log.Printf("Çeviri hatası: %v", err)
				continue
			}

			// 6. Core katmanını tetikle! (driving port)
			// adaptor, servise emri veriyor.
			go b.service.ProcessIncomingCandle(candle)
		}

	}
}

// DTO - Yardımcı Struct
type BinanceKlineEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		StartTime int64  `json:"t"`
		EndTime   int64  `json:"T"`
		Symbol    string `json:"s"`
		Interval  string `json:"i"`

		// FİYATLAR
		Open  string `json:"o"`
		Close string `json:"c"`
		High  string `json:"h"`

		// --- KRİTİK DÜZELTME BURADA ---
		Low         string `json:"l"` // Küçük 'l': Fiyat (String)
		LastTradeID int64  `json:"L"` // Büyük 'L': İşlem ID (Sayı) - Bunu ekleyince karışıklık biter!
		// -----------------------------

		Volume   string `json:"v"`
		IsClosed bool   `json:"x"`
	} `json:"k"`
}

// ToDomain: Binance formatını bizim temiz Domain Candle yapısına çevirir.
func (e *BinanceKlineEvent) ToDomain() (domain.Candle, error) {
	// String gelen fiyatları Float'a çevirmemiz lazım
	open, _ := strconv.ParseFloat(e.Kline.Open, 64)
	closePrice, _ := strconv.ParseFloat(e.Kline.Close, 64)
	high, _ := strconv.ParseFloat(e.Kline.High, 64)
	low, _ := strconv.ParseFloat(e.Kline.Low, 64)
	volume, _ := strconv.ParseFloat(e.Kline.Volume, 64)

	return domain.Candle{
		Symbol:   e.Symbol,
		Interval: e.Kline.Interval,
		Open:     open,
		Close:    closePrice,
		High:     high,
		Low:      low,
		Volume:   volume,
		// Unix milisaniyeyi -> Go Time nesnesine çevir
		EventTime: time.UnixMilli(e.EventTime),
	}, nil
}
