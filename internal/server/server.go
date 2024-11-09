package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/1337Bart/smol-crypto-api/internal/config"
	"github.com/1337Bart/smol-crypto-api/internal/service"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Server struct {
	cfg        *config.Config
	grpcServer *grpc.Server
	httpServer *http.Server
	cryptoSvc  *service.CryptoService
	tracer     trace.Tracer
}

func New(cfg *config.Config, cryptoSvc *service.CryptoService) *Server {
	return &Server{
		cfg:       cfg,
		cryptoSvc: cryptoSvc,
	}
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		if err := s.startGRPC(); err != nil {
			fmt.Printf("Failed to start gRPC server: %v\n", err)
		}
	}()

	go func() {
		if err := s.startHTTP(); err != nil {
			fmt.Printf("Failed to start HTTP server: %v\n", err)
		}
	}()

	<-ctx.Done()
	return s.Shutdown()
}

func (s *Server) startGRPC() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.cfg.Server.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	return s.grpcServer.Serve(lis)
}

func (s *Server) startHTTP() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	s.grpcServer.GracefulStop()
	return s.httpServer.Shutdown(context.Background())
}
