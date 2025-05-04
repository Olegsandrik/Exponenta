package delivery

import (
	"context"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/utils"
	"github.com/gorilla/mux"
	"net/http"
)

const collectionID = "collectionID"

type MainPageUsecase interface {
	GetRecipesByDishType(ctx context.Context, dishType string, page int) (dto.RecipePage, error)
	GetRecipesByDiet(ctx context.Context, diet string, page int) (dto.RecipePage, error)
	GetCollectionByID(ctx context.Context, collectionID int, page int) (dto.RecipePage, error)
	GetAllCollections(ctx context.Context) ([]dto.Collection, error)
}

type MainPageHandler struct {
	usecase MainPageUsecase
	router  *mux.Router
}

func NewMainPageHandler(usecase MainPageUsecase) *MainPageHandler {
	return &MainPageHandler{
		usecase: usecase,
		router:  mux.NewRouter(),
	}
}

func (h *MainPageHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/main").Subrouter()
	{
		h.router.Handle("/collection/all",
			http.HandlerFunc(h.GetCollections)).Methods("GET", "OPTIONS")
		h.router.Handle("/collection/{collectionID}",
			http.HandlerFunc(h.GetCollectionByID)).Methods("GET", "OPTIONS")
		h.router.Handle("/recipe/diet",
			http.HandlerFunc(h.GetRecipesByDiet)).Methods("GET", "OPTIONS")
		h.router.Handle("/recipe/types",
			http.HandlerFunc(h.GetRecipesByDishType)).Methods("GET", "OPTIONS")
	}
}

func (h *MainPageHandler) GetCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	collections, err := h.usecase.GetAllCollections(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не удалось получить коллекции рецептов",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   collections,
	})
}

func (h *MainPageHandler) GetCollectionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	collID, err := dto.GetIntURLParam(r, collectionID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр collectionID",
		})
		return
	}

	pageParam, err := dto.GetIntQueryParam(r, page)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "page query parameter error",
			MsgRus: "некорректный параметр page",
		})
		return
	}

	recipePage, err := h.usecase.GetCollectionByID(ctx, collID, pageParam)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не удалось получить рецепты из коллекции",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipePage,
	})
}

func (h *MainPageHandler) GetRecipesByDishType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pageParam, err := dto.GetIntQueryParam(r, page)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "page query parameter error",
			MsgRus: "некорректный параметр page",
		})
		return
	}

	dishType, err := dto.GetDishTypeData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "dishType parameter error",
			MsgRus: "требуется передать параметр dishType",
		})
		return
	}

	recipePage, err := h.usecase.GetRecipesByDishType(ctx, dishType, pageParam)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не удалось получить рецепты по типу блюда",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipePage,
	})
}

func (h *MainPageHandler) GetRecipesByDiet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pageParam, err := dto.GetIntQueryParam(r, page)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "page query parameter error",
			MsgRus: "некорректный параметр page",
		})
		return
	}

	diet, err := dto.GetDietData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "diet parameter error",
			MsgRus: "требуется передать параметр diet",
		})
		return
	}

	recipePage, err := h.usecase.GetRecipesByDiet(ctx, diet, pageParam)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не удалось получить рецепты по диете",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipePage,
	})
}
