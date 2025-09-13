// Package prettylog provides a pretty-printed structured logging handler for Go's standard library slog package.
//
// prettylog offers customizable formatting, coloring, and modular entry writers for beautiful console output.
// It implements the slog.Handler interface and can be used as a drop-in replacement for other slog handlers.
//
// # Features
//
//   - Colorized output with terminal detection
//   - Modular entry writers for different log components
//   - Configurable time, function, and file/line formatting
//   - Pretty-printed JSON for structured data
//   - Thread-safe operation with buffer pooling
//   - Extensive customization through variadic options
//
// # Basic Usage
//
// The simplest way to use prettylog is to create a handler with default settings:
//
//	handler := prettylog.New()
//	logger := slog.New(handler)
//	logger.Info("Hello, world!", "key", "value")
//
// # Configuration Options
//
// prettylog supports extensive customization through options:
//
//	handler := prettylog.New(
//	    prettylog.WithPackageName("myapp"),
//	    prettylog.WithLevel(slog.LevelDebug),
//	    prettylog.WithOutput(os.Stdout),
//	    prettylog.WithAddSource(true),
//	)
//
// Available configuration options:
//
//   - WithPackageName(string): Set package name for function trimming
//   - WithOutput(io.Writer): Set output destination
//   - WithLevel(slog.Level): Set minimum log level
//   - WithAddSource(bool): Enable/disable source information
//   - WithReplaceAttr(func): Set attribute replacement function
//   - WithHandlerOptions(*slog.HandlerOptions): Set complete handler options
//   - WithColor(bool): Enable/disable colored output
//   - WithNoColor(): Disable colored output (convenience function)
//   - WithPoolSize(int): Set buffer pool size
//
// # Writer Management
//
// You can customize log components by managing entry writers:
//
//	// Replace all writers with custom ones
//	handler := prettylog.New(
//	    prettylog.WithWriters(myLevelWriter, myMessageWriter),
//	)
//
//	// Add additional writers to defaults
//	customWriter := prettylog.NewCommonWriter(myFormatter).
//	    WithStaticKey("Custom").
//	    WithValueColorizer(prettylog.BoldColoredStyler)
//	
//	handler := prettylog.New(
//	    prettylog.WithAdditionalWriters(customWriter),
//	)
//
//	// Remove specific writers from defaults
//	handler := prettylog.New(
//	    prettylog.WithoutWriters(prettylog.DefaultTimeWriter),
//	)
//
//	// Replace specific writers
//	handler := prettylog.New(
//	    prettylog.ReplaceWriter(prettylog.DefaultLevelWriter, myLevelWriter),
//	)
//
// Writer management options:
//
//   - WithWriters(...EntryWriter): Replace all writers
//   - WithAdditionalWriters(...EntryWriter): Add to existing writers
//   - WithoutWriters(...EntryWriter): Remove specific writers
//   - ReplaceWriter(old, new): Replace a specific writer
//
// # Entry Writers
//
// prettylog uses a modular system of EntryWriter components:
//
//   - DefaultLevelWriter: Log level with background color
//   - DefaultMessageWriter: Log message with level-based color  
//   - DefaultTimeWriter: Timestamp in configurable format
//   - DefaultFunctionWriter: Function name with optional package trimming
//   - DefaultFileLineWriter: File path and line number
//   - DefaultPrettyJSONWriter: Pretty-printed JSON for structured data
//
// Each writer can be individually customized using their With* methods or replaced entirely.
//
// # Color Support
//
// prettylog automatically detects terminal capabilities and enables colors when appropriate.
// Color support can be explicitly controlled:
//
//	// Force colors off
//	handler := prettylog.New(prettylog.WithNoColor())
//	
//	// Force colors on
//	handler := prettylog.New(prettylog.WithColor(true))
//
// # Performance
//
// prettylog uses buffer pooling to minimize memory allocations during logging.
// The pool size can be configured:
//
//	handler := prettylog.New(
//	    prettylog.WithPoolSize(32 * 1024 * 1024), // 32MB pool
//	)
package prettylog
