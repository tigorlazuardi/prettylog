package prettylog

import (
	"log/slog"
	"runtime"
)

type RecordInfo struct {
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

	// DisableColor indicates if color output should be disabled.
	//
	// If true, it means the text output must not have ANSI color codes.
	DisableColor bool

	// KeyFieldLength is the length of the longest key field from all the [EntryWriter] after
	// including ANSI color codes.
	//
	// This value is zero when passed into [EntryWriter.Key] (because it is not known yet).
	KeyFieldLength int
}
