package prettylog

import (
	"bytes"
	"context"
	"log/slog"
	"runtime"
)

// RecordInfo contains contextual information about the current log record being processed.
// It is passed to formatters, stylers, and entry writers to provide all necessary
// context for generating the final log output.
type RecordInfo struct {
	// Context is the context passed to the [slog.Handler.Handle] method.
	Context context.Context
	// Record is the current [slog.Record] being processed.
	//
	// Do note that [Formatter]s and [Styler]s are called for every
	// log record. It is in your best interest to minimize calling potentially
	// expensive operations like [slog.Record.Attrs] or [slog.Record.Clone].
	Record slog.Record
	// PackageName is the name of the package given by setting [WithPackageName] option
	// on [New] or [Handler.Clone].
	PackageName string
	// Frame is the runtime.Frame information about the caller.
	//
	// It's cached here to avoid multiple calls to [runtime.CallerFrames] or [runtime.FuncForPC] inside formatters and stylers.
	//
	// Do note that if [runtime.Frame.Func] is nil, it means that the function information is not available. Ensure to take note of that into account when using it.
	Frame runtime.Frame

	// Color indicates if output should be colored or not.
	//
	// If false, it means the text output must not have ANSI color codes.
	Color bool

	// KeyFieldLength is the length of the longest key field from all the [EntryWriter] after
	// including ANSI color codes.
	//
	// This value is zero when passed into [EntryWriter.Key] (because it is not known yet).
	KeyFieldLength int

	// HandlerOptions is the handler options passed to the handler.
	//
	// This field may be nil and it's members can also be nil.
	HandlerOptions *slog.HandlerOptions

	// Buffer is a buffer that can be used by [EntryWriter]s to write
	// data to.
	//
	// The Buffer is shared between all registered [EntryWriter]s, so
	// you can peek into what previous writers have written. To avoid
	// unwanted log output, avoid calling methods that move the internal
	// cursor backwards like [bytes.Buffer.Reset], [bytes.Buffer.Truncate],
	// or Unread methods outside what is strictly necessary.
	//
	// Do not hold references to this buffer outside the scope as
	// the buffer will be reused for future log records.
	//
	// If you have to keep hold of the value, make a copy of the buffer
	Buffer *bytes.Buffer
}
