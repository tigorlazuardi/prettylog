package prettylog

import (
	"time"
)

var _ EntryWriter = (*TimeWriter)(nil)

// DefaultTimeWriter is the default entry writer for timestamps.
// It uses time-only format and includes the "Time" key.
var DefaultTimeWriter = NewTimeWriter()

// TimeOnlyTimeFormat formats the timestamp in time.TimeOnly format ("15:04:05").
func TimeOnlyTimeFormat(info RecordInfo) string {
	return info.Record.Time.Format(time.TimeOnly)
}

// RFC3339TimeFormat formats the timestamp in RFC3339 format.
func RFC3339TimeFormat(info RecordInfo) string {
	return info.Record.Time.Format(time.RFC3339)
}

// TimeWriter is a specialized entry writer for timestamp output.
// It extends CommonWriter with time-specific formatting options.
type TimeWriter struct {
	*CommonWriter
}

// NewTimeWriter creates a new TimeWriter with time-only format and "Time" key.
func NewTimeWriter() *TimeWriter {
	return &TimeWriter{
		NewCommonWriter(TimeOnlyTimeFormat).WithStaticKey("Time"),
	}
}

// WithTimeOnlyFormat sets the time formatter to use time.TimeOnly format ("15:04:05").
func (tw *TimeWriter) WithTimeOnlyFormat() *TimeWriter {
	tw.CommonWriter.Valuer = TimeOnlyTimeFormat
	return tw
}

// WithRFC3339Format sets the time formatter to use RFC3339 format.
func (tw *TimeWriter) WithRFC3339Format() *TimeWriter {
	tw.CommonWriter.Valuer = RFC3339TimeFormat
	return tw
}

// WithTimeFormat sets the time formatter to use a custom time layout.
// The layout parameter follows Go's time formatting conventions.
func (tw *TimeWriter) WithTimeFormat(layout string) *TimeWriter {
	tw.CommonWriter.Valuer = func(info RecordInfo) string {
		return info.Record.Time.Format(layout)
	}
	return tw
}

func (tw *TimeWriter) Write(info RecordInfo) {
	if info.Record.Time.IsZero() {
		return
	}
	tw.CommonWriter.Write(info)
}
