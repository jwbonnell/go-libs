package web

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

type App struct {
	log     *logx.Logger
	mux     *http.ServeMux
	mw      []httpx.Middleware
	origins []string
	prefix  string
}

func NewApp(log *logx.Logger, mw ...httpx.Middleware) *App {
	return &App{
		log: log,
		mux: http.NewServeMux(),
		mw:  mw,
	}
}

func (a *App) WithPrefix(prefix string) {
	a.prefix = prefix
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.origins != nil {
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range a.origins {
			if allowedOrigin == "*" || origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
	}

	// Max-age is set to 2 years, and is suffixed with
	// preload, which is necessary for inclusion in all major web browsers' HSTS
	// preload lists, like Chromium, Edge, and Firefox.
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	a.mux.ServeHTTP(w, r)
}

func (a *App) Group(prefix string, mw ...httpx.Middleware) *Group {
	return &Group{
		app:    a,
		prefix: httpx.JoinPaths(a.prefix, prefix),
		mw:     mw,
	}
}

func (a *App) HandleFunc(method string, path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	handler = httpx.Wrap(mw, handler)   // route level mw
	handler = httpx.Wrap(a.mw, handler) // app level mw
	fullPath := httpx.BuildPath(method, httpx.JoinPaths(a.prefix, path))

	h := func(w http.ResponseWriter, r *http.Request) {
		v := httpx.Values{
			TraceID: uuid.NewString(),
			Now:     time.Now().UTC(),
		}
		ctx := context.WithValue(r.Context(), httpx.CtxKey, &v)

		resp := handler(w, r)

		if err := resp.Respond(ctx, w, r); err != nil {
			a.log.Info(ctx, "web-respond", "ERROR", err)
			return
		}
	}

	a.mux.HandleFunc(fullPath, h)
}

func (a *App) WithCORS(origins ...string) {
	a.origins = origins
}

/*func (a *App) WithTracing(tracer Tracer) {
	a.tracer = tracer
}
*/
