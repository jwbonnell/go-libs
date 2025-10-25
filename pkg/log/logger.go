package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

type Level = slog.Level

// TraceIDFn represents a function that can return the trace id from
// the specified context.
type TraceIDFn func(ctx context.Context) string

type Logger struct {
	discard   bool
	handler   slog.Handler
	traceIDFn TraceIDFn
}

// Debug logs at LevelDebug with the given context.
func (log *Logger) Debug(ctx context.Context, msg string, args ...any) {
	log.write(ctx, slog.LevelDebug, 3, msg, args...)
}

// Debugc logs the information at the specified call stack position.
func (log *Logger) Debugc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, slog.LevelDebug, caller, msg, args...)
}

// Info logs at LevelInfo with the given context.
func (log *Logger) Info(ctx context.Context, msg string, args ...any) {
	log.write(ctx, slog.LevelInfo, 3, msg, args...)
}

// Infoc logs the information at the specified call stack position.
func (log *Logger) Infoc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, slog.LevelInfo, caller, msg, args...)
}

// Warn logs at LevelWarn with the given context.
func (log *Logger) Warn(ctx context.Context, msg string, args ...any) {
	log.write(ctx, slog.LevelWarn, 3, msg, args...)
}

// Warnc logs the information at the specified call stack position.
func (log *Logger) Warnc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, slog.LevelWarn, caller, msg, args...)
}

// Error logs at LevelError with the given context.
func (log *Logger) Error(ctx context.Context, msg string, args ...any) {
	log.write(ctx, slog.LevelError, 3, msg, args...)
}

// Errorc logs the information at the specified call stack position.
func (log *Logger) Errorc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, slog.LevelError, caller, msg, args...)
}

func (log *Logger) write(ctx context.Context, level Level, caller int, msg string, args ...any) {
	if log.discard {
		return
	}

	if !log.handler.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:])

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])

	if log.traceIDFn != nil {
		args = append(args, "trace_id", log.traceIDFn(ctx))
	}
	r.Add(args...)

	log.handler.Handle(ctx, r)
}

// New constructs a new log for application use.
func New(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn) *Logger {
	return new(w, minLevel, serviceName, traceIDFn)
}

func new(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn) *Logger {
	// Convert the file name to just the name.ext when this key/value will
	// be logged.
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	// Construct the slog JSON handler for use.
	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true, Level: minLevel, ReplaceAttr: f}))

	// Attributes to add to every log.
	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}

	// Add those attributes and capture the final handler.
	handler = handler.WithAttrs(attrs)

	return &Logger{
		discard:   w == io.Discard,
		handler:   handler,
		traceIDFn: traceIDFn,
	}
}
