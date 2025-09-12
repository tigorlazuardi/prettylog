package prettylog

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Formatter func(info RecordInfo) string

// DefaultLevelFormatter returns the string representation of the log level.
func DefaultLevelFormatter(info RecordInfo) string {
	return info.Record.Level.String()
}

// TrimPrefixFunctionFormatter tries to trim the given package name prefix from the function name.
func TrimPrefixFunctionFormatter(info RecordInfo) string {
	if info.Frame.Func == nil {
		return ""
	}
	return strings.TrimPrefix(info.Frame.Function, info.PackageName)
}

// FullFunctionFormatter returns the full function name.
func FullFunctionFormatter(info RecordInfo) string {
	if info.Frame.Func == nil {
		return ""
	}
	return info.Frame.Function
}

// ShortFileLineFormatter tries to shorten the file path by trimming the current working directory from it.
// If the current working directory cannot be determined or is not a prefix of the file path,
// the full file path is returned.
func ShortFileLineFormatter(info RecordInfo) string {
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

// FullFileLineFormatter returns the full file path with line number.
func FullFileLineFormatter(info RecordInfo) string {
	if info.Frame.Func == nil {
		return ""
	}
	return info.Frame.File + ":" + strconv.Itoa(info.Frame.Line)
}

// TimeOnlyFormatter returns the time in [time.TimeOnly] format.
func TimeOnlyFormatter(info RecordInfo) string {
	return info.Record.Time.Format(time.TimeOnly)
}

// RFC3339TimeFormatter returns the time in [time.RFC3339] format.
func RFC3339TimeFormatter(info RecordInfo) string {
	return info.Record.Time.Format(time.RFC3339)
}

// DefaultMessageFormatter returns the log message as is.
func DefaultMessageFormatter(info RecordInfo) string {
	return info.Record.Message
}

// Static returns a static string as the formatted value.
func Static(key string) Formatter {
	return func(RecordInfo) string {
		return key
	}
}
