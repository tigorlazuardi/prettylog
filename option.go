package prettylog

type Option func(h *Handler)

// WithPackageName supplies the Handler with the package name to be used.
func WithPackageName(name string) Option {
	return func(h *Handler) {
		h.packageName = name
	}
}
