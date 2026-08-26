package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"

	"go.uber.org/fx"
)

func NewListener(cfg *Config) (net.Listener, error) {
	return net.Listen("tcp", ":"+cfg.Port)
}

func NewHTTPServer(lc fx.Lifecycle, ln net.Listener, mux *http.ServeMux, cfg *Config, logger *slog.Logger) *http.Server {
	srv := &http.Server{
		Handler:           mux,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting server", "addr", ln.Addr().String())
			go srv.Serve(ln)
			return nil
		},

		OnStop: func(ctx context.Context) error {
			logger.Info("stopping server")
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}
