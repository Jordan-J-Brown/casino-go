package engineclient

import (
	"context"
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/httpx"
	"github.com/Jordan-J-Brown/casino-go/pkg/mimex"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/Jordan-J-Brown/casino-go/pkg/natsx"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Client interface {
	Setup(ctx context.Context, engineID string, request enginemsg.SetupRequest) (*enginemsg.SetupResponse, *errorx.Error)
	Play(ctx context.Context, engineID string, request enginemsg.PlayRequest) (*enginemsg.RoundResponse, *errorx.Error)
	Next(ctx context.Context, engineID string, request enginemsg.NextRequest) (*enginemsg.RoundResponse, *errorx.Error)
	Close(ctx context.Context, engineID string, request enginemsg.CloseRequest) (*enginemsg.CloseResponse, *errorx.Error)
}

type client struct {
	nc      *nats.Conn
	nclient *natsx.Client
	timeout time.Duration
}

func New(
	nc *nats.Conn,
	tracer trace.Tracer,
	timeout time.Duration,
) Client {
	nclient := natsx.NewClient(nc).
		SetContentType(mimex.ContentTypeJSON).
		SetRetry(3, time.Millisecond*100, time.Second).
		SetTimeout(3 * time.Second)

	client := &client{
		nc:      nc,
		nclient: nclient,
		timeout: timeout,
	}
	return client
}

func (c *client) Setup(ctx context.Context, engineID string, request enginemsg.SetupRequest) (*enginemsg.SetupResponse, *errorx.Error) {
	var response enginemsg.SetupResponse
	var errResponse natsx.ErrorResponse

	req, err := c.nclient.R().
		SetHeaders(c.getHeaders(ctx)).
		SetBody(&request).
		SetResult(&response).
		SetError(&errResponse).
		Exec(ctx, enginemsg.CreateEngineSubject(engineID, "setup"))
	if err != nil {
		return nil, errorx.NewFromError(err)
	}
	if req.IsError() {
		return nil, convertErrorResponse(&errResponse)
	}

	return &response, nil
}

func (c *client) Play(ctx context.Context, engineID string, request enginemsg.PlayRequest) (*enginemsg.RoundResponse, *errorx.Error) {
	var response enginemsg.RoundResponse
	var errResponse natsx.ErrorResponse

	req, err := c.nclient.R().
		SetHeaders(c.getHeaders(ctx)).
		SetBody(&request).
		SetResult(&response).
		SetError(&errResponse).
		Exec(ctx, enginemsg.CreateEngineSubject(engineID, "play"))
	if err != nil {
		return nil, errorx.NewFromError(err)
	}
	if req.IsError() {
		return nil, convertErrorResponse(&errResponse)
	}

	return &response, nil
}

func (c *client) Next(ctx context.Context, engineID string, request enginemsg.NextRequest) (*enginemsg.RoundResponse, *errorx.Error) {
	var response enginemsg.RoundResponse
	var errResponse natsx.ErrorResponse

	req, err := c.nclient.R().
		SetHeaders(c.getHeaders(ctx)).
		SetBody(&request).
		SetResult(&response).
		SetError(&errResponse).
		Exec(ctx, enginemsg.CreateEngineSubject(engineID, "next"))
	if err != nil {
		return nil, errorx.NewFromError(err)
	}
	if req.IsError() {
		return nil, convertErrorResponse(&errResponse)
	}

	return &response, nil
}

func (c *client) Close(ctx context.Context, engineID string, request enginemsg.CloseRequest) (*enginemsg.CloseResponse, *errorx.Error) {
	var response enginemsg.CloseResponse
	var errResponse natsx.ErrorResponse

	req, err := c.nclient.R().
		SetHeaders(c.getHeaders(ctx)).
		SetBody(&request).
		SetResult(&response).
		SetError(&errResponse).
		Exec(ctx, enginemsg.CreateEngineSubject(engineID, "close"))
	if err != nil {
		return nil, errorx.NewFromError(err)
	}
	if req.IsError() {
		return nil, convertErrorResponse(&errResponse)
	}

	return &response, nil
}

func (c *client) getHeaders(ctx context.Context) nats.Header {
	h := nats.Header{}

	otel.GetTextMapPropagator().Inject(ctx, &natsx.NatsOtelHeaderCarrier{Header: h})

	return h
}

func convertErrorResponse(err *natsx.ErrorResponse) *errorx.Error {
	var code errorx.Code
	if httpx.IsClientError(err.StatusCode) {
		code = errorx.ClientError
	} else {
		code = errorx.ServerError
	}

	return errorx.New(err.Message, errorx.WithCode(code), errorx.WithType(err.Type))
}
