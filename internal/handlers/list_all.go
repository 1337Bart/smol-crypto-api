package handlers

import (
	"context"
	"time"

	v1 "github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"github.com/1337Bart/smol-crypto-api/internal/model"
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
	filter := model.CryptoFilter{
		Page:  1,
		Limit: 10,
	}

	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			filter.Page = int(req.Pagination.Page)
		}
		if req.Pagination.Limit > 0 && req.Pagination.Limit <= 100 {
			filter.Limit = int(req.Pagination.Limit)
		}
	}

	if req.Symbol != "" {
		filter.Symbol = req.Symbol
	}

	if req.StartTime != nil {
		startTime := req.StartTime.AsTime()
		filter.StartTime = &startTime
	}

	if req.EndTime != nil {
		endTime := req.EndTime.AsTime()
		filter.EndTime = &endTime
	}

	if filter.StartTime != nil && filter.EndTime != nil {
		if filter.StartTime.After(*filter.EndTime) {
			return nil, status.Error(codes.InvalidArgument, "start_time must be before end_time")
		}
	}

	cryptos, total, err := h.service.ListCryptos(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch cryptos: %v", err)
	}

	response := &v1.ListCryptosResponse{
		Cryptos:     make([]*v1.Crypto, 0, len(cryptos)),
		TotalCount:  int32(total),
		CurrentPage: int32(filter.Page),
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
