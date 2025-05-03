package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Olegsandrik/Exponenta/internal/utils"
)

type UserRepo interface {
	GetUserIDBySessionID(ctx context.Context, sessionID string) (uint, error)
}

func NewAuthMiddleware(adapter UserRepo) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cookie, err := r.Cookie("session_id")
			if err != nil {
				ctx = context.WithValue(ctx, utils.UserID{}, 0)
			} else {
				userID, err := adapter.GetUserIDBySessionID(ctx, cookie.Value)
				if err != nil {
					ctx = context.WithValue(ctx, utils.UserID{}, 0)
				}
				ctx = context.WithValue(ctx, utils.UserID{}, userID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
