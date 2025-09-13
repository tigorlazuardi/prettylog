package prettylog_test

import (
	"log/slog"
	"os"

	"github.com/tigorlazuardi/prettylog"
)

// ExampleNew demonstrates basic usage of prettylog with default settings.
func ExampleNew() {
	// Create a handler with default settings
	handler := prettylog.New()
	logger := slog.New(handler)

	// Log some messages
	logger.Info("Application started", "version", "1.0.0")
	logger.Warn("This is a warning", "component", "auth")
	logger.Error("Something went wrong", "error", "connection timeout")

	// Output will be colorized and formatted with default writers
}

// ExampleNew_customized demonstrates advanced configuration options.
func ExampleNew_customized() {
	handler := prettylog.New(
		prettylog.WithPackageName("myapp"),
		prettylog.WithLevel(slog.LevelDebug),
		prettylog.WithOutput(os.Stdout),
	)

	logger := slog.New(handler)
	logger.Debug("Debug message with custom formatting")
	logger.Info("Info message", "user_id", 12345, "action", "login")
}

// ExampleNew_customWriter demonstrates creating and using custom entry writers.
func ExampleNew_customWriter() {
	// Create a custom writer for request IDs
	requestIDWriter := prettylog.NewCommonWriter(func(info prettylog.RecordData) string {
		// Extract request ID from log record attributes
		info.Record.Attrs(func(a slog.Attr) bool {
			if a.Key == "request_id" {
				return false // Stop iteration, we found it
			}
			return true // Continue iteration
		})
		return "" // Would return the actual request ID in practice
	}).WithStaticKey("RequestID").WithValueColorizer(prettylog.PlainStyler)

	handler := prettylog.New(
		prettylog.WithAdditionalWriters(requestIDWriter),
	)

	logger := slog.New(handler)
	logger.Info("Request processed", "request_id", "abc-123", "duration_ms", 42)
}

// ExampleWithWriters demonstrates replacing the default writers entirely.
func ExampleWithWriters() {
	// Create custom writers
	levelWriter := prettylog.NewCommonWriter(prettylog.DefaultLevelValuer).
		WithStaticKey("LEVEL").
		WithValueColorizer(prettylog.SimpleColoredStyler)

	messageWriter := prettylog.NewCommonWriter(prettylog.DefaultMessageValuer).
		WithValueColorizer(prettylog.PlainStyler)

	handler := prettylog.New(
		prettylog.WithWriters(levelWriter, messageWriter),
	)

	logger := slog.New(handler)
	logger.Info("Custom writer configuration")
}

// ExampleWithoutWriters demonstrates removing specific default writers.
func ExampleWithoutWriters() {
	handler := prettylog.New(
		// Remove time and function writers, keep others
		prettylog.WithoutWriters(
			prettylog.DefaultTimeWriter,
			prettylog.DefaultFunctionWrtier,
		),
	)

	logger := slog.New(handler)
	logger.Info("No timestamp or function info", "data", "important")
}

// ExampleHandler_Clone demonstrates cloning handlers with modifications.
func ExampleHandler_Clone() {
	// Create base handler
	baseHandler := prettylog.New(
		prettylog.WithPackageName("myapp"),
		prettylog.WithLevel(slog.LevelInfo),
	)

	// Clone for debug logging with additional options
	debugHandler := baseHandler.Clone(
		prettylog.WithLevel(slog.LevelDebug),
	)

	baseLogger := slog.New(baseHandler)
	debugLogger := slog.New(debugHandler)

	baseLogger.Info("Base logger message")
	debugLogger.Debug("Debug logger message") // This will be shown due to debug level
}

// ExampleTimeWriter demonstrates custom time writer configuration.
func ExampleTimeWriter() {
	// RFC3339 format
	timeWriterRFC3339 := prettylog.NewTimeWriter().WithRFC3339Format()
	handler1 := prettylog.New(
		prettylog.ReplaceWriter(prettylog.DefaultTimeWriter, timeWriterRFC3339),
	)

	// Custom format
	timeWriterCustom := prettylog.NewTimeWriter().WithTimeFormat("2006-01-02 15:04:05")
	handler2 := prettylog.New(
		prettylog.ReplaceWriter(prettylog.DefaultTimeWriter, timeWriterCustom),
	)

	// Time-only format (default)
	timeWriterOnly := prettylog.NewTimeWriter().WithTimeOnlyFormat()
	handler3 := prettylog.New(
		prettylog.ReplaceWriter(prettylog.DefaultTimeWriter, timeWriterOnly),
	)

	logger1 := slog.New(handler1)
	logger2 := slog.New(handler2)
	logger3 := slog.New(handler3)

	logger1.Info("RFC3339 timestamp")
	logger2.Info("Custom timestamp format")
	logger3.Info("Time-only timestamp")
}

// ExampleFunctionWriter demonstrates function name formatting options.
func ExampleFunctionWriter() {
	// Short function names (package prefix trimmed)
	functionWriterShort := prettylog.NewFunctionWriter().WithShort(true)
	handler1 := prettylog.New(
		prettylog.WithPackageName("myapp"),
		prettylog.ReplaceWriter(prettylog.DefaultFunctionWrtier, functionWriterShort),
	)

	// Full function names
	functionWriterFull := prettylog.NewFunctionWriter().WithShort(false)
	handler2 := prettylog.New(
		prettylog.ReplaceWriter(prettylog.DefaultFunctionWrtier, functionWriterFull),
	)

	logger1 := slog.New(handler1)
	logger2 := slog.New(handler2)

	logger1.Info("Short function name")
	logger2.Info("Full function name")
}

// ExamplePrettyJSONWriter demonstrates structured data formatting.
func ExamplePrettyJSONWriter() {
	handler := prettylog.New()
	logger := slog.New(handler)

	// Log with structured data - will be pretty-printed as JSON
	logger.Info("User action",
		"user", slog.GroupValue(
			slog.String("id", "12345"),
			slog.String("name", "John Doe"),
			slog.String("email", "john@example.com"),
		),
		"action", "login",
		"metadata", slog.GroupValue(
			slog.String("ip", "192.168.1.1"),
			slog.String("user_agent", "Mozilla/5.0..."),
			slog.Int("session_duration", 3600),
		),
	)
}

// ExampleCanColor demonstrates color detection functionality.
func ExampleCanColor() {
	// Check if a writer supports color
	if prettylog.CanColor(os.Stdout) {
		handler := prettylog.New(
			prettylog.WithOutput(os.Stdout),
			prettylog.WithColor(true),
		)
		logger := slog.New(handler)
		logger.Info("Colorized output to stdout")
	} else {
		handler := prettylog.New(
			prettylog.WithOutput(os.Stdout),
			prettylog.WithColor(false),
		)
		logger := slog.New(handler)
		logger.Info("Plain output to stdout")
	}
}
