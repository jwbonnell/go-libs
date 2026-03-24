package web

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

type Group struct {
	app    *App
	prefix string
	mw     []httpx.Middleware
}

func (g *Group) Group(prefix string, mw ...httpx.Middleware) *Group {
	return &Group{
		app:    g.app,
		prefix: httpx.JoinPaths(g.prefix, prefix),
		mw:     append([]httpx.Middleware{}, append(g.mw, mw...)...),
	}
}

func (g *Group) HandleFunc(method string, path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	handler = httpx.Wrap(mw, handler)       // route level mw
	handler = httpx.Wrap(g.mw, handler)     // group level mw
	handler = httpx.Wrap(g.app.mw, handler) // app level mw

	fullPath := httpx.BuildPath(method, httpx.JoinPaths(g.prefix, path))

	h := func(w http.ResponseWriter, r *http.Request) {
		v := httpx.Values{
			TraceID: uuid.NewString(),
			Now:     time.Now().UTC(),
		}
		ctx := context.WithValue(r.Context(), httpx.CtxKey, &v)
		resp := handler(w, r)

		if err := resp.Respond(ctx, w, r); err != nil {
			g.app.log.Error(ctx, "web-respond", "ERROR", err)
			return
		}
	}

	g.app.mux.HandleFunc(fullPath, h)
}
