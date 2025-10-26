package web

import (
	"context"
	"time"
)

type ctxKey int

const key ctxKey = 1

type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

func SetStatusCode(ctx context.Context, statusCode int) {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return
	}

	v.StatusCode = statusCode
}
