package natsx

import (
	"context"
	"slices"
	"sync/atomic"

	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/nats-io/nats.go"
)

type MessageHub struct {
	nc                 *nats.Conn
	locked             atomic.Bool
	inboundMiddleware  []Middleware
	outboundMiddleware []Middleware
}

func NewMessageHub(nc *nats.Conn) *MessageHub {
	return &MessageHub{
		nc: nc,
	}
}

func (m *MessageHub) SetInboundMiddleware(mw ...Middleware) *MessageHub {
	if !m.locked.Load() {
		m.inboundMiddleware = mw
		return m
	}
	panic("attempted to set inbound middleware after message hub has been locked")
}

func (m *MessageHub) SetOutboundMiddleware(mw ...Middleware) *MessageHub {
	if !m.locked.Load() {
		m.outboundMiddleware = mw
		return m
	}
	panic("attempted to set outbound middleware after message hub has been locked")
}

func (m *MessageHub) Conn() *nats.Conn {
	return m.nc
}

type SubConfig[T any] struct {
	Subject    string
	Handler    func(ctx context.Context, msg *Msg[T]) error
	ErrHandler func(ctx context.Context, msg *nats.Msg, err error)
	Decoder    Decoder[T]
	Middleware []Middleware //Optional
}

func (m *MessageHub) Subscribe[T any](ctx context.Context, cfg SubConfig[T]) error {
	if cfg.Subject == "" {
		return errorx.New("missing subject")
	}
	if cfg.Handler == nil {
		return errorx.New("missing handler")
	}
	if cfg.ErrHandler == nil {
		return errorx.New("missing error handler")
	}
	if cfg.Decoder == nil {
		return errorx.New("missing decoder")
	}

	m.locked.Store(true)

	sub := &subscription[T]{
		handler:    cfg.Handler,
		errHandler: cfg.ErrHandler,
		subject:    cfg.Subject,
	}

	decoder := cfg.Decoder
	handler := RawMsgHandler(func(ctx context.Context, rawMsg *nats.Msg) error {
		msg := &Msg[T]{
			Msg: rawMsg,
		}
		if err := decoder.Decode(msg); err != nil {
			return err
		}

		err := sub.handler(ctx, msg)
		if err != nil {
			return err
		}

		return nil
	})
	middleware := slices.Concat(m.inboundMiddleware, cfg.Middleware)
	for i := range slices.Backward(middleware) {
		handler = middleware[i](handler)
	}

	natsSub, err := m.nc.Subscribe(sub.subject, func(msg *nats.Msg) {
		err := handler(ctx, msg)
		if err != nil {
			sub.errHandler(ctx, msg, err)
		}
	})
	if err != nil {
		return err
	}

	sub.Subscription = natsSub

	return nil
}

// Publish a message with optional headers. Outbound middleware will pre-process the message before publishing it.
func (m *MessageHub) Publish(ctx context.Context, subject string, data []byte, header ...nats.Header) error {
	var h nats.Header
	if len(header) == 1 {
		h = header[0]
	}

	msg := &nats.Msg{
		Subject: subject,
		Header:  h,
		Data:    data,
	}

	if len(m.outboundMiddleware) > 0 {
		mwStack := RawMsgHandler(func(ctx context.Context, msg *nats.Msg) error { return nil })
		for i := range slices.Backward(m.outboundMiddleware) {
			mwStack = m.outboundMiddleware[i](mwStack)
		}

		err := mwStack(ctx, msg)
		if err != nil {
			return err
		}
	}

	return m.nc.PublishMsg(msg)
}

func (m *MessageHub) Close() error {
	err := m.nc.Drain()
	if err != nil {
		return errorx.New("failed to drain nats connection", errorx.WithChildren(err))
	}

	return nil
}
