package web

import (
	"fmt"
	"net/http"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

type App struct {
	log     *logx.Logger
	mux     *http.ServeMux
	mw      []httpx.Middleware
	origins []string
}

type AppConfig struct {
	allowedOrigins []string
}

func NewApp(log *logx.Logger, mw ...httpx.Middleware) *App {
	return &App{
		log: log,
		mux: http.NewServeMux(),
		mw:  mw,
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Max-age is set to 2 years, and is suffixed with
	// preload, which is necessary for inclusion in all major web browsers' HSTS
	// preload lists, like Chromium, Edge, and Firefox.
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

	a.mux.ServeHTTP(w, r)
}

func (a *App) HandleFunc(method string, path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	handler = httpx.Wrap(mw, handler)
	handler = httpx.Wrap(a.mw, handler)
	path = fmt.Sprintf("%s %s", method, path)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		//ctx := setTracer(r.Context(), a.tracer)
		//ctx = setWriter(ctx, w)

		//otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

		resp := handler(w, r)

		if err := httpx.Respond(ctx, w, resp); err != nil {
			a.log.Info(ctx, "web-respond", "ERROR", err)
			return
		}
	}

	a.mux.HandleFunc(path, h)
}

func (a *App) WithCORS(origins ...string) {
	a.origins = origins
}

/*func (a *App) WithTracing(tracer Tracer) {
	a.tracer = tracer
}
*/
