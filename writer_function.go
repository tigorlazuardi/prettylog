package prettylog

import (
	"strings"
)

var _ EntryWriter = (*FunctionWriter)(nil)

// DefaultFunctionWrtier is the default entry writer for function names.
// It uses short function format and includes the "Function" key.
var DefaultFunctionWrtier = NewFunctionWriter()

// ShortFunctionFormat returns the function name with package prefix trimmed.
func ShortFunctionFormat(info RecordData) string {
	return strings.TrimPrefix(info.Frame.Function, info.PackageName)
}

// FullFunctionFormat returns the complete function name including package.
func FullFunctionFormat(info RecordData) string {
	return info.Frame.Function
}

// FunctionWriter is a specialized entry writer for function name output.
// It extends CommonWriter with function-specific formatting options.
type FunctionWriter struct {
	*CommonWriter
}

// NewFunctionWriter creates a new FunctionWriter with short format and "Function" key.
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

// WithFormat sets a custom formatter for the function name output.
// This allows complete control over how function names are formatted.
func (fw *FunctionWriter) WithFormat(format Formatter) *FunctionWriter {
	fw.CommonWriter.Valuer = format
	return fw
}

func (fu FunctionWriter) Write(info RecordData) {
	// Keep consistent with slog contract to not write anything if caller info is not available.
	if info.Frame.Func == nil {
		return
	}
	fu.CommonWriter.Write(info)
}
