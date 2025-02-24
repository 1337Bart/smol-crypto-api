package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/1337Bart/smol-crypto-api/internal/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const logEveryNRecords = 25

type ICryptoRepository interface {
	BatchSave(ctx context.Context, prices []model.CryptoData) error
	ListCryptos(ctx context.Context, filter model.CryptoFilter, offset int) ([]model.CryptoData, int, error)
}

type cryptoRepository struct {
	db     *sql.DB
	tracer trace.Tracer
}

func NewCryptoRepository(db *sql.DB, tracer trace.Tracer) ICryptoRepository {
	_, err := db.Exec(` 
        CREATE INDEX IF NOT EXISTS idx_crypto_prices_symbol_timestamp  
        ON crypto_prices (symbol, timestamp DESC) 
    `)
	if err != nil {
		log.Printf("Failed to create index: %v", err)
	}

	return &cryptoRepository{
		db:     db,
		tracer: tracer,
	}
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

func (r *cryptoRepository) ListCryptos(ctx context.Context, filter model.CryptoFilter, offset int) ([]model.CryptoData, int, error) {
	ctx, span := r.tracer.Start(ctx, "db_list_cryptos")
	defer span.End()
	span.SetAttributes(
		attribute.String("filter.symbol", filter.Symbol),
		attribute.Int("filter.limit", filter.Limit),
		attribute.Int("filter.offset", offset),
	)

	log.Printf("Created span with trace ID: %s", span.SpanContext().TraceID().String())

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	query := ` 
        SELECT id, symbol, name, timestamp, current_price, high_24h, low_24h, 
               total_volume, market_cap, market_rank, price_change_24h, 
               price_change_percentage_24h, circulating_supply, total_supply 
        FROM crypto_prices 
        WHERE 1=1 
    `
	countQuery := "SELECT COUNT(*) FROM crypto_prices WHERE 1=1"

	params := []interface{}{}
	paramCount := 1

	if filter.Symbol != "" {
		whereClause := fmt.Sprintf(" AND symbol = $%d", paramCount)
		query += whereClause
		countQuery += whereClause
		params = append(params, filter.Symbol)
		paramCount++
	}

	if filter.StartTime != nil {
		whereClause := fmt.Sprintf(" AND timestamp >= $%d", paramCount)
		query += whereClause
		countQuery += whereClause
		params = append(params, filter.StartTime)
		paramCount++
	}

	if filter.EndTime != nil {
		whereClause := fmt.Sprintf(" AND timestamp <= $%d", paramCount)
		query += whereClause
		countQuery += whereClause
		params = append(params, filter.EndTime)
		paramCount++
	}

	query += " ORDER BY market_rank ASC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount, paramCount+1)

	queryParams := append(params, filter.Limit, offset)

	var total int
	err = r.db.QueryRowContext(ctx, countQuery, params...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query cryptos: %w", err)
	}
	defer rows.Close()

	cryptos := make([]model.CryptoData, 0, filter.Limit)

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
			return nil, 0, fmt.Errorf("failed to scan crypto row: %w", err)
		}
		cryptos = append(cryptos, crypto)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating crypto rows: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return cryptos, total, nil
}
