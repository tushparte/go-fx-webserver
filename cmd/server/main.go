package main

import (
	"net/http"

	"go-fx-webserver/internal/config"
	"go-fx-webserver/internal/db"
	"go-fx-webserver/internal/httpserver"
	"go-fx-webserver/internal/logger"
	"go-fx-webserver/internal/user"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		logger.Module,
		db.Module,
		httpserver.Module,
		user.Module,
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
