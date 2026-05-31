package metrics

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

// TracingMiddleware returns an httpx.Middleware that wraps each request in an OTel span.
// It extracts upstream W3C trace context from headers, records HTTP attributes,
// and marks the span as error for 5xx responses.
func (p *Provider) TracingMiddleware() httpx.Middleware {
	tracer := p.tp.Tracer("http")
	return func(next httpx.HandlerFunc) httpx.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) httpx.Responder {
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.Start(ctx, spanName,
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
				oteltrace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(r.Method),
					semconv.URLPath(r.URL.Path),
				),
			)
			defer span.End()

			resp := next(w, r.WithContext(ctx))

			span.SetAttributes(semconv.HTTPResponseStatusCode(resp.Status()))
			if resp.Status() >= 500 {
				span.SetStatus(codes.Error, http.StatusText(resp.Status()))
			}

			return resp
		}
	}
}
