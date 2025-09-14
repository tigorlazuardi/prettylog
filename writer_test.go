package prettylog

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/fatih/color"
)

func TestEnsureWriteDoesNotPanicWithNegativeRepeats(t *testing.T) {
	color.NoColor = false
	buf := &bytes.Buffer{}
	cw := NewCommonWriter(func(info RecordData) string {
		return info.Record.Message
	}).WithStaticKey("key")
	handler := New(
		WithOutput(buf),
		WithWriters(cw, DefaultPrettyJSONWriter),
		WithColor(true),
	)

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)

	handler.Handle(context.Background(), record)
}
