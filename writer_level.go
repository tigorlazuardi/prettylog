package prettylog

import (
	"bytes"
	"strings"

	"github.com/fatih/color"
)

var _ EntryWriter = (*LevelWriter)(nil)

var DefaultLevelWriter = NewLevelWriter()

type LevelWriter struct {
	key            Formatter
	keyColorizer   func(info RecordInfo, key string) string
	valueColorizer func(info RecordInfo, value string) string
	keyColoredLen  int
}

func NewLevelWriter() *LevelWriter {
	return &LevelWriter{
		key: Static(""),
		keyColorizer: func(info RecordInfo, key string) string {
			if key == "" {
				return ""
			}
			return getColoredText(info.Record.Level).Add(color.Bold).Sprint(key)
		},
		valueColorizer: func(info RecordInfo, value string) string {
			return getColoredBackground(info.Record.Level).Sprint("", value, "")
		},
	}
}

func (lw *LevelWriter) WithKey(f Formatter) *LevelWriter {
	lw.key = f
	lw.keyColoredLen = 0
	return lw
}

func (lw *LevelWriter) WithKeyColorizer(c func(info RecordInfo, key string) string) *LevelWriter {
	lw.keyColorizer = c
	lw.keyColoredLen = 0
	return lw
}

func (lw *LevelWriter) WithValueColorizer(c func(info RecordInfo, value string) string) *LevelWriter {
	lw.valueColorizer = c
	return lw
}

// Key returns the key to be used for this entry.
// If colored, this should also include the ANSI color codes
// as length of key.
func (le *LevelWriter) KeyLen(info RecordInfo) int {
	key := le.key(info)
	if !info.DisableColor {
		if le.keyColoredLen == 0 {
			le.keyColoredLen = len(le.keyColorizer(info, key))
		}
		return le.keyColoredLen
	}
	return len(key)
}

// Write writes the entry to the given buffer. Information
// about wether to format in pretty mode or not is given
// in [info]. To respect it or not is up to the implementation.
//
// Unless you are aware of what you are doing, avoid calling
// [buf.Reset], [buf.Truncate], [buf.Unread]s or similar methods that
// causes the internal buffer cursor to move backwards since they can mess
// with the log output.
//
// You can call [buf.Len] and [buf.Bytes] if needed to peek
// into the current buffer content.
func (le *LevelWriter) Write(info RecordInfo, buf *bytes.Buffer) {
	if info.DisableColor {
		if le.key != "" {
			if buf.Len() > 0 {
				buf.WriteByte('\n')
			}
		}
		buf.WriteString(le.key)
		buf.WriteString(strings.Repeat(" ", info.KeyFieldLength-len(le.key)+1))
		buf.WriteString(info.Record.Level.String())
		return
	}
	if le.key != "" {
		buf.WriteString(le.keyColorizer(info, le.key))
		buf.WriteString(": ")
	}
	buf.WriteString(le.valueColorizer(info, info.Record.Level.String()))
}
