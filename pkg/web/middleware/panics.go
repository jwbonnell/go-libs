package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func Panics() Middleware {
	m := func(next HandlerFunc) HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))

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
