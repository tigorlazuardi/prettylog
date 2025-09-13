package prettylog

import (
	"strings"
)

var _ EntryWriter = (*FunctionWriter)(nil)

var DefaultFunctionWrtier = NewFunctionWriter()

func ShortFunctionFormat(info RecordInfo) string {
	return strings.TrimPrefix(info.Frame.Function, info.PackageName)
}

func FullFunctionFormat(info RecordInfo) string {
	return info.Frame.Function
}

type FunctionWriter struct {
	*CommonWriter
}

func NewFunctionWriter() *FunctionWriter {
	return &FunctionWriter{
		CommonWriter: NewCommonWriter(ShortFunctionFormat).WithStaticKey("Function"),
	}
}

// WithShort sets whether to use the short function name if possible.
//
// If true, the package name is stripped from the function name if it matches
// the package name given by [WithPackageName] option in [New] or [Handler.Clone].
func (fw *FunctionWriter) WithShort(short bool) *FunctionWriter {
	if short {
		fw.CommonWriter.Valuer = ShortFunctionFormat
	} else {
		fw.CommonWriter.Valuer = FullFunctionFormat
	}
	return fw
}

func (fw *FunctionWriter) WithFormat(format Formatter) *FunctionWriter {
	fw.CommonWriter.Valuer = format
	return fw
}

func (fu FunctionWriter) Write(info RecordInfo) {
	// Keep consistent with slog contract to not write anything if caller info is not available.
	if info.Frame.Func == nil {
		return
	}
	fu.CommonWriter.Write(info)
}
