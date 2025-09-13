package prettylog

// DefaultMessageValuer extracts the log message from RecordData.
// This is the default value formatter used by DefaultMessageWriter.
var DefaultMessageValuer = func(info RecordData) string {
	return info.Record.Message
}

// DefaultMessageWriter is the default entry writer for log messages.
// It displays the log message with simple color styling based on log level.
var DefaultMessageWriter = NewCommonWriter(DefaultMessageValuer).
	WithValueColorizer(SimpleColoredStyler)
