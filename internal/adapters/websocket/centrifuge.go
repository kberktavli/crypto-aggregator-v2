package websocket

import (
	"context"
	"encoding/json"
	"log"

	"v2-trading-bot/internal/core/domain"

	"github.com/centrifugal/centrifuge"
)

type SocketService struct {
	Node *centrifuge.Node
}

func NewSocketService() *SocketService {
	cfg := centrifuge.Config{}

	node, err := centrifuge.New(cfg)
	if err != nil {
		log.Fatalf("Centrifuge motoru çalışmadı: %v", err)
	}

	// 👇 KRİTİK DÜZELTME BURADA 👇
	// Gelen kullanıcıya bir ID vermezsek sistem "Bad Request" verir!
	node.OnConnecting(func(ctx context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		return centrifuge.ConnectReply{
			// Herkese "misafir" veya rastgele bir ID veriyoruz.
			Credentials: &centrifuge.Credentials{
				UserID: "misafir_kullanici",
			},
		}, nil
	})

	node.OnConnect(func(client *centrifuge.Client) {
		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			cb(centrifuge.SubscribeReply{}, nil)
		})
		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			cb(centrifuge.PublishReply{}, nil)
		})
	})

	if err := node.Run(); err != nil {
		log.Fatalf("Centrifuge Run Hatası: %v", err)
	}

	return &SocketService{Node: node}
}

func (s *SocketService) PublishCandle(candle domain.Candle) error {
	data, err := json.Marshal(candle)
	if err != nil {
		return err
	}
	_, err = s.Node.Publish("kline", data)
	return err
}

func (s *SocketService) PublishSignal(signal domain.TradeSignal) error {
	data, err := json.Marshal(signal)
	if err != nil {
		return err
	}
	_, err = s.Node.Publish("signals", data)
	return err
}

func (s *SocketService) PublishWallet(update domain.WalletUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}
	_, err = s.Node.Publish("wallet", data)
	return err
}
