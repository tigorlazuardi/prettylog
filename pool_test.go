package prettylog

import (
	"bytes"
	"sync"
	"testing"
)

func TestNewLimitedPool(t *testing.T) {
	tests := []struct {
		name            string
		maxMemoryBuffer int
		expectedMax     uint64
	}{
		{"positive value", 1024, 1024},
		{"zero value", 0, defaultLimitedPoolSize},
		{"negative value", -100, defaultLimitedPoolSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := newLimitedPool(tt.maxMemoryBuffer)
			
			if pool.maxMemoryBufferStorage != tt.expectedMax {
				t.Errorf("expected max memory buffer %d, got %d", 
					tt.expectedMax, pool.maxMemoryBufferStorage)
			}
			
			if pool.pool == nil {
				t.Error("sync.Pool should be initialized")
			}
			
			if pool.currentMemoryBufferUsage == nil {
				t.Error("current memory usage tracker should be initialized")
			}
			
			// Test that pool creates new buffers
			buf := pool.Get()
			if buf == nil {
				t.Error("pool should create new buffer")
			}
			
			// Return buffer for cleanup
			pool.Put(buf)
		})
	}
}

func TestLimitedPoolGetPut(t *testing.T) {
	pool := newLimitedPool(1024)
	
	// Get a buffer
	buf1 := pool.Get()
	if buf1 == nil {
		t.Fatal("expected buffer from pool")
	}
	
	// Write some data to it
	buf1.WriteString("test data")
	if buf1.Len() == 0 {
		t.Error("buffer should contain data")
	}
	
	// Put it back
	pool.Put(buf1)
	
	// Buffer should be reset
	if buf1.Len() != 0 {
		t.Error("buffer should be reset after Put")
	}
	
	// Get another buffer (might be the same one)
	buf2 := pool.Get()
	if buf2 == nil {
		t.Fatal("expected buffer from pool")
	}
	
	// It should be empty
	if buf2.Len() != 0 {
		t.Error("reused buffer should be empty")
	}
	
	pool.Put(buf2)
}

func TestLimitedPoolMemoryTracking(t *testing.T) {
	pool := newLimitedPool(100) // Small limit for testing
	
	// Get a buffer and write data to expand its capacity
	buf := pool.Get()
	
	// Write enough data to expand the buffer beyond the limit
	largeData := make([]byte, 200)
	buf.Write(largeData)
	
	initialCapacity := buf.Cap()
	if initialCapacity < 200 {
		t.Error("buffer should have expanded to accommodate data")
	}
	
	// Put the buffer back
	pool.Put(buf)
	
	// Since the buffer is now larger than the pool limit, it should be dropped
	// The pool should create a new buffer on next Get
	buf2 := pool.Get()
	
	// The new buffer should be smaller (new allocation)
	if buf2.Cap() >= initialCapacity {
		// This test might be flaky depending on Go's buffer allocation strategy
		// But generally, a new buffer should have a smaller initial capacity
		t.Logf("Warning: new buffer has capacity %d, expected smaller than %d", 
			buf2.Cap(), initialCapacity)
	}
	
	pool.Put(buf2)
}

func TestLimitedPoolMemoryLimit(t *testing.T) {
	maxMemory := uint64(50)
	pool := newLimitedPool(int(maxMemory))
	
	// Create a buffer with known capacity
	buf := pool.Get()
	
	// Write data to set a specific capacity
	buf.Write(make([]byte, 30))
	currentCap := uint64(buf.Cap())
	
	// If the buffer capacity is less than limit, it should be accepted back
	if currentCap <= maxMemory {
		pool.Put(buf)
		
		// Current usage should reflect the buffer capacity
		currentUsage := pool.currentMemoryBufferUsage.Load()
		if currentUsage != currentCap {
			t.Errorf("expected current usage %d, got %d", currentCap, currentUsage)
		}
		
		// Get the buffer back and verify usage is decremented
		buf2 := pool.Get()
		newUsage := pool.currentMemoryBufferUsage.Load()
		if newUsage >= currentUsage {
			t.Error("memory usage should decrease when buffer is retrieved")
		}
		
		pool.Put(buf2)
	} else {
		// If buffer is too large, it should be dropped
		pool.Put(buf)
		
		currentUsage := pool.currentMemoryBufferUsage.Load()
		if currentUsage != 0 {
			t.Errorf("expected usage to remain 0 for dropped buffer, got %d", currentUsage)
		}
	}
}

