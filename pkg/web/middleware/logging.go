package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

func Logger(log *logx.Logger) httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Response {
			resp := next(w, r)

			if resp.Err == nil {
				ctx := r.Context()
				v := httpx.GetValues(ctx)

				path := r.URL.Path
				if r.URL.RawQuery != "" {
					path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
				}

				log.Info(ctx, "request completed", "trace_id", v.TraceID, "method", r.Method, "path", path,
					"remoteaddr", r.RemoteAddr, "statuscode", v.StatusCode, "since", time.Since(v.Now))
			}

			return resp
		}

		return h
	}

	return m
}
