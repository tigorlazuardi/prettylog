package prettylog

import (
	"os"
	"strconv"
	"strings"
)

var wd, _ = os.Getwd()

var DefaultFileLineWriter = NewFileLineWriter()

func ShortFileLineFormat(info RecordInfo) string {
	return strings.TrimPrefix(info.Frame.File, wd) + ":" + strconv.Itoa(info.Frame.Line)
}

func LongFileLineFormat(info RecordInfo) string {
	return info.Frame.File + ":" + strconv.Itoa(info.Frame.Line)
}

type FileLineWriter struct {
	*CommonWriter
}

func NewFileLineWriter() *FileLineWriter {
	return &FileLineWriter{
		CommonWriter: NewCommonWriter(ShortFileLineFormat).WithStaticKey("File"),
	}
}

func (fw *FileLineWriter) WithShortFormat() *FileLineWriter {
	fw.CommonWriter.Valuer = ShortFileLineFormat
	return fw
}

func (fw *FileLineWriter) WithLongFormat() *FileLineWriter {
	fw.CommonWriter.Valuer = LongFileLineFormat
	return fw
}

func (fw *FileLineWriter) WithFormat(format Formatter) *FileLineWriter {
	fw.CommonWriter.Valuer = format
	return fw
}
