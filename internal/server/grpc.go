package server

import (
	"context"
	cryptov1 "github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (s *Server) initGRPC() error {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor()),
	)

	// Register services
	cryptov1.RegisterCryptoServiceServer(server, s.cryptoSvc)

	// Enable reflection for grpcurl
	reflection.Register(server)

	s.grpcServer = server
	return nil
}

func (s *Server) unaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Add telemetry, logging, etc.
		return handler(ctx, req)
	}
}
