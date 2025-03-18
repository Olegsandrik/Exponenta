package delivery

import (
	"context"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/utils"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type CookingRecipeUsecase interface {
	GetAllRecipe(context.Context, int) ([]dto.RecipeDto, error)
	GetRecipeByID(context.Context, int) ([]dto.RecipeDto, error)
	StartCookingRecipe(context.Context, int) error
	EndCookingRecipe(context.Context) error
	GetCurrentRecipe(context.Context) (dto.CurrentRecipeDto, error)
	NextStepRecipe(context.Context) (dto.CurrentStepRecipeDto, error)
	PreviousStepRecipe(context.Context) (dto.CurrentStepRecipeDto, error)
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
		h.router.Handle("/{id}", http.HandlerFunc(h.GetRecipeByID)).Methods("GET")
		h.router.Handle("/start", http.HandlerFunc(h.StartCookingRecipe)).Methods("POST")
		h.router.Handle("/end", http.HandlerFunc(h.EndCookingRecipe)).Methods("POST")
		h.router.Handle("/next", http.HandlerFunc(h.NextStepCookingRecipe)).Methods("POST")
		h.router.Handle("/prev", http.HandlerFunc(h.PrevStepCookingRecipe)).Methods("POST")
		h.router.Handle("/timers", http.HandlerFunc(h.GetAllTimersCookingRecipe)).Methods("GET")
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
			Msg:    "id not found",
		})
		return
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "num must be int",
		})
		return
	}

	recipeData, err := h.usecase.GetAllRecipe(ctx, num)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "Failed to get recipes",
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
	recipeIdStr, ok := mux.Vars(r)["id"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "id not found",
		})
		return
	}

	recipeId, err := strconv.Atoi(recipeIdStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "id must be int",
		})
		return
	}

	recipe, err := h.usecase.GetRecipeByID(ctx, recipeId)
	if err != nil || len(recipe) == 0 {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "Failed to get recipe",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   recipe[0],
	})
}

func (h *CookingRecipeHandler) GetCurrentRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recipeData, err := h.usecase.GetCurrentRecipe(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "Failed to get recipe",
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
	idStr := r.FormValue("id")
	if idStr == "" {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "id not found",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "id must be int",
		})
		return
	}

	err = h.usecase.StartCookingRecipe(ctx, id)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "failed to start recipe",
		})
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *CookingRecipeHandler) EndCookingRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.usecase.EndCookingRecipe(ctx)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    "failed to end recipe",
		})
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
			Msg:    "Failed next step recipe",
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
			Msg:    "Failed previous step recipe",
		})
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   prevStepData,
	})
}

func (h *CookingRecipeHandler) GetAllTimersCookingRecipe(w http.ResponseWriter, r *http.Request) {

}

func (h *CookingRecipeHandler) AddTimerCookingRecipe(w http.ResponseWriter, r *http.Request) {

}

func (h *CookingRecipeHandler) FinishTimerCookingRecipe(w http.ResponseWriter, r *http.Request) {

}
