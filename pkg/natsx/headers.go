package natsx

import (
	"maps"
	"slices"

	"github.com/nats-io/nats.go"
)

const (
	HeaderStatus = "Status"
)

type NatsOtelHeaderCarrier struct {
	nats.Header
}

func (n *NatsOtelHeaderCarrier) Keys() []string {
	return slices.Collect(maps.Keys(n.Header))
}
