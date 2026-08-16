package main

import (
	"net/http"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		// register dependencies
		fx.Provide(
			NewConfig,
			NewLogger,
			NewMux,
			newHTTPServer,
		),
		// start up new http server which is returned by newHTTPServer
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
