package middleware

import (
	"errors"
	"net/http"

	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

// MaxBodySize rejects requests whose body exceeds maxBytes with a 413. It enforces
// the limit in two ways: an early rejection when Content-Length is present and
// exceeds the limit, and wrapping r.Body with http.MaxBytesReader so reads that
// exceed the limit during decoding also fail with a clear error rather than silent
// truncation.
func MaxBodySize(maxBytes int64) httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Responder {
			if r.ContentLength > maxBytes {
				return httpx.ErrorResponse(
					httpx.NewTrustedError("request body too large", errors.New("content-length exceeds limit")),
					http.StatusRequestEntityTooLarge,
				)
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			return next(w, r)
		}
		return h
	}
	return m
}
