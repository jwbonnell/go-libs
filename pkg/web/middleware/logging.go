package middleware

import (
	"net/http"
	"net/url"
	"time"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

// Logger middleware logs completed requests that produced no error.
//
// Use WithSanitizer to provide a function that produces the path string that is logged.
// Without it, only r.URL.Path is logged with no query string.
func Logger(log *logx.Logger, opts ...Option) httpx.Middleware {
	o := applyOptions(opts)
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Responder {
			resp := next(w, r)

			if resp.Error() == nil {
				ctx := r.Context()
				v := httpx.GetValues(ctx)
				log.Info(ctx, "request completed", "trace_id", v.TraceID, "method", r.Method,
					"path", logPath(r.URL, o.sanitize), "remoteaddr", r.RemoteAddr,
					"statuscode", resp.Status(), "since", time.Since(v.Now))
			}

			return resp
		}

		return h
	}

	return m
}

// logPath returns the URL string to log. If sanitize is non-nil it is called to
// produce the string; otherwise only the path component is returned.
func logPath(u *url.URL, sanitize func(*url.URL) string) string {
	if sanitize != nil {
		return sanitize(u)
	}
	return u.Path
}
