package types

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListCryptosResponse struct {
	Cryptos     []CryptoResponse `json:"cryptos"`
	TotalCount  int              `json:"total_count"`
	CurrentPage int              `json:"current_page"`
}

type CryptoResponse struct {
	ID                    string  `json:"id"`
	Symbol                string  `json:"symbol"`
	Name                  string  `json:"name"`
	CurrentPrice          float64 `json:"current_price"`
	High24h               float64 `json:"high_24h"`
	Low24h                float64 `json:"low_24h"`
	TotalVolume           float64 `json:"total_volume"`
	MarketCap             float64 `json:"market_cap"`
	MarketRank            int     `json:"market_rank"`
	PriceChange24h        float64 `json:"price_change_24h"`
	PriceChangePercent24h float64 `json:"price_change_percentage_24h"`
	CirculatingSupply     float64 `json:"circulating_supply"`
	TotalSupply           float64 `json:"total_supply"`
	Timestamp             string  `json:"timestamp"`
}
