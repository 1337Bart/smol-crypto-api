package server

import (
	cryptov1 "github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"github.com/1337Bart/smol-crypto-api/internal/handlers"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func (s *Server) initGRPC() error {
	grpcHandler := handlers.NewCryptoHandler(*s.CryptoService)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor()),
	)

	cryptov1.RegisterCryptoServiceServer(server, grpcHandler)

	reflection.Register(server)

	s.GrpcServer = server
	return nil
}

func (s *Server) unaryInterceptor() grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
	)
}
