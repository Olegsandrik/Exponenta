package middleware

import (
	"fmt"
	"github.com/Olegsandrik/Exponenta/internal/utils"
	"net/http"

	"github.com/Olegsandrik/Exponenta/logger"
)

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(r.Context(), fmt.Sprintf("panic recovered: %+v", err))

					utils.JSONResponse(r.Context(), w, 200, utils.ErrResponse{
						Status: http.StatusInternalServerError,
						Msg:    "internal server error",
						MsgRus: "что-то пошло не так",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
}
