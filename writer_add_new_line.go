package prettylog

func AddNewLineFormat(info RecordData) string {
	return "\n"
}

var DefaultNewLineWriter = NewCommonWriter(AddNewLineFormat)
