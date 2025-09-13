package prettylog

import (
	"log/slog"

	"github.com/fatih/color"
)

// Styler is a function type that applies styling (colors, formatting) to formatted strings.
// It receives the RecordData for context (like log level) and the formatted string,
// returning the styled version.
type Styler func(info RecordData, formatted string) (styled string)

// SimpleColoredStyler applies color styling based on the log level.
// Colors: Error=Red, Warn=Yellow, Info=Green, Debug=Cyan, others=White.
func SimpleColoredStyler(info RecordData, s string) string {
	return getColoredText(info.Record.Level).Sprint(s)
}

// BoldColoredStyler applies bold color styling based on the log level.
// Same colors as SimpleColoredStyler but with bold formatting.
func BoldColoredStyler(info RecordData, s string) string {
	return getColoredText(info.Record.Level).Add(color.Bold).Sprint(s)
}

// PlainStyler returns the string unchanged without any styling.
// Useful when you want to disable styling for specific components.
func PlainStyler(info RecordData, s string) string {
	return s
}

// BackgroundBoldColoredStyler applies background color with bold white text based on log level.
// Provides high contrast styling suitable for important information like log levels.
func BackgroundBoldColoredStyler(info RecordData, s string) string {
	return getColoredBackground(info.Record.Level).Add(color.Bold).Sprint(s)
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
