package prettylog

import (
	"time"
)

var _ EntryWriter = (*TimeWriter)(nil)

var DefaultTimeWriter = NewTimeWriter()

func TimeOnlyTimeFormat(info RecordInfo) string {
	return info.Record.Time.Format(time.TimeOnly)
}

func RFC3339TimeFormat(info RecordInfo) string {
	return info.Record.Time.Format(time.RFC3339)
}

type TimeWriter struct {
	*CommonWriter
}

func NewTimeWriter() *TimeWriter {
	return &TimeWriter{
		NewCommonWriter(TimeOnlyTimeFormat).WithStaticKey("Time"),
	}
}

func (tw *TimeWriter) WithTimeOnlyFormat() *TimeWriter {
	tw.CommonWriter.Valuer = TimeOnlyTimeFormat
	return tw
}

func (tw *TimeWriter) WithRFC3339Format() *TimeWriter {
	tw.CommonWriter.Valuer = RFC3339TimeFormat
	return tw
}

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
