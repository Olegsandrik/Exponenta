package delivery

import (
	"context"
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/utils"
	"github.com/gorilla/mux"
	"net/http"
)

type CookingRecipeUsecase interface {
	GetAllRecipe(context.Context, string) ([]dto.RecipeDto, error)
	GetRecipeByID(context.Context, string) ([]dto.RecipeDto, error)
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

	recipeData, err := h.usecase.GetAllRecipe(ctx, numStr)
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

func (h *CookingRecipeHandler) GetCurrentRecipe(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
}

func (h *CookingRecipeHandler) GetRecipeByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeID, ok := mux.Vars(r)["id"]
	if !ok {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "id not found",
		})
		return
	}

	recipe, err := h.usecase.GetRecipeByID(ctx, recipeID)
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

func (h *CookingRecipeHandler) StartCookingRecipe(w http.ResponseWriter, r *http.Request) {

}

func (h *CookingRecipeHandler) EndCookingRecipe(w http.ResponseWriter, r *http.Request) {

}

func (h *CookingRecipeHandler) NextStepCookingRecipe(w http.ResponseWriter, r *http.Request) {

}

func (h *CookingRecipeHandler) PrevStepCookingRecipe(w http.ResponseWriter, r *http.Request) {

}

func (h *CookingRecipeHandler) GetAllTimersCookingRecipe(w http.ResponseWriter, r *http.Request) {

}

func (h *CookingRecipeHandler) AddTimerCookingRecipe(w http.ResponseWriter, r *http.Request) {

}

func (h *CookingRecipeHandler) FinishTimerCookingRecipe(w http.ResponseWriter, r *http.Request) {

}
