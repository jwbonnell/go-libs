package middleware

import "net/http"

type HandlerFunc func(http.ResponseWriter, *http.Request) error
type Middleware func(HandlerFunc) HandlerFunc

func wrap(mw []Middleware, h HandlerFunc) HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		f := mw[i]
		if f != nil {
			h = f(h)
		}
	}
	return h
}
