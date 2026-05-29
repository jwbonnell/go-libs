package httpx

import (
	"context"
	"net/http"
)

// Responder handles writing itself to the response writer
type Responder interface {
	Error() error
	Status() int
	Respond(ctx context.Context, w http.ResponseWriter, req *http.Request) error
}

type BaseResponder struct {
	Err        error
	StatusCode int
}

func (br BaseResponder) Error() error {
	return br.Err
}

func (br BaseResponder) Status() int {
	return br.StatusCode
}

// ErrorResponder is for responding with an error
type ErrorResponder struct {
	BaseResponder
	ContentType string
}

func ErrorResponse(err error, statusCode int) ErrorResponder {
	contentType := "plaintext"
	return ErrorResponder{
		ContentType: contentType,
		BaseResponder: BaseResponder{
			Err:        err,
			StatusCode: statusCode,
		},
	}
}

func ErrorResponseWithContentType(err error, statusCode int, contentType string) ErrorResponder {
	if contentType == "" {
		contentType = "text/plain"
	}

	return ErrorResponder{
		ContentType: contentType,
		BaseResponder: BaseResponder{
			Err:        err,
			StatusCode: statusCode,
		},
	}
}

func (r ErrorResponder) Respond(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	var encoder Encoder
	switch r.ContentType {
	case "application/json":
		encoder = NewEncoder("json")
	case "application/xml":
		encoder = NewEncoder("xml")
	default:
		encoder = NewEncoder("plaintext")
	}

	w.WriteHeader(r.StatusCode)
	data, err := encoder.Encode(r.Error().Error())
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// JSONResponder for responding with JSON
type JSONResponder struct {
	BaseResponder
	Data any
}

// JSONResponse is a helper function for responding with JSON
func JSONResponse(statusCode int, data any) Responder {
	return JSONResponder{
		Data: data,
		BaseResponder: BaseResponder{
			StatusCode: statusCode,
		},
	}
}

func (r JSONResponder) Respond(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.StatusCode)

	encoder := NewEncoder("json")
	data, err := encoder.Encode(r.Data)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// XMLResponder is for responding with XML
type XMLResponder struct {
	BaseResponder
	Data any
}

func XMLResponse(statusCode int, data any) Responder {
	return XMLResponder{
		Data: data,
		BaseResponder: BaseResponder{
			StatusCode: statusCode,
		},
	}
}

func (r XMLResponder) Respond(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(r.StatusCode)

	encoder := NewEncoder("xml")
	data, err := encoder.Encode(r.Data)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// PlainTextResponder for responding with plain text
type PlainTextResponder struct {
	BaseResponder
	Data any
}

func PlainTextResponse(statusCode int, data string) Responder {
	return PlainTextResponder{
		Data: data,
		BaseResponder: BaseResponder{
			StatusCode: statusCode,
		},
	}
}

func (r PlainTextResponder) Respond(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(r.StatusCode)

	encoder := NewEncoder("plaintext")
	data, err := encoder.Encode(r.Data)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// FileResponder for serving files with custom headers
type FileResponder struct {
	BaseResponder
	FilePath    string
	ContentType string
	Headers     map[string]string
}

func FileResponse(statusCode int, filePath string, contentType string, headers map[string]string) Responder {
	return FileResponder{
		FilePath:    filePath,
		ContentType: contentType,
		Headers:     headers,
		BaseResponder: BaseResponder{
			StatusCode: statusCode,
		},
	}
}

func (r FileResponder) Respond(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	w.Header().Set("Content-Type", r.ContentType)

	for k, v := range r.Headers {
		w.Header().Set(k, v)
	}

	http.ServeFile(w, req, r.FilePath)
	return nil
}

// StreamResponder for streaming data
type StreamResponder struct {
	BaseResponder
	ContentType string
	Stream      func(w http.ResponseWriter) error
}

func StreamResponse(statusCode int, contentType string, stream func(w http.ResponseWriter) error) Responder {
	return StreamResponder{
		ContentType: contentType,
		Stream:      stream,
		BaseResponder: BaseResponder{
			StatusCode: statusCode,
		},
	}
}

func (r StreamResponder) Respond(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	w.Header().Set("Content-Type", r.ContentType)
	w.WriteHeader(r.StatusCode)
	return r.Stream(w)
}

// RawResponder for complete control
type RawResponder struct {
	BaseResponder
	Body        []byte
	ContentType string
	Headers     map[string]string
}

func RawResponse(statusCode int, body []byte, contentType string) Responder {
	return RawResponder{
		Body:        body,
		ContentType: contentType,
		Headers:     make(map[string]string),
		BaseResponder: BaseResponder{
			StatusCode: statusCode,
		},
	}
}

func (r RawResponder) Respond(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	w.Header().Set("Content-Type", r.ContentType)
	for k, v := range r.Headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(r.StatusCode)
	_, err := w.Write(r.Body)
	return err
}

// NoContentResponder is for responding with no content
type NoContentResponder struct {
	BaseResponder
}

func NoContentResponse() Responder {
	return NoContentResponder{
		BaseResponder: BaseResponder{
			StatusCode: http.StatusNoContent,
		},
	}
}

func (r NoContentResponder) Respond(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	w.WriteHeader(r.StatusCode)
	return nil

}

// NoopResponder is for not responding with anything
type NoopResponder struct {
	BaseResponder
}

func NoopResponse() Responder {
	return NoopResponder{}
}

func (r NoopResponder) Respond(_ context.Context, _ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// CustomResponder is for when custom response implementations are required
type CustomResponder struct {
	BaseResponder
	Fn func(ctx context.Context, w http.ResponseWriter, req *http.Request) error
}

func Custom(statusCode int, fn func(ctx context.Context, w http.ResponseWriter, req *http.Request) error) CustomResponder {
	return CustomResponder{
		Fn: fn,
		BaseResponder: BaseResponder{
			StatusCode: statusCode,
		},
	}
}

func (r CustomResponder) Respond(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	SetContextStatusCode(ctx, r.StatusCode)
	return r.Fn(ctx, w, req)
}
