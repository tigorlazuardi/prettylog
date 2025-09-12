package prettylog

import "github.com/fatih/color"

type Styler func(info RecordInfo, formatted string) (styled string)

func DefaultTimeKeyStyler(info RecordInfo, key string) string {
	return getColoredText(info.Record.Level).Add(color.Bold).Sprint(key)
}

func DefaultTimeValueStyler(info RecordInfo, s string) string {
	return s
}

func DefaultFunctionKeyStyler(info RecordInfo, key string) string {
	return getColoredText(info.Record.Level).Add(color.Bold).Sprint(key)
}

func DefaultFunctionValueStyler(info RecordInfo, s string) string {
	return s
}

func DefaultFileKeyStyler(info RecordInfo, key string) string {
	return getColoredText(info.Record.Level).Add(color.Bold).Sprint(key)
}

func DefaultFileValueStyler(info RecordInfo, s string) string {
	return s
}

func DefaultMessageStyler(info RecordInfo, s string) string {
	return getColoredText(info.Record.Level).Sprint(s)
}
