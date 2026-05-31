package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/metrics"
	"github.com/jwbonnell/go-libs/pkg/web"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var store = []Item{
	{ID: 1, Name: "Widget"},
	{ID: 2, Name: "Gadget"},
	{ID: 3, Name: "Doohickey"},
}

func main() {
	ctx := context.Background()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	tempoEndpoint := os.Getenv("TEMPO_ENDPOINT")
	if tempoEndpoint == "" {
		tempoEndpoint = "localhost:4317"
	}

	prov, err := metrics.New(ctx, metrics.Config{
		ServiceName:    "sample-api",
		ServiceVersion: "0.1.0",
		Environment:    "development",
		TraceEndpoint:  tempoEndpoint,
		TraceInsecure:  true,
		SampleRate:     1.0,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init metrics: %v\n", err)
		os.Exit(1)
	}
	defer prov.Shutdown(ctx)

	log := logx.New(os.Stdout, slog.LevelInfo, "sample-api", prov.TraceIDFn())

	meter := prov.Meter("sample-api")
	itemRequests, _ := meter.Int64Counter("sample_api.items.requests",
		metric.WithDescription("Total requests to item endpoints"),
	)

	tracer := prov.Tracer("sample-api")

	app := web.NewApp(log, prov.TracingMiddleware())

	app.GET("/health", func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		return httpx.JSONResponse(http.StatusOK, map[string]string{"status": "ok"})
	})

	app.GET("/metrics", func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		return httpx.Custom(http.StatusOK, func(_ context.Context, w http.ResponseWriter, r *http.Request) error {
			prov.PrometheusHandler().ServeHTTP(w, r)
			return nil
		})
	})

	v1 := app.Group("/api/v1")

	v1.GET("/items", func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		itemRequests.Add(r.Context(), 1,
			metric.WithAttributes(attribute.String("method", "GET")),
		)
		log.Info(r.Context(), "list-items", "count", len(store))
		return httpx.JSONResponse(http.StatusOK, store)
	})

	v1.GET("/items/{id}", func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		id := r.PathValue("id")

		spanCtx, span := tracer.Start(r.Context(), "db.lookup",
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithAttributes(attribute.String("item.id", id)),
		)
		defer span.End()
		time.Sleep(5 * time.Millisecond)

		itemRequests.Add(r.Context(), 1,
			metric.WithAttributes(attribute.String("method", "GET")),
		)

		for _, item := range store {
			if fmt.Sprint(item.ID) == id {
				log.Info(spanCtx, "get-item", "id", id)
				return httpx.JSONResponse(http.StatusOK, item)
			}
		}

		log.Info(spanCtx, "item-not-found", "id", id)
		return httpx.ErrorResponse(fmt.Errorf("item %s not found", id), http.StatusNotFound)
	})

	v1.POST("/items", func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return httpx.ErrorResponse(fmt.Errorf("invalid request body"), http.StatusBadRequest)
		}
		if req.Name == "" {
			return httpx.ErrorResponse(fmt.Errorf("name is required"), http.StatusBadRequest)
		}

		newItem := Item{ID: len(store) + 1, Name: req.Name}
		store = append(store, newItem)

		itemRequests.Add(r.Context(), 1,
			metric.WithAttributes(attribute.String("method", "POST")),
		)
		log.Info(r.Context(), "create-item", "id", newItem.ID, "name", newItem.Name)
		return httpx.JSONResponse(http.StatusCreated, newItem)
	})

	server := &http.Server{
		Addr:    addr,
		Handler: app,
	}

	go func() {
		log.Info(ctx, "server-starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(ctx, "server-error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(ctx, "server-shutting-down")
	shutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutCtx); err != nil {
		log.Error(ctx, "shutdown-error", "err", err)
	}
}
