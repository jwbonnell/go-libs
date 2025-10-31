package httpx

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type sampleJSON struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type sampleXML struct {
	XMLName struct{} `xml:"person"`
	Name    string   `xml:"name"`
	Age     int      `xml:"age"`
}

func newRequest(body, contentType string) *http.Request {
	return &http.Request{
		Header: http.Header{
			"Content-Type": []string{contentType},
		},
		Body: ioNopCloser{strings.NewReader(body)},
	}
}

// ioNopCloser implements io.ReadCloser around an io.Reader for tests.
type ioNopCloser struct {
	*strings.Reader
}

func (r ioNopCloser) Close() error { return nil }

func TestDecodeJSON_Success(t *testing.T) {
	req := newRequest(`{"name":"Alice","age":30}`, "application/json")
	var out sampleJSON
	err := Decode(req, &out)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", out.Name)
	assert.Equal(t, 30, out.Age)
}

func TestDecodeJSON_UnknownField_Fails(t *testing.T) {
	req := newRequest(`{"name":"Bob","age":25,"extra":true}`, "application/json")
	var out sampleJSON
	err := Decode(req, &out)
	assert.Error(t, err)
}

func TestDecodeXML_Success(t *testing.T) {
	req := newRequest(`<person><name>Carol</name><age>40</age></person>`, "application/xml")
	var out sampleXML
	err := Decode(req, &out)
	assert.NoError(t, err)
	assert.Equal(t, "Carol", out.Name)
	assert.Equal(t, 40, out.Age)
}

func TestDecodeText_String(t *testing.T) {
	req := newRequest("hello world", "text/plain")
	var out string
	err := Decode(req, &out)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", out)
}

func TestDecodeText_DefaultWhenEmptyContentType(t *testing.T) {
	req := newRequest("default text", "")
	var out string
	err := Decode(req, &out)
	assert.NoError(t, err)
	assert.Equal(t, "default text", out)
}

func TestUnsupportedContentType(t *testing.T) {
	req := newRequest("{}", "application/octet-stream")
	var out map[string]any
	err := Decode(req, &out)
	assert.ErrorIs(t, err, ErrUnsupportedContentType)
}
