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

func TestHandlerEnabled(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		testWith slog.Level
		expected bool
	}{
		{"info level allows info", slog.LevelInfo, slog.LevelInfo, true},
		{"info level allows warn", slog.LevelInfo, slog.LevelWarn, true},
		{"info level allows error", slog.LevelInfo, slog.LevelError, true},
		{"info level blocks debug", slog.LevelInfo, slog.LevelDebug, false},
		{"warn level allows warn", slog.LevelWarn, slog.LevelWarn, true},
		{"warn level allows error", slog.LevelWarn, slog.LevelError, true},
		{"warn level blocks info", slog.LevelWarn, slog.LevelInfo, false},
		{"warn level blocks debug", slog.LevelWarn, slog.LevelDebug, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(WithLevel(tt.level))
			result := handler.Enabled(context.Background(), tt.testWith)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHandlerEnabledNoLevel(t *testing.T) {
	handler := &Handler{
		opts: &slog.HandlerOptions{},
	}
	
	// Should default to slog.LevelInfo
	if !handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected info level to be enabled by default")
	}
	if handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected debug level to be disabled by default")
	}
}

func TestHandlerHandle(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := New(
		WithOutput(buf),
		WithWriters(&testWriter{name: "test"}),
	)

	pc, _, _, _ := runtime.Caller(0)
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", pc)
	err := handler.Handle(context.Background(), record)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test") {
		t.Errorf("expected output to contain 'test', got: %s", output)
	}
}

func TestHandlerHandleEmptyBuffer(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := New(
		WithOutput(buf),
		WithWriters(), // No writers means empty buffer
	)

	pc, _, _, _ := runtime.Caller(0)
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", pc)
	err := handler.Handle(context.Background(), record)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if buf.Len() > 0 {
		t.Errorf("expected empty output, got: %s", buf.String())
	}
}

func TestHandlerWithAttrs(t *testing.T) {
	handler := &Handler{
		attrs: []slog.Attr{slog.String("existing", "value")},
	}

	newAttrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	newHandler := handler.WithAttrs(newAttrs)

	// Original handler should be unchanged
	if len(handler.attrs) != 1 {
		t.Errorf("original handler attrs changed, expected 1, got %d", len(handler.attrs))
	}

	// New handler should have combined attributes
	h := newHandler.(*Handler)
	if len(h.attrs) != 3 {
		t.Errorf("new handler should have 3 attrs, got %d", len(h.attrs))
	}

	// Verify attributes are correct
	if h.attrs[0].Key != "existing" || h.attrs[0].Value.String() != "value" {
		t.Error("existing attribute not preserved correctly")
	}
	if h.attrs[1].Key != "key1" || h.attrs[1].Value.String() != "value1" {
		t.Error("first new attribute not added correctly")
	}
	if h.attrs[2].Key != "key2" || h.attrs[2].Value.String() != "42" {
		t.Error("second new attribute not added correctly")
	}
}

func TestHandlerWithGroup(t *testing.T) {
	handler := &Handler{
		groups: []string{"existing"},
	}

	newHandler := handler.WithGroup("newgroup")

	// Original handler should be unchanged
	if len(handler.groups) != 1 {
		t.Errorf("original handler groups changed, expected 1, got %d", len(handler.groups))
	}

	// New handler should have both groups
	h := newHandler.(*Handler)
	if len(h.groups) != 2 {
		t.Errorf("new handler should have 2 groups, got %d", len(h.groups))
	}

	if h.groups[0] != "existing" {
		t.Errorf("expected first group 'existing', got '%s'", h.groups[0])
	}
	if h.groups[1] != "newgroup" {
		t.Errorf("expected second group 'newgroup', got '%s'", h.groups[1])
	}
}

func TestHandlerWithGroupEmpty(t *testing.T) {
	handler := &Handler{}

	newHandler := handler.WithGroup("")

	// Empty group name should return the same handler
	if newHandler != handler {
		t.Error("empty group name should return the same handler")
	}
}

