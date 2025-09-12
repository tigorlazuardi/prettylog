package prettylog

import (
	"log/slog"

	"github.com/fatih/color"
)

// DefaultLevelStyler is the default styler for log levels.
func DefaultLevelStyler(lvl slog.Level, s string) string {
	return getColoredBackground(lvl).Add(color.Bold).Sprintf(" %s ", s)
}

// DefaultSourceStyler is the default styler for source (function and file/line).
func DefaultSourceStyler(lvl slog.Level, s string) string {
	return s
}

func DefaultTimeStyler(lvl slog.Level, s string) string {
	return getColoredBackground(lvl).Sprint(s)
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
