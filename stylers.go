package prettylog

import (
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// DefaultLevelStyler is the default styler for log levels.
func DefaultLevelStyler(lvl slog.Level, s string) string {
	return getColoredBackground(lvl).Add(color.Bold).Sprintf(" %s ", s)
}

// DefaultLevelFormatter is the default formatter for log levels.
func DefaultLevelFormatter(lvl slog.Level) string {
	return lvl.String()
}

// DefaultMessageFormatter is the default formatter for log messages.
func DefaultMessageFormatter(lvl slog.Level, msg string) string {
	return msg
}

// TrimPrefixFunctionFormatter tries to trim the given package name prefix from the function name.
func TrimPrefixFunctionFormatter(_ slog.Level, packageNamePrefix string, frame runtime.Frame) string {
	return strings.TrimPrefix(frame.Function, packageNamePrefix)
}

// FullFunctionFormatter returns the full function name.
func FullFunctionFormatter(_ slog.Level, _ string, frame runtime.Frame) string {
	return frame.Function
}

// ShortFileLineFormatter tries to shorten the file path by trimming the current working directory from it.
// If the current working directory cannot be determined or is not a prefix of the file path,
// the full file path is returned.
func ShortFileLineFormatter(_ slog.Level, frame runtime.Frame) string {
	wd, _ := os.Getwd()
	s, trimmed := strings.CutPrefix(frame.File, wd)
	if trimmed {
		return strings.TrimPrefix(s, "/") + ":" + strconv.Itoa(frame.Line)
	}
	return s + ":" + strconv.Itoa(frame.Line)
}

// FullFileLineFormatter returns the full file path with line number.
func FullFileLineFormatter(_ slog.Level, frame runtime.Frame) string {
	return frame.File + ":" + strconv.Itoa(frame.Line)
}

// ShortTimeFormatter returns the time formatted as HH:MM:SS.
func ShortTimeFormatter(_ slog.Level, t time.Time) string {
	return t.Format(time.TimeOnly)
}

// RFC3339TimeFormatter returns the time formatted as RFC3339.
func RFC3339TimeFormatter(_ slog.Level, t time.Time) string {
	return t.Format(time.RFC3339)
}

// DefaultSourceStyler is the default styler for source (function and file/line).
func DefaultSourceStyler(lvl slog.Level, s string) string {
	return s
}

func DefaultTimeStyler(lvl slog.Level, s string) string {
	return getColoredBackground(lvl).Sprint(s)
}

func DefaultMessageStyler(lvl slog.Level, s string) string {
	return getColoredText(lvl).Add(color.Bold).Sprint(s)
}

// getColoredText returns a color.Color with text color based on the log level.
//
// Should be used with transparent background color.
func getColoredText(lvl slog.Level) *color.Color {
	switch {
	case lvl >= slog.LevelError:
		return color.New(color.FgRed)
	case lvl >= slog.LevelWarn:
		return color.New(color.FgYellow)
	case lvl >= slog.LevelInfo:
		return color.New(color.FgGreen)
	case lvl >= slog.LevelDebug:
		return color.New(color.FgCyan)
	default:
		return color.New(color.FgWhite)
	}
}

// getColoredBackground returns a color.Color with background color based on the log level and contrasting text color.
func getColoredBackground(lvl slog.Level) *color.Color {
	switch {
	case lvl >= slog.LevelError:
		return color.New(color.BgRed, color.FgWhite)
	case lvl >= slog.LevelWarn:
		return color.New(color.BgYellow, color.FgWhite)
	case lvl >= slog.LevelInfo:
		return color.New(color.BgGreen, color.FgWhite)
	case lvl >= slog.LevelDebug:
		return color.New(color.BgCyan, color.FgWhite)
	default:
		return color.New(color.BgWhite, color.FgBlack)
	}
}
