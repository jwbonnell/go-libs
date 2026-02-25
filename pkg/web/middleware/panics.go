package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

func Panics() httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) (resp httpx.Responder) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					resp = httpx.ErrorResponse(
						fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace)),
						http.StatusInternalServerError,
					)

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
