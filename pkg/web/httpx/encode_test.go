package httpx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type sample struct {
	Name    string `json:"name" xml:"name"`
	Age     int    `json:"age" xml:"age"`
	Unknown int    `json:"unknown" xml:"unknown"`
}

func TestJSONEncoder_Success(t *testing.T) {
	e := &JSONEncoder{}
	s := sample{Name: "Alice", Age: 30}
	data, err := e.Encode(s)
	require.NoError(t, err)

	// decode body and compare
	var got sample
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, s, got)
}

func TestXMLEncoder_Success(t *testing.T) {
	e := &XMLEncoder{}
	s := sample{Name: "Bob", Age: 42}
	data, err := e.Encode(s)
	require.NoError(t, err)

	// decode body and compare
	var got sample
	err = xml.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, s, got)
}

func TestPlainTextEncoder_Success(t *testing.T) {
	e := &PlainTextEncoder{}
	data, err := e.Encode("something")
	require.NoError(t, err)
	require.Equal(t, string(data), "something")
}

func TestPlainTextEncoder_Error(t *testing.T) {
	e := &PlainTextEncoder{}
	data, err := e.Encode(struct{ Name string }{Name: "Z"})
	require.Error(t, err)
	require.Equal(t, "encoder data is not a string", err.Error())
	require.Nil(t, data)
}

// Test that encoders produce deterministic formatting for simple types (optional)
// JSON should be valid and XML should contain the root element.
func TestEncoders_BasicOutputChecks(t *testing.T) {
	j := &JSONEncoder{}
	x := &XMLEncoder{}

	data, err := j.Encode(map[string]string{"k": "v"})
	require.NoError(t, err)
	require.True(t, json.Valid(data))

	// Use a struct for XML (maps unsupported)
	type kv struct {
		XMLName xml.Name `xml:"kv"`
		K       string   `xml:"k"`
		V       string   `xml:"v"`
	}

	data, err = x.Encode(kv{K: "k", V: "v"})
	require.NoError(t, err)

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
