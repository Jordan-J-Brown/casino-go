package natsx

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Middleware func(next RawMsgHandler) RawMsgHandler

type RawMsgHandler func(ctx context.Context, msg *nats.Msg) error
