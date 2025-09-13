package prettylog

import (
	"context"
	"log/slog"

	"github.com/tidwall/pretty"
)

var _ EntryWriter = (*PrettyJSONWriter)(nil)

// DefaultPrettyJSONWriter is the default entry writer for pretty-printed JSON output
// of log attributes. It excludes standard slog keys (time, level, message, source)
// and formats the remaining attributes as colored JSON.
var DefaultPrettyJSONWriter = NewPrettyJSONWriter()

// NewPrettyJSONWriter creates a new PrettyJSONWriter with default pretty-printing
// options and terminal styling.
func NewPrettyJSONWriter() *PrettyJSONWriter {
	return &PrettyJSONWriter{
		options: pretty.DefaultOptions,
		style:   pretty.TerminalStyle,
		pool:    newLimitedPool(16 * 1024), // 16KB
	}
}

// PrettyJSONWriter is a specialized entry writer for pretty-printed JSON output.
// It formats log attributes as colored, indented JSON, excluding standard slog keys.
//
// It excludes standard slog keys (time, level, message, source)
// and formats the remaining attributes as colored JSON.
type PrettyJSONWriter struct {
	options *pretty.Options
	style   *pretty.Style

	pool *limitedPool
}

// WithPrettyOptions sets the pretty-printing options for JSON formatting.
// This controls indentation, sorting, and other formatting aspects.
func (pr *PrettyJSONWriter) WithPrettyOptions(opts *pretty.Options) *PrettyJSONWriter {
	pr.options = opts
	return pr
}

// WithStyle sets the color styling for JSON output.
// This controls the colors used for different JSON elements like keys, values, etc.
func (pr *PrettyJSONWriter) WithStyle(style *pretty.Style) *PrettyJSONWriter {
	pr.style = style
	return pr
}

// KeyLen implements [EntryWriter] interface. Always returns 0.
func (pr *PrettyJSONWriter) KeyLen(info RecordData) int {
	return 0
}

// Write implements [EntryWriter] interface.
func (pr *PrettyJSONWriter) Write(info RecordData) {
	if info.Buffer.Len() > 0 {
		info.Buffer.WriteByte('\n')
	}
	placeholder := pr.pool.Get()
	defer pr.pool.Put(placeholder)

	opt := cloneHandlerOptions(info.HandlerOptions)
	if opt == nil {
		opt = &slog.HandlerOptions{}
	}
	opt.ReplaceAttr = pr.buildReplaceAttr(opt.ReplaceAttr)
	jsonSerializer := slog.NewJSONHandler(placeholder, opt)
	jsonSerializer.Handle(context.Background(), info.Record)

	b := placeholder.Bytes()
	b = pretty.PrettyOptions(b, pr.options)
	if info.Color {
		b = pretty.Color(b, pr.style)
	}
	info.Buffer.Write(b)
}

type replaceAttrFunc = func(group []string, a slog.Attr) slog.Attr

func (pr *PrettyJSONWriter) buildReplaceAttr(parent replaceAttrFunc) replaceAttrFunc {
	return func(group []string, a slog.Attr) slog.Attr {
		if len(group) == 0 {
			switch a.Key {
			case slog.TimeKey, slog.LevelKey, slog.MessageKey, slog.SourceKey:
				return slog.Attr{}
			}
		}
		if parent != nil {
			return parent(group, a)
		}
		return a
	}
}
