package myerrors

import (
	"fmt"
)

type InvalidPasswordError struct {
	Password string
}

func (ip *InvalidPasswordError) Error() string {
	return fmt.Sprintf("password %s invalid", ip.Password)
}

func NewInvalidPasswordError(password string) error {
	return &InvalidPasswordError{
		Password: password,
	}
}
