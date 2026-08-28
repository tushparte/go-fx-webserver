package main

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/fx"
)

func NewDB(lc fx.Lifecycle, cfg *Config, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err != nil {
				return err
			}

			logger.Info("connected to mysql")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("closing mysql connection")
			return db.Close()
		},
	})

	return db, nil
}
