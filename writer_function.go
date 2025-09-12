package prettylog

import (
	"bytes"
	"strings"
)

type FunctionWriter struct {
	key            string
	keyColorizer   func(info RecordInfo, key string) string
	valueColorizer func(info RecordInfo, value string) string

	keyColoredLen int
	short         bool
}

func NewFunctionWriter(key string) *FunctionWriter {
	return &FunctionWriter{
		key:            key,
		keyColorizer:   DefaultFunctionKeyStyler,
		valueColorizer: DefaultFunctionValueStyler,
		short:          true,
	}
}

func (fw *FunctionWriter) WithKeyColorizer(c func(info RecordInfo, key string) string) *FunctionWriter {
	fw.keyColorizer = c
	fw.keyColoredLen = 0
	return fw
}

func (fw *FunctionWriter) WithValueColorizer(c func(info RecordInfo, value string) string) *FunctionWriter {
	fw.valueColorizer = c
	return fw
}

// WithShort sets whether to use the short function name if possible.
func (fw *FunctionWriter) WithShort(short bool) *FunctionWriter {
	fw.short = short
	return fw
}

// Key returns the key to be used for this entry.
// If colored, this should also include the ANSI color codes
// as length of key.
func (fu FunctionWriter) KeyLen(info RecordInfo) int {
	if fu.key == "" {
		return 0
	}
	if !info.DisableColor {
		if fu.keyColoredLen == 0 {
			fu.keyColoredLen = len(fu.keyColorizer(info, fu.key))
		}
		return fu.keyColoredLen
	}
	return len(fu.key)
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
func (fu FunctionWriter) Write(info RecordInfo, buf *bytes.Buffer) {
	if info.Frame.Func == nil {
		return
	}
	if info.DisableColor {
		buf.WriteString(fu.key)
		buf.WriteString(strings.Repeat(" ", info.KeyFieldLength-len(fu.key)+1))
		if fu.short {
			buf.WriteString(strings.TrimPrefix(info.Frame.Function, info.PackageName))
		} else {
			buf.WriteString(info.Frame.Function)
		}
		return
	}
	buf.WriteString(fu.keyColorizer(info, fu.key))
	buf.WriteString(strings.Repeat(" ", info.KeyFieldLength-len(fu.keyColorizer(info, fu.key))+1))
	if fu.short {
		buf.WriteString(fu.valueColorizer(info, strings.TrimPrefix(info.Frame.Function, info.PackageName)))
	} else {
		buf.WriteString(fu.valueColorizer(info, info.Frame.Function))
	}
}
