package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"

	"go.uber.org/fx"
)

func newHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, cfg *Config, logger *slog.Logger) *http.Server {
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			logger.Info("Starting server", "addr", srv.Addr)
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
