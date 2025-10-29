package httpx

import (
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) Response
type Middleware func(HandlerFunc) HandlerFunc

func Wrap(mw []Middleware, h HandlerFunc) HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		f := mw[i]
		if f != nil {
			h = f(h)
		}
	}
	return h
}
