package web

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jwbonnell/go-libs/pkg/log"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// simple logger helper for tests
func testLogger(t *testing.T) *log.Logger {
	var buf bytes.Buffer
	return log.New(&buf, slog.LevelDebug, "test-service", mockTraceIDFn)
}

func mockTraceIDFn(ctx context.Context) string {
	return "test-trace-id"
}

func TestNewServer_Defaults(t *testing.T) {
	logger := testLogger(t)

	srv, err := NewServer("svc", logger)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// default port 8080 -> Addr ":8080"
	assert.Equal(t, ":8080", srv.httpServer.Addr)
	assert.Equal(t, 8080, srv.config.Port)
	// default origins wildcard
	assert.Equal(t, []string{"*"}, srv.config.Origins)
	// read/write timeouts default 10s
	assert.Equal(t, 10*time.Second, srv.httpServer.ReadTimeout)
	assert.Equal(t, 10*time.Second, srv.httpServer.WriteTimeout)
}

func TestNewServer_WithOptions(t *testing.T) {
	logger := testLogger(t)

	srv, err := NewServer("svc", logger, func(c *ServerConfig) {
		c.Port = 9090
		c.ReadTimeout = 1 * time.Second
		c.WriteTimeout = 2 * time.Second
		c.IdleTimeout = 3 * time.Second
		c.Origins = []string{"https://example.com"}
	})
	require.NoError(t, err)
	require.NotNil(t, srv)

	assert.Equal(t, ":9090", srv.httpServer.Addr)
	assert.Equal(t, 9090, srv.config.Port)
	assert.Equal(t, 1*time.Second, srv.httpServer.ReadTimeout)
	assert.Equal(t, 2*time.Second, srv.httpServer.WriteTimeout)
	assert.Equal(t, 3*time.Second, srv.httpServer.IdleTimeout)
	assert.Equal(t, []string{"https://example.com"}, srv.config.Origins)
}

func TestServeHTTP_SetsHSTSHeaderAndDelegates(t *testing.T) {
	logger := testLogger(t)
	srv, err := NewServer("svc", logger)
	require.NoError(t, err)

	// Register a simple handler using HandleFunc directly
	srv.HandleFunc("GET", "/ping", func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return httpx.PlainTextResponse(http.StatusOK, "pong")
	})

	req := httptest.NewRequest("GET", "http://example.test/ping", nil)
	// The mux registered paths are stored as "METHOD path" so ServeHTTP should forward to mux
	// Create ResponseRecorder
	rr := httptest.NewRecorder()

	// Call ServeHTTP
	srv.ServeHTTP(rr, req)

	// Check HSTS header present
	expected := "max-age=63072000; includeSubDomains; preload"
	assert.Equal(t, expected, rr.Header().Get("Strict-Transport-Security"))

	// Verify body and status forwarded from inner handler
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "pong", rr.Body.String())
}

func TestShutdown_Noop(t *testing.T) {
	logger := testLogger(t)
	srv, err := NewServer("svc", logger)
	require.NoError(t, err)

	// Shutdown currently returns nil (noop)
	err = srv.Shutdown(context.Background())
	assert.NoError(t, err)
}
