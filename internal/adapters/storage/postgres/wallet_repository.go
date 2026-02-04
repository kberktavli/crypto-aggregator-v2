package postgres

import (
	"context"
	"fmt"
	"time"
	"v2-trading-bot/internal/core/domain"
)

// InıtWalletTable : Cüzdan tablosunu oluşturur ve içine 1k$ başlangıc parası koyar.
// Bunu Repository.go içindeki NewRepository fonksiyonunda çağıracagız.
func (r *Repository) InitWalletTable() error {
	ctx := context.Background()
	// 1. tabloyu olustur
	query := `
	CREATE TABLE IF NOT EXISTS wallets (
		id TEXT PRIMARY KEY,
		usdt_balance DOUBLE PRECISION,
		coin_balance DOUBLE PRECISION,
		updated_at TIMESTAMPTZ
	);
	`
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("Wallet tablo hatası: %w", err)
	}

	// 2. başlangıc cüzdanını olusturuyoruz eğer yoksa
	// ıd:"demo", bakiye: 1k$ , coin : 0
	initQuery := `INSERT INTO wallets (id, usdt_balance, coin_balance, updated_at)
	VALUES ('demo', 1000.0, 0.0, NOW())
	ON CONFLICT (id) DO NOTHING;
	`

	_, err = r.db.Exec(ctx, initQuery)
	if err != nil {
		return fmt.Errorf("wallet init hatası: %w", err)
	}

	return nil

}

// GetWallet: Veritabanından cüzdanı çeker.
func (r *Repository) GetWallet() (*domain.Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, usdt_balance, coin_balance FROM wallets WHERE id = 'demo'`

	var w domain.Wallet
	err := r.db.QueryRow(ctx, query).Scan(&w.ID, &w.USDTBalance, &w.CoinBalance)
	if err != nil {
		return nil, fmt.Errorf("cüzdan bulunamadı: %w", err)
	}
	return &w, nil
}

// UpdateWallet: İşlem sonrası yeni bakiyeyi kaydeder.
func (r *Repository) UpdateWallet(w domain.Wallet) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	UPDATE wallets 
	SET usdt_balance = $1, coin_balance = $2, updated_at = NOW() 
	WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, w.USDTBalance, w.CoinBalance, w.ID)
	if err != nil {
		return fmt.Errorf("cüzdan güncellenemedi: %w", err)
	}
	return nil
}
