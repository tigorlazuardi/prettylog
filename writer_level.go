package prettylog

// DefaultLevelValuer extracts the log level as a string from RecordData.
// This is the default value formatter used by DefaultLevelWriter.
func DefaultLevelValuer(info RecordData) string {
	return info.Record.Level.String()
}

// DefaultLevelWriter is the default entry writer for log levels.
// It displays the log level with background bold colored styling.
var DefaultLevelWriter = NewCommonWriter(DefaultLevelValuer).
	WithValueColorizer(BackgroundBoldColoredStyler)
