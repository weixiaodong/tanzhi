package log

import (
	"context"
)

type ctxTraceIdKey struct{}

func WithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, ctxTraceIdKey{}, traceId)
}

func traceIdFromCtx(ctx context.Context) (traceId string) {
	traceId, _ = ctx.Value(ctxTraceIdKey{}).(string)
	return
}

type ctxModuleKey struct{}

func WithModule(ctx context.Context, module string) context.Context {
	return context.WithValue(ctx, ctxModuleKey{}, module)
}

func moduleFromCtx(ctx context.Context) (module string) {
	module, _ = ctx.Value(ctxModuleKey{}).(string)
	return
}

func ctxFields(c context.Context, fields []interface{}) []interface{} {
	if n := len(fields); n&0x1 == 1 { // odd number
		fields = fields[:n-1]
	}
	traceId := traceIdFromCtx(c)
	module := moduleFromCtx(c)
	fields = append(
		[]interface{}{"traceId", traceId, "module", module},
		fields...,
	)

	return fields
}
