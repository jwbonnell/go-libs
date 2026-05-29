package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jwbonnell/go-libs/pkg/web/httpx"
	"github.com/stretchr/testify/require"
)

func TestMaxBodySize_ContentLengthExceedsLimit(t *testing.T) {
	mw := MaxBodySize(10)
	called := false
	next := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		called = true
		return httpx.NoContentResponse()
	})
	h := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello world"))
	req.ContentLength = 11
	w := httptest.NewRecorder()

	resp := h(w, req)
	require.False(t, called)
	require.Error(t, resp.Error())
	require.Equal(t, http.StatusRequestEntityTooLarge, resp.Status())
	require.True(t, httpx.IsTrustedError(resp.Error()))
	require.Equal(t, "request body too large", resp.Error().Error())
}

func TestMaxBodySize_BodyWithinLimit(t *testing.T) {
	mw := MaxBodySize(100)
	next := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		return httpx.NoContentResponse()
	})
	h := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("small"))
	w := httptest.NewRecorder()

	resp := h(w, req)
	require.NoError(t, resp.Error())
	require.Equal(t, http.StatusNoContent, resp.Status())
}

func TestMaxBodySize_BodyExceedsLimitDuringRead(t *testing.T) {
	mw := MaxBodySize(5)
	var readErr error
	next := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) httpx.Responder {
		_, readErr = io.ReadAll(r.Body)
		return httpx.NoContentResponse()
	})
	h := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("more than five bytes"))
	req.ContentLength = -1 // unknown length; skip early rejection, let MaxBytesReader enforce
	w := httptest.NewRecorder()

	resp := h(w, req)
	// next was called; the error surfaces on the read, not on the response
	require.NoError(t, resp.Error())
	require.Error(t, readErr)
	var maxBytesErr *http.MaxBytesError
	require.ErrorAs(t, readErr, &maxBytesErr)
}
