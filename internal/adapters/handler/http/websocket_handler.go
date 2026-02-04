package handler

import (
	"log"
	"net/http"

	"github.com/centrifugal/centrifuge"
)

func StartWebSocketServer(node *centrifuge.Node) {

	// Centrifuge Handler
	wsHandler := centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// CORS: Herkese kapı açık
		CheckOrigin: func(r *http.Request) bool {
			log.Println("📞 TIK TIK! Biri kapıyı çalıyor: " + r.RemoteAddr)
			return true
		},
	})

	mux := http.NewServeMux()

	// Yolu tekrar netleştiriyoruz.
	mux.Handle("/connection/websocket", wsHandler)

	go func() {
		log.Println("🦅 WebSocket Sunucusu 8085 portunda (Path: /connection/websocket) başlatıldı...")
		// Yeni Port: 8085
		if err := http.ListenAndServe(":8085", mux); err != nil {
			log.Fatalf("WebSocket sunucusu başlatılamadı: %v", err)
		}
	}()
}
