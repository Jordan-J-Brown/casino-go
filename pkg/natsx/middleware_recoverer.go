package natsx

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/nats-io/nats.go"
)

func MiddlewareRecoverer(next RawMsgHandler) RawMsgHandler {
	return RawMsgHandler(func(ctx context.Context, msg *nats.Msg) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorx.Newf("Panic recover: %v", r)
			}
		}()

		return next(ctx, msg)
	})
}
