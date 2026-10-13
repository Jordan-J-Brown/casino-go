package basehandler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Jordan-J-Brown/casino-go/pkg/casino/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"go.opentelemetry.io/otel/trace"
)

type Base struct {
	Logger logsx.Logger
	Tracer trace.Tracer
}

func (base *Base) SetJsonContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func DecodeJSON[T any](r io.Reader) (*T, error) {
	value := new(T)
	err := json.NewDecoder(r).Decode(value)
	return value, err
}

func (base *Base) HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	if errX, ok := err.(*errorx.Error); ok {
		var statusCode int
		if errX.Code().IsClientError() {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusInternalServerError
		}

		errType := errX.Type()
		if errType == "" {
			errType = "UNKNOWN"
		}

		base.ReplyErrorJSON(ctx, w, errType, statusCode)

		return
	}

	base.ReplyErrorJSON(ctx, w, "UNKNOWN", http.StatusInternalServerError)
}

func (base *Base) ReplyErrorJSON(ctx context.Context, w http.ResponseWriter, errType string, httpCode int, msg ...string) {
	var message string
	if len(msg) == 1 {
		message = msg[0]
	}

	base.SetJsonContentType(w)

	w.WriteHeader(httpCode)

	err := json.NewEncoder(w).Encode(responses.Error{
		Type:    errType,
		Message: message,
	})
	if err != nil {
		base.Logger.ErrorContext(ctx, "failed to reply with error", "error", err)
	}
}

func (base *Base) ReplyJSON(w http.ResponseWriter, r *http.Request, response any) {
	base.SetJsonContentType(w)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		base.Logger.ErrorContext(r.Context(), "failed to reply with JSON",
			"route", r.URL.Path,
			"error", err,
		)
	}
}
