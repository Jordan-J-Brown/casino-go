package natsx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jordan-J-Brown/casino-go/pkg/contextx"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func HandleRPC[Req, Res any](
	ctx context.Context,
	logger logsx.Logger,
	tracer trace.Tracer,
	handler HandlerFunc[Req, Res],
) nats.MsgHandler {
	return func(msg *nats.Msg) {
		respond := func(status Status, response any) {
			var err error
			defer func() {
				if err != nil {
					logger.ErrorContext(ctx, "failed to respond to message",
						"subject", msg.Subject,
						"reply_subject", msg.Reply,
						"error", err.Error(),
					)
				}
			}()

			header := nats.Header{
				"Status": []string{status},
			}
			otel.GetTextMapPropagator().Inject(ctx, &NatsOtelHeaderCarrier{header})

			data, err := json.Marshal(response)
			if err != nil {
				return
			}

			err = msg.RespondMsg(&nats.Msg{
				Header: header,
				Data:   data,
			})
		}

		defer func() {
			if r := recover(); r != nil {
				respond(StatusError, ErrorResponse{
					StatusCode: http.StatusInternalServerError,
					Message:    fmt.Sprintf("Panic recover: %v", r),
				})
			}
		}()

		cid := msg.Header.Get(HeaderCorrelationID)
		if cid != "" {
			ctx = context.WithValue(ctx, contextx.CtxKeyCorrelationID, cid)
		}

		ctx = otel.GetTextMapPropagator().Extract(ctx, &NatsOtelHeaderCarrier{msg.Header})
		ctx, span := tracer.Start(ctx, msg.Subject)
		defer span.End()

		var request Req
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			respond(StatusError, ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    "failed to unmarshal message data: " + err.Error(),
			})
			return
		}

		response, err := handler(ctx, request)
		if err != nil {
			respond(StatusError, ErrorResponse{
				StatusCode: StatusCodeFromError(err),
				Type:       err.Type(),
				Message:    "error from message handler: " + err.Error(),
			})
			return
		}

		respond(StatusSuccess, response)
	}
}

func StatusCodeFromError(err *errorx.Error) int {
	switch err.Code() {
	case errorx.Unknown:
		return http.StatusInternalServerError
	case errorx.ClientError:
		return http.StatusBadRequest
	case errorx.NotFound:
		return http.StatusNotFound
	case errorx.ServerError:
		return http.StatusInternalServerError
	default:
		panic("unexpected errorx.Code")
	}
}
