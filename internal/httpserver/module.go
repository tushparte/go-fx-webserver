package httpserver

import (
	"net/http"

	"go.uber.org/fx"
)

type RouterRegistrar interface {
	RegisterRoutes(mux *http.ServeMux)
}

func NewMux(registrars []RouterRegistrar) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	for _, r := range registrars {
		r.RegisterRoutes(mux)
	}

	return mux
}

var Module = fx.Module(
	"httpserver",
	fx.Provide(
		NewListener,
		NewHTTPServer,
		fx.Annotate(NewMux, fx.ParamTags(`group:"routes"`)),
	),
)
