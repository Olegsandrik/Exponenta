package models

import (
	"time"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
)

type User struct {
	ID           uint
	Name         string
	SurName      string
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserVK struct {
	ID        uint
	VKID      uint
	Name      string
	SurName   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserSession struct {
	UserID    uint
	SessionID string
}

type VKLoginData struct {
	Code     string
	State    string
	DeviceID string
}

func ConvertUserModelToDTO(user User) dto.User {
	return dto.User{
		ID:        user.ID,
		Name:      user.Name,
		SurName:   user.SurName,
		Login:     user.Login,
		CreatedAt: user.CreatedAt,
	}
}

func ConvertVKLoginDataDtoToModel(loginData dto.VKLoginData) VKLoginData {
	return VKLoginData{
		Code:     loginData.Code,
		State:    loginData.State,
		DeviceID: loginData.DeviceID,
	}
}
