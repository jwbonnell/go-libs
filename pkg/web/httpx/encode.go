package httpx

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

type Encoder interface {
	Encode(data any) (encoded []byte, contentType string, err error)
}

type (
	PlainTextEncoder struct{}
	JSONEncoder      struct{}
	XMLEncoder       struct{}
	FileEncoder      struct{}
)

func NewEncoder(id string) Encoder {
	var e Encoder
	switch id {
	case "plaintext":
		e = &PlainTextEncoder{}
	case "json":
		e = &JSONEncoder{}
	case "xml":
		e = &XMLEncoder{}
	case "file":
		e = &FileEncoder{}
	default:
		return nil
	}
	return e
}

func (e *PlainTextEncoder) Encode(data any) ([]byte, string, error) {
	str, ok := data.(string)
	if !ok {
		return nil, "", fmt.Errorf("encoder data is not a string")
	}
	return []byte(str), "text/plain", nil
}

func (e *JSONEncoder) Encode(data any) ([]byte, string, error) {
	enc, err := json.Marshal(data)
	return enc, "application/json", err
}

func (e *XMLEncoder) Encode(data any) ([]byte, string, error) {
	enc, err := xml.Marshal(data)
	return enc, "application/xml", err
}

func (e *FileEncoder) Encode(data any) ([]byte, string, error) {
	//TODO
	return []byte{}, "", nil
}
