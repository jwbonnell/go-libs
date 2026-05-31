package metrics

import (
	"context"
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/jwbonnell/go-libs/pkg/logx"
)

// Config holds configuration for setting up OpenTelemetry.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string  // e.g. "production", "staging"
	TraceEndpoint  string  // OTLP gRPC address, e.g. "tempo:4317"
	TraceInsecure  bool    // skip TLS; common for internal deployments
	SampleRate     float64 // 0.0–1.0; 0 defaults to 1.0
}

// Provider wraps OpenTelemetry setup and exposes helpers for tracing, metrics, and logx integration.
type Provider struct {
	tp      *sdktrace.TracerProvider
	mp      *sdkmetric.MeterProvider
	promReg *prometheus.Registry
	shut    []func(context.Context) error
}

// New initializes OTel tracing (→ Tempo via OTLP gRPC) and metrics (→ Prometheus).
// The caller must defer provider.Shutdown(ctx) after a successful return.
func New(ctx context.Context, cfg Config) (*Provider, error) {
	p := &Provider{}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentName(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	traceOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.TraceEndpoint),
	}
	if cfg.TraceInsecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
	}

	traceExp, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return nil, err
	}

	sampleRate := cfg.SampleRate
	if sampleRate <= 0 {
		sampleRate = 1.0
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	p.tp = tp
	p.shut = append(p.shut, tp.Shutdown)

	promReg := prometheus.NewRegistry()
	promExp, err := promexporter.New(promexporter.WithRegisterer(promReg))
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExp),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	p.mp = mp
	p.promReg = promReg
	p.shut = append(p.shut, mp.Shutdown)

	return p, nil
}

// Shutdown flushes all exporters and releases resources in reverse registration order.
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error
	for i := len(p.shut) - 1; i >= 0; i-- {
		if err := p.shut[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// TraceIDFn returns a logx.TraceIDFn that reads the active OTel span's trace ID.
// Pass this to logx.New() to correlate Loki log lines with Tempo traces in Grafana.
func (p *Provider) TraceIDFn() logx.TraceIDFn {
	return func(ctx context.Context) string {
		span := trace.SpanFromContext(ctx)
		if span.SpanContext().IsValid() {
			return span.SpanContext().TraceID().String()
		}
		return ""
	}
}

// PrometheusHandler returns an http.Handler for a /metrics scrape endpoint.
func (p *Provider) PrometheusHandler() http.Handler {
	return promhttp.HandlerFor(p.promReg, promhttp.HandlerOpts{})
}

// Tracer returns a named tracer backed by the configured TracerProvider.
func (p *Provider) Tracer(name string) trace.Tracer {
	return p.tp.Tracer(name)
}

// Meter returns a named meter backed by the configured MeterProvider.
func (p *Provider) Meter(name string) metric.Meter {
	return p.mp.Meter(name)
}
