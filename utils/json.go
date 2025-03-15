package utils

import (
	"context"
	"encoding/json"
	"github.com/Olegsandrik/Exponenta/logger"
	"net/http"
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
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(message)
	if err != nil {
		logger.Error(ctx, "marshall error: "+err.Error())
	}
}
