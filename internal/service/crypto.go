package service

import (
	"context"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/client/coingecko"
	"github.com/1337Bart/smol-crypto-api/internal/repository/postgres"
	"github.com/1337Bart/smol-crypto-api/internal/repository/redis"
	"log"
	"time"
)

const (
	cacheTTLinHours = 4
)

type CryptoUpdateService struct {
	gecko      *coingecko.CoinGeckoClient
	cache      redis.CryptoCache
	repository postgres.CryptoRepository
	coinIDs    []string
}

func NewCryptoUpdateService(cache redis.CryptoCache, repository postgres.CryptoRepository, coinIDs []string) *CryptoUpdateService {
	return &CryptoUpdateService{
		gecko:      coingecko.NewCoinGeckoClient(),
		cache:      cache,
		repository: repository,
		coinIDs:    coinIDs,
	}
}

func (s *CryptoUpdateService) StartPeriodicUpdates(ctx context.Context) {
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

// testing one-time update, in prod this will be replaced by StartPeriodicUpdates
func (s *CryptoUpdateService) UpdateCryptosSingle(ctx context.Context) {
	if err := s.updatePrices(ctx); err != nil {
		log.Printf("Initial price update failed: %v", err)
	}
}

func (s *CryptoUpdateService) updatePrices(ctx context.Context) error {
	cryptoData, err := s.gecko.FetchRecentCoinsData()
	if err != nil {
		return fmt.Errorf("failed to fetch cryptoData: %w", err)
	}

	// Store in Redis (hot data)
	if err := s.cache.BatchSave(ctx, cryptoData); err != nil {
		log.Printf("Failed to store cryptoData in Redis: %v", err)
		// Continue execution even if Redis fails
	}

	// Store in PostgreSQL (historical data)
	if err := s.repository.BatchSave(ctx, cryptoData); err != nil {
		return fmt.Errorf("failed to store cryptoData in PostgreSQL: %w", err)
	}

	return nil
}
