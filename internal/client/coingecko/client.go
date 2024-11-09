package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL              = "https://api.coingecko.com/api/v3"
	simplePriceEndpoint  = "/simple/price"
	coinsMarketsEndpoint = "/coins/markets"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type PriceResponse map[string]map[string]float64

type MarketData struct {
	ID                       string    `json:"id"`
	Symbol                   string    `json:"symbol"`
	Name                     string    `json:"name"`
	CurrentPrice             float64   `json:"current_price"`
	MarketCap                float64   `json:"market_cap"`
	Volume24h                float64   `json:"total_volume"`
	PriceChange24h           float64   `json:"price_change_24h"`
	PriceChangePercentage24h float64   `json:"price_change_percentage_24h"`
	LastUpdated              time.Time `json:"last_updated"`
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

// GetBatchPrices fetches prices for multiple cryptocurrencies in one call
func (c *Client) GetBatchPrices(ctx context.Context, ids []string, vsCurrency string) (PriceResponse, error) {
	url := fmt.Sprintf("%s%s?ids=%s&vs_currencies=%s",
		c.baseURL,
		simplePriceEndpoint,
		strings.Join(ids, ","),
		vsCurrency,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var priceResponse PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return priceResponse, nil
}

// GetMarketData fetches comprehensive market data for multiple cryptocurrencies
func (c *Client) GetMarketData(ctx context.Context, vsCurrency string, ids []string, perPage int) ([]MarketData, error) {
	url := fmt.Sprintf("%s%s?vs_currency=%s&ids=%s&order=market_cap_desc&per_page=%d&sparkline=false",
		c.baseURL,
		coinsMarketsEndpoint,
		vsCurrency,
		strings.Join(ids, ","),
		perPage,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var marketData []MarketData
	if err := json.NewDecoder(resp.Body).Decode(&marketData); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return marketData, nil
}

// RetryableClient wraps the basic client with retry functionality
type RetryableClient struct {
	client      *Client
	maxRetries  int
	backoffBase time.Duration
}

func NewRetryableClient(timeout time.Duration, maxRetries int, backoffBase time.Duration) *RetryableClient {
	return &RetryableClient{
		client:      NewClient(timeout),
		maxRetries:  maxRetries,
		backoffBase: backoffBase,
	}
}

func (rc *RetryableClient) GetBatchPricesWithRetry(ctx context.Context, ids []string, vsCurrency string) (PriceResponse, error) {
	var lastErr error
	for attempt := 0; attempt < rc.maxRetries; attempt++ {
		prices, err := rc.client.GetBatchPrices(ctx, ids, vsCurrency)
		if err == nil {
			return prices, nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(rc.backoffBase * time.Duration(attempt+1)):
			continue
		}
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (rc *RetryableClient) GetMarketDataWithRetry(ctx context.Context, vsCurrency string, ids []string, perPage int) ([]MarketData, error) {
	var lastErr error
	for attempt := 0; attempt < rc.maxRetries; attempt++ {
		data, err := rc.client.GetMarketData(ctx, vsCurrency, ids, perPage)
		if err == nil {
			return data, nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(rc.backoffBase * time.Duration(attempt+1)):
			continue
		}
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// RateLimitedClient adds rate limiting to the retryable client
type RateLimitedClient struct {
	*RetryableClient
	rateLimiter chan struct{}
}

func NewRateLimitedClient(timeout time.Duration, maxRetries int, backoffBase time.Duration, requestsPerMinute int) *RateLimitedClient {
	client := &RateLimitedClient{
		RetryableClient: NewRetryableClient(timeout, maxRetries, backoffBase),
		rateLimiter:     make(chan struct{}, requestsPerMinute),
	}

	// Refill rate limiter
	go func() {
		ticker := time.NewTicker(time.Minute / time.Duration(requestsPerMinute))
		defer ticker.Stop()

		for range ticker.C {
			select {
			case client.rateLimiter <- struct{}{}:
			default:
			}
		}
	}()

	return client
}

func (rlc *RateLimitedClient) GetBatchPrices(ctx context.Context, ids []string, vsCurrency string) (PriceResponse, error) {
	select {
	case <-rlc.rateLimiter:
		return rlc.RetryableClient.GetBatchPricesWithRetry(ctx, ids, vsCurrency)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (rlc *RateLimitedClient) GetMarketData(ctx context.Context, vsCurrency string, ids []string, perPage int) ([]MarketData, error) {
	select {
	case <-rlc.rateLimiter:
		return rlc.RetryableClient.GetMarketDataWithRetry(ctx, vsCurrency, ids, perPage)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
