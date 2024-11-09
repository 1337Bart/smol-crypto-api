package server

import (
	"context"
	"net/http"

	cryptov1 "github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *Server) initHTTP() error {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := "localhost:" + s.cfg.Server.GRPCPort

	if err := cryptov1.RegisterCryptoServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		return err
	}

	s.httpServer = &http.Server{
		Addr:    ":" + s.cfg.Server.HTTPPort,
		Handler: mux,
	}

	return nil
}
