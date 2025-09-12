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

	headerWriters []EntryWriter
	bodyWriter    EntryWriter

	// Formatters
	functionKeyFormatter   Formatter
	functionValueFormatter Formatter
	fileKeyFormatter       Formatter
	fileValueFormatter     Formatter
	timeKeyFormatter       Formatter
	timeValueFormatter     Formatter
	levelFormatter         Formatter
	messageFormatter       Formatter

	// Stylers
	sourceStyler        Styler
	functionKeyStyler   Styler
	functionValueStyler Styler
	timeValueStyler     Styler
	timeKeyStyler       Styler
	levelValueStyler    Styler
	messageStyler       Styler

	// JSON prettier options
	prettyOption *pretty.Options
	prettyColor  *pretty.Style

	packageName string
	addNewLine  bool
	pool        *limitedPool
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

	frame, _ := runtime.CallersFrames([]uintptr{rec.PC}).Next()
	info := RecordInfo{
		Record:      rec,
		PackageName: ha.packageName,
		Frame:       frame,
	}

	if lvlString := ha.formatLevel(info); len(lvlString) > 0 {
		buf.WriteString(lvlString)
	}
	if msg := ha.formatMessage(info); len(msg) > 0 {
		if buf.Len() > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(msg)
	}
	timeKey := ha.formatTimeKey(info)
	functionKey := ha.formatFunctionKey(info)
	fileKey := ha.formatFileKey(info)
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

func (ha *Handler) formatLevel(info RecordInfo) string {
	var levelText string
	if ha.levelFormatter != nil {
		levelText = ha.levelFormatter(info)
	}
	if levelText == "" {
		return levelText
	}
	if ha.levelValueStyler != nil {
		levelText = ha.levelValueStyler(info, levelText)
	}
	return levelText
}

func (ha *Handler) formatMessage(info RecordInfo) string {
	var msg string
	if ha.messageFormatter != nil {
		msg = ha.messageFormatter(info)
	}
	if msg == "" {
		return msg
	}
	if ha.messageStyler != nil {
		msg = ha.messageStyler(info, msg)
	}
	return msg
}

func (ha *Handler) formatTimeKey(info RecordInfo) string {
	var timeKey string
	if ha.timeKeyFormatter != nil {
		timeKey = ha.timeKeyFormatter(info)
	}
	if timeKey == "" {
		return timeKey
	}
	if ha.timeKeyStyler != nil {
		timeKey = ha.timeKeyStyler(info, timeKey)
	}
	return timeKey
}

func (ha *Handler) formatTimeValue(info RecordInfo) string {
	var timeText string
	if ha.timeValueFormatter != nil {
		timeText = ha.timeValueFormatter(info)
	}
	if timeText == "" {
		return timeText
	}
	if ha.timeValueStyler != nil {
		timeText = ha.timeValueStyler(info, timeText)
	}
	return timeText
}

func (ha *Handler) formatFunctionKey(info RecordInfo) string {
	var s string
	if ha.functionKeyFormatter != nil {
		s = ha.functionKeyFormatter(info)
	}
	if s == "" {
		return s
	}
	if ha.functionKeyStyler != nil {
		s = ha.functionKeyStyler(info, s)
	}
	return s
}

func (ha *Handler) formatFunctionValue(info RecordInfo) string {
	var s string
	if ha.functionValueFormatter != nil {
		s = ha.functionValueFormatter(info)
	}
	if s == "" {
		return s
	}
	if ha.functionValueStyler != nil {
		s = ha.functionValueStyler(info, s)
	}
	return s
}

func (ha *Handler) formatTimeValue(lvl slog.Level, t time.Time) string {
	var timeText string
	if ha.timeKeyFormatter != nil {
		timeText = ha.timeKeyFormatter(lvl, t)
	} else {
		timeText = t.Format(time.RFC3339Nano)
	}
	if timeText == "" {
		return timeText
	}
	if ha.timeValueStyler != nil {
		timeText = ha.timeValueStyler(lvl, timeText)
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
		attrs:                  append([]slog.Attr{}, handler.attrs...), // Must copy values to detach references from original.
		groups:                 append([]string{}, handler.groups...),   // Must copy values to detach references from original.
		writer:                 handler.writer,
		opts:                   handler.opts,
		functionValueFormatter: handler.functionValueFormatter,
		timeKeyFormatter:       handler.timeKeyFormatter,
		pool:                   handler.pool,
		levelFormatter:         handler.levelFormatter,
		addNewLine:             handler.addNewLine,
		packageName:            handler.packageName,
		messageFormatter:       handler.messageFormatter,
		sourceStyler:           handler.sourceStyler,
		timeValueStyler:        handler.timeValueStyler,
		levelValueStyler:       handler.levelValueStyler,
		messageStyler:          handler.messageStyler,
		pretty:                 handler.pretty,
		fileKeyFormatter:       handler.fileKeyFormatter,
		prettyOption:           handler.prettyOption,
		prettyColor:            handler.prettyColor,
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
