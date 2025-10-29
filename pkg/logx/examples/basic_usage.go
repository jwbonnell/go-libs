package main

import (
	"context"
	"github.com/jwbonnell/go-libs/pkg/logx"
	"log/slog"
	"os"
)

// traceIDFromContext is an example implementation of TraceIDFn
func traceIDFromContext(ctx context.Context) string {
	// In a real-world scenario, you might extract a trace ID from the context
	return "123e4567-e89b-12d3-a456-426614174000"
}

func main() {
	// Create a logger that writes to stdout
	logger := logx.New(
		os.Stdout,          // Write to standard output
		slog.LevelDebug,    // Set minimum logx level to Debug
		"my-service",       // Service name
		traceIDFromContext, // Trace ID function
	)

	// Create a context (can be background or with values)
	ctx := context.Background()

	// Demonstrate different logx levels
	logger.Debug(ctx, "This is a debug message",
		"user_id", 12345,
		"action", "login")

	logger.Info(ctx, "User logged in",
		"username", "johndoe",
		"ip", "192.168.1.100")

	logger.Warn(ctx, "Potential performance issue",
		"response_time_ms", 250,
		"threshold_ms", 200)

	logger.Error(ctx, "Failed to connect to database",
		"error", "connection timeout",
		"retry_count", 3)

	// Demonstrate logging with specific caller
	logger.Debugc(ctx, 2, "Detailed debug with specific caller",
		"extra_info", "more context")
}
