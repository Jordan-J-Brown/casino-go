package contextx

import (
	"context"
)

func CorrelationIdFromContext(ctx context.Context) string {
	id, _ := ctx.Value(CtxKeyCorrelationID).(string)
	return id
}

func IdempotencyKeyFromContext(ctx context.Context) string {
	id, _ := ctx.Value(CtxKeyIdempotencyKey).(string)
	return id
}
