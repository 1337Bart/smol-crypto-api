package server

import (
	"fmt"
	"net/http"

	"github.com/1337Bart/smol-crypto-api/internal/handlers/http_handler"
	"github.com/go-chi/chi/v5"
)

func (s *Server) initHTTP() error {
	httpHandler := http_handler.NewCryptoHandler(*s.CryptoService)
	router := chi.NewRouter()

	router.Get("/api/v1/crypto", httpHandler.ListCryptos)

	s.HttpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.Cfg.Server.HTTPPort),
		Handler: router,
	}

	return nil
}
