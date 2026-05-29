package middleware

import (
	"net/http"
	"time"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

func Logger(log *logx.Logger) httpx.Middleware {
	m := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		h := func(w http.ResponseWriter, r *http.Request) httpx.Responder {
			resp := next(w, r)

			if resp.Error() == nil {
				ctx := r.Context()
				v := httpx.GetValues(ctx)
				log.Info(ctx, "request completed", "trace_id", v.TraceID, "method", r.Method, "path", r.URL.Path,
					"remoteaddr", r.RemoteAddr, "statuscode", resp.Status(), "since", time.Since(v.Now))
			}

			return resp
		}

		return h
	}

	return m
}
