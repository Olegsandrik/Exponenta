package delivery

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/utils"

	"github.com/gorilla/mux"
)

type SearchUsecase interface {
	Search(ctx context.Context, query string, diet string, dishType string, maxTime int) (dto.SearchResponseDto, error)
	Suggest(ctx context.Context, query string) (dto.SuggestResponseDto, error)
	GetFilter(ctx context.Context) (dto.FiltersDto, error)
}

type SearchHandler struct {
	router  *mux.Router
	usecase SearchUsecase
}

func NewSearchHandler(usecase SearchUsecase) *SearchHandler {
	return &SearchHandler{
		router:  mux.NewRouter(),
		usecase: usecase,
	}
}

func (h *SearchHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/search").Subrouter()
	{
		h.router.HandleFunc("", h.Search).Methods("GET")
		h.router.HandleFunc("/suggest", h.Suggest).Methods("GET")
		h.router.HandleFunc("/filters", h.GetAllFilters).Methods("GET")
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("query")
	diet := r.URL.Query().Get("diet")
	dishType := r.URL.Query().Get("dishType")
	maxTimeStr := r.URL.Query().Get("maxTime")

	if query == "" {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "query parameter not found",
			MsgRus: "не найдена строчка поискового запроса",
		})
		return
	}

	maxTime, err := strconv.Atoi(maxTimeStr)

	if err != nil && maxTimeStr != "" {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "max time must be a int",
			MsgRus: "максимальное время должно быть целым числом",
		})
		return
	}

	searchResponse, err := h.usecase.Search(ctx, query, diet, dishType, maxTime)

	if err != nil {
		if errors.Is(err, utils.ErrNoFound) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusNotFound,
				Msg:    "results not found",
				MsgRus: "по запросу ничего не найдено",
			})
			return
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось произвести поиск",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   searchResponse,
	})
}

func (h *SearchHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("query")
	if query == "" {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "query parameter not found",
			MsgRus: "не найдена строчка поискового запроса",
		})
		return
	}

	suggestResponse, err := h.usecase.Suggest(ctx, query)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось найти подсказку",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   suggestResponse,
	})
}

func (h *SearchHandler) GetAllFilters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filtersData, err := h.usecase.GetFilter(ctx)

	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось получить фильтры",
		})
		return
	}

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   filtersData,
	})
}
