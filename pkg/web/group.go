package web

import (
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
	g.app.mux.HandleFunc(fullPath, serve(g.app.log, handler))
}
