package http

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type sample struct {
	Name string `json:"name" xml:"name"`
	Age  int    `json:"age" xml:"age"`
}

// errorWriter simulates a ResponseWriter that returns an error on Write.
type errorWriter struct {
	header http.Header
}

func (e *errorWriter) Header() http.Header {
	if e.header == nil {
		e.header = make(http.Header)
	}
	return e.header
}

func (e *errorWriter) Write(b []byte) (int, error) {
	return 0, errors.New("write failed")
}

func (e *errorWriter) WriteHeader(statusCode int) {}

func TestJSONEncoder_Success(t *testing.T) {
	e := &JSONEncoder{}
	w := httptest.NewRecorder()

	s := sample{Name: "Alice", Age: 30}
	err := e.Encode(w, s)
	require.NoError(t, err)

	// check content type
	ct := w.Header().Get("Content-Type")
	require.Equal(t, ct, "application/json")

	// decode body and compare
	var got sample
	err = json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	require.Equal(t, s, got)
}

func TestXMLEncoder_Success(t *testing.T) {
	e := &XMLEncoder{}
	w := httptest.NewRecorder()

	s := sample{Name: "Bob", Age: 42}
	err := e.Encode(w, s)
	require.NoError(t, err)

	// check content type
	ct := w.Header().Get("Content-Type")
	require.Equal(t, ct, "application/xml")

	// decode XML body and compare
	var got sample
	err = xml.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	require.Equal(t, s, got)
}

// Test that encoder returns the write error when the writer fails.
func TestEncoders_WriteError(t *testing.T) {
	tests := []struct {
		name    string
		encoder Encoder
	}{
		{"json", &JSONEncoder{}},
		{"xml", &XMLEncoder{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &errorWriter{}
			err := tt.encoder.Encode(w, sample{Name: "Z", Age: 1})
			require.Error(t, err)
			require.True(t, strings.Contains(err.Error(), "write failed"))
		})
	}
}

// Test that encoders produce deterministic formatting for simple types (optional)
// JSON should be valid and XML should contain the root element.
func TestEncoders_BasicOutputChecks(t *testing.T) {
	j := &JSONEncoder{}
	x := &XMLEncoder{}

	wj := httptest.NewRecorder()
	err := j.Encode(wj, map[string]string{"k": "v"})
	require.NoError(t, err)
	require.True(t, json.Valid(wj.Body.Bytes()))

	// Use a struct for XML (maps unsupported)
	type kv struct {
		XMLName xml.Name `xml:"kv"`
		K       string   `xml:"k"`
		V       string   `xml:"v"`
	}
	wx := httptest.NewRecorder()
	err = x.Encode(wx, kv{K: "k", V: "v"})
	require.NoError(t, err)

	// parse to ensure valid XML
	dec := xml.NewDecoder(bytes.NewReader(wx.Body.Bytes()))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}
}
