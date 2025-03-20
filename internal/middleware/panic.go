package middleware

import (
	"fmt"
	"net/http"

	"github.com/Olegsandrik/Exponenta/logger"
	"github.com/Olegsandrik/Exponenta/utils"
)

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(r.Context(), fmt.Sprintf("panic recovered: %e", err))

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
