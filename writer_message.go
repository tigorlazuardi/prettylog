package prettylog

var DefaultMessageValuer = func(info RecordInfo) string {
	return info.Record.Message
}

var DefaultMessageWriter = NewCommonWriter(DefaultMessageValuer).
	WithValueColorizer(SimpleColoredStyler)
