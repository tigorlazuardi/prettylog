package prettylog

import (
	"context"
	"log/slog"

	"github.com/tidwall/pretty"
)

var _ EntryWriter = (*PrettyJSONWriter)(nil)

var DefaultPrettyJSONWriter = NewPrettyJSONWriter()

func NewPrettyJSONWriter() *PrettyJSONWriter {
	return &PrettyJSONWriter{
		options: pretty.DefaultOptions,
		style:   pretty.TerminalStyle,
		pool:    newLimitedPool(16 * 1024), // 16KB
	}
}

type PrettyJSONWriter struct {
	options *pretty.Options
	style   *pretty.Style

	pool *limitedPool
}

func (pr *PrettyJSONWriter) WithPrettyOptions(opts *pretty.Options) *PrettyJSONWriter {
	pr.options = opts
	return pr
}

func (pr *PrettyJSONWriter) WithStyle(style *pretty.Style) *PrettyJSONWriter {
	pr.style = style
	return pr
}

func (pr *PrettyJSONWriter) KeyLen(info RecordInfo) int {
	return 0
}

func (pr *PrettyJSONWriter) Write(info RecordInfo) {
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
