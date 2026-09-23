package natsx

import "github.com/nats-io/nats.go"

type Msg[T any] struct {
	*nats.Msg
	ContentType string
	Body        T
}
