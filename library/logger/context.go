package logger

import (
	"context"
)

type contextKey uint64

const (
	contextLogID contextKey = iota
	contextHTTPLogFields
	contextTraceID
)

// WithLogID inject log id to context
func WithLogID(ctx context.Context, val interface{}) context.Context {
	return context.WithValue(ctx, contextLogID, val)
}

// ValueLogID extract log id from context
func ValueLogID(ctx context.Context) string {
	val := ctx.Value(contextLogID)
	logID, ok := val.(string)
	if !ok {
		return ""
	}
	return logID
}

// WithTraceID inject trace_id id to context
func WithTraceID(ctx context.Context, val interface{}) context.Context {
	return context.WithValue(ctx, contextTraceID, val)
}

// ValueTraceID extract trace id from context
func ValueTraceID(ctx context.Context) string {
	val := ctx.Value(contextTraceID)
	logID, ok := val.(string)
	if !ok {
		return ""
	}
	return logID
}

// WithHTTPFields inject common http log fields to context
func WithHTTPFields(ctx context.Context, fields Fields) context.Context {
	return context.WithValue(ctx, contextHTTPLogFields, fields)
}

// ValueHTTPFields extrect common http log fields from context
func ValueHTTPFields(ctx context.Context) Fields {
	val := ctx.Value(contextHTTPLogFields)
	fields, ok := val.(Fields)
	if !ok {
		return Fields{}
	}
	return fields
}
