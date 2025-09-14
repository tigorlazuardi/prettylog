package prettylog

import (
	"context"
	"io"
	"log/slog"
	"runtime"
)

// Handler is a slog.Handler that provides cutomizeable, modular logging capabilities.
//
// Handler only implements the slog.Handler interface and pass the actual writing to
// a list of [EntryWriter]s. This allows for flexible and modular logging behavior.
type Handler struct {
	attrs  []slog.Attr
	groups []string
	writer WriteLocker
	opts   *slog.HandlerOptions
	color  bool

	writers []EntryWriter

	packageName string
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

	frame, _ := runtime.CallersFrames([]uintptr{rec.PC}).Next()
	info := RecordData{
		Context:        ctx,
		Record:         rec,
		PackageName:    ha.packageName,
		Frame:          frame,
		HandlerOptions: cloneHandlerOptions(ha.opts),
		Color:          ha.color,
		KeyFieldLength: 0,
		Buffer:         buf,
	}
	keyFieldLength := 0
	for _, w := range ha.writers {
		if l := w.KeyLen(info); l > keyFieldLength {
			keyFieldLength = l
		}
	}
	info.KeyFieldLength = keyFieldLength
	for _, w := range ha.writers {
		w.Write(info)
	}
	if buf.Len() == 0 {
		return nil
	}
	ha.writer.Lock()
	defer ha.writer.Unlock()
	_, err := io.Copy(ha.writer, buf)
	return err
}

// WithAttrs implements [slog.Handler] interface.
func (ha *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cloned := ha.Clone()
	cloned.attrs = append(cloned.attrs, attrs...)
	return cloned
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
// Attrs and Groups collected from [WithAttrs] and [WithGroup] are copied
// and detached from the original handler.
//
// Other fields (like output WriteLocker) are shallow copied and may still reference the same underlying
// values. If you want to change those fields, use the appropriate options
// when calling Clone.
func (handler *Handler) Clone(opts ...Option) *Handler {
	h := &Handler{
		attrs:       append([]slog.Attr{}, handler.attrs...), // Must copy values to detach references from original.
		groups:      append([]string{}, handler.groups...),   // Must copy values to detach references from original.
		writer:      handler.writer,
		opts:        handler.opts,
		pool:        handler.pool,
		packageName: handler.packageName,
		color:       handler.color,
		writers:     handler.writers,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(h)
	}
	return h
}

func cloneHandlerOptions(opts *slog.HandlerOptions) *slog.HandlerOptions {
	if opts == nil {
		return nil
	}
	return &slog.HandlerOptions{
		Level:       opts.Level,
		AddSource:   opts.AddSource,
		ReplaceAttr: opts.ReplaceAttr,
	}
}
