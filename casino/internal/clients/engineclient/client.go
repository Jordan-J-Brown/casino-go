package engineclient

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/msg"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/nats-io/nats.go"
)

type Client interface {
	Setup(ctx context.Context, engineID string, request enginemsg.SetupRequest) (*enginemsg.SetupResponse, error)
	Play(ctx context.Context, engineID string, request enginemsg.PlayRequest) (*enginemsg.RoundResponse, error)
	Next(ctx context.Context, engineID string, request enginemsg.NextRequest) (*enginemsg.RoundResponse, error)
	Close(ctx context.Context, engineID string, request enginemsg.CloseRequest) (*enginemsg.CloseResponse, error)
}

type client struct {
	nc      *nats.Conn
	timeout time.Duration
}

func New(nc *nats.Conn, timeout time.Duration) Client {
	client := &client{
		nc:      nc,
		timeout: timeout,
	}
	return client
}

func (c *client) Setup(ctx context.Context, engineID string, request enginemsg.SetupRequest) (*enginemsg.SetupResponse, error) {
	response, err := (&Request[enginemsg.SetupResponse]{
		nc:      c.nc,
		timeout: c.timeout,
		header:  c.getHeaders(ctx),
		subject: enginemsg.CreateEngineSubject(engineID, "setup"),
	}).Execute(request)
	return response, err
}

func (c *client) Play(ctx context.Context, engineID string, request enginemsg.PlayRequest) (*enginemsg.RoundResponse, error) {
	response, err := (&Request[enginemsg.RoundResponse]{
		nc:      c.nc,
		timeout: c.timeout,
		header:  c.getHeaders(ctx),
		subject: enginemsg.CreateEngineSubject(engineID, "play"),
	}).Execute(request)
	return response, err
}

func (c *client) Next(ctx context.Context, engineID string, request enginemsg.NextRequest) (*enginemsg.RoundResponse, error) {
	response, err := (&Request[enginemsg.RoundResponse]{
		nc:      c.nc,
		timeout: c.timeout,
		header:  c.getHeaders(ctx),
		subject: enginemsg.CreateEngineSubject(engineID, "next"),
	}).Execute(request)
	return response, err
}

func (c *client) Close(ctx context.Context, engineID string, request enginemsg.CloseRequest) (*enginemsg.CloseResponse, error) {
	response, err := (&Request[enginemsg.CloseResponse]{
		nc:      c.nc,
		timeout: c.timeout,
		header:  c.getHeaders(ctx),
		subject: enginemsg.CreateEngineSubject(engineID, "close"),
	}).Execute(request)
	return response, err
}

func (c *client) getHeaders(ctx context.Context) nats.Header {
	// TODO: propagate otel headers
	return nats.Header{}
}

type Request[Result any] struct {
	nc      *nats.Conn
	timeout time.Duration
	header  nats.Header
	subject string
}

func (r *Request[Result]) Execute(body any) (*Result, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	reply, err := r.nc.RequestMsg(&nats.Msg{
		Subject: r.subject,
		Header:  r.header,
		Data:    data,
	}, r.timeout)
	if err != nil {
		return nil, err
	}

	status := reply.Header.Get(msg.HeaderStatus)
	if status == msg.StatusError {
		var err msg.ErrorResponse
		if err := json.Unmarshal(reply.Data, &err); err != nil {
			return nil, err
		}
		return nil, &Error{Code: ErrEngineError, Message: err.Message}
	}

	var result Result
	if err := json.Unmarshal(reply.Data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
