package prettylog

import (
	"io"
	"sync"
)

// WriteLocker combines io.Writer with sync.Locker for thread-safe writing.
// This interface ensures that write operations can be synchronized across
// multiple goroutines.
type WriteLocker interface {
	io.Writer
	sync.Locker
}

// WrapWriteLocker wraps an io.Writer to implement the WriteLocker interface.
// If the writer already implements WriteLocker, it returns the writer unchanged.
// Otherwise, it wraps it with a mutex for thread-safe writing.
func WrapWriteLocker(w io.Writer) WriteLocker {
	if wl, ok := w.(WriteLocker); ok {
		return wl
	}
	return &writeLocker{w: w}
}

type writeLocker struct {
	w  io.Writer
	mu sync.Mutex
}

func (wl *writeLocker) Write(p []byte) (n int, err error) {
	return wl.w.Write(p)
}

func (wl *writeLocker) Lock() {
	wl.mu.Lock()
}

func (wl *writeLocker) Unlock() {
	wl.mu.Unlock()
}

func (wl *writeLocker) Unwrap() io.Writer {
	return wl.w
}
