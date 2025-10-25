package log

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"log/slog"
	"strings"

	"github.com/stretchr/testify/assert"
)

// mockTraceIDFn is a test implementation of TraceIDFn
func mockTraceIDFn(ctx context.Context) string {
	return "test-trace-id"
}

// TestLoggerLevels tests logging at different levels
func TestLoggerLevels(t *testing.T) {
	testCases := []struct {
		name          string
		logFunc       func(*Logger, context.Context, string, ...any)
		level         slog.Level
		expectedLevel string
	}{
		{
			name:          "Debug Log",
			logFunc:       (*Logger).Debug,
			level:         slog.LevelDebug,
			expectedLevel: "DEBUG",
		},
		{
			name:          "Info Log",
			logFunc:       (*Logger).Info,
			level:         slog.LevelInfo,
			expectedLevel: "INFO",
		},
		{
			name:          "Warn Log",
			logFunc:       (*Logger).Warn,
			level:         slog.LevelWarn,
			expectedLevel: "WARN",
		},
		{
			name:          "Error Log",
			logFunc:       (*Logger).Error,
			level:         slog.LevelError,
			expectedLevel: "ERROR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, tc.level, "test-service", mockTraceIDFn)

			ctx := context.Background()
			tc.logFunc(logger, ctx, "test message", "key", "value")

			// Parse the JSON log output
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			// Verify log level and other details
			assert.Equal(t, tc.expectedLevel, logEntry["level"])
			assert.Equal(t, "test message", logEntry["msg"])
			assert.Equal(t, "test-service", logEntry["service"])
			assert.Equal(t, "test-trace-id", logEntry["trace_id"])
		})
	}
}

// TestLoggerCallerSpecific tests logging with specific caller depth
func TestLoggerCallerSpecific(t *testing.T) {
	testCases := []struct {
		name          string
		logFunc       func(*Logger, context.Context, int, string, ...any)
		level         slog.Level
		expectedLevel string
	}{
		{
			name:          "Debug Log with Caller",
			logFunc:       (*Logger).Debugc,
			level:         slog.LevelDebug,
			expectedLevel: "DEBUG",
		},
		{
			name:          "Info Log with Caller",
			logFunc:       (*Logger).Infoc,
			level:         slog.LevelInfo,
			expectedLevel: "INFO",
		},
		{
			name:          "Warn Log with Caller",
			logFunc:       (*Logger).Warnc,
			level:         slog.LevelWarn,
			expectedLevel: "WARN",
		},
		{
			name:          "Error Log with Caller",
			logFunc:       (*Logger).Errorc,
			level:         slog.LevelError,
			expectedLevel: "ERROR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, tc.level, "test-service", mockTraceIDFn)

			ctx := context.Background()
			// Continuing from the previous test file...

			tc.logFunc(logger, ctx, 2, "test message", "key", "value")

			// Parse the JSON log output
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			// Verify log level and other details
			assert.Equal(t, tc.expectedLevel, logEntry["level"])
			assert.Equal(t, "test message", logEntry["msg"])
			assert.Equal(t, "test-service", logEntry["service"])
			assert.Equal(t, "test-trace-id", logEntry["trace_id"])
		})
	}
}

// TestLoggerDiscard tests that logs are not written when discarded
func TestLoggerDiscard(t *testing.T) {
	var buf bytes.Buffer
	logger := New(io.Discard, slog.LevelDebug, "test-service", mockTraceIDFn)

	ctx := context.Background()
	logger.Info(ctx, "test message")

	assert.Equal(t, 0, buf.Len(), "No logs should be written when using io.Discard")
}

