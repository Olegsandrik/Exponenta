package utils

import (
	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	internalErrors "github.com/Olegsandrik/Exponenta/internal/errors"

	"regexp"
)

func ValidateName(name string) error {
	if len([]rune(name)) < 2 {
		return internalErrors.ErrTooShortUsername
	}
	return nil
}

func ValidateSurname(surName string) error {
	if len([]rune(surName)) < 2 {
		return internalErrors.ErrTooShortSurname
	}
	return nil
}

func ValidateSignUpUserData(user dto.User) error {
	if user.Name == "" || user.Login == "" || user.Password == "" || user.SurName == "" {
		return internalErrors.ErrEmptySingUpData
	}

	err := ValidateName(user.Name)
	if err != nil {
		return err
	}

	err = ValidateSurname(user.SurName)

	if err != nil {
		return err
	}

	re := regexp.MustCompile(`[!@#$&*]`)

	if len(user.Password) < 8 || len(re.FindAllString(user.Password, -1)) < 2 {
		return internalErrors.ErrTooEasyPassword
	}
	return nil
}
