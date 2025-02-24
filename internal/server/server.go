package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

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

func New(cfg *config.Config, cryptoService *service.CryptoService, tracer trace.Tracer) *Server {
	srv := &Server{
		Cfg:           cfg,
		CryptoService: cryptoService,
		Tracer:        tracer,
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
	errChan := make(chan error, 2)

	go func() {
		if err := s.startGRPC(); err != nil && !errors.Is(err, net.ErrClosed) {
			errChan <- fmt.Errorf("grpc server error: %w", err)
		}
	}()

	go func() {
		if err := s.startHTTP(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("http server error: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		return s.Shutdown()
	}
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	errChan := make(chan error, 2)

	go func() {
		s.GrpcServer.GracefulStop()
		errChan <- nil
	}()

	go func() {
		errChan <- s.HttpServer.Shutdown(ctx)
	}()

	var errs []error
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
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
