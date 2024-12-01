package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/model"
	"log"
)

const logEveryNRecords = 25

type ICryptoRepository interface {
	BatchSave(ctx context.Context, prices []model.CryptoData) error
	ListCryptos(ctx context.Context, offset, limit int) ([]model.CryptoData, error)
}

type cryptoRepository struct {
	db *sql.DB
}

func NewCryptoRepository(db *sql.DB) ICryptoRepository {
	return &cryptoRepository{db: db}
}

func (r *cryptoRepository) BatchSave(ctx context.Context, prices []model.CryptoData) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `    
		INSERT INTO crypto_prices (    
			id, symbol, name, timestamp, current_price, high_24h, low_24h,    
			total_volume, market_cap, market_rank, price_change_24h,    
			price_change_percentage_24h, circulating_supply, total_supply    
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) 
		ON CONFLICT (id, timestamp) DO NOTHING;    
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for n, price := range prices {
		_, err := stmt.ExecContext(ctx,
			price.ID, price.Symbol, price.Name, price.Timestamp,
			price.CurrentPrice, price.High24h, price.Low24h,
			price.TotalVolume, price.MarketCap, price.MarketRank,
			price.PriceChange24h, price.PriceChangePercent24h,
			price.CirculatingSupply, price.TotalSupply,
		)

		if n%logEveryNRecords == 0 {
			log.Printf("Inserted %d out of %d, records", n, len(prices))
		}

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *cryptoRepository) ListCryptos(ctx context.Context, offset, limit int) ([]model.CryptoData, error) {
	query := ` 
        SELECT DISTINCT ON (id)  
            id, symbol, name, timestamp, current_price, high_24h, low_24h, 
            total_volume, market_cap, market_cap_rank, price_change_24h, 
            price_change_percentage_24h, circulating_supply, total_supply 
        FROM crypto_data 
        WHERE timestamp >= NOW() - INTERVAL '4 hours' 
        ORDER BY id, timestamp DESC, market_cap_rank ASC 
        LIMIT $1 OFFSET $2 
    `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query cryptos: %w", err)
	}
	defer rows.Close()

	var cryptos []model.CryptoData
	for rows.Next() {
		var crypto model.CryptoData
		err := rows.Scan(
			&crypto.ID, &crypto.Symbol, &crypto.Name, &crypto.Timestamp,
			&crypto.CurrentPrice, &crypto.High24h, &crypto.Low24h,
			&crypto.TotalVolume, &crypto.MarketCap, &crypto.MarketRank,
			&crypto.PriceChange24h, &crypto.PriceChangePercent24h,
			&crypto.CirculatingSupply, &crypto.TotalSupply,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan crypto row: %w", err)
		}
		cryptos = append(cryptos, crypto)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating crypto rows: %w", err)
	}

	return cryptos, nil
}
