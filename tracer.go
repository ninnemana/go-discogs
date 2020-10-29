package discogs

import (
	"context"

	"go.opencensus.io/trace"
)

type ErrorConfig struct {
	Error      error
	Code       int32
	Message    string
	Attributes []trace.Attribute
}

func RecordError(ctx context.Context, e ErrorConfig) {
	span := trace.FromContext(ctx)
	if span == nil {
		return
	}

	if e.Error != nil {
		span.AddAttributes(trace.StringAttribute("error", e.Error.Error()))
	}

	if e.Code == 0 {
		e.Code = trace.StatusCodeUnknown
	}

	span.AddAttributes(e.Attributes...)
	span.SetStatus(trace.Status{
		Code:    e.Code,
		Message: e.Message,
	})
}
