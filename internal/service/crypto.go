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
}

func NewCryptoService(cache redis.CryptoCache, repository postgres.ICryptoRepository) *CryptoService {
	return &CryptoService{
		cache:      cache,
		repository: repository,
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

	if filter.Symbol == "" && filter.StartTime == nil && filter.EndTime == nil {
		cryptos, total, err := s.cache.GetCryptosWithPagination(ctx, offset, filter.Limit)
		if err == nil {
			return cryptos, total, nil
		}
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
