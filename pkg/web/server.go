package web

import (
	"context"
	"fmt"
	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"net/http"
	"time"
)

type Server struct {
	serviceName string
	mux         *http.ServeMux
	httpServer  *http.Server
	logger      *logx.Logger
	config      *ServerConfig
	mw          []httpx.Middleware
}

type ServerConfig struct {
	Addr         string
	Port         int
	Origins      []string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

func NewServer(serviceName string, logger *logx.Logger, opts ...func(*ServerConfig)) (*Server, error) {
	// Default configuration
	config := &ServerConfig{
		Port:         8080,
		Origins:      []string{"*"},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Create mux
	mux := http.NewServeMux()

	// Create server
	srv := &Server{
		serviceName: serviceName,
		mux:         mux,
		config:      config,
		logger:      logger,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Port),
			Handler:      mux,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}

	return srv, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Max-age is set to 2 years, and is suffixed with
	// preload, which is necessary for inclusion in all major web browsers' HSTS
	// preload lists, like Chromium, Edge, and Firefox.
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

	s.mux.ServeHTTP(w, r)
}

// Group is used to wrap a slice of routes in the same prefix and middleware
func (s *Server) Group(prefix string, gmw []httpx.Middleware, routes ...Route) {
	for _, route := range routes {
		mw := append(gmw, route.Mw...)

		if prefix != "" {
			route.Path = prefix + route.Path
		}

		s.HandleFunc(route.Method, route.Path, route.Handler, mw...)
	}
}

func (s *Server) HandleFunc(method string, path string, handler httpx.HandlerFunc, mw ...httpx.Middleware) {
	handler = httpx.Wrap(mw, handler)
	handler = httpx.Wrap(s.mw, handler)
	path = fmt.Sprintf("%s %s", method, path)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		//ctx := setTracer(r.Context(), a.tracer)
		//ctx = setWriter(ctx, w)

		//otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

		resp := handler(w, r)

		if err := httpx.Respond(ctx, w, resp); err != nil {
			s.logger.Info(ctx, "web-respond", "ERROR", err)
			return
		}
	}

	s.mux.HandleFunc(path, h)
}
