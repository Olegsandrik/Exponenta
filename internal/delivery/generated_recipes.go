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
	versionID = "versionID"
)

type GeneratedUsecase interface {
	GetAllRecipes(ctx context.Context, num int) ([]dto.RecipeDto, error)
	GetRecipeByID(ctx context.Context, recipeID int) (dto.RecipeDto, error)
	CreateRecipe(ctx context.Context, products []string, query string) (dto.RecipeDto, error)
	UpdateRecipe(ctx context.Context, query string, recipeID int, versionID int) (dto.RecipeDto, error)
	GetHistoryByID(ctx context.Context, recipeID int) ([]dto.RecipeDto, error)
	SetNewMainVersion(ctx context.Context, recipeID int, versionID int) error
	StartCookingByRecipeID(ctx context.Context, recipeID int) (dto.CurrentStepRecipeDto, error)
}

type GeneratedHandler struct {
	router  *mux.Router
	usecase GeneratedUsecase
}

func NewGeneratedHandler(usecase GeneratedUsecase) *GeneratedHandler {
	return &GeneratedHandler{
		mux.NewRouter(),
		usecase,
	}
}

func (h *GeneratedHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/generate").Subrouter()
	{
		h.router.Handle("/all", http.HandlerFunc(h.GetAllGeneratedRecipes)).Methods(http.MethodGet)
		h.router.Handle("/{recipeID}/history",
			http.HandlerFunc(h.GetGeneratedRecipeHistoryByID)).Methods(http.MethodGet)
		h.router.Handle("/{recipeID}", http.HandlerFunc(h.GetGeneratedRecipeByID)).Methods(http.MethodGet)
		h.router.Handle("/make", http.HandlerFunc(h.CreateGeneratedRecipe)).Methods(http.MethodPost)
		h.router.Handle("/{recipeID}/modern/{versionID}",
			http.HandlerFunc(h.UpgradeGeneratedRecipeByIDByVersion)).Methods(http.MethodPost)
		h.router.Handle("/{recipeID}/main/{versionID}",
			http.HandlerFunc(h.SetNewMainVersionGeneratedRecipe)).Methods(http.MethodPost)
		h.router.Handle("/{recipeID}/start",
			http.HandlerFunc(h.StartCookingGeneratedRecipe)).Methods(http.MethodPost)
	}
}

func (h *GeneratedHandler) GetAllGeneratedRecipes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	numParam, err := dto.GetIntQueryParam(r, num)

	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр num",
		})
		return
	}

	recipesData, err := h.usecase.GetAllRecipes(ctx, numParam)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		} else if errors.Is(err, internalErrors.ErrZeroRowsGet) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    internalErrors.ErrZeroRowsGet.Error(),
				MsgRus: "на данный момент у вас нет сгенерированных рецептов",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить сгенерированные рецепты",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipesData,
	})
}

func (h *GeneratedHandler) GetGeneratedRecipeByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recipeIDParam, err := dto.GetIntURLParam(r, recipeID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр recipeID",
		})
		return
	}

	recipeData, err := h.usecase.GetRecipeByID(ctx, recipeIDParam)

	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить сгенерированный рецепт",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipeData,
	})
}

func (h *GeneratedHandler) CreateGeneratedRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	generatedRecipeData, err := dto.GetGenerationData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректные данные рецепта для генерации",
		})
		return
	}

	if generatedRecipeData.Ingredients == nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "err with request",
			MsgRus: "некорректные данные рецепта для генерации",
		})
		return
	}

	generatedRecipe, err := h.usecase.CreateRecipe(ctx, generatedRecipeData.Ingredients, generatedRecipeData.Query)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		} else if errors.Is(err, internalErrors.ErrAllKeysAreUsing) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrAllKeysAreUsing.Error(),
				MsgRus: "На данный момент шеф занят, попробуйте позднее",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось сгенерировать рецепт",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   generatedRecipe,
	})
}

func (h *GeneratedHandler) GetGeneratedRecipeHistoryByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDParam, err := dto.GetIntURLParam(r, recipeID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр recipeID",
		})
		return
	}

	history, err := h.usecase.GetHistoryByID(ctx, recipeIDParam)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить историю",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   history,
	})
}

func (h *GeneratedHandler) UpgradeGeneratedRecipeByIDByVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDParam, err := dto.GetIntURLParam(r, recipeID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр recipeID",
		})
		return
	}

	versionIDParam, err := dto.GetIntURLParam(r, versionID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр versionID",
		})
		return
	}

	generatedRecipeData, err := dto.GetGenerationData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректные данные рецепта для генерации",
		})
		return
	}

	recipeData, err := h.usecase.UpdateRecipe(ctx, generatedRecipeData.Query, recipeIDParam, versionIDParam)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		} else if errors.Is(err, internalErrors.ErrAllKeysAreUsing) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrAllKeysAreUsing.Error(),
				MsgRus: "На данный момент шеф занят, попробуйте позднее",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось сгенерировать рецепт",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipeData,
	})
}

func (h *GeneratedHandler) StartCookingGeneratedRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDParam, err := dto.GetIntURLParam(r, recipeID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр recipeID",
		})
		return
	}

	currentRecipeData, err := h.usecase.StartCookingByRecipeID(ctx, recipeIDParam)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   currentRecipeData,
	})
}

func (h *GeneratedHandler) SetNewMainVersionGeneratedRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDParam, err := dto.GetIntURLParam(r, recipeID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр recipeID",
		})
		return
	}

	versionIDParam, err := dto.GetIntURLParam(r, versionID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректный параметр versionID",
		})
		return
	}

	err = h.usecase.SetNewMainVersion(ctx, recipeIDParam, versionIDParam)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		} else if errors.Is(err, internalErrors.ErrVersionNotFound) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    internalErrors.ErrVersionNotFound.Error(),
				MsgRus: "данной версии не существует",
			})
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось установить версию как главную",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}
