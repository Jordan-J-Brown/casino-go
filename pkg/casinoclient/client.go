package casinoclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
)

const (
	DefaultBaseURL string = "http://localhost:10000/api/v1"
)

var (
	DefaultClient = &Client{
		hclient: http.DefaultClient,
		baseURL: DefaultBaseURL,
	}
)

type Client struct {
	hclient *http.Client
	baseURL string
}

func New(
	baseURL string,
	httpClient *http.Client,
) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL: baseURL,
		hclient: httpClient,
	}
}

func NewDefaultClient() *Client {
	//Copies the defaults from stdlib

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return &Client{
		baseURL: DefaultBaseURL,
		hclient: &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           dialer.DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (c *Client) Auth(r requests.Auth) (*Response[responses.Auth], error) {
	return c.post[requests.Auth, responses.Auth](c.baseURL+"/auth", r)
}

func (c *Client) Balance(r requests.Balance) (*Response[responses.Balance], error) {
	return c.post[requests.Balance, responses.Balance](c.baseURL+"/users/balance", r)
}

func (c *Client) MustGetBalance(r requests.Balance) *Response[responses.Balance] {
	res, err := c.Balance(r)
	if err != nil {
		panic(err)
	}
	if res.IsError {
		panic(fmt.Sprint("error response:", res.Error))
	}
	return res
}

func (c *Client) Setup(r requests.Setup) (*Response[responses.Setup], error) {
	return c.post[requests.Setup, responses.Setup](c.baseURL+"/games/setup", r)
}

func (c *Client) Play(r requests.Play) (*Response[responses.Round], error) {
	return c.post[requests.Play, responses.Round](c.baseURL+"/games/play", r)
}

func (c *Client) MustPlay(r requests.Play) *Response[responses.Round] {
	res, err := c.Play(r)
	if err != nil {
		panic(err)
	}
	if res.IsError {
		panic(fmt.Sprint("error response:", res.Error))
	}
	return res
}

func (c *Client) Next(r requests.Next) (*Response[responses.Round], error) {
	return c.post[requests.Next, responses.Round](c.baseURL+"/games/next", r)
}

func (c *Client) MustNext(r requests.Next) *Response[responses.Round] {
	res, err := c.Next(r)
	if err != nil {
		panic(err)
	}
	if res.IsError {
		panic(fmt.Sprint("error response:", res.Error))
	}

	return res
}

func (c *Client) Close(r requests.Close) (*Response[responses.Close], error) {
	return c.post[requests.Close, responses.Close](c.baseURL+"/games/close", r)
}

func (c *Client) MustClose(r requests.Close) *Response[responses.Close] {
	res, err := c.Close(r)
	if err != nil {
		panic(err)
	}
	if res.IsError {
		panic(fmt.Sprint("error response:", res.Error))
	}

	return res
}

func (c *Client) post[Req, Res any](url string, r Req) (*Response[Res], error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	resp, err := c.hclient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, errorx.Newf("failed to post data to: %s", url, errorx.WithChildren(err))
	}
	defer resp.Body.Close()

	res := Response[Res]{
		HttpResponse: resp,
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&res.Result)
		if err != nil {
			return nil, errorx.New("failed to decode resul", errorx.WithChildren(err))
		}
	} else {
		res.IsError = true
		err := json.NewDecoder(resp.Body).Decode(&res.Error)
		if err != nil {
			return nil, errorx.Newf("failed to decode error. Status: %s", res.HttpResponse.Status, errorx.WithChildren(err))
		}
	}

	return &res, nil
}

type Response[R any] struct {
	HttpResponse *http.Response
	IsError      bool
	Error        responses.Error
	Result       R
}
