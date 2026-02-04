package main

import (
	"log"
	"time"

	"v2-trading-bot/internal/adapters/broker/binance"
	"v2-trading-bot/internal/adapters/storage/postgres"
	"v2-trading-bot/internal/adapters/websocket"
	"v2-trading-bot/internal/core/services"

	// Handler paketini import et (Kendi yoluna göre güncelle)
	httpHandler "v2-trading-bot/internal/adapters/handler/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	dbURL := "postgres://postgres:password@localhost:5432/trading_v2"

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// --- 1. DB ---
	log.Println("🔌 TimescaleDB'ye bağlanılıyor...")
	repo, err := postgres.NewRepository(dbURL)
	if err != nil {
		log.Fatalf("❌ Veritabanı hatası: %v", err)
	}

	// --- 2. WEBSOCKET (Embedded) ---
	log.Println("🔌 WebSocket Motoru başlatılıyor...")
	socketService := websocket.NewSocketService()

	// DÜZELTME 5: Handler'ı artık doğru çağırıyoruz.
	// (Node'u doğrudan vermiyoruz, handler fonksiyonunu çağırıyoruz)
	httpHandler.StartWebSocketServer(socketService.Node)

	// --- 3. CORE & BINANCE ---
	// socketService artık PublishCandle metoduna sahip olduğu için hata vermeyecek
	tradingService := services.NewTradingService(repo, repo, socketService)

	binanceAdapter := binance.NewBinanceAdapter(tradingService)

	go func() {
		log.Println("🚀 Binance WebSocket başlatılıyor (BTCUSDT)...")
		time.Sleep(2 * time.Second)
		binanceAdapter.Connect("btcusdt")
	}()

	// --- 4. START ---
	log.Println("🦅 Sunucu 3000 portunda hazır!")
	log.Fatal(app.Listen(":3000"))
}
