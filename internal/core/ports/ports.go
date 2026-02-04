package ports

import "v2-trading-bot/internal/core/domain"

// ---Driven Ports (Gidenler/Çıkış Kapıları)---
// Kodumuzun dış dünyaya ihtiyaç duydugu yerler.

// Veritabanı işlemleri için interface.
type CandleRepository interface {
	Save(candle domain.Candle) error
	GetLatestCandles(symbol string, limit int) ([]domain.Candle, error)
}

// Mesajlaşma işlemleri için interface (Centrifugo).
type EventBus interface {
	PublishCandle(candle domain.Candle) error
	PublishSignal(signal domain.TradeSignal) error

	PublishWallet(update domain.WalletUpdate) error
}

// ---Driving Ports(Gelenler/Giriş Kapıları)---
// Dış dünyanın bizim kodumuzu tetikledigi yerler.

// Uygulamanın ana beyni
// Handler'lar (HTTP veya WebSocket) bu servisi çağıracak.
type TradingService interface {
	// Dışarıdan yeni bir mum geldiginde bu çalışacak.
	ProcessIncomingCandle(candle domain.Candle) error
}

type WalletRepository interface {
	GetWallet() (*domain.Wallet, error)
	UpdateWallet(wallet domain.Wallet) error
}

/*
	Driving (Giriş): "Hey Service, al sana yeni veri!" (Dışarıdan içeriye).
	Driven (Çıkış): "Hey Database, al bunu sakla!" (İçeriden dışarıya).
*/
