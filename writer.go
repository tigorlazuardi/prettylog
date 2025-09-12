package prettylog

import "bytes"

type EntryWriter interface {
	// Key returns the key to be used for this entry.
	//
	// If colored, this should also include the ANSI color codes
	// as length of key.
	KeyLen(info RecordInfo) int

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
	Write(info RecordInfo, buf *bytes.Buffer)
}

// UnimplementedEntryWriter is an [EntryWriter] that does nothing.
// It can be embedded to have forward compatible implementations.
type UnimplementedEntryWriter struct{}

func (UnimplementedEntryWriter) KeyLen(info RecordInfo) int               { return 0 }
func (UnimplementedEntryWriter) Write(info RecordInfo, buf *bytes.Buffer) {}