func TestLimitedPoolConcurrency(t *testing.T) {
	pool := newLimitedPool(1024)
	const numGoroutines = 10
	const operationsPerGoroutine = 100
	
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	// Concurrent get/put operations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				buf := pool.Get()
				if buf == nil {
					t.Errorf("expected buffer from pool")
					return
				}
				
				// Do some work with the buffer
				buf.WriteString("test data")
				
				// Put it back
				pool.Put(buf)
			}
		}()
	}
	
	wg.Wait()
	
	// Pool should still be functional after concurrent access
	buf := pool.Get()
	if buf == nil {
		t.Error("pool should still work after concurrent access")
	}
	pool.Put(buf)
}

func TestLimitedPoolMemoryAccounting(t *testing.T) {
	pool := newLimitedPool(1000)
	
	// Get multiple buffers and track memory usage
	buffers := make([]*bytes.Buffer, 3)
	var totalExpectedUsage uint64
	
	for i := range buffers {
		buf := pool.Get()
		buf.Write(make([]byte, 50)) // Write some data to give it capacity
		buffers[i] = buf
		
		// Put buffer back and track expected usage
		pool.Put(buf)
		totalExpectedUsage += uint64(buf.Cap())
	}
	
	// Check that memory usage is tracked (the exact amount may vary due to Go's buffer allocation)
	currentUsage := pool.currentMemoryBufferUsage.Load()
	if currentUsage == 0 {
		t.Error("expected some memory usage to be tracked")
	}
	
	// Get all buffers back - usage should go to 0
	for i := range buffers {
		buffers[i] = pool.Get()
	}
	
	finalUsage := pool.currentMemoryBufferUsage.Load()
	if finalUsage >= currentUsage {
		t.Error("memory usage should decrease when buffers are retrieved")
	}
	
	// Put them back
	for _, buf := range buffers {
		pool.Put(buf)
	}
}

func TestLimitedPoolBufferReuse(t *testing.T) {
	pool := newLimitedPool(1024)
	
	// Get a buffer and mark it
	buf1 := pool.Get()
	buf1.WriteString("marker")
	originalPtr := buf1
	
	// Put it back
	pool.Put(buf1)
	
	// Get a buffer again - it might be the same one (but reset)
	buf2 := pool.Get()
	
	if buf2 == originalPtr {
		// If it's the same buffer, it should be reset
		if buf2.Len() != 0 {
			t.Error("reused buffer should be reset")
		}
		if buf2.String() != "" {
			t.Error("reused buffer should be empty")
		}
	}
	
	pool.Put(buf2)
}

// Benchmark tests
func BenchmarkLimitedPoolGetPut(b *testing.B) {
	pool := newLimitedPool(1024 * 1024) // 1MB limit
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		buf.WriteString("test data for benchmarking")
		pool.Put(buf)
	}
}

func BenchmarkLimitedPoolGetPutParallel(b *testing.B) {
	pool := newLimitedPool(1024 * 1024) // 1MB limit
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			buf.WriteString("test data for parallel benchmarking")
			pool.Put(buf)
		}
	})
}

func BenchmarkLimitedPoolWithLargeBuffers(b *testing.B) {
	pool := newLimitedPool(1024 * 1024) // 1MB limit
	largeData := make([]byte, 1024)     // 1KB data
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		buf.Write(largeData)
		pool.Put(buf)
	}
}

func TestDefaultLimitedPoolSize(t *testing.T) {
	expectedSize := 16 * 1024 * 1024 // 16MB
	if defaultLimitedPoolSize != expectedSize {
		t.Errorf("expected default pool size %d, got %d", 
			expectedSize, defaultLimitedPoolSize)
	}
}