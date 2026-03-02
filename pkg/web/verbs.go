package web

import (
	"net/http"

	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

// App level verb helpers
func (a *App) GET(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	a.HandleFunc(http.MethodGet, path, handler, mw...)
}
func (a *App) POST(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	a.HandleFunc(http.MethodPost, path, handler, mw...)
}
func (a *App) PUT(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	a.HandleFunc(http.MethodPut, path, handler, mw...)
}
func (a *App) PATCH(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	a.HandleFunc(http.MethodPatch, path, handler, mw...)
}
func (a *App) DELETE(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	a.HandleFunc(http.MethodDelete, path, handler, mw...)
}

// Group level verb helpers
func (g *Group) GET(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	g.HandleFunc(http.MethodGet, path, handler, mw...)
}
func (g *Group) POST(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	g.HandleFunc(http.MethodPost, path, handler, mw...)
}
func (g *Group) PUT(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	g.HandleFunc(http.MethodPut, path, handler, mw...)
}
func (g *Group) PATCH(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	g.HandleFunc(http.MethodPatch, path, handler, mw...)
}
func (g *Group) DELETE(path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	g.HandleFunc(http.MethodDelete, path, handler, mw...)
}
