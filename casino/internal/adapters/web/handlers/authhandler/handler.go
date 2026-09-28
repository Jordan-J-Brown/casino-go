package authhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/handlers/basehandler"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/userservice"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/responses"
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

//TODO: add real auth handling

// Demo auth handling, should just end up creating a new user+wallet with the balance and returning a session ID that links to the new user.
func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "casino-auth")
	defer span.End()

	request, err := h.Base.DecodeJSON[requests.Auth](r.Body)
	if err != nil {
		http.Error(w, "failed to unmarshal request data", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUserWithBalance(ctx, request.Username, request.Balance)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to create user", "request", request, "error", err)
		h.ReplyInternalError(w)
		return
	}
	sessionID, err := h.userService.CreateSessionForUser(ctx, user)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to create session for user", "error", err, "user", user)
		h.ReplyInternalError(w)
		return
	}

	h.ReplyJSON(w, r, responses.Auth{
		SessionID: sessionID,
	})
}

func (h *Handler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer(make([]byte, 0, r.ContentLength))
		_, err := io.Copy(buf, r.Body)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to read body", "error", err)
			h.ReplyInternalError(w)
			return
		}
		if err := r.Body.Close(); err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to close body", "error", err)
			h.ReplyInternalError(w)
			return
		}

		bodyReader := bytes.NewReader(buf.Bytes())
		r.Body = body{
			Reader: bodyReader,
		}

		var authSessionID authSessionID
		if err := json.NewDecoder(buf).Decode(&authSessionID); err != nil {
			h.Logger.ErrorContext(r.Context(), "failed to decode auth middleware data", "error", err)
			h.ReplyInternalError(w)
			return
		}

		user, err := h.userService.GetUserFromSessionID(r.Context(), authSessionID.SessionID)
		if err != nil {
			http.Error(w, "Not authenticated.", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), usermodels.CtxKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type body struct {
	*bytes.Reader
}

func (b body) Close() error { return nil }
