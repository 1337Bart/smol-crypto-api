package server

import (
	"context"
	"fmt"
	"net/http"

	cryptopb "github.com/1337Bart/smol-crypto-api/api/proto/v1"
	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
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
		page := r.URL.Query().Get("page")
		limit := r.URL.Query().Get("limit")

		q := r.URL.Query()
		if page != "" {
			q.Set("pagination.page", page)
			q.Del("page")
		}
		if limit != "" {
			q.Set("pagination.limit", limit)
			q.Del("limit")
		}
		r.URL.RawQuery = q.Encode()

		gwmux.ServeHTTP(w, r)
	})

	router := chi.NewRouter()
	router.Mount("/", wrappedHandler)

	s.HttpServer = &http.Server{
		Addr: fmt.Sprintf("%s:%s",
			s.Cfg.Server.HTTPHost,
			s.Cfg.Server.HTTPPort,
		),
		Handler: router,
	}
	return nil
}
