package web

import (
	"fmt"
	"github.com/jwbonnell/go-libs/pkg/log"
	"github.com/jwbonnell/go-libs/pkg/web/middleware"
	"net/http"
)

type App struct {
	log       log.Logger
	router    *http.ServeMux
	middlware []middleware.Middleware
}

func (a *App) HandlerFunc(method string, group string, path string, handlerFunc middleware.HandlerFunc, middlware ...middleware.Middleware) {
	handlerFunc = middleware.wrapMiddleware(middlware, handlerFunc)
	handlerFunc = middleware.wrapMiddleware(a.middlware, handlerFunc)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := setTracer(r.Context(), a.tracer)
		ctx = setWriter(ctx, w)

		//otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

		resp := handlerFunc(ctx, r)

		if err := Respond(ctx, w, resp); err != nil {
			a.log(ctx, "web-respond", "ERROR", err)
			return
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	a.mux.HandleFunc(finalPath, h)
}
