package usecaseErrors

import "fmt"

var (
	ErrEmptyPassword    = fmt.Errorf("password or newPassword is empty")
	ErrEmptySingUpData  = fmt.Errorf("name or username or login or password is empty")
	ErrEmptyName        = fmt.Errorf("new name is empty")
	ErrEmptySurname     = fmt.Errorf("new surname is empty")
	ErrEmptyLogin       = fmt.Errorf("new login is empty")
	ErrEmptyVKLoginData = fmt.Errorf("device id or code or state is empty")
)
