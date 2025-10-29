package middleware

import (
	"errors"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPanics_RecoverFromPanic(t *testing.T) {
	mw := Panics()
	panicker := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Response {
		panic("boom")
	})
	h := mw(panicker)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, resp.Err)
	require.Contains(t, resp.Err.Error(), "boom")
	require.Contains(t, resp.Err.Error(), "TRACE[")
	require.Contains(t, resp.Err.Error(), "goroutine")
}

func TestPanics_PassThroughError(t *testing.T) {
	mw := Panics()
	want := errors.New("handler error")
	bad := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return httpx.Response{
			Err: want,
		}
	})
	h := mw(bad)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, resp.Err)
	require.Equal(t, resp.Err.Error(), want.Error())
}

func TestPanics_PassThroughNil(t *testing.T) {
	mw := Panics()
	ok := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Response {
		w.WriteHeader(http.StatusTeapot) // 418
		_, _ = w.Write([]byte("ok"))
		return httpx.Response{Err: nil}
	})
	h := mw(ok)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.NoError(t, resp.Err)
	require.True(t, w.Code == http.StatusTeapot && w.Body.String() == "ok")
}

func TestCORS_AllowsSpecificOrigin(t *testing.T) {
	mw := CORS([]string{"https://example.com"})
	h := mw(okHandler)

	req := httptest.NewRequest("GET", "http://server/", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	resp := h(w, req)
	if resp.Err != nil {
		t.Fatalf("unexpected error: %v", resp.Err)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("Allow-Origin = %q; want %q", got, "https://example.com")
	}
	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("handler not invoked correctly: code=%d body=%q", w.Code, w.Body.String())
	}
}

func TestCORS_RejectsOtherOrigin(t *testing.T) {
	mw := CORS([]string{"https://example.com"})
	h := mw(okHandler)

	req := httptest.NewRequest("GET", "http://server/", nil)
	req.Header.Set("Origin", "https://something.example")

	w := httptest.NewRecorder()
	resp := h(w, req)
	if resp.Err != nil {
		t.Fatalf("unexpected error: %v", resp.Err)
	}

	// When origin doesn't match, middleware should not set the header.
	require.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), "")
	require.Equal(t, w.Header().Get("Access-Control-Allow-Methods"), "POST, PATCH, GET, OPTIONS, PUT, DELETE")
}

func TestCORS_AllowsWildcardOrigin(t *testing.T) {
	mw := CORS([]string{"*"})
	h := mw(okHandler)

	req := httptest.NewRequest("GET", "http://server/", nil)
	req.Header.Set("Origin", "https://any.origin")

	w := httptest.NewRecorder()
	resp := h(w, req)
	if resp.Err != nil {
		t.Fatalf("unexpected error: %v", resp.Err)
	}

	// When wildcard is configured, the middleware sets header to request origin.
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://any.origin" {
		t.Fatalf("Allow-Origin = %q; want %q", got, "https://any.origin")
	}
}

func TestCORS_PreservesExistingHeadersAndSetsDefaults(t *testing.T) {
	mw := CORS([]string{"https://example.com"})
	h := mw(httpx.HandlerFunc(okHandler))

	req := httptest.NewRequest("GET", "http://server/", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	// pre-set a header before middleware runs (simulate upstream)
	w.Header().Set("X-Custom", "v")

	resp := h(w, req)
	if resp.Err != nil {
		t.Fatalf("unexpected error: %v", resp.Err)
	}

	if got := w.Header().Get("X-Custom"); got != "v" {
		t.Fatalf("existing header lost: got %q want %q", got, "v")
	}
	// check required CORS headers
	if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatalf("Allow-Headers not set")
	}
	if got := w.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Fatalf("Max-Age = %q; want %q", got, "86400")
	}
}

// simple handler that records it was called
func okHandler(w http.ResponseWriter, r *http.Request) httpx.Response {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
	return httpx.Response{Err: nil}
}
