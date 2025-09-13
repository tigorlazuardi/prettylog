package prettylog

import (
	"io"
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"
)

var DefaultWriters = [...]EntryWriter{
	DefaultLevelWriter,
	DefaultMessageWriter,
	DefaultTimeWriter,
	DefaultFunctionWrtier,
	DefaultFileLineWriter,
	DefaultPrettyJSONWriter,
}

func New(opts ...Option) *Handler {
	h := &Handler{
		writer: WrapWriteLocker(os.Stderr),
		opts: &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		},
		addNewLine:  true,
		pool:        newLimitedPool(defaultLimitedPoolSize),
		attrs:       []slog.Attr{},
		groups:      []string{},
		color:       CanColor(os.Stderr),
		packageName: "",
		writers:     DefaultWriters[:],
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(h)
	}
	return h
}

// CanColor determines if the given writer supports pretty output (i.e., is a terminal and has tty).
//
// [CanColor] is called by [New] to determine if pretty output should be enabled (overrideable),
// but will not be called by [Handler.Clone].
//
// It does so by checking if the writer has an Fd() method and returns a file descriptor. This is typically implemented by [os.File].
//
// If the writer is wrapped (e.g., by a [WriteLocker]), it will attempt to unwrap if it implements interface { Unwrap() io.Writer }.
// If available, then this function will check the underlying writer. This repeats
// until a writer without Unwrap is found or a nil writer is encountered.
func CanColor(w io.Writer) bool {
	for {
		if w == nil {
			return false
		}
		if fd, ok := (w).(interface{ Fd() uintptr }); ok {
			return isatty.IsTerminal(fd.Fd())
		}
		if uw, ok := w.(interface{ Unwrap() io.Writer }); ok {
			w = uw.Unwrap()
			continue
		}
		return false
	}
}
