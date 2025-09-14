package prettylog

import (
	"log/slog"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDefaultLevelFormatter(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		expected string
	}{
		{"debug level", slog.LevelDebug, "DEBUG"},
		{"info level", slog.LevelInfo, "INFO"},
		{"warn level", slog.LevelWarn, "WARN"},
		{"error level", slog.LevelError, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := slog.NewRecord(time.Now(), tt.level, "test message", 0)
			info := RecordData{Record: record}
			
			result := DefaultLevelFormatter(info)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTrimPrefixFunctionFormatter(t *testing.T) {
	// Get current function info for testing
	pc, _, _, _ := runtime.Caller(0)
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	
	tests := []struct {
		name        string
		packageName string
		frame       runtime.Frame
		expected    string
	}{
		{
			"with package prefix",
			"github.com/tigorlazuardi/prettylog",
			frame,
			strings.TrimPrefix(frame.Function, "github.com/tigorlazuardi/prettylog"),
		},
		{
			"without package prefix",
			"different.package",
			frame,
			frame.Function, // Should return full function name since prefix doesn't match
		},
		{
			"nil function",
			"test.package",
			runtime.Frame{Func: nil},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := RecordData{
				PackageName: tt.packageName,
				Frame:       tt.frame,
			}
			
			result := TrimPrefixFunctionFormatter(info)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFullFunctionFormatter(t *testing.T) {
	// Get current function info for testing
	pc, _, _, _ := runtime.Caller(0)
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	
	tests := []struct {
		name     string
		frame    runtime.Frame
		expected string
	}{
		{
			"with function info",
			frame,
			frame.Function,
		},
		{
			"nil function",
			runtime.Frame{Func: nil},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := RecordData{Frame: tt.frame}
			
			result := FullFunctionFormatter(info)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestShortFileLineFormatter(t *testing.T) {
	// Get current function info for testing
	pc, file, line, _ := runtime.Caller(0)
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	
	// Test with current working directory
	wd, _ := os.Getwd()
	
	tests := []struct {
		name     string
		frame    runtime.Frame
		expected func() string
	}{
		{
			"with working directory prefix",
			runtime.Frame{
				Func: frame.Func,
				File: file,
				Line: line,
			},
			func() string {
				s, trimmed := strings.CutPrefix(file, wd)
				if trimmed {
					return strings.TrimPrefix(s, "/") + ":" + "42"
				}
				return file + ":" + "42"
			},
		},
		{
			"without working directory prefix",
			runtime.Frame{
				Func: frame.Func,
				File: "/some/other/path/file.go",
				Line: 123,
			},
			func() string {
				return "/some/other/path/file.go:42"
			},
		},
		{
			"nil function",
			runtime.Frame{Func: nil},
			func() string {
				return ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override line number for consistent testing
			testFrame := tt.frame
			if testFrame.Func != nil {
				testFrame.Line = 42
			}
			
			info := RecordData{Frame: testFrame}
			
			result := ShortFileLineFormatter(info)
			expected := tt.expected()
			
			if result != expected {
				t.Errorf("expected %s, got %s", expected, result)
			}
		})
	}
}

func TestFullFileLineFormatter(t *testing.T) {
	// Get current function info for testing
	pc, _, _, _ := runtime.Caller(0)
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	
	tests := []struct {
		name     string
		frame    runtime.Frame
		expected string
	}{
		{
			"with function info",
			runtime.Frame{
				Func: frame.Func,
				File: "/full/path/to/file.go",
				Line: 123,
			},
			"/full/path/to/file.go:123",
		},
		{
			"nil function",
			runtime.Frame{Func: nil},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := RecordData{Frame: tt.frame}
			
			result := FullFileLineFormatter(info)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTimeOnlyFormatter(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC)
	record := slog.NewRecord(testTime, slog.LevelInfo, "test message", 0)
	
	info := RecordData{Record: record}
	
	result := TimeOnlyFormatter(info)
	expected := "14:30:45"
	
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestRFC3339TimeFormatter(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC)
	record := slog.NewRecord(testTime, slog.LevelInfo, "test message", 0)
	
	info := RecordData{Record: record}
	
	result := RFC3339TimeFormatter(info)
	expected := "2024-03-15T14:30:45Z"
	
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestDefaultMessageFormatter(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"simple message", "hello world", "hello world"},
		{"empty message", "", ""},
		{"message with special chars", "test\nmessage\twith\rspecial", "test\nmessage\twith\rspecial"},
		{"unicode message", "héllo wørld 🌍", "héllo wørld 🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := slog.NewRecord(time.Now(), slog.LevelInfo, tt.message, 0)
			info := RecordData{Record: record}
			
			result := DefaultMessageFormatter(info)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStatic(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"simple string", "test", "test"},
		{"empty string", "", ""},
		{"special chars", "test\nvalue", "test\nvalue"},
		{"unicode", "tëst 🔥", "tëst 🔥"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := Static(tt.key)
			
			// The formatter should return the same value regardless of RecordData
			info1 := RecordData{}
			info2 := RecordData{Record: slog.NewRecord(time.Now(), slog.LevelError, "different message", 0)}
			
			result1 := formatter(info1)
			result2 := formatter(info2)
			
			if result1 != tt.expected {
				t.Errorf("expected %s, got %s for first call", tt.expected, result1)
			}
			if result2 != tt.expected {
				t.Errorf("expected %s, got %s for second call", tt.expected, result2)
			}
			if result1 != result2 {
				t.Error("Static formatter should return the same value regardless of input")
			}
		})
	}
}

// Benchmark tests for formatters
func BenchmarkDefaultLevelFormatter(b *testing.B) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	info := RecordData{Record: record}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultLevelFormatter(info)
	}
}

func BenchmarkTrimPrefixFunctionFormatter(b *testing.B) {
	pc, _, _, _ := runtime.Caller(0)
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	info := RecordData{
		PackageName: "github.com/tigorlazuardi/prettylog",
		Frame:       frame,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TrimPrefixFunctionFormatter(info)
	}
}

func BenchmarkShortFileLineFormatter(b *testing.B) {
	pc, _, _, _ := runtime.Caller(0)
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	info := RecordData{Frame: frame}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShortFileLineFormatter(info)
	}
}

func BenchmarkTimeOnlyFormatter(b *testing.B) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	info := RecordData{Record: record}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeOnlyFormatter(info)
	}
}

func BenchmarkStatic(b *testing.B) {
	formatter := Static("test")
	info := RecordData{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter(info)
	}
}