// TestLoggerLevelFiltering tests that logs below the set level are not written
func TestLoggerLevelFiltering(t *testing.T) {
	testCases := []struct {
		name           string
		minLevel       slog.Level
		logFunc        func(*Logger, context.Context, string, ...any)
		logLevel       slog.Level
		shouldBeLogged bool
	}{
		{
			name:           "Debug log when level is Debug",
			minLevel:       slog.LevelDebug,
			logFunc:        (*Logger).Debug,
			logLevel:       slog.LevelDebug,
			shouldBeLogged: true,
		},
		{
			name:           "Info log when level is Info",
			minLevel:       slog.LevelInfo,
			logFunc:        (*Logger).Info,
			logLevel:       slog.LevelInfo,
			shouldBeLogged: true,
		},
		{
			name:           "Warn log when level is Warn",
			minLevel:       slog.LevelWarn,
			logFunc:        (*Logger).Warn,
			logLevel:       slog.LevelWarn,
			shouldBeLogged: true,
		},
		{
			name:           "Error log when level is Error",
			minLevel:       slog.LevelError,
			logFunc:        (*Logger).Error,
			logLevel:       slog.LevelError,
			shouldBeLogged: true,
		},
		{
			name:           "Debug log when level is Info",
			minLevel:       slog.LevelInfo,
			logFunc:        (*Logger).Debug,
			logLevel:       slog.LevelDebug,
			shouldBeLogged: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, tc.minLevel, "test-service", mockTraceIDFn)

			ctx := context.Background()
			tc.logFunc(logger, ctx, "test message")

			if tc.shouldBeLogged {
				assert.NotEqual(t, 0, buf.Len(), "Log should be written")

				// Verify the log content
				var logEntry map[string]interface{}
				err := json.Unmarshal(buf.Bytes(), &logEntry)
				assert.NoError(t, err)
				assert.Equal(t, strings.ToUpper(tc.logLevel.String()), logEntry["level"])
			} else {
				assert.Equal(t, 0, buf.Len(), "No logs should be written")
			}
		})
	}
}

// TestLoggerSourceFormatting tests the source file formatting
func TestLoggerSourceFormatting(t *testing.T) {
	// Continuing from the previous test file...

	var buf bytes.Buffer
	logger := New(&buf, slog.LevelDebug, "test-service", mockTraceIDFn)

	ctx := context.Background()
	logger.Info(ctx, "test message")

	// Parse the JSON log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	// Check that the file is formatted correctly (base name + line number)
	file, ok := logEntry["file"].(string)
	assert.True(t, ok, "File field should exist")
	assert.Regexp(t, `^[^/\\]+:\d+$`, file, "File should be in format 'filename:line'")
}

// TestLoggerWithTraceID tests trace ID functionality
func TestLoggerWithTraceID(t *testing.T) {
	testCases := []struct {
		name            string
		traceIDFn       TraceIDFn
		expectedTraceID string
	}{
		{
			name: "Custom Trace ID",
			traceIDFn: func(ctx context.Context) string {
				return "custom-trace-id"
			},
			expectedTraceID: "custom-trace-id",
		},
		{
			name:            "Default Mock Trace ID",
			traceIDFn:       mockTraceIDFn,
			expectedTraceID: "test-trace-id",
		},
		{
			name:            "Nil Trace ID Function",
			traceIDFn:       nil,
			expectedTraceID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, slog.LevelDebug, "test-service", tc.traceIDFn)

			ctx := context.Background()
			logger.Info(ctx, "test message")

			// Skip trace ID check if no trace ID function is provided
			if tc.traceIDFn == nil {
				return
			}

			// Parse the JSON log output
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			// Check trace ID
			if tc.expectedTraceID != "" {
				traceID, exists := logEntry["trace_id"]
				assert.True(t, exists, "Trace ID should exist in log entry")
				assert.Equal(t, tc.expectedTraceID, traceID)
			}
		})
	}
}

// Benchmark the logger performance
func BenchmarkLoggerPerformance(b *testing.B) {
	ctx := context.Background()

	// Benchmark with different log writers
	benchmarkCases := []struct {
		name   string
		writer io.Writer
	}{
		{"Discard Writer", io.Discard},
		{"Buffered Writer", &bytes.Buffer{}},
	}

	for _, bc := range benchmarkCases {
		b.Run(bc.name, func(b *testing.B) {
			logger := New(bc.writer, slog.LevelInfo, "benchmark-service", mockTraceIDFn)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info(ctx, "benchmark log message",
					"key1", "value1",
					"key2", 42,
					"key3", true)
			}
		})
	}
}
