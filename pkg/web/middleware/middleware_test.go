package middleware

import (
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPanics_RecoverFromPanic(t *testing.T) {
	mw := Panics()
	panicker := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		panic("boom")
	})
	h := mw(panicker)

	w := httptest.NewRecorder()
	err := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), "boom")
	require.Contains(t, err.Error(), "TRACE[")
	require.Contains(t, err.Error(), "goroutine")
}

func TestPanics_PassThroughError(t *testing.T) {
	mw := Panics()
	want := errors.New("handler error")
	bad := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return want
	})
	h := mw(bad)

	w := httptest.NewRecorder()
	err := h(w, httptest.NewRequest("GET", "/", nil))
	require.Error(t, err)
	require.Equal(t, err.Error(), want.Error())
}

func TestPanics_PassThroughNil(t *testing.T) {
	mw := Panics()
	ok := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusTeapot) // 418
		_, _ = w.Write([]byte("ok"))
		return nil
	})
	h := mw(ok)

	w := httptest.NewRecorder()
	err := h(w, httptest.NewRequest("GET", "/", nil))
	require.NoError(t, err)
	require.True(t, w.Code == http.StatusTeapot && w.Body.String() == "ok")
}
