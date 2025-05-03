package usecase

import (
	"context"
	"regexp"

	"github.com/microcosm-cc/bluemonday"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	internalErrors "github.com/Olegsandrik/Exponenta/internal/internalerrors"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/internal/utils"
)

type UserRepo interface {
	CreateSession(ctx context.Context, uID uint) (string, error)
	DeleteSession(ctx context.Context, sID string) error
	SessionExists(ctx context.Context, sID string) bool
	GetUser(ctx context.Context, login string) (models.User, error)
	CreateUser(ctx context.Context, user models.User) (uint, error)
	DeleteUser(ctx context.Context, uID uint) error
	UpdateUser(ctx context.Context, entity string, newVal string, uID uint) error
	GetUserByID(ctx context.Context, userID uint) (models.User, error)
	GetUserPassword(ctx context.Context, userID uint) (string, error)
	IsVKUser(ctx context.Context, userID uint) bool
	LoginVK(ctx context.Context, data models.VKLoginData) (string, error)
	GetUserLoginByID(ctx context.Context, userID uint) (string, error)
}

type UserUsecase struct {
	repo UserRepo
}

func NewUserUsecase(repo UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (a *UserUsecase) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := a.repo.GetUser(ctx, login)
	if err != nil {
		return "", err
	}

	if err = utils.CheckPassword(password, user.PasswordHash); err != nil {
		return "", internalErrors.ErrInvalidPassword
	}
	return a.repo.CreateSession(ctx, user.ID)
}

func (a *UserUsecase) IsLoggedIn(ctx context.Context, sID string) bool {
	return a.repo.SessionExists(ctx, sID)
}

func (a *UserUsecase) Logout(ctx context.Context, sID string) error {
	return a.repo.DeleteSession(ctx, sID)
}

func (a *UserUsecase) GetUserByID(ctx context.Context, uID uint) (dto.User, error) {
	userModel, err := a.repo.GetUserByID(ctx, uID)
	if err != nil {
		return dto.User{}, err
	}
	userDto := models.ConvertUserModelToDTO(userModel)

	return userDto, nil
}

func (a *UserUsecase) SignUp(ctx context.Context, user dto.User) (uint, string, error) {
	err := utils.ValidateSignUpUserData(user)
	if err != nil {
		return 0, "", err
	}

	PasswordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return 0, "", err
	}

	userModel := models.User{
		Name:         user.Name,
		SurName:      user.SurName,
		Login:        user.Login,
		PasswordHash: PasswordHash,
	}

	uID, err := a.repo.CreateUser(ctx, userModel)
	if err != nil {
		return 0, "", err
	}

	sID, err := a.repo.CreateSession(ctx, uID)
	if err != nil {
		return 0, "", err
	}

	return uID, sID, nil
}

func (a *UserUsecase) UpdatePassword(ctx context.Context, userID uint, password string, newPassword string) error {
	if password == "" || newPassword == "" {
		return internalErrors.ErrEmptyPassword
	}

	prevPassword, err := a.repo.GetUserPassword(ctx, userID)
	if err != nil {
		return err
	}

	if err = utils.CheckPassword(password, prevPassword); err != nil {
		return internalErrors.ErrInvalidPassword
	}

	re := regexp.MustCompile(`[!@#$&*]`)

	if len(newPassword) < 8 || len(re.FindAllString(newPassword, -1)) < 2 {
		return internalErrors.ErrTooEasyPassword
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return a.repo.UpdateUser(ctx, "password_hash", passwordHash, userID)
}

func (a *UserUsecase) UpdateUserName(ctx context.Context, userID uint, newUsername string) error {
	sanitizer := bluemonday.UGCPolicy()

	newUsername = sanitizer.Sanitize(newUsername)
	if newUsername == "" {
		return internalErrors.ErrEmptyName
	}

	err := utils.ValidateName(newUsername)

	if err != nil {
		return err
	}

	return a.repo.UpdateUser(ctx, "name", newUsername, userID)
}

func (a *UserUsecase) UpdateUserSurname(ctx context.Context, userID uint, newUserSurName string) error {
	sanitizer := bluemonday.UGCPolicy()

	newUserSurName = sanitizer.Sanitize(newUserSurName)
	if newUserSurName == "" {
		return internalErrors.ErrEmptySurname
	}

	err := utils.ValidateSurname(newUserSurName)

	if err != nil {
		return err
	}

	return a.repo.UpdateUser(ctx, "sur_name", newUserSurName, userID)
}

func (a *UserUsecase) UpdateUserLogin(ctx context.Context, userID uint, newLogin string) error {
	sanitizer := bluemonday.UGCPolicy()

	newLogin = sanitizer.Sanitize(newLogin)

	if newLogin == "" {
		return internalErrors.ErrEmptyLogin
	}

	return a.repo.UpdateUser(ctx, "login", newLogin, userID)
}

func (a *UserUsecase) DeleteProfile(ctx context.Context, userID uint) error {
	return a.repo.DeleteUser(ctx, userID)
}

func (a *UserUsecase) IsVKUser(ctx context.Context, userID uint) bool {
	return a.repo.IsVKUser(ctx, userID)
}

func (a *UserUsecase) LoginVK(ctx context.Context, data dto.VKLoginData) (string, error) {
	if data.DeviceID == "" || data.Code == "" || data.State == "" {
		return "", internalErrors.ErrEmptyVKLoginData
	}
	return a.repo.LoginVK(ctx, models.ConvertVKLoginDataDtoToModel(data))
}

func (a *UserUsecase) GetUserLoginByID(ctx context.Context, userID uint) (string, error) {
	return a.repo.GetUserLoginByID(ctx, userID)
}
