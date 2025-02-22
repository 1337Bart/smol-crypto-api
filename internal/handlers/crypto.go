package handlers

import (
	"context"
	"time"

	"github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"github.com/1337Bart/smol-crypto-api/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CryptoHandler struct {
	service service.CryptoService
	v1.UnimplementedCryptoServiceServer
}

func NewCryptoHandler(service service.CryptoService) *CryptoHandler {
	return &CryptoHandler{
		service: service,
	}
}

func (h *CryptoHandler) ListCryptos(ctx context.Context, req *v1.ListCryptosRequest) (*v1.ListCryptosResponse, error) {
	page := int32(1)
	limit := int32(10)

	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.Limit > 0 && req.Pagination.Limit <= 100 {
			limit = req.Pagination.Limit
		}
	}

	cryptos, total, err := h.service.ListCryptos(ctx, int(page), int(limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch cryptos: %v", err)
	}

	response := &v1.ListCryptosResponse{
		Cryptos:     make([]*v1.Crypto, 0, len(cryptos)),
		TotalCount:  int32(total),
		CurrentPage: page,
	}

	for _, crypto := range cryptos {
		response.Cryptos = append(response.Cryptos, &v1.Crypto{
			Id:                        crypto.ID,
			Symbol:                    crypto.Symbol,
			Name:                      crypto.Name,
			CurrentPrice:              crypto.CurrentPrice,
			High_24H:                  crypto.High24h,
			Low_24H:                   crypto.Low24h,
			TotalVolume:               crypto.TotalVolume,
			MarketCap:                 crypto.MarketCap,
			MarketRank:                int32(crypto.MarketRank),
			PriceChange_24H:           crypto.PriceChange24h,
			PriceChangePercentage_24H: crypto.PriceChangePercent24h,
			CirculatingSupply:         crypto.CirculatingSupply,
			TotalSupply:               crypto.TotalSupply,
			Timestamp:                 crypto.Timestamp.Format(time.RFC3339),
		})
	}

	return response, nil
}
