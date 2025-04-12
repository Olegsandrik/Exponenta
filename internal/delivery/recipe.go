package delivery

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/utils"
	"github.com/gorilla/mux"
)

type CookingRecipeUsecase interface {
	GetAllRecipe(context.Context, int) ([]dto.RecipeDto, error)
	GetRecipeByID(context.Context, int) (dto.RecipeDto, error)
	StartCookingRecipe(context.Context, int) (dto.CurrentStepRecipeDto, error)
	EndCookingRecipe(context.Context) error
	GetCurrentRecipe(context.Context) (dto.CurrentRecipeDto, error)
	NextStepRecipe(context.Context) (dto.CurrentStepRecipeDto, error)
	PreviousStepRecipe(context.Context) (dto.CurrentStepRecipeDto, error)
	AddTimerRecipe(context.Context, int, int) error
	DeleteTimerRecipe(context.Context, int) error
	GetTimersRecipe(context.Context) ([]dto.TimerRecipeDto, error)
}

type CookingRecipeHandler struct {
	router  *mux.Router
	usecase CookingRecipeUsecase
}

func NewCookingRecipeHandler(usecase CookingRecipeUsecase) *CookingRecipeHandler {
	return &CookingRecipeHandler{
		router:  mux.NewRouter(),
		usecase: usecase,
	}
}

func (h *CookingRecipeHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/recipe").Subrouter()
	{
		h.router.Handle("", http.HandlerFunc(h.GetCurrentRecipe)).Methods("GET")
		h.router.Handle("/all", http.HandlerFunc(h.GetAllRecipes)).Methods("GET")
		h.router.Handle("/timers", http.HandlerFunc(h.GetAllTimersCookingRecipe)).Methods("GET")
		h.router.Handle("/{id}", http.HandlerFunc(h.GetRecipeByID)).Methods("GET")
		h.router.Handle("/start", http.HandlerFunc(h.StartCookingRecipe)).Methods("POST")
		h.router.Handle("/end", http.HandlerFunc(h.EndCookingRecipe)).Methods("POST")
		h.router.Handle("/next", http.HandlerFunc(h.NextStepCookingRecipe)).Methods("POST")
		h.router.Handle("/prev", http.HandlerFunc(h.PrevStepCookingRecipe)).Methods("POST")
		h.router.Handle("/timer/add", http.HandlerFunc(h.AddTimerCookingRecipe)).Methods("POST")
		h.router.Handle("/timer/finish", http.HandlerFunc(h.FinishTimerCookingRecipe)).Methods("POST")
	}
}

func (h *CookingRecipeHandler) GetAllRecipes(w http.ResponseWriter, r *http.Request) {
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

	recipeData, err := h.usecase.GetAllRecipe(ctx, num)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить рецепты",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipeData,
	})
}

func (h *CookingRecipeHandler) GetRecipeByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeIDStr, ok := mux.Vars(r)["id"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "id not found",
			MsgRus: "id не найден",
		})
		return
	}

	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			MsgRus: "id должен быть целым числом",
		})
		return
	}

	recipe, err := h.usecase.GetRecipeByID(ctx, recipeID)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить рецепт",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipe,
	})
}

func (h *CookingRecipeHandler) GetCurrentRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recipeData, err := h.usecase.GetCurrentRecipe(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить рецепт",
		})
		return
	}
	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipeData,
	})
}

func (h *CookingRecipeHandler) StartCookingRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recipeData, err := dto.GetCookingRecipeData(r)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid recipe id",
			MsgRus: "некорректный recipe id",
		})
		return
	}

	if recipeData.ID == 0 {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "not found recipe id",
			MsgRus: "не найден recipe id",
		})
		return
	}

	currentRecipe, err := h.usecase.StartCookingRecipe(ctx, recipeData.ID)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось начать готовку",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   currentRecipe,
	})
}

func (h *CookingRecipeHandler) EndCookingRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.usecase.EndCookingRecipe(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось завершить рецепт",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *CookingRecipeHandler) NextStepCookingRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	nextStepData, err := h.usecase.NextStepRecipe(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось перейти к следующему шагу рецепта",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nextStepData,
	})
}

func (h *CookingRecipeHandler) PrevStepCookingRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	prevStepData, err := h.usecase.PreviousStepRecipe(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось перейти к предыдущему шагу рецетпа",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   prevStepData,
	})
}

func (h *CookingRecipeHandler) GetAllTimersCookingRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	timersData, err := h.usecase.GetTimersRecipe(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить список таймеров",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   timersData,
	})
}

func (h *CookingRecipeHandler) AddTimerCookingRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	TimerData, err := dto.GetTimerRecipeData(r)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid timer data",
			MsgRus: "некорректны данные таймера",
		})
		return
	}

	if TimerData.StepNum == 0 {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "step not found",
			MsgRus: "step не найден",
		})
		return
	}

	if TimerData.Time == 0 {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "time not found",
			MsgRus: "time не найден",
		})
		return
	}

	err = h.usecase.AddTimerRecipe(ctx, TimerData.StepNum, TimerData.Time)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось добавить таймер",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *CookingRecipeHandler) FinishTimerCookingRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	TimerData, err := dto.GetTimerRecipeData(r)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid timer data",
			MsgRus: "некорректны данные таймера",
		})
		return
	}

	if TimerData.StepNum == 0 {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "step not found",
			MsgRus: "step не найден",
		})
		return
	}

	err = h.usecase.DeleteTimerRecipe(ctx, TimerData.StepNum)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось завершить таймер",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}
