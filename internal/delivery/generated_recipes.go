package delivery

import (
	"context"
	"errors"
	internalErrors "github.com/Olegsandrik/Exponenta/internal/errors"
	"github.com/Olegsandrik/Exponenta/internal/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
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
		h.router.Handle("/all", http.HandlerFunc(h.GetAllGeneratedRecipes)).Methods("GET")
		h.router.Handle("/{recipeID}/history",
			http.HandlerFunc(h.GetGeneratedRecipeHistoryByID)).Methods("GET")
		h.router.Handle("/{recipeID}", http.HandlerFunc(h.GetGeneratedRecipeByID)).Methods("GET")
		h.router.Handle("/make", http.HandlerFunc(h.CreateGeneratedRecipe)).Methods("POST")
		h.router.Handle("/{recipeID}/modern/{versionID}",
			http.HandlerFunc(h.UpgradeGeneratedRecipeByIDByVersion)).Methods("POST")
		h.router.Handle("/{recipeID}/main/{versionID}",
			http.HandlerFunc(h.SetNewMainVersionGeneratedRecipe)).Methods("POST")
		h.router.Handle("/{recipeID}/start",
			http.HandlerFunc(h.StartCookingGeneratedRecipe)).Methods("POST")
	}
}

func (h *GeneratedHandler) GetAllGeneratedRecipes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	numStr := r.URL.Query().Get("num")
	if numStr == "" {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "num not found",
			MsgRus: "num не найден",
		})
		return
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "номер должен быть целым числом",
		})
		return
	}

	recipesData, err := h.usecase.GetAllRecipes(ctx, num)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		} else if errors.Is(err, internalErrors.ErrZeroRowsGet) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    internalErrors.ErrZeroRowsGet.Error(),
				MsgRus: "на данный момент у вас нет сгенерированных рецептов",
			})
			return
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить сгенерированные рецепты",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipesData,
	})
}

func (h *GeneratedHandler) GetGeneratedRecipeByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDStr, ok := mux.Vars(r)["recipeID"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "recipeID not found",
			MsgRus: "recipeID не найден",
		})
		return
	}

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "recipeID должен быть целым числом",
		})
		return
	}

	recipeData, err := h.usecase.GetRecipeByID(ctx, recipeID)

	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить сгенерированный рецепт",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipeData,
	})
}

func (h *GeneratedHandler) CreateGeneratedRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	generatedRecipeData, err := dto.GetGenerationData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректные данные рецепта для генерации",
		})
		return
	}

	if generatedRecipeData.Ingredients == nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "errors with request",
			MsgRus: "некорректные данные рецепта для генерации",
		})
		return
	}

	generatedRecipe, err := h.usecase.CreateRecipe(ctx, generatedRecipeData.Ingredients, generatedRecipeData.Query)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		} else if errors.Is(err, internalErrors.ErrAllKeysAreUsing) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrAllKeysAreUsing.Error(),
				MsgRus: "На данный момент шеф занят, попробуйте позднее",
			})
			return
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось сгенерировать рецепт",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   generatedRecipe,
	})
}

func (h *GeneratedHandler) GetGeneratedRecipeHistoryByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDStr, ok := mux.Vars(r)["recipeID"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "recipeID not found",
			MsgRus: "recipeID не найден",
		})
		return
	}

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "recipeID должен быть целым числом",
		})
		return
	}

	history, err := h.usecase.GetHistoryByID(ctx, recipeID)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить историю",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   history,
	})
}

func (h *GeneratedHandler) UpgradeGeneratedRecipeByIDByVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDStr, ok := mux.Vars(r)["recipeID"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "recipeID not found",
			MsgRus: "recipeID не найден",
		})
		return
	}

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "recipeID должен быть целым числом",
		})
		return
	}

	versionIDStr, ok := mux.Vars(r)["versionID"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "versionID not found",
			MsgRus: "versionID не найден",
		})
		return
	}

	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "versionID должен быть целым числом",
		})
		return
	}

	generatedRecipeData, err := dto.GetGenerationData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "некорректные данные рецепта для генерации",
		})
		return
	}

	recipeData, err := h.usecase.UpdateRecipe(ctx, generatedRecipeData.Query, recipeID, versionID)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		} else if errors.Is(err, internalErrors.ErrAllKeysAreUsing) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrAllKeysAreUsing.Error(),
				MsgRus: "На данный момент шеф занят, попробуйте позднее",
			})
			return
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось сгенерировать рецепт",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipeData,
	})
}

func (h *GeneratedHandler) StartCookingGeneratedRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDStr, ok := mux.Vars(r)["recipeID"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "recipeID not found",
			MsgRus: "recipeID не найден",
		})
		return
	}

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "recipeID должен быть целым числом",
		})
		return
	}

	currentRecipeData, err := h.usecase.StartCookingByRecipeID(ctx, recipeID)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   currentRecipeData,
	})
}

func (h *GeneratedHandler) SetNewMainVersionGeneratedRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDStr, ok := mux.Vars(r)["recipeID"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "recipeID not found",
			MsgRus: "recipeID не найден",
		})
		return
	}

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "recipeID должен быть целым числом",
		})
		return
	}

	versionIDStr, ok := mux.Vars(r)["versionID"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "versionID not found",
			MsgRus: "versionID не найден",
		})
		return
	}

	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "versionID должен быть целым числом",
		})
		return
	}

	err = h.usecase.SetNewMainVersion(ctx, recipeID, versionID)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserNotAuth) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    internalErrors.ErrUserNotAuth.Error(),
				MsgRus: "пользователь не авторизован",
			})
			return
		} else if errors.Is(err, internalErrors.ErrVersionNotFound) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    internalErrors.ErrVersionNotFound.Error(),
				MsgRus: "данной версии не существует",
			})
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось установить версию как главную",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}
