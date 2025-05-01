package delivery

import (
	"context"
	"errors"
	"github.com/Olegsandrik/Exponenta/internal/repository/repoErrors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/usecaseErrors"
	"github.com/Olegsandrik/Exponenta/utils"
)

type AuthUsecase interface {
	Login(ctx context.Context, login string, password string) (string, error)
	IsLoggedIn(ctx context.Context, sID string) bool
	Logout(ctx context.Context, sID string) error
	GetUserByID(ctx context.Context, uID uint) (dto.User, error)
	SignUp(ctx context.Context, user dto.User) (uint, string, error)
	UpdatePassword(ctx context.Context, userID uint, password string, newPassword string) error
	UpdateUserName(ctx context.Context, userID uint, newUsername string) error
	UpdateUserLogin(ctx context.Context, userID uint, newLogin string) error
	UpdateUserSurname(ctx context.Context, userID uint, newUsername string) error
	DeleteProfile(ctx context.Context, userID uint) error
	IsVKUser(ctx context.Context, userID uint) bool
	LoginVK(ctx context.Context, data dto.VKLoginData) (string, error)
	GetUserLoginByID(ctx context.Context, userID uint) (string, error)
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
		h.router.Handle("/login", http.HandlerFunc(h.Login)).Methods("POST", "OPTIONS")
		h.router.Handle("/logout", http.HandlerFunc(h.Logout)).Methods("POST", "OPTIONS")
		h.router.Handle("/profile", http.HandlerFunc(h.Profile)).Methods("GET", "OPTIONS")
		h.router.Handle("/signup", http.HandlerFunc(h.Signup)).Methods("POST")
		h.router.Handle("/profile/edit/name",
			http.HandlerFunc(h.EditName)).Methods("POST", "OPTIONS")
		h.router.Handle("/profile/edit/surname",
			http.HandlerFunc(h.EditSurname)).Methods("POST", "OPTIONS")
		h.router.Handle("/profile/edit/password",
			http.HandlerFunc(h.EditPassword)).Methods("POST", "OPTIONS")
		h.router.Handle("/profile/edit/login",
			http.HandlerFunc(h.EditLogin)).Methods("POST", "OPTIONS")
		h.router.Handle("/profile/delete",
			http.HandlerFunc(h.DeleteProfile)).Methods("POST", "OPTIONS")
		h.router.Handle("/login/vk",
			http.HandlerFunc(h.LoginWithVK)).Methods("POST", "OPTIONS")
	}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signupData, err := dto.GetSignupData(r)
	if err != nil {
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
			Status: http.StatusBadRequest,
			Msg:    "invalid signup data",
			MsgRus: "некорректные данные регистрации",
		})
		return
	}

	_, sID, err := h.authUsecase.SignUp(ctx, signupData)
	if err != nil {
		if errors.Is(err, repoErrors.ErrUserWithThisLoginAlreadyExists) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    err.Error(),
				MsgRus: "пользователь с таким логином уже существует",
			})
			return
		} else if errors.Is(err, usecaseErrors.ErrTooEasyPassword) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "password is too easy",
				MsgRus: "пароль должен иметь длину не менее 8 символов, а также " +
					"содержать не менее 2 спецсимволов (!@#$&*)",
			})
			return
		}
		utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
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

	utils.JSONResponse(ctx, w, 200, utils.SuccessResponse{
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
		if errors.Is(err, repoErrors.ErrFailToGetUser) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    err.Error(),
				MsgRus: "пользователя с таким логином не существует",
			})
			return
		} else if errors.Is(err, usecaseErrors.ErrInvalidPassword) {
			utils.JSONResponse(ctx, w, 200, utils.ErrResponse{
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

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.authUsecase.GetUserByID(ctx, uID)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusUnauthorized,
			Msg:    err.Error(),
			MsgRus: "не получилось найти профиль пользователя",
		})
		return
	}

	user.IsVKUser = h.authUsecase.IsVKUser(ctx, uID)

	if !user.IsVKUser {
		login, err := h.authUsecase.GetUserLoginByID(ctx, uID)
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

func (h *AuthHandler) EditName(w http.ResponseWriter, r *http.Request) {
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

	err = h.authUsecase.UpdateUserName(ctx, uID, editData.NewName)
	if err != nil {
		if errors.Is(err, usecaseErrors.ErrEmptyName) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new name is required",
				MsgRus: "новое имя не найдено",
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

func (h *AuthHandler) EditLogin(w http.ResponseWriter, r *http.Request) {
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

	ok := h.authUsecase.IsVKUser(ctx, uID)
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

	err = h.authUsecase.UpdateUserLogin(ctx, uID, editData.NewLogin)
	if err != nil {
		if errors.Is(err, usecaseErrors.ErrEmptyLogin) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new login is required",
				MsgRus: "новый login не найден",
			})
			return
		} else if errors.Is(err, repoErrors.ErrLoginAlreadyUsed) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "login is already used",
				MsgRus: "этот логин использует другой пользователь",
			})
			return
		} else {
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

func (h *AuthHandler) EditSurname(w http.ResponseWriter, r *http.Request) {
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

	err = h.authUsecase.UpdateUserSurname(ctx, uID, editData.NewSurname)
	if err != nil {
		if errors.Is(err, usecaseErrors.ErrEmptySurname) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new surname is required",
				MsgRus: "новая фамилия не найдена",
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

func (h *AuthHandler) EditPassword(w http.ResponseWriter, r *http.Request) {
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

	ok := h.authUsecase.IsVKUser(ctx, uID)
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

	err = h.authUsecase.UpdatePassword(ctx, uID, editData.Password, editData.NewPassword)
	if err != nil {
		if errors.Is(err, usecaseErrors.ErrEmptyPassword) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "new password or password is required",
				MsgRus: "новый пароль не найден",
			})
			return
		} else if errors.Is(err, usecaseErrors.ErrInvalidPassword) {
			utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
				Status: http.StatusBadRequest,
				Msg:    "password is invalid",
				MsgRus: "пароль некорректен",
			})
			return
		} else if errors.Is(err, usecaseErrors.ErrTooEasyPassword) {
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

	err = h.authUsecase.Logout(ctx, cookie.Value)
	if err != nil {
		utils.JSONResponse(ctx, w, http.StatusOK, utils.ErrResponse{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
			MsgRus: "не получилось выйти из аккаунта пользователя",
		})
	}

	utils.JSONResponse(ctx, w, http.StatusOK, utils.SuccessResponse{
		Status: http.StatusOK,
		Data:   nil,
	})
}

func (h *AuthHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
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

	err = h.authUsecase.DeleteProfile(ctx, uID)
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
		if errors.Is(err, usecaseErrors.ErrEmptyVKLoginData) {
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
