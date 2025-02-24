package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/1337Bart/smol-crypto-api/internal/coingecko_client"
	"github.com/1337Bart/smol-crypto-api/internal/model"
	"github.com/1337Bart/smol-crypto-api/internal/repository/postgres"
	"github.com/1337Bart/smol-crypto-api/internal/repository/redis"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ICryptoService interface {
	// coingecko operations
	StartPeriodicUpdates(ctx context.Context)
	UpdateCryptosSingle(ctx context.Context)

	// handler operations
	ListCryptos(ctx context.Context, filter model.CryptoFilter) ([]model.CryptoData, int, error)
}

type CryptoService struct {
	cache      redis.CryptoCache
	repository postgres.ICryptoRepository
	tracer     trace.Tracer
}

func NewCryptoService(cache redis.CryptoCache, repository postgres.ICryptoRepository, tracer trace.Tracer) *CryptoService {
	return &CryptoService{
		cache:      cache,
		repository: repository,
		tracer:     tracer,
	}
}

func (s *CryptoService) StartPeriodicUpdates(ctx context.Context) {
	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	if err := s.updatePrices(ctx); err != nil {
		log.Printf("Initial price update failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.updatePrices(ctx); err != nil {
				log.Printf("Price update failed: %v", err)
				continue
			}
		}
	}
}

func (s *CryptoService) UpdateCryptosSingle(ctx context.Context) {
	if err := s.updatePrices(ctx); err != nil {
		log.Printf("Initial price update failed: %v", err)
	}
}

func (s *CryptoService) updatePrices(ctx context.Context) error {
	gecko := coingecko_client.NewCoinGeckoClient()
	cryptoData, err := gecko.FetchRecentCoinsData()
	if err != nil {
		return fmt.Errorf("failed to fetch cryptoData: %w", err)
	}

	// store in Redis (hot data)
	if err := s.cache.SetCryptos(ctx, cryptoData); err != nil {
		log.Printf("Failed to store cryptoData in Redis: %v", err)
	}

	if err := s.repository.BatchSave(ctx, cryptoData); err != nil {
		return fmt.Errorf("failed to store cryptoData in PostgreSQL: %w", err)
	}

	return nil
}

func (s *CryptoService) ListCryptos(ctx context.Context, filter model.CryptoFilter) ([]model.CryptoData, int, error) {
	if filter.Page < 1 {
		return nil, 0, fmt.Errorf("page must be greater than 0")
	}
	if filter.Limit < 1 {
		return nil, 0, fmt.Errorf("limit must be greater than 0")
	}

	offset := (filter.Page - 1) * filter.Limit

	//if filter.Symbol == "" && filter.StartTime == nil && filter.EndTime == nil {
	//	// to zwraca co potrzebuje ale bez paginacji
	//	cryptos, err := s.cache.GetAllCryptos(ctx)
	//	// to zwraca gowno
	//	//cryptos, total, err := s.cache.GetCryptosWithPagination(ctx, offset, filter.Limit)
	//	if err == nil {
	//		return cryptos, 2, nil
	//	}
	//}
	if filter.Symbol == "" && filter.StartTime == nil && filter.EndTime == nil {
		ctx, span := s.tracer.Start(ctx, "cache_get_cryptos")
		defer span.End()

		cryptos, err := s.cache.GetAllCryptos(ctx)
		if err == nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			return cryptos, len(cryptos), nil
		}
		span.SetAttributes(attribute.Bool("cache.hit", false))
		span.RecordError(err)
	}

	cryptos, total, err := s.repository.ListCryptos(ctx, filter, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list cryptos from repository: %w", err)
	}

	// only cache if we're getting the latest data (no filters)
	if filter.Symbol == "" && filter.StartTime == nil && filter.EndTime == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.cache.SetCryptos(ctx, cryptos); err != nil {
				log.Printf("failed to update cache: %v", err)
			}
		}()
	}

	return cryptos, total, nil
}
