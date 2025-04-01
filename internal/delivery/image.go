package delivery

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/utils"

	"github.com/gorilla/mux"
)

const (
	recipe    = "recipe"
	equipment = "equipment"
)

type ImageUsecase interface {
	GetImageByID(ctx context.Context, id int, entity string) (dto.Image, error)
}

type ImageHandler struct {
	router  *mux.Router
	usecase ImageUsecase
}

func NewImageHandler(usecase ImageUsecase) *ImageHandler {
	return &ImageHandler{
		mux.NewRouter(),
		usecase,
	}
}

func (h *ImageHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/image").Subrouter()
	{
		h.router.Handle("/recipe/{id}", http.HandlerFunc(h.GetRecipeImageByID)).Methods("GET")
		h.router.Handle("/equipment/{id}", http.HandlerFunc(h.GetEquipmentImageByID)).Methods("GET")
	}
}

func (h *ImageHandler) GetRecipeImageByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "id not found",
			MsgRus: "id не найден",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "id должен быть целым",
		})
		return
	}

	imageData, err := h.usecase.GetImageByID(ctx, id, recipe)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить фотографию рецепта",
		})
		return
	}

	if closer, ok := imageData.Image.(io.Closer); ok {
		defer closer.Close()
	}

	formatFile := strings.TrimPrefix(imageData.ContentType, "image/")
	filename := fmt.Sprintf("%s/%s.%s", recipe, idStr, formatFile)

	w.Header().Set("Content-Type", imageData.ContentType)
	http.ServeContent(w, r, filename, time.Now(), imageData.Image)
}

func (h *ImageHandler) GetEquipmentImageByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "id not found",
			MsgRus: "id не найден",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "id должен быть целым",
		})
		return
	}

	imageData, err := h.usecase.GetImageByID(ctx, id, equipment)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить фотографию рецепта",
		})
		return
	}

	if closer, ok := imageData.Image.(io.Closer); ok {
		defer closer.Close()
	}

	formatFile := strings.TrimPrefix(imageData.ContentType, "image/")
	filename := fmt.Sprintf("%s/%s.%s", equipment, idStr, formatFile)

	w.Header().Set("Content-Type", imageData.ContentType)
	http.ServeContent(w, r, filename, time.Now(), imageData.Image)
}
