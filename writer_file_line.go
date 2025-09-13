package prettylog

import (
	"os"
	"strconv"
	"strings"
)

var wd, _ = os.Getwd()

// DefaultFileLineWriter is the default entry writer for file and line information.
// It uses short format and includes the "File" key.
var DefaultFileLineWriter = NewFileLineWriter()

// ShortFileLineFormat returns file path and line number with working directory trimmed.
func ShortFileLineFormat(info RecordInfo) string {
	return strings.TrimPrefix(info.Frame.File, wd) + ":" + strconv.Itoa(info.Frame.Line)
}

// LongFileLineFormat returns the complete file path with line number.
func LongFileLineFormat(info RecordInfo) string {
	return info.Frame.File + ":" + strconv.Itoa(info.Frame.Line)
}

// FileLineWriter is a specialized entry writer for file and line number output.
// It extends CommonWriter with file/line specific formatting options.
type FileLineWriter struct {
	*CommonWriter
}

// NewFileLineWriter creates a new FileLineWriter with short format and "File" key.
func NewFileLineWriter() *FileLineWriter {
	return &FileLineWriter{
		CommonWriter: NewCommonWriter(ShortFileLineFormat).WithStaticKey("File"),
	}
}

// WithShortFormat sets the formatter to use short file paths with working directory trimmed.
func (fw *FileLineWriter) WithShortFormat() *FileLineWriter {
	fw.CommonWriter.Valuer = ShortFileLineFormat
	return fw
}

// WithLongFormat sets the formatter to use complete file paths.
func (fw *FileLineWriter) WithLongFormat() *FileLineWriter {
	fw.CommonWriter.Valuer = LongFileLineFormat
	return fw
}

// WithFormat sets a custom formatter for the file and line number output.
// This allows complete control over how file/line information is formatted.
func (fw *FileLineWriter) WithFormat(format Formatter) *FileLineWriter {
	fw.CommonWriter.Valuer = format
	return fw
}
