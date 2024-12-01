package postgres

import (
	"context"
	"database/sql"
	"github.com/1337Bart/smol-crypto-api/internal/model"
	"log"
	"time"
)

// todo - nietestowane

type CryptoRepository interface {
	SaveOne(ctx context.Context, crypto *model.CryptoData) error
	GetByID(ctx context.Context, id string, timestamp time.Time) (*model.CryptoData, error)
	GetLatestByID(ctx context.Context, id string) (*model.CryptoData, error)
	//GetTimeRange(ctx context.Context, id string, start, end time.Time) ([]model.CryptoData, error)
	BatchSave(ctx context.Context, prices []model.CryptoData) error
}

type cryptoRepository struct {
	db *sql.DB
}

func NewCryptoRepository(db *sql.DB) CryptoRepository {
	return &cryptoRepository{db: db}
}

func (r *cryptoRepository) SaveOne(ctx context.Context, crypto *model.CryptoData) error {
	query := `  
        INSERT INTO crypto_prices (  
            id, symbol, name, timestamp, current_price, high_24h, low_24h,  
            total_volume, market_cap, market_rank, price_change_24h,  
            price_change_percentage_24h, circulating_supply, total_supply  
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, query,
		crypto.ID, crypto.Symbol, crypto.Name, crypto.Timestamp, crypto.CurrentPrice,
		crypto.High24h, crypto.Low24h, crypto.TotalVolume, crypto.MarketCap,
		crypto.MarketRank, crypto.PriceChange24h, crypto.PriceChangePercent24h,
		crypto.CirculatingSupply, crypto.TotalSupply)

	return err
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

		if n%10 == 0 {
			log.Printf("Inserted %d out of %d, records", n, len(prices))
		}

		if err != nil {
			log.Printf("Error inserting crypto %s (%s): %v", price.Name, price.ID, err)
			log.Printf("Values: current_price=%.8f, market_cap=%.2f, total_volume=%.2f",
				price.CurrentPrice, price.MarketCap, price.TotalVolume)
			return err
		}
	}

	return tx.Commit()
}

// GetByID retrieves a crypto price record for a specific ID and timestamp
func (r *cryptoRepository) GetByID(ctx context.Context, id string, timestamp time.Time) (*model.CryptoData, error) {
	query := `  
        SELECT * FROM crypto_prices   
        WHERE id = $1 AND timestamp = $2`

	crypto := &model.CryptoData{}
	err := r.db.QueryRowContext(ctx, query, id, timestamp).Scan(
		&crypto.ID, &crypto.Symbol, &crypto.Name, &crypto.Timestamp,
		&crypto.CurrentPrice, &crypto.High24h, &crypto.Low24h,
		&crypto.TotalVolume, &crypto.MarketCap, &crypto.MarketRank,
		&crypto.PriceChange24h, &crypto.PriceChangePercent24h,
		&crypto.CirculatingSupply, &crypto.TotalSupply)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return crypto, err
}

// GetLatestByID retrieves the most recent price for a crypto
func (r *cryptoRepository) GetLatestByID(ctx context.Context, id string) (*model.CryptoData, error) {
	query := `  
        SELECT * FROM crypto_prices   
        WHERE id = $1   
        ORDER BY timestamp DESC   
        LIMIT 1`

	crypto := &model.CryptoData{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&crypto.ID, &crypto.Symbol, &crypto.Name, &crypto.Timestamp,
		&crypto.CurrentPrice, &crypto.High24h, &crypto.Low24h,
		&crypto.TotalVolume, &crypto.MarketCap, &crypto.MarketRank,
		&crypto.PriceChange24h, &crypto.PriceChangePercent24h,
		&crypto.CirculatingSupply, &crypto.TotalSupply)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return crypto, err
}
