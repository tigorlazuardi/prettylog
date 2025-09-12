package prettylog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/tidwall/pretty"
)

type Handler struct {
	attrs  []slog.Attr
	groups []string
	writer WriteLocker
	opts   *slog.HandlerOptions
	pretty bool

	// Keys
	timeKey     string
	fileKey     string
	functionKey string

	// Formatters
	functionFormatter   func(lvl slog.Level, packageNamePrefix string, frame runtime.Frame) string
	fileLineFormatter   func(lvl slog.Level, frame runtime.Frame) string
	timeFormatter       func(lvl slog.Level, t time.Time) string
	levelValueFormatter func(lvl slog.Level) string
	messageFormatter    func(lvl slog.Level, msg string) string

	// Stylers
	sourceStyler     func(lvl slog.Level, s string) string
	sourceKeyStyler  func(lvl slog.Level, s string) string
	timeStyler       func(lvl slog.Level, s string) string
	timeKeyStyler    func(lvl slog.Level, s string) string
	levelValueStyler func(lvl slog.Level, s string) string
	messageStyler    func(lvl slog.Level, s string) string

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

// Handle implements [slog.Handler] interface.
func (ha *Handler) Handle(ctx context.Context, rec slog.Record) error {
	buf := ha.pool.Get()
	defer ha.pool.Put(buf)

	frame, _ := runtime.CallersFrames([]uintptr{rec.PC}).Next()

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

	if lvlString := ha.formatLevel(rec.Level); len(lvlString) > 0 {
		buf.WriteString(lvlString)
	}
	if msg := ha.formatMessage(rec.Level, rec.Message); len(msg) > 0 {
		if buf.Len() > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(msg)
	}
	timeKey := ha.formatTimeKey(rec.Level, ha.timeKey)
	timeStr := ha.formatTimeValue(rec.Level, rec.Time)
	functionKey := ha.formatFunctionKey(rec.Level, ha.functionKey, frame)
	fieldLength := max(len(timeKey), len(functionKey))
	if len(timeStr) > 0 {
		if buf.Len() > 0 {
			buf.WriteRune('\n')
		}
		buf.WriteString(timeKey)
		buf.WriteString(strings.Repeat(" ", fieldLength-len(timeKey)+1))
		buf.WriteString(timeStr)
	}
	buf.WriteRune('\n')
	if len(timeKey) > 0 {
	}

	return nil
}

func (ha *Handler) formatLevel(lvl slog.Level) string {
	var levelText string
	if ha.levelValueFormatter != nil {
		levelText = ha.levelValueFormatter(lvl)
	} else {
		levelText = lvl.String()
	}
	if levelText == "" {
		return levelText
	}
	if ha.levelValueStyler != nil {
		levelText = ha.levelValueStyler(lvl, levelText)
	}
	return levelText
}

func (ha *Handler) formatMessage(lvl slog.Level, msg string) string {
	if ha.messageFormatter != nil {
		msg = ha.messageFormatter(lvl, msg)
	}
	if msg == "" {
		return msg
	}
	if ha.messageStyler != nil {
		msg = ha.messageStyler(lvl, msg)
	}
	return msg
}

func (ha *Handler) formatTimeKey(lvl slog.Level, s string) string {
	if ha.timeKeyStyler != nil {
		s = ha.timeKeyStyler(lvl, s)
	}
	return s
}

func (ha *Handler) formatFunctionKey(lvl slog.Level, s string, frame runtime.Frame) string {
	if frame.PC == 0 {
		// If PC is not found, we should skip writing Source field.
		return ""
	}
	if ha.functionFormatter != nil {
		s = ha.functionFormatter(lvl, ha.packageNamePrefix, frame)
	}
	return s
}

// func (ha *Handler) formatSource(lvl slog.Level, pc uintptr) string {
// 	if ha.formatSource != nil {
// 		return ha.format
// 	}
// 	return ShortFileLineFormatter(lvl, )
// }

func (ha *Handler) formatTimeValue(lvl slog.Level, t time.Time) string {
	var timeText string
	if ha.timeFormatter != nil {
		timeText = ha.timeFormatter(lvl, t)
	} else {
		timeText = t.Format(time.RFC3339Nano)
	}
	if timeText == "" {
		return timeText
	}
	if ha.timeStyler != nil {
		timeText = ha.timeStyler(lvl, timeText)
	}
	return timeText
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
		attrs:               append([]slog.Attr{}, handler.attrs...), // Must copy values to detach references from original.
		groups:              append([]string{}, handler.groups...),   // Must copy values to detach references from original.
		writer:              handler.writer,
		opts:                handler.opts,
		functionFormatter:   handler.functionFormatter,
		timeFormatter:       handler.timeFormatter,
		pool:                handler.pool,
		levelValueFormatter: handler.levelValueFormatter,
		addNewLine:          handler.addNewLine,
		packageNamePrefix:   handler.packageNamePrefix,
		messageFormatter:    handler.messageFormatter,
		sourceStyler:        handler.sourceStyler,
		timeStyler:          handler.timeStyler,
		levelValueStyler:    handler.levelValueStyler,
		messageStyler:       handler.messageStyler,
		pretty:              handler.pretty,
		fileLineFormatter:   handler.fileLineFormatter,
		prettyOption:        handler.prettyOption,
		prettyColor:         handler.prettyColor,
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
