package prettylog

import (
	"io"
	"log/slog"
	"slices"
)

// Option is a function type for configuring Handler instances.
// Options are applied during Handler creation or cloning.
type Option func(h *Handler)

// WithPackageName supplies the Handler with the package name to be used.
// The package name is used by function formatters to trim package prefixes
// from function names, providing cleaner output.
func WithPackageName(name string) Option {
	return func(h *Handler) {
		h.packageName = name
	}
}

// WithOutput sets the output writer for the handler.
// The writer will be wrapped with [WriteLocker] if it doesn't already implement it.
func WithOutput(w io.Writer) Option {
	return func(h *Handler) {
		h.writer = WrapWriteLocker(w)
	}
}

// WithLevel sets the minimum log level for the handler.
func WithLevel(level slog.Level) Option {
	return func(h *Handler) {
		if h.opts == nil {
			h.opts = &slog.HandlerOptions{}
		}
		h.opts.Level = level
	}
}

// WithAddSource enables or disables source information (file, line, function) in log output.
func WithAddSource(addSource bool) Option {
	return func(h *Handler) {
		if h.opts == nil {
			h.opts = &slog.HandlerOptions{}
		}
		h.opts.AddSource = addSource
	}
}

// WithReplaceAttr sets a function to replace or modify attributes before logging.
func WithReplaceAttr(replaceAttr func(groups []string, a slog.Attr) slog.Attr) Option {
	return func(h *Handler) {
		if h.opts == nil {
			h.opts = &slog.HandlerOptions{}
		}
		h.opts.ReplaceAttr = replaceAttr
	}
}

// WithHandlerOptions sets the complete slog.HandlerOptions for the handler.
func WithHandlerOptions(opts *slog.HandlerOptions) Option {
	return func(h *Handler) {
		h.opts = opts
	}
}

// WithColor enables or disables colored output.
func WithColor(enabled bool) Option {
	return func(h *Handler) {
		h.color = enabled
	}
}

// WithWriters sets the complete list of entry writers for the handler.
// This replaces all existing writers.
func WithWriters(writers ...EntryWriter) Option {
	return func(h *Handler) {
		h.writers = writers
	}
}

// WithAdditionalWriters adds new entry writers to the existing list.
// The new writers are appended to the current writer list.
func WithAdditionalWriters(writers ...EntryWriter) Option {
	return func(h *Handler) {
		h.writers = append(h.writers, writers...)
	}
}

// WithoutWriters removes writers from the handler by comparing function pointers.
// This is useful for removing specific default writers while keeping others.
func WithoutWriters(writers ...EntryWriter) Option {
	return func(h *Handler) {
		for _, existing := range h.writers {
			h.writers = slices.DeleteFunc(h.writers, func(w EntryWriter) bool {
				return w == existing
			})
		}
	}
}

// ReplaceWriter replaces an existing writer with a new one by comparing function pointers.
// If the old writer is not found, the new writer is appended to the list.
func ReplaceWriter(oldWriter, newWriter EntryWriter) Option {
	return func(h *Handler) {
		for i, w := range h.writers {
			if w == oldWriter {
				h.writers[i] = newWriter
				return
			}
		}
		// If oldWriter not found, append newWriter
		h.writers = append(h.writers, newWriter)
	}
}

// WithPoolSize sets the maximum memory buffer storage for the internal buffer pool.
// This controls how much memory the handler will use for buffering log output.
func WithPoolSize(maxSize int) Option {
	return func(h *Handler) {
		h.pool = newLimitedPool(maxSize)
	}
}

// AddWritersBefore add writers before given tgt and unshift tgt up if found.
//
// If not found or tgt is nil, the writers are appended to the end.
func AddWritersBefore(tgt EntryWriter, writers ...EntryWriter) Option {
	return func(h *Handler) {
		if len(writers) == 0 {
			return
		}
		if tgt == nil {
			h.writers = append(h.writers, writers...)
			return
		}
		index := slices.Index(h.writers, tgt)
		if index == -1 {
			h.writers = append(h.writers, writers...)
			return
		}
		h.writers = slices.Insert(h.writers, index, writers...)
	}
}

// AddWritersAfter add writers after given tgt and add the writer
// after the tgt, unshifting the rest of the writers.
//
// If not found or tgt is nil, the writers are appended to the end.
func AddWritersAfter(tgt EntryWriter, writers ...EntryWriter) Option {
	return func(h *Handler) {
		if len(writers) == 0 {
			return
		}
		if tgt == nil {
			h.writers = append(h.writers, writers...)
			return
		}
		h.writers = slices.Insert(h.writers,
			slices.Index(h.writers, tgt)+1,
			writers...,
		)
	}
}
