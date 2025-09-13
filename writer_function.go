package prettylog

import (
	"bytes"
	"strings"
)

var _ EntryWriter = (*FunctionWriter)(nil)

var DefaultFunctionWrtier = NewFunctionWriter()

type FunctionWriter struct {
	key            Formatter
	keyColorizer   Styler
	valueColorizer Styler

	keyColoredLen int
	short         bool
}

func NewFunctionWriter() *FunctionWriter {
	return &FunctionWriter{
		key:            Static("Function"),
		keyColorizer:   DefaultFunctionKeyStyler,
		valueColorizer: DefaultFunctionValueStyler,
		short:          true,
	}
}

func (fw *FunctionWriter) WithKey(f Formatter) *FunctionWriter {
	fw.key = f

func (fw *FunctionWriter) WithKeyColorizer(c Styler) *FunctionWriter {
	fw.keyColorizer = c
	fw.keyColoredLen = 0
	return fw
}

func (fw *FunctionWriter) WithValueColorizer(c Styler) *FunctionWriter {
	fw.valueColorizer = c
	return fw
}

// WithShort sets whether to use the short function name if possible.
//
// If true, the package name is stripped from the function name if it matches
// the package name given by [WithPackageName] option in [New] or [Handler.Clone].
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
