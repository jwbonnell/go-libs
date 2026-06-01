package middleware

import (
	"errors"
	"net/http"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

// Errors middleware handles errors that bubble up from the next middleware or handler.
// By default, errors are intercepted and replaced with generic HTTP statuses and error messages
// to avoid leaking API implementation details. TrustedErrors are an opt-in way to return error information
// to the API consumer.
//
// Use WithSanitizer to provide a function that produces the path string logged on errors.
// Without it, only r.URL.Path is logged with no query string.
func Errors(log *logx.Logger, opts ...Option) httpx.Middleware {
	o := applyOptions(opts)
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Responder {
			ctx := r.Context()
			resp := next(w, r)
			if resp.Error() != nil {
				v := httpx.GetValues(ctx)

				// unwrap to log the internal error; fall back to resp.Error() for non-TrustedErrors
				logErr := errors.Unwrap(resp.Error())
				if logErr == nil {
					logErr = resp.Error()
				}

				log.Error(ctx, "request error", "trace_id", v.TraceID, "statuscode", resp.Status(),
					"error", logErr, "method", r.Method, "path", logPath(r.URL, o.sanitize))

				statusCode := resp.Status()
				if statusCode < 400 {
					statusCode = http.StatusInternalServerError
				}

				switch {
				case httpx.IsTrustedError(resp.Error()):
					// resp.Error() returns te.Msg — the safe message — after the TrustedError fix
					return httpx.ErrorResponse(resp.Error(), statusCode)
				default:
					statusText := http.StatusText(statusCode)
					if statusText == "" {
						statusText = http.StatusText(http.StatusInternalServerError)
					}
					return httpx.ErrorResponse(errors.New(statusText), statusCode)
				}
			}

			return resp
		}

		return h
	}

	return m
}
