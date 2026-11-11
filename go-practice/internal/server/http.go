package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// HTTPServer owns transport lifecycle; Gin owns routing and HTTP handlers.
type HTTPServer struct {
	*http.Server
	listener net.Listener
}

func NewHTTP(lc fx.Lifecycle, shutdown fx.Shutdowner, cfg config.Config, logger *slog.Logger, router *gin.Engine) *HTTPServer {
	srv := &HTTPServer{Server: &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       cfg.RequestTimeout + 3*time.Second,
		WriteTimeout:      cfg.RequestTimeout + 3*time.Second,
		IdleTimeout:       30 * time.Second,
	}}
	manage(lc, shutdown, logger, srv)
	return srv
}

func (s *HTTPServer) Endpoint() (*url.URL, error) {
	if s.listener == nil {
		listener, err := net.Listen("tcp", s.Addr)
		if err != nil {
			return nil, err
		}
		s.listener = listener
	}
	return &url.URL{Scheme: "http", Host: s.listener.Addr().String()}, nil
}

func (s *HTTPServer) Start(context.Context) error {
	err := s.Serve(s.listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if err := s.Shutdown(ctx); err != nil {
		_ = s.Close()
		return err
	}
	return nil
}
