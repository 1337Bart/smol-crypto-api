package model

import "time"

type CryptoPrice struct {
	Symbol    string    `json:"symbol"`
	PriceUSD  float64   `json:"price_usd"`
	Timestamp time.Time `json:"timestamp"`
}

type BatchPriceData struct {
	Prices []CryptoPrice `json:"prices"`
}

type MarketSentiment struct {
	Symbol         string    `json:"symbol"`
	SentimentScore float64   `json:"sentiment_score"`
	Timestamp      time.Time `json:"timestamp"`
}
