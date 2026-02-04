package postgres

import (
	"context"
	"fmt"
	"time"
	"v2-trading-bot/internal/core/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository: TimescaleDB ile konusan adaptörümüz.
// ports.CandleRepository interface'ini implemente eder.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(connString string) (*Repository, error) {
	// Veritabanı baglantı havuzu oluşturuyoruz.
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("DB bağlantı hatası: %w", err)
	}

	// Veritabanı tablo oluşturma
	// TimescaleDB için "event_time" kolonu önemlidir.
	query := `CREATE TABLE IF NOT EXISTS candles (
		time        TIMESTAMPTZ NOT NULL,
		symbol      TEXT NOT NULL,
		interval    TEXT NOT NULL,
		open        DOUBLE PRECISION,
		high        DOUBLE PRECISION,
		low         DOUBLE PRECISION,
		close       DOUBLE PRECISION,
		volume      DOUBLE PRECISION,
		PRIMARY KEY (time, symbol, interval)
	);
	`

	_, err = pool.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Tablo oluşturma hatası: %w", err)
	}

	// Tabloyu HyperTable'a çeviriyoruz (timescaleDB büyüsü imiş)
	// Bu komut standart Postgres tablosunu zaman serisi tablosuna çevirir.
	// Hata verirse (zaten hypertable ise) yoksayabiliriz ama temiz kod için kontrol etmek iyidir.
	// Basitlik adına direkt çalıştırıyoruz, hata verirse "zaten hypertable" diye varsayıyoruz şimdilik.
	hypertableQuery := `SELECT create_hypertable('candles','time', if_not_exists => TRUE);`
	_, err = pool.Exec(ctx, hypertableQuery)
	if err != nil {
		fmt.Printf("Hypertable uyarısı (önemli olmayabilir): %v\n", err)
	}

	// Cüzdan Tablosunu Hazırla (Paper Trading)
	if err := (&Repository{db: pool}).InitWalletTable(); err != nil {
		return nil, fmt.Errorf("cüzdan tablosu oluşturulamadı: %w", err)
	}
	return &Repository{db: pool}, nil
}

// Save : Tek bir mumu veritabanına yazar.
func (r *Repository) Save(candle domain.Candle) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	INSERT INTO candles (time, symbol, interval, open, high, low, close, volume)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT DO NOTHING;
	`

	// domain nesnesini sql parametrelerine çeviriyoruz
	_, err := r.db.Exec(ctx, query,
		candle.EventTime,
		candle.Symbol,
		candle.Interval,
		candle.Open,
		candle.High,
		candle.Low,
		candle.Close,
		candle.Volume,
	)
	if err != nil {
		return fmt.Errorf("Kayıt hatası: %w", err)
	}
	return nil
}

// GetLastCandles: Strateji hesaplaması için geçmiş veriyi çeker.
func (r *Repository) GetLatestCandles(symbol string, limit int) ([]domain.Candle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Zamanı Ters çevirip (desc) son gelenleri alıyoruz, sonra tekrar düzeltmek gerekebilir, şimdilik en yenileri en üste geliyor.
	query := `
	SELECT time, symbol, interval, open, high, low, close, volume
	FROM candles
	WHERE symbol = $1
	ORDER BY time DESC
	LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []domain.Candle
	for rows.Next() {
		var c domain.Candle
		err := rows.Scan(
			&c.EventTime,
			&c.Symbol,
			&c.Interval,
			&c.Open,
			&c.High,
			&c.Low,
			&c.Close,
			&c.Volume,
		)
		if err != nil {
			return nil, err
		}
		candles = append(candles, c)
	}
	return candles, nil
}
