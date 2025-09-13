package prettylog

func AddNewLineFormat(info RecordInfo) string {
	return "\n"
}

var DefaultNewLineWriter = NewCommonWriter(AddNewLineFormat)
