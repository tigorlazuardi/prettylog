package prettylog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"runtime"
	"time"

	"github.com/tidwall/pretty"
)

type Handler struct {
	attrs  []slog.Attr
	groups []string
	writer WriteLocker
	opts   *slog.HandlerOptions
	pretty bool

	// Formatters
	functionFormatter    func(lvl slog.Level, packageNamePrefix string, frame runtime.Frame) string
	fileLineFormatter    func(lvl slog.Level, frame runtime.Frame) string
	timeFormatter        func(lvl slog.Level, t time.Time) string
	headerLevelFormatter func(lvl slog.Level) string
	messageFormatter     func(lvl slog.Level, msg string) string

	// Stylers
	sourceStyler      func(lvl slog.Level, s string) string
	timeStyler        func(lvl slog.Level, s string) string
	headerLevelStyler func(lvl slog.Level, s string) string
	messageStyler     func(lvl slog.Level, s string) string

	// JSON prettier options
	prettyOption *pretty.Options
	prettyColor  *pretty.Style

	packageNamePrefix string
	addNewLine        bool
	pool              *limitedPool
}

// Enabled implements [slog.Handler] interface.
func (h *Handler) Enabled(ctx context.Context, lvl slog.Level) bool {
	// Copied from slog.commonHandler.enabled
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return lvl >= minLevel
}

// Handle handles the Record.
// It will only be called when Enabled returns true.
// The Context argument is as for Enabled.
// It is present solely to provide Handlers access to the context's values.
// Canceling the context should not affect record processing.
// (Among other things, log messages may be necessary to debug a
// cancellation-related problem.)
//
// Handle methods that produce output should observe the following rules:
//   - If r.Time is the zero time, ignore the time.
//   - If r.PC is zero, ignore it.
//   - Attr's values should be resolved.
//   - If an Attr's key and value are both the zero value, ignore the Attr.
//     This can be tested with attr.Equal(Attr{}).
//   - If a group's key is empty, inline the group's Attrs.
//   - If a group has no Attrs (even if it has a non-empty key),
//     ignore it.
func (ha *Handler) Handle(ctx context.Context, rec slog.Record) error {
	buf := ha.pool.Get()
	defer ha.pool.Put(buf)

	if !ha.pretty { // If not pretty, just pass everything to slog.JSONHandler.
		serializer := ha.createSerializer(buf)
		if err := serializer.Handle(ctx, rec); err != nil {
			return err
		}
		ha.writer.Lock()
		defer ha.writer.Unlock()
		_, err := io.Copy(ha.writer, buf)
		return err
	}

	return nil
}

// WithAttrs implements [slog.Handler] interface.
func (ha *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cloned := ha.Clone()
	cloned.attrs = append(cloned.attrs, attrs...)
	return cloned
}

func (ha *Handler) createSerializer(buf *bytes.Buffer) *slog.JSONHandler {
	serializer := slog.NewJSONHandler(buf, ha.opts)
	// Satify the [slog.Handler.WithGroup] contract by applying groups in reverse order.
	// That is, the last group added should be the outermost group.
	// See [slog.Handler.WithGroup] documentation for details.
	for i := len(ha.groups) - 1; i >= 0; i-- {
		serializer = serializer.WithGroup(ha.groups[i]).(*slog.JSONHandler)
	}
	if len(ha.attrs) > 0 {
		serializer = serializer.WithAttrs(ha.attrs).(*slog.JSONHandler)
	}
	return serializer
}

// WithGroup implements [slog.Handler] interface.
func (ha *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		// satisfy the slog.Handler.WithGroup contract.
		//
		// according to slog.Handler interface, empty name must return the same handler
		return ha
	}
	cloned := ha.Clone()
	cloned.groups = append(cloned.groups, name)
	return cloned
}

// Clone makes a copy of the handler and applies the given options to it.
//
// Other than Attrs and Groups stored in the handler, all other fields are
// shallow copied.
func (handler *Handler) Clone(opts ...Option) *Handler {
	h := &Handler{
		attrs:                append([]slog.Attr{}, handler.attrs...), // Must copy values to detach references from original.
		groups:               append([]string{}, handler.groups...),   // Must copy values to detach references from original.
		writer:               handler.writer,
		opts:                 handler.opts,
		functionFormatter:    handler.functionFormatter,
		timeFormatter:        handler.timeFormatter,
		pool:                 handler.pool,
		headerLevelFormatter: handler.headerLevelFormatter,
		addNewLine:           handler.addNewLine,
		packageNamePrefix:    handler.packageNamePrefix,
		messageFormatter:     handler.messageFormatter,
		sourceStyler:         handler.sourceStyler,
		timeStyler:           handler.timeStyler,
		headerLevelStyler:    handler.headerLevelStyler,
		messageStyler:        handler.messageStyler,
		pretty:               handler.pretty,
		fileLineFormatter:    handler.fileLineFormatter,
		prettyOption:         handler.prettyOption,
		prettyColor:          handler.prettyColor,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(h)
	}
	return h
}

func (handler *Handler) buildReplaceAttr() func(group []string, a slog.Attr) slog.Attr {
	return func(group []string, a slog.Attr) slog.Attr {
		if handler.pretty && len(group) == 0 {
			switch a.Key {
			// These keys are managed by the handler itself. The json serializer
			// must not handle them when pretty is enabled.
			case slog.TimeKey, slog.LevelKey, slog.MessageKey, slog.SourceKey:
				return slog.Attr{}
			}
		}
		if handler.opts != nil && handler.opts.ReplaceAttr != nil {
			a = handler.opts.ReplaceAttr(group, a)
		}
		return a
	}
}
