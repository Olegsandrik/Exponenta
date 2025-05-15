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

type CookingHistoryUsecase interface {
	GetRecipesFromHistory(ctx context.Context, uID uint, page int) ([]dto.RecipeDto, error)
}

type CookingHistoryHandler struct {
	usecase CookingHistoryUsecase
	router  *mux.Router
}

func NewCookingHistoryHandler(usecase CookingHistoryUsecase) *CookingHistoryHandler {
	return &CookingHistoryHandler{
		usecase: usecase,
		router:  mux.NewRouter(),
	}
}

func (h *CookingHistoryHandler) InitRouter(router *mux.Router) {
	h.router = router.PathPrefix("/user").Subrouter()
	{
		h.router.Handle("/history", http.HandlerFunc(h.GetAllRecipesFromHistory)).Methods(http.MethodGet, http.MethodOptions)
	}
}

func (h *CookingHistoryHandler) GetAllRecipesFromHistory(w http.ResponseWriter, r *http.Request) {
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

	recipes, err := h.usecase.GetRecipesFromHistory(ctx, uID, pageParam)
	if err != nil {
		if errors.Is(err, internalErrors.ErrZeroRowsGet) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    err.Error(),
				MsgRus: "на данный момент вы еще не готовили",
			})
			return
		} else if errors.Is(err, internalErrors.ErrGetZeroRowsWithPageGreaterThanOne) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    err.Error(),
				MsgRus: "больше приготовленных рецептов у вас нет",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не удалось получить рецепты из истории",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipes,
	})
}
