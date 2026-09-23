package natsx_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Jordan-J-Brown/casino-go/pkg/natsx"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

// Should test that a middleware chain is ran in the correct order
func Test_Middleware_Chain(t *testing.T) {
	const sequence = "sequence"

	type Foo struct {
		Bar int `json:"bar"`
	}

	ctx := context.WithValue(context.Background(), sequence, make([]int, 0, 4))

	testMiddleware := func(n int) natsx.Middleware {
		return func(next natsx.RawMsgHandler) natsx.RawMsgHandler {
			return natsx.RawMsgHandler(func(ctx context.Context, msg *nats.Msg) error {
				seq := ctx.Value(sequence).([]int)
				ctx = context.WithValue(ctx, sequence, append(seq, n))
				return next(ctx, msg)
			})
		}
	}

	server := natsserver.RunRandClientPortServer()
	defer server.Shutdown()

	addr := server.Addr().String()
	port := strings.Split(addr, ":")[1]

	nc, err := nats.Connect("nats://127.0.0.1:" + port)
	if err != nil {
		t.Fatal(err)
	}

	mh := natsx.NewMessageHub(nc)
	mh.SetInboundMiddleware(testMiddleware(1))
	defer func() {
		if err := mh.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	seq := make(chan []int)

	err = mh.Subscribe(ctx, natsx.SubConfig[Foo]{
		Subject: "subject.foo",
		Handler: func(ctx context.Context, msg *natsx.Msg[Foo]) error {
			seq <- ctx.Value(sequence).([]int)
			return nil
		},
		ErrHandler: func(ctx context.Context, msg *nats.Msg, err error) {
			t.Error(err)
			seq <- nil
		},
		Decoder: natsx.DecodeFunc[Foo](func(msg *natsx.Msg[Foo]) error {
			return json.Unmarshal(msg.Data, &msg.Body)
		}),
		Middleware: []natsx.Middleware{
			testMiddleware(2),
			testMiddleware(3),
			testMiddleware(4),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = nc.Publish("subject.foo", []byte(`{"bar": 42}`))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []int{1, 2, 3, 4}, <-seq)
}
