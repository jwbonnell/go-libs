package httpx

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

type Encoder interface {
	Encode(data any) (encoded []byte, err error)
}

type (
	PlainTextEncoder struct{}
	JSONEncoder      struct{}
	XMLEncoder       struct{}
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
	default:
		panic("httpx: unknown encoder id: " + id)
	}
	return e
}

func (e *PlainTextEncoder) Encode(data any) ([]byte, error) {
	str, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("encoder data is not a string")
	}
	return []byte(str), nil
}

func (e *JSONEncoder) Encode(data any) ([]byte, error) {
	enc, err := json.Marshal(data)
	return enc, err
}

func (e *XMLEncoder) Encode(data any) ([]byte, error) {
	enc, err := xml.Marshal(data)
	return enc, err
}
