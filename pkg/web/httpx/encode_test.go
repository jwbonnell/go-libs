package httpx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type sample struct {
	Name string `json:"name" xml:"name"`
	Age  int    `json:"age" xml:"age"`
}

func TestJSONEncoder_Success(t *testing.T) {
	e := &JSONEncoder{}
	s := sample{Name: "Alice", Age: 30}
	data, contextType, err := e.Encode(s)
	require.NoError(t, err)
	require.Equal(t, contextType, "application/json")

	// decode body and compare
	var got sample
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, s, got)
}

func TestXMLEncoder_Success(t *testing.T) {
	e := &XMLEncoder{}
	s := sample{Name: "Bob", Age: 42}
	data, contextType, err := e.Encode(s)
	require.NoError(t, err)
	require.Equal(t, contextType, "application/xml")

	// decode body and compare
	var got sample
	err = xml.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, s, got)
}

func TestPlainTextEncoder_Success(t *testing.T) {
	e := &PlainTextEncoder{}
	data, contextType, err := e.Encode("something")
	require.NoError(t, err)
	require.Equal(t, contextType, "text/plain")
	require.Equal(t, string(data), "something")
}

// Test that encoder returns write error
func TestEncoders_WriteError(t *testing.T) {
	tests := []struct {
		name        string
		encoder     Encoder
		contentType string
	}{
		{"json", &JSONEncoder{}, "application/json"},
		{"xml", &XMLEncoder{}, "application/xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, contentType, err := tt.encoder.Encode(sample{Name: "Z", Age: 1})
			require.Error(t, err)
			require.Equal(t, contentType, tt.contentType)
			require.True(t, strings.Contains(err.Error(), "write failed"))
		})
	}
}

func TestPlainTextEncoder_Error(t *testing.T) {
	e := &PlainTextEncoder{}
	data, contextType, err := e.Encode(struct{ Name string }{Name: "Z"})
	require.Error(t, err)
	require.Equal(t, contextType, "")
	require.Equal(t, "encoder data is not a string", err.Error())
	require.Nil(t, data)
}

// Test that encoders produce deterministic formatting for simple types (optional)
// JSON should be valid and XML should contain the root element.
func TestEncoders_BasicOutputChecks(t *testing.T) {
	j := &JSONEncoder{}
	x := &XMLEncoder{}

	data, contentType, err := j.Encode(map[string]string{"k": "v"})
	require.NoError(t, err)
	require.Equal(t, "application/json", contentType)
	require.True(t, json.Valid(data))

	// Use a struct for XML (maps unsupported)
	type kv struct {
		XMLName xml.Name `xml:"kv"`
		K       string   `xml:"k"`
		V       string   `xml:"v"`
	}

	data, contentType, err = x.Encode(kv{K: "k", V: "v"})
	require.NoError(t, err)
	require.Equal(t, "application/xml", contentType)

	// parse to ensure valid XML
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}
}
