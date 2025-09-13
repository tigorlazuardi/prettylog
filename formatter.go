package prettylog

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Formatter is a function type that extracts and formats information from a RecordData.
// Formatters are used by writers to generate the string representation of log components.
type Formatter func(info RecordData) string

// DefaultLevelFormatter returns the string representation of the log level.
// It uses the standard slog level string representation.
func DefaultLevelFormatter(info RecordData) string {
	return info.Record.Level.String()
}

// TrimPrefixFunctionFormatter returns the function name with the package name prefix trimmed.
// If the function information is not available, it returns an empty string.
func TrimPrefixFunctionFormatter(info RecordData) string {
	if info.Frame.Func == nil {
		return ""
	}
	return strings.TrimPrefix(info.Frame.Function, info.PackageName)
}

// FullFunctionFormatter returns the complete function name including package.
// If the function information is not available, it returns an empty string.
func FullFunctionFormatter(info RecordData) string {
	if info.Frame.Func == nil {
		return ""
	}
	return info.Frame.Function
}

// ShortFileLineFormatter returns the file path and line number with the current working
// directory trimmed when possible. If the working directory cannot be determined or is not
// a prefix of the file path, the full file path is returned.
// Format: "path/to/file.go:123"
func ShortFileLineFormatter(info RecordData) string {
	if info.Frame.Func == nil {
		return ""
	}
	wd, _ := os.Getwd()
	s, trimmed := strings.CutPrefix(info.Frame.File, wd)
	if trimmed {
		return strings.TrimPrefix(s, "/") + ":" + strconv.Itoa(info.Frame.Line)
	}
	return info.Frame.File + ":" + strconv.Itoa(info.Frame.Line)
}

// FullFileLineFormatter returns the complete file path with line number.
// Format: "/full/path/to/file.go:123"
func FullFileLineFormatter(info RecordData) string {
	if info.Frame.Func == nil {
		return ""
	}
	return info.Frame.File + ":" + strconv.Itoa(info.Frame.Line)
}

// TimeOnlyFormatter returns the time in time.TimeOnly format ("15:04:05").
func TimeOnlyFormatter(info RecordData) string {
	return info.Record.Time.Format(time.TimeOnly)
}

// RFC3339TimeFormatter returns the time in time.RFC3339 format.
func RFC3339TimeFormatter(info RecordData) string {
	return info.Record.Time.Format(time.RFC3339)
}

// DefaultMessageFormatter returns the log message without any modifications.
func DefaultMessageFormatter(info RecordData) string {
	return info.Record.Message
}

// Static returns a Formatter that always returns the same static string,
// regardless of the RecordData content.
func Static(key string) Formatter {
	return func(RecordData) string {
		return key
	}
}
