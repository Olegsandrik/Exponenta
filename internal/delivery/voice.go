package delivery

import (
	"fmt"
	"github.com/Olegsandrik/Exponenta/logger"
	"io"
	"net/http"

	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/utils"
	"github.com/gorilla/mux"
)

type VoiceHandler struct {
	config *config.Config
	router *mux.Router
}

func NewVoiceHandler(cfg *config.Config) *VoiceHandler {
	return &VoiceHandler{cfg, mux.NewRouter()}
}

func (h *VoiceHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/voice").Subrouter()
	{
		h.router.Handle("", http.HandlerFunc(h.ServeHTTP)).Methods("GET")
	}
}

func (h *VoiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	APIURL := h.config.DeepSeekAPIURL
	APIKey := h.config.DeepSeekAPIKey
	voiceData, err := dto.GetVoiceData(r)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "Bad Request",
			MsgRus: "не найден text",
		})
		return
	}

	req, err := utils.BuildRequest(ctx, voiceData.Text, APIURL, APIKey)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Build request fail with text: %s , URL: %s", voiceData.Text, APIURL))
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "internal server error",
		})
		return
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Request failed: %v", err))
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "internal server error",
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, fmt.Sprintf("Do req fail with req: %v", req))
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "internal server error",
		})
		return
	}

	id, err := dto.ConvertVoiceData(ctx, resp.Body)

	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		logger.Error(ctx, fmt.Sprintf("Unexpected status code: %d, body: %s", resp.StatusCode, string(body)))
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "internal server error",
		})
		return
	}

	logger.Info(ctx, fmt.Sprintf("Text: %s, recognize like: %v", voiceData.Text, id))
	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: 200,
		Data:   id,
	})
}
