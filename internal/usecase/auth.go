package usecase

import (
	"context"
	"errors"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/utils"
)

type AuthRepo interface {
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

type AuthUsecase struct {
	repo AuthRepo
}

func NewAuthUsecase(repo AuthRepo) *AuthUsecase {
	return &AuthUsecase{repo: repo}
}

func (a *AuthUsecase) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := a.repo.GetUser(ctx, login)
	if err != nil {
		return "", err
	}
	if err = utils.CheckPassword(password, user.PasswordHash); err != nil {
		return "", err
	}
	return a.repo.CreateSession(ctx, user.ID)
}

func (a *AuthUsecase) IsLoggedIn(ctx context.Context, sID string) bool {
	return a.repo.SessionExists(ctx, sID)
}

func (a *AuthUsecase) Logout(ctx context.Context, sID string) error {
	return a.repo.DeleteSession(ctx, sID)
}

func (a *AuthUsecase) GetUserByID(ctx context.Context, uID uint) (dto.User, error) {
	userModel, err := a.repo.GetUserByID(ctx, uID)
	if err != nil {
		return dto.User{}, err
	}
	userDto := models.ConvertUserModelToDTO(userModel)

	return userDto, nil
}

func (a *AuthUsecase) SignUp(ctx context.Context, user dto.User) (uint, string, error) {
	if user.Name == "" || user.Login == "" || user.Password == "" || user.SurName == "" {
		return 0, "", errors.New("name or username or login or password is empty")
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

func (a *AuthUsecase) UpdatePassword(ctx context.Context, userID uint, password string, newPassword string) error {
	if password == "" || newPassword == "" {
		return errors.New("password or newPassword is empty")
	}

	prevPassword, err := a.repo.GetUserPassword(ctx, userID)
	if err != nil {
		return err
	}

	if err = utils.CheckPassword(password, prevPassword); err != nil {
		return err
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return a.repo.UpdateUser(ctx, "password_hash", passwordHash, userID)
}

func (a *AuthUsecase) UpdateUserName(ctx context.Context, userID uint, newUsername string) error {
	if newUsername == "" {
		return errors.New("new name is empty")
	}

	return a.repo.UpdateUser(ctx, "name", newUsername, userID)
}

func (a *AuthUsecase) UpdateUserSurname(ctx context.Context, userID uint, newUsername string) error {
	if newUsername == "" {
		return errors.New("new surname is empty")
	}

	return a.repo.UpdateUser(ctx, "sur_name", newUsername, userID)
}

func (a *AuthUsecase) UpdateUserLogin(ctx context.Context, userID uint, newLogin string) error {
	if newLogin == "" {
		return errors.New("new email is empty")
	}

	return a.repo.UpdateUser(ctx, "login", newLogin, userID)
}

func (a *AuthUsecase) DeleteProfile(ctx context.Context, userID uint) error {
	return a.repo.DeleteUser(ctx, userID)
}

func (a *AuthUsecase) IsVKUser(ctx context.Context, userID uint) bool {
	return a.repo.IsVKUser(ctx, userID)
}

func (a *AuthUsecase) LoginVK(ctx context.Context, data dto.VKLoginData) (string, error) {
	if data.DeviceID == "" || data.Code == "" || data.State == "" {
		return "", errors.New("device id or code or state is empty")
	}
	return a.repo.LoginVK(ctx, models.ConvertVKLoginDataDtoToModel(data))
}

func (a *AuthUsecase) GetUserLoginByID(ctx context.Context, userID uint) (string, error) {
	return a.repo.GetUserLoginByID(ctx, userID)
}
