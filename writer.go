package prettylog

import (
	"bytes"
	"strings"
)

type EntryWriter interface {
	// Key returns the key to be used for this entry.
	//
	// If colored, this should also include the ANSI color codes
	// as length of key
	KeyLen(info RecordInfo) int

	// Write writes the entry to the given buffer to [RecordInfo.Buffer] inside info.
	//
	// There are information about whether to use color format or not inside info by
	// referencing [RecordInfo.Color].
	//
	// To respect it or not is up to the implementation.
	Write(info RecordInfo)
}

// UnimplementedEntryWriter is an [EntryWriter] that does nothing.
// It can be embedded to have forward compatible implementations.
type UnimplementedEntryWriter struct{}

func (UnimplementedEntryWriter) KeyLen(info RecordInfo) int               { return 0 }
func (UnimplementedEntryWriter) Write(info RecordInfo, buf *bytes.Buffer) {}

func DefaultPrefix(info RecordInfo, this *CommonWriter) string {
	if info.Buffer.Len() == 0 {
		return ""
	}
	if key := this.Key(info); key != "" {
		return "\n"
	}
	return " "
}

type PrefixFunc func(info RecordInfo, this *CommonWriter) string

// CommonWriter is a common implementation of [EntryWriter] that writes a key-value pair.
//
// Please use [NewCommonWriter] to create a new instance so fields are properly initialized
// then modify it with the provided methods like [CommonWriter.WithKey], [CommonWriter.WithStaticKey],
type CommonWriter struct {
	Key         Formatter
	Valuer      Formatter
	Prefix      PrefixFunc
	KeyStyler   Styler
	ValueStyler Styler
}

// NewCommonWriter creates a new CommonWriter with the given value formatter to
// extract the value from the RecordInfo.
//
// The key is empty by default, and can be set by [WithKey] or [WithStaticKey].
func NewCommonWriter(valuer Formatter) *CommonWriter {
	return &CommonWriter{
		Key:         Static(""),
		Valuer:      valuer,
		Prefix:      DefaultPrefix,
		KeyStyler:   BoldColoredStyler,
		ValueStyler: PlainStyler,
	}
}

func (cw *CommonWriter) WithKey(f Formatter) *CommonWriter {
	cw.Key = f
	return cw
}

func (cw *CommonWriter) WithValuer(f Formatter) *CommonWriter {
	cw.Valuer = f
	return cw
}

func (cw *CommonWriter) WithStaticKey(key string) *CommonWriter {
	cw.Key = Static(key)
	return cw
}

func (cw *CommonWriter) WithPrefix(f PrefixFunc) *CommonWriter {
	cw.Prefix = f
	return cw
}

func (cw *CommonWriter) WithKeyColorizer(c Styler) *CommonWriter {
	cw.KeyStyler = c
	return cw
}

func (cw *CommonWriter) WithValueColorizer(c Styler) *CommonWriter {
	cw.ValueStyler = c
	return cw
}

func (cw *CommonWriter) KeyLen(info RecordInfo) int {
	key := cw.Key(info)
	if key == "" {
		return 0
	}
	if !info.Color {
		return len(cw.KeyStyler(info, key))
	}
	return len(key)
}

func (cw *CommonWriter) Write(info RecordInfo) {
	info.Buffer.WriteString(cw.Prefix(info, cw))
	key := cw.Key(info)
	value := cw.Valuer(info)
	if info.Color {
		if key != "" {
			key = cw.KeyStyler(info, key)
		}
		value = cw.ValueStyler(info, value)
	}
	info.Buffer.WriteString(key)
	info.Buffer.WriteString(strings.Repeat(" ", info.KeyFieldLength-len(key)+1))
	info.Buffer.WriteString(value)
}
