package middleware

import (
	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"net/http"
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
				log.Error(ctx, "ERROR", "trace_id", "TODO", "message", resp.Err)

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
				/*if err := httpx.Respond(ctx, w, er, status); err != nil {
					return err
				}*/
			}

			return resp
		}

		return h
	}

	return m
}
