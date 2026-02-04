package services

import (
	"fmt"
	"v2-trading-bot/internal/core/domain"
	"v2-trading-bot/internal/core/ports"
)

type TradingService struct {
	repo       ports.CandleRepository
	publisher  ports.EventBus
	walletRepo ports.WalletRepository
}

// NewTradingService : Servisi oluşturmak için kullanılan "constructor" fonksiyonudur.
func NewTradingService(repo ports.CandleRepository, walletRepo ports.WalletRepository, publisher ports.EventBus) *TradingService {
	return &TradingService{
		repo:       repo,
		publisher:  publisher,
		walletRepo: walletRepo,
	}
}

func (s *TradingService) ProcessIncomingCandle(candle domain.Candle) error {
	// 1. Log ve Yayın (Mevcut kodlar)
	// fmt.Printf(...) kaldırabilirsin, kirlilik yapmasın.

	// Veritabanına kaydet
	err := s.repo.Save(candle)
	if err != nil {
		return fmt.Errorf("veritabanı kayıt hatası: %v", err)
	}

	// Frontend'e canlı mumu gönder
	_ = s.publisher.PublishCandle(candle)

	// --- STRATEJİ BÖLÜMÜ (YENİ EKLENEN KISIM) ---

	// 2. Analiz için geçmiş veriyi çek (Örn: Son 20 mum lazım)
	// Stratejimiz RSI(14) kullanacak, o yüzden en az 15-20 mum lazım.
	pastCandles, err := s.repo.GetLatestCandles(candle.Symbol, 20)
	if err != nil {
		fmt.Printf("Geçmiş veri çekilemedi: %v\n", err)
		return nil // Akışı bozma
	}

	// Veritabanından veriler "Yeniden -> Eskiye" (DESC) gelir.
	// Ama hesaplama için bize "Eskiden -> Yeniye" (ASC) lazım.
	// O yüzden sıralamayı ters çevirmemiz (Reverse) gerekebilir.
	// (Basitlik adına şimdilik reverse fonksiyonu yazmadan, repository'de ORDER BY time ASC yapabiliriz
	//  AMA şimdilik DESC geldiğini varsayıp, hesaplamayı ona göre yapalım ya da basit bir reverse yapalım).

	// NOT: Repository'deki kodumuz DESC (En yeni en üstte) getiriyor.
	// İndikatör hesaplarken diziyi ters çevirmek en sağlıklısıdır.
	reverseCandles(pastCandles)

	// Yeterli veri var mı?
	if len(pastCandles) < 3 {
		fmt.Println("⚠️ Strateji için yeterli veri yok, veri birikmesi bekleniyor...")
		return nil
	}

	// 3. İndikatörleri Hesapla
	rsi := CalculateRSI(pastCandles, 3)
	sma := CalculateSMA(pastCandles, 3)

	fmt.Printf("📊 ANALİZ: %s | Fiyat: %.2f | RSI: %.2f | SMA: %.2f\n",
		candle.Symbol, candle.Close, rsi, sma)

	// 4. Karar Mekanizması (Basit Strateji)
	// Kural: RSI 30'un altındaysa (Oversold) -> AL
	// Kural: RSI 70'in üstündeyse (Overbought) -> SAT

	var signal domain.TradeSignal

	if rsi < 30 {
		signal = domain.TradeSignal{
			Symbol:    candle.Symbol,
			Action:    domain.SignalBuy,
			Price:     candle.Close,
			Timestamp: candle.EventTime,
			Reason:    fmt.Sprintf("RSI Aşırı Satım (%.2f < 30)", rsi),
		}
	} else if rsi > 70 {
		signal = domain.TradeSignal{
			Symbol:    candle.Symbol,
			Action:    domain.SignalSell,
			Price:     candle.Close,
			Timestamp: candle.EventTime,
			Reason:    fmt.Sprintf("RSI Aşırı Alım (%.2f > 70)", rsi),
		}
	}

	// 5. Eğer bir sinyal üretildiyse, bunu yayınla!
	if signal.Action != "" {
		fmt.Printf("🚨 SİNYAL ÜRETİLDİ: %s %s\n", signal.Action, signal.Reason)
		_ = s.publisher.PublishSignal(signal)
		// İleride buraya: s.exchange.ExecuteOrder(signal) gelecek (Paper Trading)
		s.ExecutePaperTrade(signal)
	}

	return nil
}

// Yardımcı Fonksiyon: Slice'ı ters çevirir
func reverseCandles(candles []domain.Candle) {
	for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
		candles[i], candles[j] = candles[j], candles[i]
	}
}
func (s *TradingService) ExecutePaperTrade(signal domain.TradeSignal) {
	// 1. Cüzdanı getir
	wallet, err := s.walletRepo.GetWallet()
	if err != nil {
		fmt.Printf("Cüzdan hatası : %v\n", err)
		return // Hata varsa devam etme
	}

	fmt.Printf("Cüzdan öncesi: %.2f USDT | %.5f BTC\n", wallet.USDTBalance, wallet.CoinBalance)

	// Değişiklik oldu mu diye kontrol etmek için bayrak
	tradeHappened := false

	// 2. İşlem mantıgı
	if signal.Action == domain.SignalBuy {
		// Alım: Tüm paramızla alıyoruz (All-in strategy)
		if wallet.USDTBalance > 10 { // en az 10 dolarımız varsa
			amountToBuy := wallet.USDTBalance / signal.Price // Kaç adet btc eder ?
			wallet.CoinBalance += amountToBuy
			wallet.USDTBalance = 0 // hepsini harcadık

			fmt.Printf("🟢 Alım yapıldı %.5f BTC alındı (Fiyat : %.2f)\n", amountToBuy, signal.Price)

			// Veritabanını güncelle
			s.walletRepo.UpdateWallet(*wallet)
			tradeHappened = true
		} else {
			fmt.Println("!! Yetersiz bakiye (usdt)")
		}

	} else if signal.Action == domain.SignalSell {
		// Satım: Elimdeki tüm btc'yi sat
		if wallet.CoinBalance > 0.0001 {
			amountUsdt := wallet.CoinBalance * signal.Price
			wallet.USDTBalance += amountUsdt
			wallet.CoinBalance = 0

			fmt.Printf("🔴 Satış Yapıldı: %.2f USDT kazanıldı Fiyat: %.2f\n", amountUsdt, signal.Price)

			// Veritabanını güncelle
			s.walletRepo.UpdateWallet(*wallet)
			tradeHappened = true
		} else {
			fmt.Println("!! Satılacak coin yok")
		}
	}

	fmt.Printf("Cüzdan sonrası: %.2f USDT | %.5f BTC\n", wallet.USDTBalance, wallet.CoinBalance)

	// 👇 KRİTİK EKLEME BURASI ŞEF 👇
	// Eğer işlem gerçekleştiyse, yeni bakiyeyi WebSocket'ten gönder
	if tradeHappened {
		update := domain.WalletUpdate{
			USDT: wallet.USDTBalance,
			BTC:  wallet.CoinBalance,
		}

		// Publisher üzerinden React'a fırlatıyoruz
		if err := s.publisher.PublishWallet(update); err != nil {
			fmt.Printf("⚠️ Cüzdan yayını başarısız: %v\n", err)
		} else {
			fmt.Println("📡 Cüzdan güncellendi ve frontend'e gönderildi.")
		}
	}
}
