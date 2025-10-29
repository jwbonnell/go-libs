package middleware

import (
	"fmt"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"net/http"
	"runtime/debug"
)

func Panics() httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) (resp httpx.Response) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					resp = httpx.Response{
						Err: fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace)),
					}

					//TODO
					//metrics.AddPanics(ctx)
				}
			}()

			return next(w, r)
		}

		return h
	}

	return m
}
