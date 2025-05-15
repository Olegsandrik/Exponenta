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

type AuthUsecase interface {
	Login(ctx context.Context, login string, password string) (string, error)
	IsLoggedIn(ctx context.Context, sID string) bool
	Logout(ctx context.Context, sID string) error
	SignUp(ctx context.Context, user dto.User) (uint, string, error)
	DeleteProfile(ctx context.Context, userID uint) error
	IsVKUser(ctx context.Context, userID uint) bool
	LoginVK(ctx context.Context, data dto.VKLoginData) (string, error)
}

type AuthHandler struct {
	authUsecase AuthUsecase
	router      *mux.Router
}

func NewAuthHandler(authUsecase AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		router:      mux.NewRouter(),
	}
}

func (h *AuthHandler) InitRouter(r *mux.Router) {
	h.router = r.PathPrefix("/auth").Subrouter()
	{
		h.router.Handle("/login", http.HandlerFunc(h.Login)).Methods(http.MethodPost, http.MethodOptions)
		h.router.Handle("/logout", http.HandlerFunc(h.Logout)).Methods(http.MethodPost, http.MethodOptions)
		h.router.Handle("/signup", http.HandlerFunc(h.Signup)).Methods(http.MethodPost)
		h.router.Handle("/login/vk",
			http.HandlerFunc(h.LoginWithVK)).Methods(http.MethodPost, http.MethodOptions)
	}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signupData, err := dto.GetSignupData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid signup data",
			MsgRus: "некорректные данные регистрации",
		})
		return
	}

	_, sID, err := h.authUsecase.SignUp(ctx, signupData)
	if err != nil {
		if errors.Is(err, internalErrors.ErrUserWithThisLoginAlreadyExists) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    err.Error(),
				MsgRus: "пользователь с таким логином уже существует",
			})
			return
		} else if errors.Is(err, internalErrors.ErrTooEasyPassword) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "password is too easy",
				MsgRus: "пароль должен иметь длину не менее 8 символов, а также " +
					"содержать не менее 2 спецсимволов (!@#$&*)",
			})
			return
		} else if errors.Is(err, internalErrors.ErrTooShortUsername) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    internalErrors.ErrTooShortUsername.Error(),
				MsgRus: "имя пользователя должно содержать от 2 до 25 символов",
			})
			return
		} else if errors.Is(err, internalErrors.ErrTooShortSurname) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    internalErrors.ErrTooShortSurname.Error(),
				MsgRus: "фамилия пользователя должна содержать от 2 до 25 символов",
			})
			return
		}

		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось зарегистрироваться",
		})
		return
	}

	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sID,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &cookie)

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	loginData, err := dto.GetLoginData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid login data",
			MsgRus: "некорректные данные для авторизации",
		})
		return
	}

	sID, err := h.authUsecase.Login(ctx, loginData.Login, loginData.Password)
	if err != nil {
		if errors.Is(err, internalErrors.ErrFailToGetUser) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    err.Error(),
				MsgRus: "пользователя с таким логином не существует",
			})
			return
		} else if errors.Is(err, internalErrors.ErrInvalidPassword) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    err.Error(),
				MsgRus: "неверный пароль для данного логина",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось авторизоваться",
		})
		return
	}

	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sID,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &cookie)

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusUnauthorized,
				Msg:    "user not authenticated",
				MsgRus: "пользователь не авторизован",
			})
			return
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusUnauthorized,
			Msg:    "failed to get cookie",
			MsgRus: "ошибка получения cookie",
		})
		return
	}

	ok := h.authUsecase.IsLoggedIn(ctx, cookie.Value)
	if !ok {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusUnauthorized,
			Msg:    "user not authenticated",
			MsgRus: "пользователь не авторизован",
		})
		return
	}

	err = h.authUsecase.Logout(ctx, cookie.Value)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось выйти из аккаунта",
		})
		return
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *AuthHandler) LoginWithVK(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	loginVKData, err := dto.GetLoginVKData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid request",
			MsgRus: "некорректные данные для авторизации",
		})
		return
	}

	sID, err := h.authUsecase.LoginVK(ctx, loginVKData)
	if err != nil {
		if errors.Is(err, internalErrors.ErrEmptyVKLoginData) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "invalid request",
				MsgRus: "некорректные данные для авторизации",
			})
		}
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось авторизоваться",
		})
		return
	}

	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sID,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &cookie)

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}
