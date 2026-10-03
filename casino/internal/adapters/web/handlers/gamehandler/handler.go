package gamehandler

import (
	"net/http"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/handlers/basehandler"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/gameservice"
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

// Setup returns game configuration and initial round state.
// @Tags Games
// @Summary Set up a game
// @ID gameSetup
// @Accept json
// @Produce json
// @Param request body requests.Setup true "Session and game identifier"
// @Success 200 {object} responses.Setup
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /games/setup [post]
func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "casino-game-setup")
	defer span.End()

	request, err := basehandler.DecodeJSON[requests.Setup](r.Body)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to decode setup request", "error", err)
		h.HandleError(ctx, w, err)
		return
	}

	user := ctx.Value(usermodels.CtxKeyUser).(*usermodels.User)

	result, setupErr := h.gameService.Setup(ctx, gameservice.SetupRequest{
		EngineID: request.GameID,
		User:     user,
	})
	if setupErr != nil {
		h.Logger.ErrorContext(ctx, "failed to retrieve setup result", "error", setupErr)
		h.HandleError(ctx, w, setupErr)
		return
	}

	h.ReplyJSON(w, r, responses.Setup{
		Bets:       result.Bets,
		GameConfig: result.GameConfig,
		MetaData:   result.MetaData,
		RoundData:  result.PublicState,
	})
}

// Play starts a game round.
// @Tags Games
// @Summary Play a game
// @ID gamePlay
// @Accept json
// @Produce json
// @Param request body requests.Play true "Session, game identifier, bet, and game data"
// @Success 200 {object} responses.Round
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /games/play [post]
func (h *Handler) Play(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "casino-game-play")
	defer span.End()

	request, err := basehandler.DecodeJSON[requests.Play](r.Body)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to decode play request", "error", err)
		h.HandleError(ctx, w, err)
		return
	}

	user := ctx.Value(usermodels.CtxKeyUser).(*usermodels.User)

	result, playErr := h.gameService.Play(ctx, gameservice.PlayRequest{
		EngineID: request.GameID,
		Bet:      request.Bet,
		User:     user,
		MetaData: request.GameData,
	})
	if playErr != nil {
		h.Logger.ErrorContext(ctx, "failed to retrieve play result from game service", "error", playErr)
		h.HandleError(ctx, w, playErr)
		return
	}

	h.ReplyJSON(w, r, responses.Round{
		Balance:          result.Balance,
		Bet:              result.Bet,
		PayoutMultiplier: result.PayoutMutliplier,
		RoundData:        result.PublicState,
	})
}

// Next advances the current game round.
// @Tags Games
// @Summary Advance a game
// @ID gameNext
// @Accept json
// @Produce json
// @Param request body requests.Next true "Session, game identifier, and game data"
// @Success 200 {object} responses.Round
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /games/next [post]
func (h *Handler) Next(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "casino-game-next")
	defer span.End()

	request, err := basehandler.DecodeJSON[requests.Next](r.Body)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to decode next request", "error", err)
		h.HandleError(ctx, w, err)
		return
	}

	user := ctx.Value(usermodels.CtxKeyUser).(*usermodels.User)

	result, nextErr := h.gameService.Next(ctx, gameservice.NextRequest{
		EngineID: request.GameID,
		User:     user,
		MetaData: request.GameData,
	})
	if nextErr != nil {
		h.Logger.ErrorContext(ctx, "failed to retrieve next result from game service", "error", nextErr)
		h.HandleError(ctx, w, nextErr)
		return
	}

	h.ReplyJSON(w, r, responses.Round{
		Balance:          result.Balance,
		Bet:              result.Bet,
		PayoutMultiplier: result.PayoutMutliplier,
		RoundData:        result.PublicState,
	})
}

// Close ends the current game session.
// @Tags Games
// @Summary Close a game
// @ID gameClose
// @Accept json
// @Produce json
// @Param request body requests.Close true "Session and game identifier"
// @Success 200 {object} responses.Close
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /games/close [post]
func (h *Handler) Close(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "casino-game-close")
	defer span.End()

	request, err := basehandler.DecodeJSON[requests.Close](r.Body)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to decode close request", "error", err)
		h.HandleError(ctx, w, err)
		return
	}

	user := ctx.Value(usermodels.CtxKeyUser).(*usermodels.User)

	result, closeErr := h.gameService.Close(ctx, gameservice.CloseRequest{
		EngineID: request.GameID,
		User:     user,
	})
	if closeErr != nil {
		h.Logger.ErrorContext(ctx, "failed to retrieve close result", "error", closeErr)
		h.HandleError(ctx, w, closeErr)
		return
	}

	h.ReplyJSON(w, r, responses.Close{
		Balance: result.Balance,
	})
}
