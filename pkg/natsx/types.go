package natsx

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
)

type Status = string

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Type       string `json:"type"`
	Message    string `json:"message,omitempty"`
}

func (err *ErrorResponse) Error() string {
	return err.Message
}

type HandlerFunc[Req, Res any] func(ctx context.Context, r Req) (*Res, *errorx.Error)
