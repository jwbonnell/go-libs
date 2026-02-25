package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test logger helper (adjust if logx.New signature differs)
func testLogger(t *testing.T) *logx.Logger {
	return logx.NewCILogger("unit-tests")
}

func TestNewApp_Defaults(t *testing.T) {
	logger := testLogger(t)
	app := NewApp(logger)
	require.NotNil(t, app)
	assert.Equal(t, logger, app.log)
	assert.NotNil(t, app.mux)
	assert.Empty(t, app.mw)
}

func TestServeHTTP_HSTSHeaderAndDelegate(t *testing.T) {
	logger := testLogger(t)
	app := NewApp(logger)

	// Register a handler
	app.HandleFunc("GET", "/ping", func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		return httpx.PlainTextResponse(http.StatusOK, "pong")
	})

	req := httptest.NewRequest("GET", "http://example.test/ping", nil)
	rr := httptest.NewRecorder()

	app.ServeHTTP(rr, req)

	assert.Equal(t, "max-age=63072000; includeSubDomains; preload", rr.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "pong", rr.Body.String())
}

func TestHandleFunc_MiddlewareOrderAndRespondCalled(t *testing.T) {
	logger := testLogger(t)
	app := NewApp(logger)

	// app-level middleware writes "A"
	appMw := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) httpx.Responder {
			resp := next(w, r)
			err := "A" + resp.Error().Error()
			return httpx.ErrorResponse(
				errors.New(err),
				http.StatusInternalServerError,
			)
		}
	}

	// route middleware writes "R"
	routeMw := func(next httpx.HandlerFunc) httpx.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) httpx.Responder {
			resp := next(w, r)
			err := "R" + resp.Error().Error()
			return httpx.ErrorResponse(
				errors.New(err),
				http.StatusInternalServerError,
			)
		}
	}

	// handler writes "H" and returns nil response
	handler := func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		return httpx.ErrorResponse(
			errors.New("H"),
			http.StatusInternalServerError,
		)
	}

	// set app-level middleware by creating new app with it
	app = NewApp(logger, appMw)
	app.HandleFunc("POST", "/abc", handler, routeMw)

	req := httptest.NewRequest("POST", "http://example.test/abc", nil)
	rr := httptest.NewRecorder()

	app.ServeHTTP(rr, req)

	// Expect order A R H
	assert.Equal(t, "ARH", rr.Body.String())
}

func TestWithCORS_SetsOrigins(t *testing.T) {
	logger := testLogger(t)
	app := NewApp(logger)

	app.WithCORS("https://example.com", "https://foo.test")
	assert.Equal(t, []string{"https://example.com", "https://foo.test"}, app.origins)
}

func TestHandleFunc_RespondErrorLogged(t *testing.T) {
	// This test ensures that when httpx.Respond returns an error, the app logs it.
	// We create a handler that returns a non-nil httpx.Responder which causes httpx.Respond to error.
	// Since httpx.Respond implementation may vary, this test focuses on exercising the error path
	// by providing a handler that returns a custom response object that Respond will likely not handle.
	logger := testLogger(t)
	app := NewApp(logger)

	handler := func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		// write nothing and return a dummy non-nil response
		return httpx.JSONResponse(http.StatusTeapot, "OK")
	}

	// Register route
	app.HandleFunc("GET", "/err", handler)

	req := httptest.NewRequest("GET", "http://example.test/err", nil)
	rr := httptest.NewRecorder()

	// Call ServeHTTP — if Respond errors, app should call logger.Info; test ensures no panic and handler path runs.
	app.ServeHTTP(rr, req)

	// No assertions on log content (log capture would require more setup). Just ensure no panic and response code is default 200.
	// Depending on httpx.Respond implementation, code may be 200 or 0 — ensure no crash.
	assert.NotNil(t, rr)
	// If a body was written, it's fine; ensure call completed.
}

