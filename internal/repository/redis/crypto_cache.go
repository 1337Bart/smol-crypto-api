package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/model"
	"github.com/go-redis/redis/v8"
	"time"
)

const ttlInHours = 4

type CryptoCache interface {
	BatchSave(ctx context.Context, prices []model.CryptoData) error
	GetAllCryptos(ctx context.Context) ([]model.CryptoData, error)
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

func (c *cryptoCache) BatchSave(ctx context.Context, cryptos []model.CryptoData) error {
	pipe := c.client.Pipeline()

	for _, crypto := range cryptos {
		key := fmt.Sprintf("crypto:id:%s", crypto.ID)
		data, err := json.Marshal(crypto)
		if err != nil {
			return fmt.Errorf("failed to marshal price data: %w", err)
		}

		pipe.Set(ctx, key, data, c.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// zrobic tu paginacje, zeby requesty byly jeszcze szybsze? zwrotka tylko tego, co paginacja potrzebuje
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
