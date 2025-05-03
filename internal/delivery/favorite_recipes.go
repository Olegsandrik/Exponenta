package delivery

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	internalErrors "github.com/Olegsandrik/Exponenta/internal/internalerrors"
	"github.com/Olegsandrik/Exponenta/internal/utils"
)

const (
	recipeID = "recipeID"
	page     = "page"
)

type FavoriteRecipesUsecase interface {
	AddRecipeToFavorite(ctx context.Context, userID uint, recipeID int) error
	DeleteRecipeFromFavorite(ctx context.Context, userID uint, recipeID int) error
	GetFavoriteRecipes(ctx context.Context, userID uint, page int) ([]dto.RecipeDto, error)
}

type FavoriteRecipesHandler struct {
	favoriteRecipesUsecase FavoriteRecipesUsecase
	router                 *mux.Router
}

func NewFavoriteRecipesHandler(favoriteRecipesUsecase FavoriteRecipesUsecase) *FavoriteRecipesHandler {
	return &FavoriteRecipesHandler{
		favoriteRecipesUsecase: favoriteRecipesUsecase,
		router:                 mux.NewRouter(),
	}
}

func (h *FavoriteRecipesHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/favorite").Subrouter()
	{
		h.router.Handle("/all",
			http.HandlerFunc(h.GetAllFavoriteRecipes)).Methods("GET", "OPTIONS")
		h.router.Handle("/add/{recipeID}",
			http.HandlerFunc(h.AddRecipeToFavorite)).Methods("POST", "OPTIONS")
		h.router.Handle("/delete/{recipeID}",
			http.HandlerFunc(h.DeleteRecipeFromFavorite)).Methods("POST", "OPTIONS")
	}
}

func (h *FavoriteRecipesHandler) GetAllFavoriteRecipes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusUnauthorized,
			Msg:    "user not authenticated",
			MsgRus: "пользователь не авторизован",
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

	recipes, err := h.favoriteRecipesUsecase.GetFavoriteRecipes(ctx, uID, pageParam)
	if err != nil {
		if errors.Is(err, internalErrors.ErrZeroRowsGet) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    err.Error(),
				MsgRus: "на данный момент у вас нет никаких рецептов в избранном",
			})
			return
		} else if errors.Is(err, internalErrors.ErrGetZeroRowsWithPageGreaterThanOne) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    err.Error(),
				MsgRus: "больше избранных рецептов у вас нет",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не удалось получить избранные рецепты",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipes,
	})
}

func (h *FavoriteRecipesHandler) AddRecipeToFavorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusUnauthorized,
			Msg:    "user not authenticated",
			MsgRus: "пользователь не авторизован",
		})
		return
	}

	rID, err := dto.GetIntURLParam(r, recipeID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр recipeID",
		})
		return
	}

	err = h.favoriteRecipesUsecase.AddRecipeToFavorite(ctx, uID, rID)
	if err != nil {
		if errors.Is(err, internalErrors.ErrDuplicateRow) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    err.Error(),
				MsgRus: "рецепт уже в избранном",
			})
			return
		} else if errors.Is(err, internalErrors.ErrRecipeWithThisIDDoesNotExist) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    err.Error(),
				MsgRus: "рецепта с таким id не существует",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не удалось добавить рецепт в избранное",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *FavoriteRecipesHandler) DeleteRecipeFromFavorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusUnauthorized,
			Msg:    "user not authenticated",
			MsgRus: "пользователь не авторизован",
		})
		return
	}

	rID, err := dto.GetIntURLParam(r, recipeID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр recipeID",
		})
	}

	err = h.favoriteRecipesUsecase.DeleteRecipeFromFavorite(ctx, uID, rID)
	if err != nil {
		if errors.Is(err, internalErrors.ErrZeroRowsDeleted) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    err.Error(),
				MsgRus: "данный рецепт отсутствует в избранных",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось удалить рецепт из избранных",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}
