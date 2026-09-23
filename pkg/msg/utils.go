package msg

import (
	"context"
	"encoding/json"

	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/nats-io/nats.go"
)

func HandleRPC[Req, Res any](ctx context.Context, logger logsx.Logger, handler HandlerFunc[Req, Res]) nats.MsgHandler {
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

			data, err := json.Marshal(response)
			if err != nil {
				return
			}

			err = msg.RespondMsg(&nats.Msg{
				Header: nats.Header{
					"Status": []string{status},
				},
				Data: data,
			})
		}

		//TODO: retrieve otel tracing data from nats headers

		var request Req
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			respond(StatusError, ErrorResponse{
				Message: "failed to unmarshal message data: " + err.Error(),
			})
			return
		}

		response, err := handler(ctx, request)
		if err != nil {
			respond(StatusError, ErrorResponse{
				Message: "error from message handler: " + err.Error(),
			})
			return
		}

		respond(StatusSuccess, response)
	}
}
