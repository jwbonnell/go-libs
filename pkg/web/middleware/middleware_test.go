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
	panicker := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		panic("boom")
	})
	h := mw(panicker)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, resp.Error())
	require.Contains(t, resp.Error().Error(), "boom")
	require.Contains(t, resp.Error().Error(), "TRACE[")
	require.Contains(t, resp.Error().Error(), "goroutine")
}

func TestPanics_PassThroughError(t *testing.T) {
	mw := Panics()
	want := errors.New("handler error")
	bad := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		return httpx.ErrorResponse(want, http.StatusInternalServerError)
	})
	h := mw(bad)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, resp.Error())
	require.Equal(t, resp.Error().Error(), want.Error())
}

func TestPanics_PassThroughNil(t *testing.T) {
	mw := Panics()
	ok := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		w.WriteHeader(http.StatusTeapot) // 418
		_, _ = w.Write([]byte("ok"))
		return httpx.JSONResponse(http.StatusOK, "ok")
	})
	h := mw(ok)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.NoError(t, resp.Error())
	require.True(t, w.Code == http.StatusTeapot && w.Body.String() == "ok")
}

// simple handler that records it was called
func okHandler(w http.ResponseWriter, r *http.Request) httpx.Responder {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
	return httpx.JSONResponse(http.StatusOK, "ok")
}

func TestErrorMiddleware(t *testing.T) {
	mw := Errors(logx.NewCILogger("unit-tests"))
	ok := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		return httpx.ErrorResponse(
			httpx.NewTrustedError("my trusted error", errors.New("internal error message")),
			http.StatusTeapot,
		)
	})
	h := mw(ok)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, resp.Error())
	require.Equal(t, "my trusted error", resp.Error().Error())
}

func TestErrorMiddleware_NonTrustedError(t *testing.T) {
	mw := Errors(logx.NewCILogger("unit-tests"))
	ok := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		return httpx.ErrorResponse(
			errors.New("some non-trusted error"),
			http.StatusTeapot,
		)
	})
	h := mw(ok)

	w := httptest.NewRecorder()
	resp := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, resp.Error())
	require.Equal(t, "I'm a teapot", resp.Error().Error()) //non-trusted error is logged by not sent in response
}
