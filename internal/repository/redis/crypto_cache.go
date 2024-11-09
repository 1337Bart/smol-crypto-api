package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"

	"github.com/1337Bart/smol-crypto-api/internal/model"
)

type CryptoCache interface {
	GetPrice(ctx context.Context, symbol string) (*model.CryptoPrice, error)
	SetPrice(ctx context.Context, price *model.CryptoPrice, ttl time.Duration) error
	SetBatchPrices(ctx context.Context, prices []*model.CryptoPrice, ttl time.Duration) error
	GetBatchPrices(ctx context.Context, symbols []string) ([]*model.CryptoPrice, error)
	DeletePrice(ctx context.Context, symbol string) error
	ClearAllPrices(ctx context.Context) error
	GetAllPrices(ctx context.Context) ([]*model.CryptoPrice, error)
}

type cryptoCache struct {
	client *redis.Client
}

func NewCryptoCache(client *redis.Client) CryptoCache {
	return &cryptoCache{client: client}
}

func (c *cryptoCache) GetPrice(ctx context.Context, symbol string) (*model.CryptoPrice, error) {
	key := getPriceKey(symbol)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var price model.CryptoPrice
	err = json.Unmarshal(data, &price)
	if err != nil {
		return nil, err
	}

	return &price, nil
}

func (c *cryptoCache) SetPrice(ctx context.Context, price *model.CryptoPrice, ttl time.Duration) error {
	data, err := json.Marshal(price)
	if err != nil {
		return err
	}

	key := getPriceKey(price.Symbol)
	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *cryptoCache) SetBatchPrices(ctx context.Context, prices []*model.CryptoPrice, ttl time.Duration) error {
	pipe := c.client.Pipeline()

	for _, price := range prices {
		data, err := json.Marshal(price)
		if err != nil {
			return err
		}

		key := getPriceKey(price.Symbol)
		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *cryptoCache) GetBatchPrices(ctx context.Context, symbols []string) ([]*model.CryptoPrice, error) {
	pipe := c.client.Pipeline()

	// Create a map to store futures
	futures := make(map[string]*redis.StringCmd)

	// Queue up all the gets
	for _, symbol := range symbols {
		key := getPriceKey(symbol)
		futures[symbol] = pipe.Get(ctx, key)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	// Process results
	var prices []*model.CryptoPrice
	for symbol, future := range futures {
		data, err := future.Bytes()
		if err == redis.Nil {
			continue // Skip if key doesn't exist
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get price for symbol %s: %w", symbol, err)
		}

		var price model.CryptoPrice
		if err := json.Unmarshal(data, &price); err != nil {
			return nil, fmt.Errorf("failed to unmarshal price for symbol %s: %w", symbol, err)
		}
		prices = append(prices, &price)
	}

	return prices, nil
}

func (c *cryptoCache) DeletePrice(ctx context.Context, symbol string) error {
	key := getPriceKey(symbol)
	return c.client.Del(ctx, key).Err()
}

func (c *cryptoCache) ClearAllPrices(ctx context.Context) error {
	pattern := "crypto:price:*"
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

func (c *cryptoCache) GetAllPrices(ctx context.Context) ([]*model.CryptoPrice, error) {
	pattern := "crypto:price:*"
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) == 0 {
		return []*model.CryptoPrice{}, nil
	}

	pipe := c.client.Pipeline()
	futures := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		futures[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	var prices []*model.CryptoPrice
	for _, future := range futures {
		data, err := future.Bytes()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get price data: %w", err)
		}

		var price model.CryptoPrice
		if err := json.Unmarshal(data, &price); err != nil {
			return nil, fmt.Errorf("failed to unmarshal price data: %w", err)
		}
		prices = append(prices, &price)
	}

	return prices, nil
}

// Helper function to check if Redis is available
func (c *cryptoCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Helper function to get TTL of a key
func (c *cryptoCache) GetTTL(ctx context.Context, symbol string) (time.Duration, error) {
	key := getPriceKey(symbol)
	return c.client.TTL(ctx, key).Result()
}

func getPriceKey(symbol string) string {
	return fmt.Sprintf("crypto:price:%s", symbol)
}
