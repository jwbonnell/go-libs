package web

import (
	"context"
	"fmt"
	"github.com/jwbonnell/go-libs/pkg/log"
	"net/http"
	"time"
)

type Server struct {
	serviceName string
	router      *mux.Router
	httpServer  *http.Server
	logger      *log.Logger
	config      *ServerConfig
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

func NewServer(serviceName string, logger *log.Logger, opts ...func(*ServerConfig)) (*Server, error) {
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

	// Create router
	router := mux.NewRouter()

	// Create server
	srv := &Server{
		serviceName: serviceName,
		router:      router,
		config:      config,
		logger:      logger,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Port),
			Handler:      router,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}

	// Setup routes
	srv.setupRoutes()

	return srv, nil
}
