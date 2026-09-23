package natsx

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/pkg/contextx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/nats-io/nats.go"
)

// Injects a logger into the context.
//
// [contextx.LoggerFromContext] can be used to retrieve the logger.
func MiddlewareLogger(logger logsx.Logger) Middleware {
	return func(next RawMsgHandler) RawMsgHandler {
		return func(ctx context.Context, msg *nats.Msg) error {
			return next(context.WithValue(ctx, contextx.CtxKeyLogger, logger), msg)
		}
	}
}
