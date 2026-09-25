package natsx

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

func (err *ErrorResponse) Error() string {
	return err.Message
}

type HandlerFunc[Req, Res any] func(ctx context.Context, r Req) (*Res, error)
