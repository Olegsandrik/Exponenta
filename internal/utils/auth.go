package utils

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	internalErrors "github.com/Olegsandrik/Exponenta/internal/internalerrors"
)

type UserID struct{}

func GetUserIDFromContext(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(UserID{}).(uint)
	if !ok || userID == 0 {
		return 0, internalErrors.ErrUserNotAuth
	}
	return userID, nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return internalErrors.ErrInvalidPassword
	}
	return nil
}
