package natsx

import (
	"context"

	"github.com/nats-io/nats.go"
)

type subscription[T any] struct {
	*nats.Subscription
	subject    string
	handler    func(ctx context.Context, msg *Msg[T]) error
	errHandler func(ctx context.Context, msg *nats.Msg, err error)
}
