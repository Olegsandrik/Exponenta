package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"

	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/internal/repository/dao"
	"github.com/Olegsandrik/Exponenta/internal/repository/repoErrors"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
	"github.com/Olegsandrik/Exponenta/utils"
)

const (
	APIURL = "https://api.vk.com/method/users.get?fields=photo_50,about&access_token=%s&v=5.131"
)

type RedisAdapter interface {
	Get(key string) (uint, error)
	Set(key string, value uint) error
	Delete(key string) error
}

type PostgresAdapter interface {
	Exec(ctx context.Context, q string, args ...interface{}) (sql.Result, error)
	Select(ctx context.Context, dest interface{}, q string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type AuthRepo struct {
	RedisAdapter    RedisAdapter
	PostgresAdapter PostgresAdapter
	Config          *config.Config
}

func NewAuthRepo(redisAdapter RedisAdapter, postgresAdapter PostgresAdapter, config *config.Config) *AuthRepo {
	return &AuthRepo{
		RedisAdapter:    redisAdapter,
		PostgresAdapter: postgresAdapter,
		Config:          config,
	}
}

func (repo *AuthRepo) CreateSession(ctx context.Context, uID uint) (string, error) {
	timeStart := time.Now()
	sID := uuid.New().String()

	err := repo.RedisAdapter.Set(sID, uID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to create session %v", err))
		return "", repoErrors.ErrFailToCreateSession
	}

	logger.Info(ctx, fmt.Sprintf("create sessionID %s", time.Since(timeStart)))
	return sID, nil
}

func (repo *AuthRepo) DeleteSession(ctx context.Context, sID string) error {
	timeStart := time.Now()
	err := repo.RedisAdapter.Delete(sID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to delete session %v", err))
		return repoErrors.ErrFailToDeleteSession
	}
	logger.Info(ctx, fmt.Sprintf("delete userID by sessionID in %s", time.Since(timeStart)))
	return nil
}

func (repo *AuthRepo) SessionExists(ctx context.Context, sID string) bool {
	timeStart := time.Now()
	uID, err := repo.RedisAdapter.Get(sID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get session %v", err))
		return false
	}

	logger.Info(ctx, fmt.Sprintf("get userID by sessionID in %s", time.Since(timeStart)))
	return uID != 0
}

func (repo *AuthRepo) GetUserIDBySessionID(ctx context.Context, sID string) (uint, error) {
	timeStart := time.Now()
	uID, err := repo.RedisAdapter.Get(sID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get userID by sessionID %v", err))
		return 0, repoErrors.ErrFailToUserIDBySessionID
	}

	logger.Info(ctx, fmt.Sprintf("get userID by sessionID in %s", time.Since(timeStart)))
	return uID, nil
}

func (repo *AuthRepo) GetUser(ctx context.Context, login string) (models.User, error) {
	q := "SELECT id, password_hash FROM Users WHERE login = $1"

	var userTable []dao.User

	err := repo.PostgresAdapter.Select(ctx, &userTable, q, login)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get user rows: %v", err))
		return models.User{}, repoErrors.ErrFailToGetUser
	}

	if len(userTable) == 0 {
		logger.Error(ctx, fmt.Sprintf("zero value get with login: %s", login))
		return models.User{}, repoErrors.ErrFailToGetUser
	}

	userModel := dao.ConvertUserTableToModel(userTable)

	logger.Info(ctx, fmt.Sprintf("success get user with login: %s", login))

	return userModel[0], nil
}

func (repo *AuthRepo) CreateUser(ctx context.Context, user models.User) (uint, error) {
	var userID uint
	q := `INSERT INTO Users(name, sur_name, login, password_hash) VALUES ($1, $2, $3, $4) returning id`

	err := repo.PostgresAdapter.QueryRow(ctx, q, user.Name, user.SurName, user.Login, user.PasswordHash).Scan(&userID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to create user: %v", err))
		return 0, repoErrors.ErrFailToCreateUser
	}

	return userID, nil
}

func (repo *AuthRepo) IsExistsUserVK(ctx context.Context, VKID uint) (bool, uint) {
	var userID uint
	q := `SELECT id FROM Users WHERE vk_id = $1`

	err := repo.PostgresAdapter.QueryRow(ctx, q, VKID).Scan(&userID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get id: %v, by VKID: %d", err, VKID))
		return false, 0
	}

	logger.Info(ctx, fmt.Sprintf("success get id: %v with VKID: %d", userID, VKID))

	return userID != 0, userID
}

func (repo *AuthRepo) CreateUserVK(ctx context.Context, user models.UserVK) (uint, error) {
	var userID uint
	q := `INSERT INTO Users(vk_id, name, sur_name) VALUES ($1, $2, $3) returning id`

	err := repo.PostgresAdapter.QueryRow(ctx, q, user.VKID, user.Name, user.SurName).Scan(&userID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to create user: %v", err))
		return 0, repoErrors.ErrFailToCreateUser
	}

	return userID, nil
}

func (repo *AuthRepo) DeleteUser(ctx context.Context, uID uint) error {
	result, err := repo.PostgresAdapter.Exec(ctx, "DELETE FROM Users WHERE id = $1", uID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to delete user: %v", err))
		return repoErrors.ErrFailToDeleteUser
	}

	count, err := result.RowsAffected()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get affected rows user: %v", err))
		return repoErrors.ErrFailToDeleteUser
	}

	if count == 0 {
		logger.Info(ctx, fmt.Sprintf("zero rows affected: %v", err))
		return repoErrors.ErrFailToDeleteUser
	}

	logger.Info(ctx, fmt.Sprintf("delete user with id %d", uID))
	return nil
}

