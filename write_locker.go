package prettylog

import (
	"io"
	"sync"
)

type WriteLocker interface {
	io.Writer
	sync.Locker
}

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