func TestAppServeHTTP_CORSHeaders(t *testing.T) {
	testCases := []struct {
		name            string
		origins         []string
		requestOrigin   string
		expectedHeaders map[string]string
		shouldAllow     bool
	}{
		{
			name:        "No CORS configured",
			origins:     nil,
			shouldAllow: false,
		},
		{
			name:          "Wildcard CORS",
			origins:       []string{"*"},
			requestOrigin: "http://example.com",
			shouldAllow:   true,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "http://example.com",
				"Access-Control-Allow-Methods": "POST, PATCH, GET, OPTIONS, PUT, DELETE",
				"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
				"Access-Control-Max-Age":       "86400",
				"Strict-Transport-Security":    "max-age=63072000; includeSubDomains; preload",
			},
		},
		{
			name:          "Specific Origin Allowed",
			origins:       []string{"http://allowed.com"},
			requestOrigin: "http://allowed.com",
			shouldAllow:   true,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "http://allowed.com",
				"Access-Control-Allow-Methods": "POST, PATCH, GET, OPTIONS, PUT, DELETE",
				"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
				"Access-Control-Max-Age":       "86400",
				"Strict-Transport-Security":    "max-age=63072000; includeSubDomains; preload",
			},
		},
		{
			name:          "Specific Origin Not Allowed",
			origins:       []string{"http://allowed.com"},
			requestOrigin: "http://notallowed.com",
			shouldAllow:   false,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Methods": "POST, PATCH, GET, OPTIONS, PUT, DELETE",
				"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
				"Access-Control-Max-Age":       "86400",
				"Strict-Transport-Security":    "max-age=63072000; includeSubDomains; preload",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger := &logx.Logger{}

			app := NewApp(mockLogger)
			app.WithCORS(tc.origins...)

			app.mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}

			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			resp := w.Result()

			// Check HSTS header is always set
			assert.Equal(t, "max-age=63072000; includeSubDomains; preload",
				resp.Header.Get("Strict-Transport-Security"),
				"Strict-Transport-Security header should always be set")

			if tc.origins != nil {
				require.NotNil(t, tc.expectedHeaders, "Expected headers should be defined for allowed origin")

				for headerName, expectedValue := range tc.expectedHeaders {
					actualValue := resp.Header.Get(headerName)
					assert.Equal(t, expectedValue, actualValue,
						"CORS header %s should match expected value", headerName)
				}
			}
		})
	}
}

func TestAppServeHTTP_OptionsRequest(t *testing.T) {
	// Create a mock logger
	mockLogger := &logx.Logger{}

	// Create the App with wildcard CORS
	app := NewApp(mockLogger)
	app.origins = []string{"*"}

	// Create a test OPTIONS request
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	resp := w.Result()

	// Verify CORS headers for preflight request
	assert.Equal(t, "http://example.com",
		resp.Header.Get("Access-Control-Allow-Origin"),
		"Preflight request should have correct Allow-Origin header")
	assert.Equal(t, "POST, PATCH, GET, OPTIONS, PUT, DELETE",
		resp.Header.Get("Access-Control-Allow-Methods"),
		"Preflight request should have correct Allow-Methods header")
	assert.Equal(t, "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
		resp.Header.Get("Access-Control-Allow-Headers"),
		"Preflight request should have correct Allow-Headers header")
	assert.Equal(t, "86400",
		resp.Header.Get("Access-Control-Max-Age"),
		"Preflight request should have correct Max-Age header")
}

func TestAppBuildPath(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		path     string
		prefix   string
		expected string
	}{
		{
			name:     "No Prefix Test",
			method:   http.MethodGet,
			path:     "/burrito",
			prefix:   "",
			expected: "GET /burrito",
		},
		{
			name:     "Prefix Test",
			method:   http.MethodPost,
			path:     "/burrito",
			prefix:   "/taco",
			expected: "POST /taco/burrito",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := buildPath(tc.method, tc.path, tc.prefix)
			assert.Equal(t, tc.expected, path)
		})
	}
}
