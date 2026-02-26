package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

// Errors middleware handles errors that bubble up from the next middleware or handler.
// By default, errors are intercepted and replaced with generic HTTP statuses and error messages
// to avoid leaking API implementation details. TrustedErrors are an opt-in way to return error information
// to the API consumer.
func Errors(log *logx.Logger) httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Responder {
			ctx := r.Context()
			resp := next(w, r)
			if resp.Error() != nil {
				v := httpx.GetValues(ctx)

				path := r.URL.Path
				if r.URL.RawQuery != "" {
					path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
				}

				log.Info(ctx, "request error", "trace_id", v.TraceID, "statuscode", resp.Status(), "error", resp.Error(), "method", r.Method, "path", path)

				var statusCode = resp.Status()
				if statusCode < 400 {
					statusCode = http.StatusInternalServerError //force an error status
				}

				switch {
				case httpx.IsTrustedError(resp.Error()):
					te := httpx.GetTrustedError(resp.Error())
					return httpx.ErrorResponse(
						errors.New(te.Msg),
						statusCode,
					)
				default:
					statusText := http.StatusText(statusCode)
					if statusText == "" {
						statusText = http.StatusText(http.StatusInternalServerError)
					}

					return httpx.ErrorResponse(
						errors.New(statusText),
						statusCode,
					)
				}
			}

			return resp
		}

		return h
	}

	return m
}
