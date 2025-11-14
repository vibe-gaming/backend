package server

import (
	"context"
	"net/http"

	"github.com/vibe-gaming/backend/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.HttpServer.Port,
			Handler:      handler,
			ReadTimeout:  cfg.HttpServer.Timeout,
			WriteTimeout: cfg.HttpServer.Timeout,
			IdleTimeout:  cfg.HttpServer.IdleTimeout,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
