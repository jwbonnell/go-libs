package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *logx.Logger) httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Response {
			ctx := r.Context()
			resp := next(w, r)
			if resp.Err != nil {
				v := httpx.GetValues(ctx)

				path := r.URL.Path
				if r.URL.RawQuery != "" {
					path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
				}

				log.Info(ctx, "request error", "trace_id", v.TraceID, "error", resp.Err, "method", r.Method, "path", path)

				/*var er v1.ErrorResponse
				var status int

				switch {
				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					er = v1.ErrorResponse{
						Error:  "data validation error",
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				case v1.IsRequestError(err):
					reqErr := v1.GetRequestError(err)
					er = v1.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				case auth.IsAuthError(err):
					er = v1.ErrorResponse{
						Error: http.StatusText(http.StatusUnauthorized),
					}
					status = http.StatusUnauthorized

				default:
					er = v1.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}
				*/
				return httpx.ErrorResponse(
					errors.New(http.StatusText(http.StatusInternalServerError)),
					http.StatusInternalServerError,
				)
			}

			return resp
		}

		return h
	}

	return m
}
