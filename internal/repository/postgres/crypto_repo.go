package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/1337Bart/smol-crypto-api/internal/model"
)

type CryptoRepository struct {
	db *sql.DB
}

func NewCryptoRepository(db *sql.DB) *CryptoRepository {
	return &CryptoRepository{db: db}
}

func (r *CryptoRepository) SavePrice(ctx context.Context, price *model.CryptoPrice) error {
	query := ` 
        INSERT INTO crypto_prices (symbol, price, timestamp) 
        VALUES ($1, $2, $3) 
        ON CONFLICT (symbol, timestamp) DO UPDATE 
        SET price = EXCLUDED.price 
    `

	_, err := r.db.ExecContext(ctx, query, price.Symbol, price.PriceUSD, price.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to save price: %w", err)
	}

	return nil
}

func (r *CryptoRepository) GetPriceHistory(ctx context.Context, symbol string, from, to time.Time) ([]*model.CryptoPrice, error) {
	query := ` 
        SELECT symbol, price, timestamp 
        FROM crypto_prices 
        WHERE symbol = $1 AND timestamp BETWEEN $2 AND $3 
        ORDER BY timestamp DESC 
    `

	rows, err := r.db.QueryContext(ctx, query, symbol, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query price history: %w", err)
	}
	defer rows.Close()

	var prices []*model.CryptoPrice
	for rows.Next() {
		price := &model.CryptoPrice{}
		err := rows.Scan(&price.Symbol, &price.PriceUSD, &price.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price row: %w", err)
		}
		prices = append(prices, price)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	return prices, nil
}

func (r *CryptoRepository) SaveBatchPrices(ctx context.Context, prices []*model.CryptoPrice) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, ` 
        INSERT INTO crypto_prices (symbol, price, timestamp) 
        VALUES ($1, $2, $3) 
        ON CONFLICT (symbol, timestamp) DO UPDATE 
        SET price = EXCLUDED.price 
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, price := range prices {
		_, err = stmt.ExecContext(ctx, price.Symbol, price.PriceUSD, price.Timestamp)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
