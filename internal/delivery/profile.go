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

type ProfileUsecase interface {
	IsLoggedIn(ctx context.Context, sID string) bool
	Logout(ctx context.Context, sID string) error
	GetUserByID(ctx context.Context, uID uint) (dto.User, error)
	UpdatePassword(ctx context.Context, userID uint, password string, newPassword string) error
	UpdateUserName(ctx context.Context, userID uint, newUsername string) error
	UpdateUserLogin(ctx context.Context, userID uint, newLogin string) error
	UpdateUserSurname(ctx context.Context, userID uint, newUsername string) error
	DeleteProfile(ctx context.Context, userID uint) error
	IsVKUser(ctx context.Context, userID uint) bool
	GetUserLoginByID(ctx context.Context, userID uint) (string, error)
}

type ProfileHandler struct {
	profileUsecase ProfileUsecase
	router         *mux.Router
}

func NewProfileHandler(profileUsecase ProfileUsecase) *ProfileHandler {
	return &ProfileHandler{
		profileUsecase: profileUsecase,
		router:         mux.NewRouter(),
	}
}

func (h *ProfileHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/profile").Subrouter()
	{
		h.router.Handle("", http.HandlerFunc(h.Profile)).Methods(http.MethodGet, http.MethodOptions)
		h.router.Handle("/edit/name",
			http.HandlerFunc(h.EditName)).Methods(http.MethodPost, http.MethodOptions)
		h.router.Handle("/edit/surname",
			http.HandlerFunc(h.EditSurname)).Methods(http.MethodPost, http.MethodOptions)
		h.router.Handle("/edit/password",
			http.HandlerFunc(h.EditPassword)).Methods(http.MethodPost, http.MethodOptions)
		h.router.Handle("/edit/login",
			http.HandlerFunc(h.EditLogin)).Methods(http.MethodPost, http.MethodOptions)
		h.router.Handle("/delete",
			http.HandlerFunc(h.DeleteProfile)).Methods(http.MethodPost, http.MethodOptions)
	}
}

func (h *ProfileHandler) Profile(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.profileUsecase.GetUserByID(ctx, uID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusUnauthorized,
			Msg:    err.Error(),
			MsgRus: "не получилось найти профиль пользователя",
		})
		return
	}

	user.IsVKUser = h.profileUsecase.IsVKUser(ctx, uID)

	if !user.IsVKUser {
		login, err := h.profileUsecase.GetUserLoginByID(ctx, uID)
		if err != nil {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusInternalServerError,
				Msg:    err.Error(),
				MsgRus: "не получилось получить login пользователя",
			})
			return
		}
		user.Login = login
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   user,
	})
}

func (h *ProfileHandler) EditName(w http.ResponseWriter, r *http.Request) {
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

	editData, err := dto.GetEditData(r)

	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid edit data",
			MsgRus: "некорректные данные для обновления",
		})
		return
	}

	err = h.profileUsecase.UpdateUserName(ctx, uID, editData.NewName)
	if err != nil {
		if errors.Is(err, internalErrors.ErrEmptyName) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new name is required",
				MsgRus: "новое имя не найдено",
			})
			return
		} else if errors.Is(err, internalErrors.ErrTooShortUsername) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new name is too short",
				MsgRus: "новое имя должно содержать от 2 до 25 символов",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось обновить данные пользователя",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *ProfileHandler) EditLogin(w http.ResponseWriter, r *http.Request) {
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

	ok := h.profileUsecase.IsVKUser(ctx, uID)
	if ok {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "VK user can not edit login",
			MsgRus: "пользователь VK не может менять login",
		})
		return
	}

	editData, err := dto.GetEditData(r)

	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid edit data",
			MsgRus: "некорректные данные для обновления",
		})
		return
	}

	err = h.profileUsecase.UpdateUserLogin(ctx, uID, editData.NewLogin)
	if err != nil {
		if errors.Is(err, internalErrors.ErrEmptyLogin) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new login is required",
				MsgRus: "новый login не найден",
			})
			return
		} else if errors.Is(err, internalErrors.ErrLoginAlreadyUsed) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "login is already used",
				MsgRus: "этот логин использует другой пользователь",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось обновить данные пользователя",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *ProfileHandler) EditSurname(w http.ResponseWriter, r *http.Request) {
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

	editData, err := dto.GetEditData(r)

	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid edit data",
			MsgRus: "некорректные данные для обновления",
		})
		return
	}

	err = h.profileUsecase.UpdateUserSurname(ctx, uID, editData.NewSurname)
	if err != nil {
		switch {
		case errors.Is(err, internalErrors.ErrEmptySurname):
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new surname is required",
				MsgRus: "новая фамилия не найдена",
			})
			return
		case errors.Is(err, internalErrors.ErrTooShortSurname):
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new surname is too short",
				MsgRus: "новая фамилия должна содержать от 2 до 25 символов",
			})
			return
		default:
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusInternalServerError,
				Msg:    err.Error(),
				MsgRus: "не получилось обновить данные пользователя",
			})
			return
		}
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *ProfileHandler) EditPassword(w http.ResponseWriter, r *http.Request) {
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

	ok := h.profileUsecase.IsVKUser(ctx, uID)
	if ok {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "VK user can not edit password",
			MsgRus: "пользователь VK не может менять password",
		})
		return
	}

	editData, err := dto.GetEditData(r)

	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid edit data",
			MsgRus: "некорректные данные для обновления",
		})
		return
	}

	err = h.profileUsecase.UpdatePassword(ctx, uID, editData.Password, editData.NewPassword)
	if err != nil {
		if errors.Is(err, internalErrors.ErrEmptyPassword) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new password or password is required",
				MsgRus: "новый пароль не найден",
			})
			return
		} else if errors.Is(err, internalErrors.ErrInvalidPassword) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "password is invalid",
				MsgRus: "пароль некорректен",
			})
			return
		} else if errors.Is(err, internalErrors.ErrTooEasyPassword) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new password is too easy",
				MsgRus: "новый пароль должен иметь длину не менее 8 символов, а также " +
					"содержать не менее 2 спецсимволов (!@#$&*)",
			})
			return
		}

		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось обновить данные пользователя",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
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

	err = h.profileUsecase.DeleteProfile(ctx, uID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось удалить профиль пользователя",
		})
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}
