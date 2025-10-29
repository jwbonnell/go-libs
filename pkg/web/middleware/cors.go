package middleware

import (
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"net/http"
)

func CORS(origins []string) httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Response {
			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range origins {
				if allowedOrigin == "*" || origin == allowedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Additional CORS headers...
			return next(w, r)
		}

		return h
	}

	return m
}
