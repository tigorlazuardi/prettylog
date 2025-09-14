package prettylog

import (
	"bytes"
	"io"
	"sync"
	"testing"
)

func TestWrapWriteLockerAlreadyImplements(t *testing.T) {
	// Test with something that already implements WriteLocker
	buf := &bytes.Buffer{}
	original := WrapWriteLocker(buf)

	// Wrap it again
	wrapped := WrapWriteLocker(original)

	// Should return the same instance
	if wrapped != original {
		t.Error("wrapping an existing WriteLocker should return the same instance")
	}
}

func TestWrapWriteLockerNewWriter(t *testing.T) {
	buf := &bytes.Buffer{}

	wrapped := WrapWriteLocker(buf)

	// Should be a writeLocker
	wl, ok := wrapped.(*writeLocker)
	if !ok {
		t.Error("expected wrapped writer to be *writeLocker")
	}

	// Should wrap the original buffer
	if wl.Unwrap() != buf {
		t.Error("expected unwrapped writer to be the original buffer")
	}
}

func TestWriteLockerWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	wl := WrapWriteLocker(buf).(*writeLocker)

	testData := []byte("test data")
	n, err := wl.Write(testData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != len(testData) {
		t.Errorf("expected to write %d bytes, got %d", len(testData), n)
	}
	if buf.String() != "test data" {
		t.Errorf("expected buffer content 'test data', got '%s'", buf.String())
	}
}

func TestWriteLockerLocking(t *testing.T) {
	buf := &bytes.Buffer{}
	wl := WrapWriteLocker(buf)

	// Test that Lock and Unlock don't panic
	wl.Lock()
	wl.Unlock()

	// Test concurrent access
	const numGoroutines = 10
	const writesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			for range writesPerGoroutine {
				wl.Lock()
				_, err := wl.Write([]byte("x"))
				wl.Unlock()

				if err != nil {
					t.Errorf("goroutine %d: unexpected error: %v", id, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// All writes should have succeeded
	expectedLength := numGoroutines * writesPerGoroutine
	if buf.Len() != expectedLength {
		t.Errorf("expected buffer length %d, got %d", expectedLength, buf.Len())
	}
}

func TestWriteLockerUnwrap(t *testing.T) {
	buf := &bytes.Buffer{}
	wl := WrapWriteLocker(buf).(*writeLocker)

	unwrapped := wl.Unwrap()
	if unwrapped != buf {
		t.Error("Unwrap should return the original writer")
	}
}

// Test with an io.Writer that returns an error
type errorWriter struct {
	err error
}

func (ew *errorWriter) Write(p []byte) (n int, err error) {
	return 0, ew.err
}

func TestWriteLockerWriteError(t *testing.T) {
	expectedErr := io.ErrShortWrite
	ew := &errorWriter{err: expectedErr}
	wl := WrapWriteLocker(ew).(*writeLocker)

	n, err := wl.Write([]byte("test"))

	if n != 0 {
		t.Errorf("expected 0 bytes written, got %d", n)
	}
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

// Test that WriteLocker interface is correctly implemented
func TestWriteLockerInterface(t *testing.T) {
	buf := &bytes.Buffer{}

	// This should compile without issues if the interface is correctly implemented
	var _ WriteLocker = WrapWriteLocker(buf)
	var _ io.Writer = WrapWriteLocker(buf)
	var _ sync.Locker = WrapWriteLocker(buf)
}

// Custom WriteLocker implementation for testing
type customWriteLocker struct {
	writer io.Writer
	mu     sync.RWMutex
}

func (cwl *customWriteLocker) Write(p []byte) (n int, err error) {
	return cwl.writer.Write(p)
}

func (cwl *customWriteLocker) Lock() {
	cwl.mu.Lock()
}

func (cwl *customWriteLocker) Unlock() {
	cwl.mu.Unlock()
}

func TestWrapWriteLockerCustomImplementation(t *testing.T) {
	buf := &bytes.Buffer{}
	custom := &customWriteLocker{writer: buf}

	wrapped := WrapWriteLocker(custom)

	// Should return the same instance since it already implements WriteLocker
	if wrapped != custom {
		t.Error("expected to get the same custom WriteLocker instance")
	}
}

// Benchmark tests
func BenchmarkWriteLockerWrite(b *testing.B) {
	buf := &bytes.Buffer{}
	wl := WrapWriteLocker(buf)
	data := []byte("benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		wl.Write(data)
	}
}

func BenchmarkWriteLockerConcurrentWrite(b *testing.B) {
	buf := &bytes.Buffer{}
	wl := WrapWriteLocker(buf)
	data := []byte("x")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wl.Lock()
			wl.Write(data)
			wl.Unlock()
		}
	})
}

func BenchmarkWrapWriteLocker(b *testing.B) {
	buf := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapWriteLocker(buf)
	}
}

