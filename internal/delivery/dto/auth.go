package dto

import (
	"encoding/json"
	"net/http"
	"time"
)

type User struct {
	ID        uint      `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	SurName   string    `json:"surname,omitempty"`
	Login     string    `json:"login,omitempty"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	IsVKUser  bool      `json:"isVKUser"`
}

type EditUser struct {
	Password    string `json:"password,omitempty"`
	NewPassword string `json:"newPassword,omitempty"`
	NewLogin    string `json:"newLogin,omitempty"`
	NewName     string `json:"newName,omitempty"`
	NewSurname  string `json:"newSurname,omitempty"`
}

type VKLoginData struct {
	Code     string `json:"code"`
	State    string `json:"state"`
	DeviceID string `json:"deviceId"`
}

type UserName struct {
	Name string `json:"name"`
}

func GetSignupData(r *http.Request) (User, error) {
	var userDTO User

	err := json.NewDecoder(r.Body).Decode(&userDTO)

	if err != nil {
		return User{}, err
	}

	return userDTO, nil
}

func GetLoginData(r *http.Request) (User, error) {
	var userDTO User

	err := json.NewDecoder(r.Body).Decode(&userDTO)

	if err != nil {
		return User{}, err
	}

	return userDTO, nil
}

func GetEditData(r *http.Request) (EditUser, error) {
	var userDTO EditUser

	err := json.NewDecoder(r.Body).Decode(&userDTO)

	if err != nil {
		return EditUser{}, err
	}

	return userDTO, nil
}

func GetLoginVKData(r *http.Request) (VKLoginData, error) {
	var loginData VKLoginData

	err := json.NewDecoder(r.Body).Decode(&loginData)

	if err != nil {
		return VKLoginData{}, err
	}

	return loginData, nil
}