func (repo *AuthRepo) UpdateUser(ctx context.Context, entity string, newVal string, uID uint) error {
	q := fmt.Sprintf("UPDATE Users SET %s = $1 WHERE id = $2", entity)
	result, err := repo.PostgresAdapter.Exec(ctx, q, newVal, uID)
	if err != nil {
		if entity == "login" {
			logger.Error(ctx, fmt.Sprintf("fail to update user login: %v", err))
			return repoErrors.ErrLoginAlreadyUsed
		}
		logger.Error(ctx, fmt.Sprintf("fail to update user: %v", err))
		return repoErrors.ErrFailToUpdateUser
	}

	count, err := result.RowsAffected()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf(
			"fail to get affected rows user: %v, entity: %s, uID: %d",
			err,
			entity,
			uID,
		))
		return repoErrors.ErrFailToUpdateUser
	}

	if count == 0 {
		logger.Info(ctx, fmt.Sprintf("zero rows affected: %v, entity: %s, uID: %d",
			err,
			entity,
			uID,
		))
		return repoErrors.ErrFailToUpdateUser
	}

	logger.Info(ctx, fmt.Sprintf("update user %s with id %d", entity, uID))

	return nil
}

func (repo *AuthRepo) GetUserLoginByID(ctx context.Context, userID uint) (string, error) {
	q := "SELECT login FROM Users WHERE id = $1"
	var login string
	err := repo.PostgresAdapter.QueryRow(ctx, q, userID).Scan(&login)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get user rows: %v, uID: %d", err, userID))
		return "", repoErrors.ErrFailToGetUser
	}

	if login == "" {
		logger.Info(ctx, fmt.Sprintf("zero value get with id %d", userID))
		return "", repoErrors.ErrUserNotFound
	}

	logger.Info(ctx, fmt.Sprintf("success get user name with id %d", userID))

	return login, nil
}

func (repo *AuthRepo) GetUserByID(ctx context.Context, userID uint) (models.User, error) {
	q := "SELECT name, sur_name, created_at FROM Users WHERE id = $1"
	var userTable []dao.User
	err := repo.PostgresAdapter.Select(ctx, &userTable, q, userID)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get user rows: %v, uID: %d", err, userID))
		return models.User{}, repoErrors.ErrFailToGetUser
	}

	if len(userTable) == 0 {
		logger.Info(ctx, fmt.Sprintf("zero value get with id %d", userID))
		return models.User{}, repoErrors.ErrUserNotFound
	}

	logger.Info(ctx, fmt.Sprintf("success get user name with id %d", userID))

	userModel := dao.ConvertUserTableToModel(userTable)

	return userModel[0], nil
}

func (repo *AuthRepo) GetUserPassword(ctx context.Context, userID uint) (string, error) {
	q := "SELECT password_hash FROM Users WHERE id = $1"
	var userTable []dao.User
	err := repo.PostgresAdapter.Select(ctx, &userTable, q, userID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get user rows: %v", err))
		return "", repoErrors.ErrFailToGetUser
	}
	if len(userTable) == 0 {
		logger.Info(ctx, fmt.Sprintf("zero value get with id %d", userID))
		return "", repoErrors.ErrUserNotFound
	}

	logger.Info(ctx, fmt.Sprintf("success get user password with id %d", userID))

	return userTable[0].PasswordHash, nil
}

func (repo *AuthRepo) IsVKUser(ctx context.Context, userID uint) bool {
	q := "SELECT vk_id FROM users WHERE id = $1"
	var userTable dao.User
	err := repo.PostgresAdapter.QueryRow(ctx, q, userID).Scan(&userTable.VKID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("fail to get user rows: %v", err))
		return false
	}

	logger.Info(ctx, fmt.Sprintf("success get vk user with id %d, vk_id: %v", userID, userTable.VKID))

	if !userTable.VKID.Valid {
		return false
	}
	return true
}

func (repo *AuthRepo) LoginVK(ctx context.Context, data models.VKLoginData) (string, error) {
	token, err := utils.ExchangeToken(
		ctx,
		data.Code,
		data.DeviceID,
		data.State,
		repo.Config,
	)

	if err != nil {
		return "", repoErrors.ErrFailedToGetToken
	}

	conf := &oauth2.Config{
		ClientID:     repo.Config.OauthAppID,
		ClientSecret: repo.Config.OauthAppSecret,
		Endpoint:     vk.Endpoint,
	}

	client := conf.Client(ctx, token)
	resp, err := client.Get(fmt.Sprintf(APIURL, token.AccessToken))
	if err != nil {
		return "", repoErrors.ErrFailedToGetDataByToken
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println(string(body))
	var respData struct {
		Response []struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			ID        uint   `json:"id"`
		} `json:"response"`
	}

	if err = json.Unmarshal(body, &respData); err != nil || len(respData.Response) == 0 {
		return "", repoErrors.ErrFailedToUnmarshalJSON
	}

	exist, uID := repo.IsExistsUserVK(ctx, respData.Response[0].ID)
	if !exist {
		uID, err = repo.CreateUserVK(ctx, models.UserVK{
			VKID:    respData.Response[0].ID,
			Name:    respData.Response[0].FirstName,
			SurName: respData.Response[0].LastName,
		})
		if err != nil {
			return "", repoErrors.ErrFailToCreateVKUser
		}
	}
	sID, err := repo.CreateSession(ctx, uID)
	if err != nil {
		return "", repoErrors.ErrFailToCreateSession
	}

	return sID, nil
}
