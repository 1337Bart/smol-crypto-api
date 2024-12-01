package server

import (
	"context"
	cryptov1 "github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"github.com/1337Bart/smol-crypto-api/internal/handlers/grpc_handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (s *Server) initGRPC() error {
	grpcHandler := grpc_handler.NewCryptoHandler(*s.CryptoService)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor()),
	)

	cryptov1.RegisterCryptoServiceServer(server, grpcHandler)

	// Enable reflection for grpcurl - co to robi??
	reflection.Register(server)

	s.GrpcServer = server
	return nil
}

func (s *Server) unaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Add telemetry, logging, etc.
		return handler(ctx, req)
	}
}
