package prettylog

import (
	"bytes"
	"context"
	"log/slog"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestEnsureBufferIsNotNil(t *testing.T) {
	writer := NewCommonWriter(func(info RecordData) string {
		if info.Buffer == nil {
			t.Fatal("info.Buffer should not be nil")
		}
		return "ok"
	})
	handler := New(
		WithWriters(writer),
	)
	handler.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0))
}

func TestUnimplementedEntryWriter(t *testing.T) {
	writer := &UnimplementedEntryWriter{}
	info := RecordData{Buffer: &bytes.Buffer{}}

	// Should return 0 for key length
	if writer.KeyLen(info) != 0 {
		t.Error("UnimplementedEntryWriter should return 0 for KeyLen")
	}

	// Should do nothing for Write
	writer.Write(info)
	if info.Buffer.Len() != 0 {
		t.Error("UnimplementedEntryWriter should not write anything")
	}
}

func TestDefaultPrefix(t *testing.T) {
	tests := []struct {
		name         string
		bufferLen    int
		key          string
		expectedText string
	}{
		{"empty buffer", 0, "testkey", ""},
		{"buffer with content, no key", 10, "", " "},
		{"buffer with content, with key", 10, "testkey", "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			if tt.bufferLen > 0 {
				buf.WriteString(strings.Repeat("x", tt.bufferLen))
			}

			info := RecordData{Buffer: buf}
			writer := &CommonWriter{Key: Static(tt.key)}

			result := DefaultPrefix(info, writer)
			if result != tt.expectedText {
				t.Errorf("expected %q, got %q", tt.expectedText, result)
			}
		})
	}
}

func TestNewCommonWriter(t *testing.T) {
	valuer := Static("test-value")
	writer := NewCommonWriter(valuer)

	if writer.Key == nil {
		t.Error("Key should be initialized")
	}
	// Test that valuer works by calling it
	info := RecordData{}
	if writer.Valuer(info) != "test-value" {
		t.Error("Valuer should be set to provided formatter")
	}
	if writer.Prefix == nil {
		t.Error("Prefix should be initialized")
	}
	if writer.KeyStyler == nil {
		t.Error("KeyStyler should be initialized")
	}
	if writer.ValueStyler == nil {
		t.Error("ValueStyler should be initialized")
	}

	// Test that default key returns empty string
	keyInfo := RecordData{}
	if writer.Key(keyInfo) != "" {
		t.Error("default key should return empty string")
	}
}

func TestCommonWriterWithKey(t *testing.T) {
	writer := NewCommonWriter(Static("value"))
	keyFormatter := Static("testkey")

	result := writer.WithKey(keyFormatter)

	// Should return the same instance for chaining
	if result != writer {
		t.Error("WithKey should return the same instance")
	}

	// Test that key was updated by calling it
	info := RecordData{}
	if writer.Key(info) != "testkey" {
		t.Error("Key should be updated")
	}
}

func TestCommonWriterWithValuer(t *testing.T) {
	writer := NewCommonWriter(Static("oldvalue"))
	newValuer := Static("newvalue")

	result := writer.WithValuer(newValuer)

	// Should return the same instance for chaining
	if result != writer {
		t.Error("WithValuer should return the same instance")
	}

	// Test that valuer was updated by calling it
	info := RecordData{}
	if writer.Valuer(info) != "newvalue" {
		t.Error("Valuer should be updated")
	}
}

func TestCommonWriterWithStaticKey(t *testing.T) {
	writer := NewCommonWriter(Static("value"))

	result := writer.WithStaticKey("statickey")

	// Should return the same instance for chaining
	if result != writer {
		t.Error("WithStaticKey should return the same instance")
	}

	// Test the key works
	info := RecordData{}
	if writer.Key(info) != "statickey" {
		t.Error("Key should return 'statickey'")
	}
}

func TestCommonWriterWithPrefix(t *testing.T) {
	writer := NewCommonWriter(Static("value"))
	customPrefix := func(info RecordData, this *CommonWriter) string {
		return "custom-prefix"
	}

	result := writer.WithPrefix(customPrefix)

	// Should return the same instance for chaining
	if result != writer {
		t.Error("WithPrefix should return the same instance")
	}

	// Test the prefix works
	info := RecordData{Buffer: &bytes.Buffer{}}
	prefix := writer.Prefix(info, writer)
	if prefix != "custom-prefix" {
		t.Error("Prefix should return 'custom-prefix'")
	}
}

func TestCommonWriterWithKeyColorizer(t *testing.T) {
	writer := NewCommonWriter(Static("value"))
	customStyler := func(info RecordData, text string) string {
		return "styled-" + text
	}

	result := writer.WithKeyColorizer(customStyler)

	// Should return the same instance for chaining
	if result != writer {
		t.Error("WithKeyColorizer should return the same instance")
	}

	// Styler should be updated
	if writer.KeyStyler == nil {
		t.Error("KeyStyler should be set")
	}

	// Test the styler works
	info := RecordData{}
	styled := writer.KeyStyler(info, "test")
	if styled != "styled-test" {
		t.Error("KeyStyler should return 'styled-test'")
	}
}

func TestCommonWriterWithValueColorizer(t *testing.T) {
	writer := NewCommonWriter(Static("value"))
	customStyler := func(info RecordData, text string) string {
		return "styled-" + text
	}

	result := writer.WithValueColorizer(customStyler)

	// Should return the same instance for chaining
	if result != writer {
		t.Error("WithValueColorizer should return the same instance")
	}

	// Styler should be updated
	if writer.ValueStyler == nil {
		t.Error("ValueStyler should be set")
	}

	// Test the styler works
	info := RecordData{}
	styled := writer.ValueStyler(info, "test")
	if styled != "styled-test" {
		t.Error("ValueStyler should return 'styled-test'")
	}
}

