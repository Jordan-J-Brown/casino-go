package natsx

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/mimex"
	"github.com/Jordan-J-Brown/casino-go/pkg/retry"
	"github.com/nats-io/nats.go"
)

// [Client.Close] must be called upon shutdown to properly drain all connections.
type Client struct {
	contentType   string
	retry         int
	retryDelay    time.Duration
	retryMaxDelay time.Duration
	timeout       time.Duration
	header        nats.Header
	ncpool        *ConnPool
}

func NewClient(ctx context.Context, poolConfig ConnPoolConfig) (*Client, error) {
	pool, err := NewConnPool(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		ncpool: pool,
	}, nil
}

func (c *Client) SetContentType(ct string) *Client {
	c.contentType = ct
	return c
}

func (c *Client) SetTimeout(t time.Duration) *Client {
	c.timeout = t
	return c
}

func (c *Client) SetHeaders(h nats.Header) *Client {
	c.header = h
	return c
}

func (c *Client) SetRetry(count int, delay, maxDelay time.Duration) *Client {
	c.retry = count
	c.retryDelay = delay
	c.retryMaxDelay = maxDelay
	return c
}

func (c *Client) R() *Request {
	header := make(nats.Header, len(c.header))
	maps.Copy(header, c.header)
	return &Request{
		nc:            c.ncpool,
		contentType:   c.contentType,
		retry:         c.retry,
		retryDelay:    c.retryDelay,
		retryMaxDelay: c.retryMaxDelay,
		timeout:       c.timeout,
		header:        header,
	}
}

func (c *Client) Close() error {
	return c.ncpool.Close()
}

type Request struct {
	nc            *ConnPool
	isError       bool
	contentType   string
	retry         int
	retryDelay    time.Duration
	retryMaxDelay time.Duration
	timeout       time.Duration
	header        nats.Header
	subject       string
	body          any
	resultPtr     any
	errResultPtr  any
}

func (r *Request) marshalBody() ([]byte, error) {
	switch r.contentType {
	case mimex.ContentTypeJSON:
		return json.Marshal(r.body)
	default:
		return nil, ErrUnsupportedContentType
	}
}

func (r *Request) unmarshalError(data []byte) error {
	switch r.contentType {
	case mimex.ContentTypeJSON:
		return json.Unmarshal(data, r.errResultPtr)
	default:
		return ErrUnsupportedContentType
	}
}

func (r *Request) unmarshalResult(data []byte) error {
	switch r.contentType {
	case mimex.ContentTypeJSON:
		return json.Unmarshal(data, r.resultPtr)
	default:
		return ErrUnsupportedContentType
	}
}

func (r *Request) SetHeaders(h nats.Header) *Request {
	maps.Copy(r.header, h)
	return r
}

func (r *Request) SetBody(b any) *Request {
	r.body = b
	return r
}

func (r *Request) SetRetry(count int, delay, maxDelay time.Duration) *Request {
	r.retry = count
	r.retryDelay = delay
	r.retryMaxDelay = maxDelay
	return r
}

func (r *Request) SetResult(ptr any) *Request {
	r.resultPtr = ptr
	return r
}

func (r *Request) SetError(ptr any) *Request {
	r.errResultPtr = ptr
	return r
}

func (r *Request) Exec(ctx context.Context, subject string) (*Request, error) {
	rResult := reflect.ValueOf(r.resultPtr)
	if rResult.IsValid() && rResult.Kind() != reflect.Pointer {
		return r, errors.New("invalid result type set")
	}
	rErrResult := reflect.ValueOf(r.errResultPtr)
	if rErrResult.IsValid() && rErrResult.Kind() != reflect.Pointer {
		return r, errors.New("invalid error result type set")
	}

	r.subject = subject

	bodyBytes, err := r.marshalBody()
	if err != nil {
		return r, err
	}

	if r.header.Get("Content-Type") == "" {
		r.header.Set("Content-Type", r.contentType)
	}

	timeout := r.timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var reply *nats.Msg
	var replyErr error
	if r.retry > 0 {
		replyErr = retry.Backoff(ctx, r.retry, r.retryDelay, r.retryMaxDelay, func(ctx context.Context) error {
			reply, replyErr = r.nc.RequestMsg(ctx, &nats.Msg{
				Subject: subject,
				Header:  r.header,
				Data:    bodyBytes,
			}, timeout)
			return replyErr
		})
	} else {
		reply, replyErr = r.nc.RequestMsg(ctx, &nats.Msg{
			Subject: subject,
			Header:  r.header,
			Data:    bodyBytes,
		}, timeout)
	}
	if replyErr != nil {
		return r, replyErr
	}

	if reflect.ValueOf(r.errResultPtr).IsValid() {
		status := reply.Header.Get(HeaderStatus)
		if status == StatusError {
			err := r.unmarshalError(reply.Data)
			if err != nil {
				return r, err
			}
			r.isError = true
			return r, nil
		}
	}

	if rResult.IsValid() {
		err = r.unmarshalResult(reply.Data)
		if err != nil {
			return r, err
		}
	}

	return r, nil
}

func (r *Request) IsError() bool {
	return r.isError
}

type ClientErrCode string

const (
	ErrBadRequest             ClientErrCode = "ERR_BAD_REQUEST"
	ErrMissingContentType     ClientErrCode = "ERR_MISSING_CONTENT_TYPE"
	ErrUnsupportedContentType ClientErrCode = "ERR_UNSUPPORTED_CONTENT_TYPE"
	ErrServerError            ClientErrCode = "ERR_SERVER_ERROR"
)

func (code ClientErrCode) Error() string {
	return string(code)
}
