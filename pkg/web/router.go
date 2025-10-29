package web

import (
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

type Route struct {
	Method  string
	Path    string
	Handler httpx.HandlerFunc
	Mw      []httpx.Middleware
}

/*type RouteGroup struct {
	prefix string
	mw     []httpx.Middleware
}

func NewRouteGroup(prefix string, mw []httpx.Middleware) *RouteGroup {
	return &RouteGroup{prefix: prefix, mw: mw}
}

func (rg *RouteGroup) HandleFunc(method string, path string, handler http.Handler, mw ...httpx.Middleware) {
	//combine route group middleware with route middleware
	mw = append(rg.mw, mw...)
	//add group prefix to route path
	if rg.prefix != "" {
		path = rg.prefix + path
	}
	rg.router.HandleFunc(method, path, handler, mw...)
}
*/
