package delivery

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/utils"
)

type ImageUsecase interface {
	GetImageByID(ctx context.Context, fileName string, entity string) (dto.Image, error)
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
		h.router.Handle("/{entity}/{filename}", http.HandlerFunc(h.GetImageByID)).Methods(http.MethodGet)
	}
}

func (h *ImageHandler) GetImageByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entity, ok := mux.Vars(r)["entity"]
	if !ok {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "entity not found",
			MsgRus: "не найден тип картинки",
		})
		return
	}

	filename, ok := mux.Vars(r)["filename"]
	if !ok {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "filename not found",
			MsgRus: "имя файла не найдено",
		})
		return
	}

	imageData, err := h.usecase.GetImageByID(ctx, filename, entity)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить изображение",
		})
		return
	}

	if closer, ok := imageData.Image.(io.Closer); ok {
		defer closer.Close()
	}

	w.Header().Set("Content-Type", imageData.ContentType)
	http.ServeContent(w, r, filename, time.Now(), imageData.Image)
}
