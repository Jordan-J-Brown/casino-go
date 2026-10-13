package userhandler

import (
	"net/http"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/handlers/basehandler"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/userservice"
	"github.com/Jordan-J-Brown/casino-go/pkg/casino/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	basehandler.Base
	userService userservice.Service
}

func New(
	logger logsx.Logger,
	tracer trace.Tracer,
	userService userservice.Service,
) *Handler {
	return &Handler{
		Logger:      logger,
		Tracer:      tracer,
		userService: userService,
	}
}

// Balance returns the authenticated player's wallet balance.
// @Tags Users
// @Summary Get player balance
// @ID userBalance
// @Accept json
// @Produce json
// @Param request body requests.Balance true "Session identifier"
// @Success 200 {object} responses.Balance
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /users/balance [post]
func (handler *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	ctx, span := handler.Tracer.Start(r.Context(), "casino-users-balance")
	defer span.End()

	user := ctx.Value(usermodels.CtxKeyUser).(*usermodels.User)

	balance, err := handler.userService.GetUserBalance(ctx, user)
	if err != nil {
		handler.Logger.ErrorContext(ctx, "failed to get user's balance", "error", err)
		handler.HandleError(ctx, w, err)
		return
	}

	handler.ReplyJSON(w, r, responses.Balance{
		Balance: balance,
	})
}
