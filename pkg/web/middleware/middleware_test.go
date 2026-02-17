package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"github.com/stretchr/testify/require"
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

// simple handler that records it was called
func okHandler(w http.ResponseWriter, r *http.Request) httpx.Response {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
	return httpx.Response{Err: nil}
}

func TestErrorMiddleware(t *testing.T) {
	mw := Errors(logx.NewCILogger("unit-tests"))
	ok := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return httpx.ErrorResponse(
			httpx.NewTrustedError("my trusted error", errors.New("internal error message")),
			http.StatusTeapot,
		)
	})
	h := mw(ok)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, resp.Err)
	require.Equal(t, resp.Err.Error(), "my trusted error")
}
