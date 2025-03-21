package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Olegsandrik/Exponenta/logger"
)

type SuccessResponse struct {
	Status int
	Data   interface{}
}

type ErrResponse struct {
	Status int
	Msg    string
	MsgRus string
}

func JSONResponse(ctx context.Context, w http.ResponseWriter, statusCode int, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(message)
	if err != nil {
		logger.Error(ctx, "marshall error: "+err.Error())
	}
}
