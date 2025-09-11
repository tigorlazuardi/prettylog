package prettylog

import (
	"bytes"
	"sync"
	"sync/atomic"
)

const defaultLimitedPoolSize = 16 * 1024 * 1024 // 16MB

// limitedPool is a wrapper around sync.Pool that limits the total memory
// used by the buffers in the pool to maxMemoryBufferStorage.
//
// This will cause big buffers to not take over the user's memory if the
// buffer is too big. The trade-off is that we will allocate more
// often if the limit is too low, but the upside is that we will not
// explode the user's memory.
type limitedPool struct {
	pool                     *sync.Pool
	maxMemoryBufferStorage   uint64
	currentMemoryBufferUsage *atomic.Uint64
}

func (l *limitedPool) Get() *bytes.Buffer {
	buf := l.pool.Get().(*bytes.Buffer)
	l.currentMemoryBufferUsage.Add(-uint64(buf.Cap()))
	return buf
}

func (l *limitedPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	bufCap := uint64(buf.Cap())
	if l.currentMemoryBufferUsage.Load()+bufCap > l.maxMemoryBufferStorage {
		// drop the buffer
		return
	}
	l.currentMemoryBufferUsage.Add(bufCap)
	l.pool.Put(buf)
}

func newLimitedPool(maxMemoryBufferStorage int) *limitedPool {
	if maxMemoryBufferStorage <= 0 {
		maxMemoryBufferStorage = defaultLimitedPoolSize
	}
	lp := &limitedPool{
		pool: &sync.Pool{
			New: func() any {
				return &bytes.Buffer{}
			},
		},
		maxMemoryBufferStorage:   uint64(maxMemoryBufferStorage),
		currentMemoryBufferUsage: &atomic.Uint64{},
	}
	return lp
}
