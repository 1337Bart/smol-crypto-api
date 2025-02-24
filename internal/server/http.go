package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	cryptopb "github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (s *Server) initHTTP() error {
	ctx := context.Background()

	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := cryptopb.RegisterCryptoServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		fmt.Sprintf("%s:%s", s.Cfg.Server.GRPCHost, s.Cfg.Server.GRPCPort),
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if page := q.Get("page"); page != "" {
			q.Set("pagination.page", page)
			q.Del("page")
		}
		if limit := q.Get("limit"); limit != "" {
			q.Set("pagination.limit", limit)
			q.Del("limit")
		}

		if symbol := q.Get("symbol"); symbol != "" {
			q.Set("symbol", strings.TrimSpace(symbol))
		}

		if startTime := strings.TrimSpace(q.Get("start_time")); startTime != "" {
			t, err := time.Parse(time.RFC3339, startTime)
			if err != nil {
				http.Error(w, "Invalid start_time format. Use RFC3339 format (e.g., 2024-02-20T00:00:00Z)", http.StatusBadRequest)
				return
			}
			q.Set("start_time", t.Format(time.RFC3339))
		}

		if endTime := strings.TrimSpace(q.Get("end_time")); endTime != "" {
			t, err := time.Parse(time.RFC3339, endTime)
			if err != nil {
				http.Error(w, "Invalid end_time format. Use RFC3339 format (e.g., 2024-02-20T00:00:00Z)", http.StatusBadRequest)
				return
			}
			q.Set("end_time", t.Format(time.RFC3339))
		}

		r.URL.RawQuery = q.Encode()

		gwmux.ServeHTTP(w, r)
	})

	router := chi.NewRouter()
	handler := otelhttp.NewHandler(wrappedHandler, "http_server")
	router.Mount("/", handler)

	s.HttpServer = &http.Server{
		Addr: fmt.Sprintf("%s:%s",
			s.Cfg.Server.HTTPHost,
			s.Cfg.Server.HTTPPort,
		),
		Handler: router,
	}
	return nil
}
