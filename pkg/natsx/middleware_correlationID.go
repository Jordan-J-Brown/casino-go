package natsx

import (
	"context"
	"errors"

	"github.com/Jordan-J-Brown/casino-go/pkg/contextx"
	"github.com/nats-io/nats.go"
)

const (
	HeaderCorrelationID = "Correlation-ID"
)

var (
	ErrMissingCorrelationID = errors.New(HeaderCorrelationID + " header is missing")
)

// Inbound middleware that retrieves the [HeaderCorrelationID] header and places it
// within the context.
func MiddlewareCorrelationID(next RawMsgHandler) RawMsgHandler {
	return func(ctx context.Context, msg *nats.Msg) error {
		id := msg.Header.Get(HeaderCorrelationID)
		if id == "" {
			return ErrMissingCorrelationID
		}
		return next(context.WithValue(ctx, contextx.CtxKeyCorrelationID, id), msg)
	}
}

// Outbound middleware that sets the [HeaderCorrelationID] header within the NATS message
func MiddlewareSetHeaderCorrelationID(next RawMsgHandler) RawMsgHandler {
	return func(ctx context.Context, msg *nats.Msg) error {
		correlationID := contextx.CorrelationIdFromContext(ctx)
		if correlationID == "" {
			return ErrMissingCorrelationID
		}

		if msg.Header == nil {
			msg.Header = nats.Header{}
		}
		msg.Header.Add(HeaderCorrelationID, correlationID)

		return next(ctx, msg)
	}
}
