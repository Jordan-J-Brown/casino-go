package msg

import "context"

const (
	HeaderStatus = "Status"
)

type Status = string

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type HandlerFunc[Req, Res any] func(ctx context.Context, r Req) (*Res, error)
