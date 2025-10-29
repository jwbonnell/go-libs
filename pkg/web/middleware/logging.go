package middleware

import (
	"fmt"
	"github.com/jwbonnell/go-libs/pkg/log"
	"github.com/jwbonnell/go-libs/pkg/web/context"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"net/http"
	"time"
)

func Logger(log *log.Logger) httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Response {
			ctx := r.Context()
			v := context.GetValues(ctx)

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			resp := next(w, r)

			log.Info(ctx, "request completed", "trace_id", v.TraceID, "method", r.Method, "path", path,
				"remoteaddr", r.RemoteAddr, "statuscode", v.StatusCode, "since", time.Since(v.Now))

			return resp
		}

		return h
	}

	return m
}
