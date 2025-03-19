package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/Olegsandrik/Exponenta/logger"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()

		newLogger := slog.Default().With(
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
		)

		ctx := logger.WithContext(r.Context(), newLogger)
		newReq := r.WithContext(ctx)

		next.ServeHTTP(w, newReq)
	})
}
