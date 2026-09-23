package gamehandler

import (
	"context"
	"net/http"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/gameservice"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/web/handlers/basehandler"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	basehandler.Base
	gameService gameservice.Service
}

func New(
	logger logsx.Logger,
	tracer trace.Tracer,
	gameService gameservice.Service,
) *Handler {
	return &Handler{
		Logger:      logger,
		Tracer:      tracer,
		gameService: gameService,
	}
}

func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "casino-game-setup")
	defer span.End()

	request, err := h.DecodeJSON[requests.Setup](r.Body)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to decode setup request", "error", err)
		h.handleError(ctx, w, err)
		return
	}

	user := ctx.Value(usermodels.CtxKeyUser).(*usermodels.User)

	result, err := h.gameService.Setup(ctx, gameservice.SetupRequest{
		EngineID: request.GameID,
		User:     user,
	})
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to retrieve setup result", "error", err)
		h.handleError(ctx, w, err)
		return
	}

	h.ReplyJSON(w, r, responses.Setup{
		Bets:       result.Bets,
		GameConfig: result.GameConfig,
		MetaData:   result.MetaData,
		RoundData:  result.PublicState,
	})
}

func (h *Handler) Play(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "casino-game-play")
	defer span.End()

	request, err := h.DecodeJSON[requests.Play](r.Body)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to decode play request", "error", err)
		h.handleError(ctx, w, err)
		return
	}

	user := ctx.Value(usermodels.CtxKeyUser).(*usermodels.User)

	result, err := h.gameService.Play(ctx, gameservice.PlayRequest{
		EngineID: request.GameID,
		Bet:      request.Bet,
		User:     user,
		MetaData: request.GameData,
	})
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to retrieve play result from game service", "error", err)
		h.handleError(ctx, w, err)
		return
	}

	h.ReplyJSON(w, r, responses.Round{
		Balance:          result.Balance,
		PayoutMultiplier: result.PayoutMutliplier,
		RoundData:        result.PublicState,
	})
}

func (h *Handler) Next(w http.ResponseWriter, r *http.Request) {
	//TODO: implement
	panic("not implemented")
}

func (h *Handler) Close(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "casino-game-close")
	defer span.End()

	request, err := h.DecodeJSON[requests.Close](r.Body)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to decode close request", "error", err)
		h.ReplyInternalError(w)
		return
	}

	user := ctx.Value(usermodels.CtxKeyUser).(*usermodels.User)

	result, err := h.gameService.Close(ctx, gameservice.CloseRequest{
		EngineID: request.GameID,
		User:     user,
	})
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to retrieve close result", "error", err)
		h.handleError(ctx, w, err)
		return
	}

	h.ReplyJSON(w, r, responses.Close{
		Balance: result.Balance,
	})
}

func (h *Handler) handleError(ctx context.Context, w http.ResponseWriter, err error) {
	if gameError, ok := err.(*gamemodels.Error); ok {
		h.ReplyErrorJSON(ctx, w, string(gameError.Code), http.StatusBadRequest)
		return
	}

	h.ReplyErrorJSON(ctx, w, "UNKNOWN", http.StatusInternalServerError)
}
