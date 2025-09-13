package prettylog

import (
	"log/slog"

	"github.com/fatih/color"
)

type Styler func(info RecordInfo, formatted string) (styled string)

func SimpleColoredStyler(info RecordInfo, s string) string {
	return getColoredText(info.Record.Level).Sprint(s)
}

func BoldColoredStyler(info RecordInfo, s string) string {
	return getColoredText(info.Record.Level).Add(color.Bold).Sprint(s)
}

func PlainStyler(info RecordInfo, s string) string {
	return s
}

func BackgroundBoldColoredStyler(info RecordInfo, s string) string {
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