func TestHandlerClone(t *testing.T) {
	buf := &bytes.Buffer{}
	pool := newLimitedPool(1024)
	
	original := &Handler{
		attrs:       []slog.Attr{slog.String("key", "value")},
		groups:      []string{"group1"},
		writer:      WrapWriteLocker(buf),
		opts:        &slog.HandlerOptions{Level: slog.LevelDebug},
		color:       true,
		writers:     []EntryWriter{DefaultLevelWriter},
		packageName: "testpkg",
		pool:        pool,
	}

	cloned := original.Clone()

	// Verify shallow copies of shared fields
	if cloned.writer != original.writer {
		t.Error("writer should be shallow copied")
	}
	if cloned.opts != original.opts {
		t.Error("opts should be shallow copied")
	}
	if cloned.pool != original.pool {
		t.Error("pool should be shallow copied")
	}
	if cloned.packageName != original.packageName {
		t.Error("packageName should be copied")
	}
	if cloned.color != original.color {
		t.Error("color should be copied")
	}

	// Verify deep copies of slices
	if &cloned.attrs == &original.attrs {
		t.Error("attrs slice should be deep copied")
	}
	if len(cloned.attrs) != len(original.attrs) {
		t.Error("attrs content should be copied")
	}
	if cloned.attrs[0].Key != original.attrs[0].Key {
		t.Error("attrs values should be copied")
	}

	if &cloned.groups == &original.groups {
		t.Error("groups slice should be deep copied")
	}
	if len(cloned.groups) != len(original.groups) {
		t.Error("groups content should be copied")
	}
	if cloned.groups[0] != original.groups[0] {
		t.Error("groups values should be copied")
	}

	// Modify cloned slices and verify original is unchanged
	cloned.attrs = append(cloned.attrs, slog.String("new", "attr"))
	cloned.groups = append(cloned.groups, "newgroup")

	if len(original.attrs) != 1 {
		t.Error("original attrs should not be affected by clone modification")
	}
	if len(original.groups) != 1 {
		t.Error("original groups should not be affected by clone modification")
	}
}

func TestHandlerCloneWithOptions(t *testing.T) {
	original := &Handler{
		color: false,
		packageName: "oldpkg",
	}

	cloned := original.Clone(
		WithColor(true),
		WithPackageName("newpkg"),
	)

	if original.color != false {
		t.Error("original color should not be changed")
	}
	if original.packageName != "oldpkg" {
		t.Error("original packageName should not be changed")
	}

	if cloned.color != true {
		t.Error("cloned color should be updated")
	}
	if cloned.packageName != "newpkg" {
		t.Error("cloned packageName should be updated")
	}
}

func TestHandlerCloneWithNilOptions(t *testing.T) {
	original := &Handler{packageName: "test"}

	cloned := original.Clone(
		WithPackageName("changed"),
		nil, // Should be ignored
		WithColor(true),
	)

	if cloned.packageName != "changed" {
		t.Error("packageName should be changed despite nil option")
	}
	if cloned.color != true {
		t.Error("color should be set despite nil option")
	}
}

func TestCloneHandlerOptions(t *testing.T) {
	replaceFunc := func(groups []string, a slog.Attr) slog.Attr { return a }
	
	original := &slog.HandlerOptions{
		Level:       slog.LevelWarn,
		AddSource:   true,
		ReplaceAttr: replaceFunc,
	}

	cloned := cloneHandlerOptions(original)

	if cloned == original {
		t.Error("cloned options should be a different instance")
	}
	if cloned.Level != original.Level {
		t.Error("Level should be copied")
	}
	if cloned.AddSource != original.AddSource {
		t.Error("AddSource should be copied")
	}
	if cloned.ReplaceAttr == nil {
		t.Error("ReplaceAttr should be copied")
	}
}

func TestCloneHandlerOptionsNil(t *testing.T) {
	cloned := cloneHandlerOptions(nil)
	if cloned != nil {
		t.Error("cloning nil options should return nil")
	}
}

func TestHandlerKeyFieldLength(t *testing.T) {
	buf := &bytes.Buffer{}
	
	writer1 := &testWriterWithKeyLen{name: "short", keyLen: 5}
	writer2 := &testWriterWithKeyLen{name: "verylongname", keyLen: 12}
	writer3 := &testWriterWithKeyLen{name: "mid", keyLen: 8}

	handler := New(
		WithOutput(buf),
		WithWriters(writer1, writer2, writer3),
	)

	pc, _, _, _ := runtime.Caller(0)
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", pc)
	err := handler.Handle(context.Background(), record)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify that the longest key length (12) was used
	if writer1.lastKeyFieldLength != 12 {
		t.Errorf("expected keyFieldLength 12 for writer1, got %d", writer1.lastKeyFieldLength)
	}
	if writer2.lastKeyFieldLength != 12 {
		t.Errorf("expected keyFieldLength 12 for writer2, got %d", writer2.lastKeyFieldLength)
	}
	if writer3.lastKeyFieldLength != 12 {
		t.Errorf("expected keyFieldLength 12 for writer3, got %d", writer3.lastKeyFieldLength)
	}
}

// Test helper that tracks key field length
type testWriterWithKeyLen struct {
	name               string
	keyLen             int
	lastKeyFieldLength int
}

func (tw *testWriterWithKeyLen) KeyLen(info RecordData) int {
	return tw.keyLen
}

func (tw *testWriterWithKeyLen) Write(info RecordData) {
	tw.lastKeyFieldLength = info.KeyFieldLength
	info.Buffer.WriteString(tw.name)
}