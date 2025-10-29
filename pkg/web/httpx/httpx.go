package httpx

import (
	"context"
	context2 "github.com/jwbonnell/go-libs/pkg/web/context"
	"net/http"
)

type Response struct {
	StatusCode int
	Data       any
	Encoder    Encoder
	Err        error
}

func PlainTextResponse(statusCode int, data any) Response {
	return Response{
		StatusCode: statusCode,
		Data:       data,
		Encoder:    NewEncoder("plaintext"),
	}
}

func JSONResponse(statusCode int, data any) Response {
	return Response{
		StatusCode: statusCode,
		Data:       data,
		Encoder:    NewEncoder("json"),
	}
}

func XMLResponse(statusCode int, data any) Response {
	return Response{
		StatusCode: statusCode,
		Data:       data,
		Encoder:    NewEncoder("xml"),
	}
}

func SendErrorResponse(err error, statusCode int) Response {
	return Response{
		StatusCode: statusCode,
		Err:        err,
		Encoder:    NewEncoder("json"),
	}
}

func Respond(ctx context.Context, w http.ResponseWriter, resp Response) error {
	context2.SetStatusCode(ctx, resp.StatusCode)

	if resp.StatusCode == http.StatusNoContent {
		w.WriteHeader(resp.StatusCode)
		return nil
	}

	data, contentType, err := resp.Encoder.Encode(resp.Data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(resp.StatusCode)

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}
