package middleware

import (
	"context"
	"net/http"
	"uuid"

	"github.com/Jordan-J-Brown/casino-go/pkg/contextx"
)

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextx.CtxKeyCorrelationID, uuid.New().String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