func TestCommonWriterKeyLen(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		color    bool
		expected int
	}{
		{"empty key", "", false, 0},
		{"empty key with color", "", true, 0},
		{"key without color", "testkey", false, 8}, // length of "testkey" after styling
		{"key with color", "testkey", true, 7},     // original length without styling
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewCommonWriter(Static("value")).WithStaticKey(tt.key)
			writer.KeyStyler = func(info RecordData, text string) string {
				return text + "!" // Add one character for styling
			}

			info := RecordData{Color: tt.color}
			result := writer.KeyLen(info)

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCommonWriterWrite(t *testing.T) {
	tests := []struct {
		name             string
		key              string
		value            string
		color            bool
		keyFieldLength   int
		bufferContent    string
		expectedContains []string
	}{
		{
			name:             "basic write without color",
			key:              "level",
			value:            "INFO",
			color:            false,
			keyFieldLength:   10,
			bufferContent:    "",
			expectedContains: []string{"level", "INFO"},
		},
		{
			name:             "basic write with color",
			key:              "level",
			value:            "INFO",
			color:            true,
			keyFieldLength:   10,
			bufferContent:    "",
			expectedContains: []string{"level", "INFO"},
		},
		{
			name:             "write with existing buffer content",
			key:              "msg",
			value:            "hello",
			color:            false,
			keyFieldLength:   5,
			bufferContent:    "existing ",
			expectedContains: []string{"existing", "\n", "msg", "hello"},
		},
		{
			name:             "write with empty key",
			key:              "",
			value:            "message",
			color:            false,
			keyFieldLength:   0,
			bufferContent:    "existing ",
			expectedContains: []string{"existing", " ", "message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			buf.WriteString(tt.bufferContent)

			writer := NewCommonWriter(Static(tt.value)).WithStaticKey(tt.key)
			writer.KeyStyler = func(info RecordData, text string) string {
				return "k:" + text
			}
			writer.ValueStyler = func(info RecordData, text string) string {
				return "v:" + text
			}

			info := RecordData{
				Buffer:         buf,
				Color:          tt.color,
				KeyFieldLength: tt.keyFieldLength,
			}

			writer.Write(info)
			result := buf.String()

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

func TestCommonWriterSpacing(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewCommonWriter(Static("INFO")).WithStaticKey("level")
	writer.KeyStyler = PlainStyler
	writer.ValueStyler = PlainStyler

	info := RecordData{
		Buffer:         buf,
		Color:          false,
		KeyFieldLength: 10, // "level" is 5 chars, so we expect 5 spaces + 1 separator
	}

	writer.Write(info)
	result := buf.String()

	// Should contain "level" followed by spaces to reach KeyFieldLength + 1, then "INFO"
	expectedSpaces := info.KeyFieldLength - len("level") + 1 // 10 - 5 + 1 = 6 spaces
	expectedResult := "level" + strings.Repeat(" ", expectedSpaces) + "INFO"

	if result != expectedResult {
		t.Errorf("expected spacing: %q, got: %q", expectedResult, result)
	}
}

func TestCommonWriterChaining(t *testing.T) {
	// Test that all methods can be chained
	writer := NewCommonWriter(Static("value")).
		WithStaticKey("key").
		WithPrefix(func(info RecordData, this *CommonWriter) string { return "prefix" }).
		WithKeyColorizer(func(info RecordData, text string) string { return "k:" + text }).
		WithValueColorizer(func(info RecordData, text string) string { return "v:" + text })

	// Test that all fields were set correctly
	info := RecordData{Buffer: &bytes.Buffer{}}

	if writer.Key(info) != "key" {
		t.Error("Key should be set")
	}
	if writer.Valuer(info) != "value" {
		t.Error("Valuer should be set")
	}
	if writer.Prefix(info, writer) != "prefix" {
		t.Error("Prefix should be set")
	}
	if writer.KeyStyler(info, "test") != "k:test" {
		t.Error("KeyStyler should be set")
	}
	if writer.ValueStyler(info, "test") != "v:test" {
		t.Error("ValueStyler should be set")
	}
}

// Integration test with real slog.Record
func TestCommonWriterIntegration(t *testing.T) {
	buf := &bytes.Buffer{}

	// Create a record like the handler would
	pc, _, _, _ := runtime.Caller(0)
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", pc)
	frame, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()

	info := RecordData{
		Context:        context.Background(),
		Record:         record,
		PackageName:    "github.com/tigorlazuardi/prettylog",
		Frame:          frame,
		Color:          true,
		KeyFieldLength: 8,
		Buffer:         buf,
	}

	// Create a writer that extracts the level
	writer := NewCommonWriter(DefaultLevelFormatter).WithStaticKey("level")
	writer.Write(info)

	result := buf.String()

	if !strings.Contains(result, "level") {
		t.Error("result should contain 'level'")
	}
	if !strings.Contains(result, "INFO") {
		t.Error("result should contain 'INFO'")
	}
}

// Benchmark tests
func BenchmarkCommonWriterWrite(b *testing.B) {
	buf := &bytes.Buffer{}
	writer := NewCommonWriter(DefaultLevelFormatter).WithStaticKey("level")

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	info := RecordData{
		Record:         record,
		Buffer:         buf,
		Color:          true,
		KeyFieldLength: 8,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		writer.Write(info)
	}
}

func BenchmarkCommonWriterKeyLen(b *testing.B) {
	writer := NewCommonWriter(Static("value")).WithStaticKey("level")
	info := RecordData{Color: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.KeyLen(info)
	}
}

