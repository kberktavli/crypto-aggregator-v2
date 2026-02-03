package domain

import "time"

// Candle: Borsadan gelen her bir fiyat mumunu temsil eder.
// Centrifugo ile frontend'e dönerken json'a çevirecegiz.

type Candle struct {
	Symbol    string    `json:"symbol"`     // Örn: BTCUSDT
	Interval  string    `json:"interval"`   // Örn: 1m, 15m, 4h
	Open      float64   `json:"open"`       // Açılış Fiyatı
	High      float64   `json:"high"`       // En Yüksek
	Low       float64   `json:"low"`        // En Düşük
	Close     float64   `json:"close"`      // Kapanış
	Volume    float64   `json:"volume"`     // Hacim
	EventTime time.Time `json:"event_time"` // Mumun kapandığı/oluştuğu zaman
}

// TradeSignal: Stratejimizin ürettiği karar.
type TradeSignal struct {
	Symbol    string     `json:"symbol"`
	Action    SignalType `json:"action"` // buy,sell,hold
	Price     float64    `json:"price"`  // Sinyalin üretildigi anki fiyat.
	Timestamp time.Time  `json:"timestamp"`
	Reason    string     `json:"reason"`
}

// SignalType : Al veya Sat emrinin yönü
type SignalType string

const (
	SignalBuy  SignalType = "BUY"
	SignalSell SignalType = "SELL"
	SignalHold SignalType = "HOLD"
)
