package httpx

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrUnsupportedContentType = fmt.Errorf("unsupported content type")

// Decode r.Body and use the Content-Type header to determine which decoder to use.
func Decode(r *http.Request, out any) error {
	ct := r.Header.Get("Content-Type")
	// normalize: drop optional charset and parameters
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}

	switch ct {
	case "application/json", "application/vnd.api+json":
		return decodeJSON(r.Body, out)
	case "application/xml", "text/xml", "application/rss+xml", "application/atom+xml":
		return decodeXML(r.Body, out)
	case "text/plain", "":
		// allow empty Content-Type as plain text
		return decodeText(r.Body, out)
	default:
		return ErrUnsupportedContentType
	}
}

func decodeJSON(body io.Reader, out any) error {
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}

func decodeXML(body io.Reader, out any) error {
	dec := xml.NewDecoder(body)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("xml decode: %w", err)
	}
	return nil
}

func decodeText(body io.Reader, out any) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	switch v := out.(type) {
	case *string:
		*v = string(b)
		return nil
	case *[]byte:
		*v = b
		return nil
	default:
		return fmt.Errorf("text decode: out must be *string or *[]byte")
	}
}
