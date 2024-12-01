package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/model"
	"github.com/go-redis/redis/v8"
	"time"
)

// todo - nietestowane

type CryptoCache interface {
	SaveOne(ctx context.Context, price *model.CryptoData) error
	BatchSave(ctx context.Context, prices []model.CryptoData) error
	Get(ctx context.Context, id string) (*model.CryptoData, error)
	GetAll(ctx context.Context) ([]model.CryptoData, error)
}

type cryptoCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCryptoCache(client *redis.Client) CryptoCache {
	return &cryptoCache{
		client: client,
		ttl:    4 * time.Hour,
	}
}

func (c *cryptoCache) SaveOne(ctx context.Context, price *model.CryptoData) error {
	key := fmt.Sprintf("crypto:price:%s", price.ID)
	data, err := json.Marshal(price)
	if err != nil {
		return fmt.Errorf("failed to marshal price data: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *cryptoCache) BatchSave(ctx context.Context, prices []model.CryptoData) error {
	pipe := c.client.Pipeline()

	for _, price := range prices {
		key := fmt.Sprintf("crypto:price:%s", price.ID)
		data, err := json.Marshal(price)
		if err != nil {
			return fmt.Errorf("failed to marshal price data: %w", err)
		}

		pipe.Set(ctx, key, data, c.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *cryptoCache) Get(ctx context.Context, id string) (*model.CryptoData, error) {
	key := fmt.Sprintf("crypto:price:%s", id)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var price model.CryptoData
	if err := json.Unmarshal(data, &price); err != nil {
		return nil, fmt.Errorf("failed to unmarshal price data: %w", err)
	}

	return &price, nil
}

func (c *cryptoCache) GetAll(ctx context.Context) ([]model.CryptoData, error) {
	keys, err := c.client.Keys(ctx, "crypto:price:*").Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []model.CryptoData{}, nil
	}

	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	prices := make([]model.CryptoData, 0, len(keys))
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue
		}

		var price model.CryptoData
		if err := json.Unmarshal(data, &price); err != nil {
			continue
		}
		prices = append(prices, price)
	}

	return prices, nil
}
