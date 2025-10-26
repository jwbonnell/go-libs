package http

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type Encoder interface {
	Encode(w http.ResponseWriter, data interface{}) error
}

type (
	JSONEncoder struct{}
	XMLEncoder  struct{}
	FileEncoder struct{}
)

func (e *JSONEncoder) Encode(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

func (e *XMLEncoder) Encode(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/xml")
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(data)
}

func (e *FileEncoder) Encode(w http.ResponseWriter, data interface{}) error {
	//TODO
	return nil
}
