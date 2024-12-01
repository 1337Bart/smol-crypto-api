package http_handler

import (
	"encoding/json"
	"github.com/1337Bart/smol-crypto-api/internal/handlers/types"
	"github.com/1337Bart/smol-crypto-api/internal/service"
	"net/http"
	"strconv"
	"time"
)

type CryptoHandler struct {
	service service.CryptoService
}

func NewCryptoHandler(service service.CryptoService) *CryptoHandler {
	return &CryptoHandler{
		service: service,
	}
}

func (h *CryptoHandler) ListCryptos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// defaults
	page := 1
	limit := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	cryptos, total, err := h.service.ListCryptos(ctx, page, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ErrorResponse{Error: "Failed to fetch cryptos"})
		return
	}

	response := types.ListCryptosResponse{
		Cryptos:     make([]types.CryptoResponse, 0, len(cryptos)),
		TotalCount:  total,
		CurrentPage: page,
	}

	for _, crypto := range cryptos {
		response.Cryptos = append(response.Cryptos, types.CryptoResponse{
			ID:                    crypto.ID,
			Symbol:                crypto.Symbol,
			Name:                  crypto.Name,
			CurrentPrice:          crypto.CurrentPrice,
			High24h:               crypto.High24h,
			Low24h:                crypto.Low24h,
			TotalVolume:           crypto.TotalVolume,
			MarketCap:             crypto.MarketCap,
			MarketRank:            crypto.MarketRank,
			PriceChange24h:        crypto.PriceChange24h,
			PriceChangePercent24h: crypto.PriceChangePercent24h,
			CirculatingSupply:     crypto.CirculatingSupply,
			TotalSupply:           crypto.TotalSupply,
			Timestamp:             crypto.Timestamp.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
