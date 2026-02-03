package services

import (
	"fmt"
	"v2-trading-bot/internal/core/domain"
	"v2-trading-bot/internal/core/ports"
)

type TradingService struct {
	repo      ports.CandleRepository
	publisher ports.EventBus
}

// NewTradingService : Servisi oluşturmak için kullanılan "constructor" fonksiyonudur.
func NewTradingService(repo ports.CandleRepository, publisher ports.EventBus) *TradingService {
	return &TradingService{
		repo:      repo,
		publisher: publisher,
	}
}

// Binance adapter'ından gelen mumu işleyen ana fonksiyon.
func (s *TradingService) ProcessIncomingCandle(candle domain.Candle) error {
	fmt.Printf("Yeni Mum Geldi : %s - Fiyat: %.2f\n", candle.Symbol, candle.Close)

	// Veriyi veritabanına kaydet (Driven port)
	err := s.repo.Save(candle)
	if err != nil {
		return fmt.Errorf("veritabanı kayıt hatası: %v", err)
	}

	// Veriyi Frontend'e (Centrifugo) ile bas (Driven port)
	err = s.publisher.PublishCandle(candle)
	if err != nil {
		return fmt.Errorf("Yayın hatası: %v\n", err)
	}

	// Strateji kontrolü - Strategy Logic, burada pyton kaynaklarından mantıgı kullanacagız. örn: son 14 mumu getir RSI hesapla rsı<30 ise s.publisher.publishsignal("AL") yapacagız
	// faz 4'de gelecek
	return nil
}
