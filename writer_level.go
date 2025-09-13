package prettylog

func DefaultLevelValuer(info RecordInfo) string {
	return info.Record.Level.String()
}

var DefaultLevelWriter = NewCommonWriter(DefaultLevelValuer).
	WithValueColorizer(BackgroundBoldColoredStyler)
