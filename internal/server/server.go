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
	Cfg           *config.Config
	GrpcServer    *grpc.Server
	HttpServer    *http.Server
	CryptoService *service.CryptoService
	Tracer        trace.Tracer
}

func New(cfg *config.Config, cryptoService *service.CryptoService) *Server {
	srv := &Server{
		Cfg:           cfg,
		CryptoService: cryptoService,
	}

	if err := srv.initGRPC(); err != nil {
		panic(fmt.Sprintf("failed to init gRPC server: %v", err))
	}
	
	if err := srv.initHTTP(); err != nil {
		panic(fmt.Sprintf("failed to init HTTP server: %v", err))
	}

	return srv
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.Cfg.Server.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	return s.GrpcServer.Serve(lis)
}

func (s *Server) startHTTP() error {
	return s.HttpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	s.GrpcServer.GracefulStop()
	return s.HttpServer.Shutdown(context.Background())
}
