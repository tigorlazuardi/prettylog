package prettylog

import (
	"strings"
)

// EntryWriter is the interface for components that write parts of log entries.
// Each EntryWriter is responsible for formatting and outputting a specific
// component of the log entry (e.g., level, message, timestamp).
type EntryWriter interface {
	// Key returns the key to be used for this entry.
	//
	// If colored, this should also include the ANSI color codes
	// as length of key
	KeyLen(info RecordData) int

	// Write writes the entry to the given buffer to [RecordData.Buffer] inside info.
	//
	// There are information about whether to use color format or not inside info by
	// referencing [RecordData.Color].
	//
	// To respect it or not is up to the implementation.
	Write(info RecordData)
}

var _ EntryWriter = (*UnimplementedEntryWriter)(nil)

// UnimplementedEntryWriter is an EntryWriter that does nothing.
// It can be embedded to have forward compatible implementations.
type UnimplementedEntryWriter struct{}

func (UnimplementedEntryWriter) KeyLen(info RecordData) int { return 0 }
func (UnimplementedEntryWriter) Write(info RecordData)      {}

// DefaultPrefix returns the default prefix between log entry components.
// It returns:
//   - Empty string if buffer is empty
//   - Newline if the writer has no key
//   - Space for subsequent entries
func DefaultPrefix(info RecordData, this *CommonWriter) string {
	if info.Buffer.Len() == 0 {
		return ""
	}
	if key := this.Key(info); key != "" {
		return "\n"
	}
	return " "
}

// PrefixFunc is a function type for generating prefixes between log entry components.
type PrefixFunc func(info RecordData, this *CommonWriter) string

// CommonWriter is a common implementation of EntryWriter that writes a key-value pair.
// It provides configurable formatting for both keys and values, along with styling options.
//
// Use NewCommonWriter to create a new instance with proper field initialization,
// then customize it using the provided methods.
type CommonWriter struct {
	Key         Formatter
	Valuer      Formatter
	Prefix      PrefixFunc
	KeyStyler   Styler
	ValueStyler Styler
}

// NewCommonWriter creates a new CommonWriter with the given value formatter.
// The value formatter extracts the actual value from the RecordData.
//
// The key is empty by default and can be set using WithKey or WithStaticKey methods.
// Default styling uses bold colored keys and plain values.
func NewCommonWriter(valuer Formatter) *CommonWriter {
	return &CommonWriter{
		Key:         Static(""),
		Valuer:      valuer,
		Prefix:      DefaultPrefix,
		KeyStyler:   BoldColoredStyler,
		ValueStyler: PlainStyler,
	}
}

// WithKey sets the key formatter for this CommonWriter.
// The key formatter determines what key text is displayed for this log component.
func (cw *CommonWriter) WithKey(f Formatter) *CommonWriter {
	cw.Key = f
	return cw
}

// WithValuer sets the value formatter for this CommonWriter.
// The value formatter extracts and formats the actual value from RecordData.
func (cw *CommonWriter) WithValuer(f Formatter) *CommonWriter {
	cw.Valuer = f
	return cw
}

// WithStaticKey sets a static string as the key for this CommonWriter.
// This is a convenience method equivalent to WithKey(Static(key)).
func (cw *CommonWriter) WithStaticKey(key string) *CommonWriter {
	cw.Key = Static(key)
	return cw
}

// WithPrefix sets the prefix function for this CommonWriter.
// The prefix function determines what text appears before this log component.
func (cw *CommonWriter) WithPrefix(f PrefixFunc) *CommonWriter {
	cw.Prefix = f
	return cw
}

// WithKeyColorizer sets the styler for the key portion of this CommonWriter.
// The key styler applies colors and formatting to the key text.
func (cw *CommonWriter) WithKeyColorizer(c Styler) *CommonWriter {
	cw.KeyStyler = c
	return cw
}

// WithValueColorizer sets the styler for the value portion of this CommonWriter.
// The value styler applies colors and formatting to the value text.
func (cw *CommonWriter) WithValueColorizer(c Styler) *CommonWriter {
	cw.ValueStyler = c
	return cw
}

func (cw *CommonWriter) KeyLen(info RecordData) int {
	key := cw.Key(info)
	if key == "" {
		return 0
	}
	if info.Color {
		return len(cw.KeyStyler(info, key))
	}
	return len(key)
}

func (cw *CommonWriter) Write(info RecordData) {
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
