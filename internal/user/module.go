package user

import (
	"go-fx-webserver/internal/httpserver"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"user",
	fx.Provide(
		NewRepository,
		fx.Annotate(
			NewHandler,
			fx.As(new(httpserver.RouterRegistrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)
