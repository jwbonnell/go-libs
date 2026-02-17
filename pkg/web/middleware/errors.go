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

				log.Info(ctx, "request error", "trace_id", v.TraceID, "statuscode", resp.StatusCode, "error", resp.Err, "method", r.Method, "path", path)

				var errResp httpx.Response
				switch {
				case httpx.IsTrustedError(resp.Err):
					te := httpx.GetTrustedError(resp.Err)
					return httpx.ErrorResponse(
						errors.New(te.Msg),
						resp.StatusCode,
					)
				default:
					errResp = httpx.ErrorResponse(
						errors.New(http.StatusText(http.StatusInternalServerError)),
						http.StatusInternalServerError,
					)
				}

				return errResp
			}

			return resp
		}

		return h
	}

	return m
}
