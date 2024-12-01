package coingecko

import (
	"encoding/json"
	"fmt"
	"github.com/1337Bart/smol-crypto-api/internal/model"
	"io"
	"net/http"
	"time"
)

const (
	baseURL              = "https://api.coingecko.com/api/v3"
	coinsMarketsEndpoint = "/coins/markets"
)

type CoinGeckoClient struct {
	client  *http.Client
	baseURL string
}

func NewCoinGeckoClient() *CoinGeckoClient {
	return &CoinGeckoClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}
}

// CoinGeckoResponse represents the raw response from CoinGecko
type CoinGeckoResponse []struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	Name                     string  `json:"name"`
	CurrentPrice             float64 `json:"current_price"`
	MarketCap                float64 `json:"market_cap"`
	MarketCapRank            int     `json:"market_cap_rank"`
	TotalVolume              float64 `json:"total_volume"`
	High24h                  float64 `json:"high_24h"`
	Low24h                   float64 `json:"low_24h"`
	PriceChange24h           float64 `json:"price_change_24h"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
	CirculatingSupply        float64 `json:"circulating_supply"`
	TotalSupply              float64 `json:"total_supply"`
	LastUpdated              string  `json:"last_updated"`
}

// FetchRecentCoinsData gets top 250 coins by market cap
// batching request to save on API calls
func (c *CoinGeckoClient) FetchRecentCoinsData() ([]model.CryptoData, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=250&page=1&sparkline=false&price_change_percentage=24h&locale=en",
		c.baseURL)

	// Make the request
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Decode the response
	var geckoResp CoinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&geckoResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Transform to our CryptoData model
	prices := make([]model.CryptoData, len(geckoResp))
	for i, coin := range geckoResp {
		// Parse the timestamp
		timestamp, err := time.Parse(time.RFC3339, coin.LastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp for %s: %w", coin.ID, err)
		}

		prices[i] = model.CryptoData{
			ID:                    coin.ID,
			Symbol:                coin.Symbol,
			Name:                  coin.Name,
			Timestamp:             timestamp,
			CurrentPrice:          coin.CurrentPrice,
			High24h:               coin.High24h,
			Low24h:                coin.Low24h,
			TotalVolume:           coin.TotalVolume,
			MarketCap:             coin.MarketCap,
			MarketRank:            coin.MarketCapRank,
			PriceChange24h:        coin.PriceChange24h,
			PriceChangePercent24h: coin.PriceChangePercentage24h,
			CirculatingSupply:     coin.CirculatingSupply,
			TotalSupply:           coin.TotalSupply,
		}
	}

	return prices, nil
}
