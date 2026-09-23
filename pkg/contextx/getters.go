package contextx

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
)

func LoggerFromContext(ctx context.Context) logsx.Logger {
	logger, ok := ctx.Value(CtxKeyLogger).(logsx.Logger)
	if !ok {
		return logsx.NopLogger{}
	}
	return logger
}

func CorrelationIdFromContext(ctx context.Context) string {
	id, _ := ctx.Value(CtxKeyCorrelationID).(string)
	return id
}

func IdempotencyKeyFromContext(ctx context.Context) string {
	id, _ := ctx.Value(CtxKeyIdempotencyKey).(string)
	return id
}
