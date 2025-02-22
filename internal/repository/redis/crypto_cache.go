package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/1337Bart/smol-crypto-api/internal/model"
	"github.com/go-redis/redis/v8"
)

const ttlInHours = 4

type CryptoCache interface {
	SetCryptos(ctx context.Context, cryptos []model.CryptoData) error

	GetAllCryptos(ctx context.Context) ([]model.CryptoData, error)
	GetCryptosWithPagination(ctx context.Context, offset, limit int) ([]model.CryptoData, int, error)
}

type cryptoCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCryptoCache(client *redis.Client) CryptoCache {
	return &cryptoCache{
		client: client,
		ttl:    ttlInHours * time.Hour,
	}
}

func (c *cryptoCache) SetCryptos(ctx context.Context, cryptos []model.CryptoData) error {
	pipe := c.client.Pipeline()

	for _, crypto := range cryptos {
		key := fmt.Sprintf("crypto:id:%s", crypto.ID)
		data, err := json.Marshal(crypto)
		if err != nil {
			return fmt.Errorf("failed to marshal crypto data: %w", err)
		}

		pipe.Set(ctx, key, data, c.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *cryptoCache) GetAllCryptos(ctx context.Context) ([]model.CryptoData, error) {
	keys, err := c.client.Keys(ctx, "crypto:id:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get crypto keys: %w", err)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no crypto data found in cache")
	}

	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	cryptos := make([]model.CryptoData, 0, len(cmds))
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			continue
		}

		var crypto model.CryptoData
		if err := json.Unmarshal([]byte(data), &crypto); err != nil {
			continue
		}

		cryptos = append(cryptos, crypto)
	}

	return cryptos, nil
}
func (c *cryptoCache) GetCryptosWithPagination(ctx context.Context, offset, limit int) ([]model.CryptoData, int, error) {
	var cursor uint64
	var keys []string
	pattern := "crypto:id:*"

	for {
		var batch []string
		batch, cursor = c.client.Scan(ctx, cursor, pattern, 100).Val()

		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return nil, 0, fmt.Errorf("no crypto data found in cache")
	}

	sort.Strings(keys)

	total := len(keys)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset >= total {
		return []model.CryptoData{}, total, nil
	}

	pageKeys := keys[offset:end]

	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(pageKeys))

	for i, key := range pageKeys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	cryptos := make([]model.CryptoData, 0, len(cmds))
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			return nil, 0, fmt.Errorf("incomplete cache data")
		}

		var crypto model.CryptoData
		if err := json.Unmarshal([]byte(data), &crypto); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal crypto data: %w", err)
		}

		if time.Since(crypto.Timestamp) > 4*time.Hour {
			return nil, 0, fmt.Errorf("cache data too old")
		}

		cryptos = append(cryptos, crypto)
	}

	sort.Slice(cryptos, func(i, j int) bool {
		return cryptos[i].MarketRank < cryptos[j].MarketRank
	})

	return cryptos, total, nil
}
