package web

import (
	"fmt"
	"github.com/jwbonnell/go-libs/pkg/web/middleware"
	"net/http"
)

type Router struct {
	mw  []middleware.Middleware
	mux http.ServeMux
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {}

func (r *Router) HandleFunc(method string, path string, handler http.Handler, mw ...middleware.Middleware) {
	handler = middleware.wrap(mw, handler)
	handler = middleware.wrap(r.mw, handler)
	path = fmt.Sprintf("%s %s", method, path)
	r.mux.HandleFunc(path, handler)
}

type Route struct {
	method  string
	path    string
	handler http.Handler
	Mw      []middleware.Middleware
}

type RouteGroup struct {
	prefix string
	router *Router
	mw     []middleware.Middleware
}

func NewRouteGroup(prefix string, router *Router, mw []middleware.Middleware) *RouteGroup {
	return &RouteGroup{prefix: prefix, router: router, mw: mw}
}

func (rg *RouteGroup) HandleFunc(method string, path string, handler http.Handler, mw ...middleware.Middleware) {
	//combine route group middleware with route middleware
	mw = append(rg.mw, mw...)
	//add group prefix to route path
	if rg.prefix != "" {
		path = rg.prefix + path
	}
	rg.router.HandleFunc(method, path, handler, mw...)
}
