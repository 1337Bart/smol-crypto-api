package service

import (
	"context"
	"fmt"
	"time"

	"github.com/1337Bart/smol-crypto-api/internal/client/coingecko"
	"github.com/1337Bart/smol-crypto-api/internal/model"
	"github.com/1337Bart/smol-crypto-api/internal/repository/postgres"
	"github.com/1337Bart/smol-crypto-api/internal/repository/redis"
)

type CryptoService struct {
	cgClient   *coingecko.Client
	pgRepo     *postgres.CryptoRepository
	redisCache *redis.CryptoCache
}

func NewCryptoService(
	cgClient *coingecko.Client,
	pgRepo *postgres.CryptoRepository,
	redisCache *redis.CryptoCache,
) *CryptoService {
	return &CryptoService{
		cgClient:   cgClient,
		pgRepo:     pgRepo,
		redisCache: redisCache,
	}
}

func (s *CryptoService) GetCurrentPrice(ctx context.Context, symbol string) (*model.CryptoPrice, error) {
	// Try cache first
	if price, err := s.redisCache.GetPrice(ctx, symbol); err == nil {
		return price, nil
	}

	// If not in cache, get from external API
	price, err := s.cgClient.GetCurrentPrice(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get price from CoinGecko: %w", err)
	}

	// Store in cache and database
	if err := s.redisCache.SetPrice(ctx, price); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to cache price: %v\n", err)
	}

	if err := s.pgRepo.SavePrice(ctx, price); err != nil {
		fmt.Printf("failed to save price to database: %v\n", err)
	}

	return price, nil
}

func (s *CryptoService) GetHistoricalPrices(ctx context.Context, symbol string, from, to time.Time) ([]*model.CryptoPrice, error) {
	// Try database first
	prices, err := s.pgRepo.GetPriceHistory(ctx, symbol, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical prices: %w", err)
	}

	return prices, nil
}

func (s *CryptoService) GetTopCryptos(ctx context.Context, limit int) ([]*model.CryptoPrice, error) {
	// Try cache first
	if prices, err := s.redisCache.GetTopCryptos(ctx, limit); err == nil {
		return prices, nil
	}

	// If not in cache, get from external API
	prices, err := s.cgClient.GetTopCryptos(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top cryptos: %w", err)
	}

	// Store in cache
	if err := s.redisCache.SetTopCryptos(ctx, prices); err != nil {
		fmt.Printf("failed to cache top cryptos: %v\n", err)
	}

	return prices, nil
}
