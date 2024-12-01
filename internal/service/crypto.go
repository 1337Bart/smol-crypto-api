package service

import (
	"context"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/coingecko_client"
	"github.com/1337Bart/smol-crypto-api/internal/model"
	"github.com/1337Bart/smol-crypto-api/internal/repository/postgres"
	"github.com/1337Bart/smol-crypto-api/internal/repository/redis"
	"log"
	"sort"
	"time"
)

type ICryptoService interface {
	// coingecko operations
	StartPeriodicUpdates(ctx context.Context)
	UpdateCryptosSingle(ctx context.Context)

	// handler operations
	ListCryptos(ctx context.Context, page, limit int) ([]model.CryptoData, int, error)
}

type CryptoService struct {
	cache      redis.CryptoCache
	repository postgres.CryptoRepository
}

func NewCryptoService(cache redis.CryptoCache, repository postgres.CryptoRepository) *CryptoService {
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

// testing one-time update, in prod this will be replaced by StartPeriodicUpdates
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

const (
	pageCacheKeyFormat = "crypto:page:%d:%d" // Format: crypto:page:{pageNum}:{limit}
	cacheDuration      = 4 * time.Hour
)

// to jest pojebane:
// trzeba dopisac metody do repo/cache
// poprawic syntax
// chce wyciagac PER PAGE, a nie WSZYSTKO wiec to jest do zaorania
// po co mi total wyników??
func (s *CryptoService) ListCryptos(ctx context.Context, page, limit int) ([]model.CryptoData, int, error) {
	// Calculate offset
	offset := (page - 1) * limit

	// Try to get from Redis first (we'll get all data and paginate in memory since it's max 250 items)
	cryptos, err := s.cache.GetAllCryptos(ctx)
	if err == nil && len(cryptos) > 0 {
		// Sort by market rank
		sort.Slice(cryptos, func(i, j int) bool {
			return cryptos[i].MarketRank < cryptos[j].MarketRank
		})

		// Calculate total and apply pagination
		total := len(cryptos)
		end := offset + limit
		if end > total {
			end = total
		}

		if offset >= total {
			return []model.CryptoData{}, total, nil
		}

		return cryptos[offset:end], total, nil
	}

	// If not in Redis, get from PostgreSQL
	cryptos, err = s.repository.ListCryptos(ctx, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list cryptos from repository: %w", err)
	}

	// Sort by market rank
	sort.Slice(cryptos, func(i, j int) bool {
		return cryptos[i].MarketRank < cryptos[j].MarketRank
	})

	return cryptos, 250, nil // hardcoded total as we know it's always 250
}
