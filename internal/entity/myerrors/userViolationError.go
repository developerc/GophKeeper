package myerrors

import (
	"fmt"
)

type UserViolationError struct {
	Err   error
	Login string
}

func (ve *UserViolationError) Error() string {
	return fmt.Sprintf("user with login %s already exists", ve.Login)
}

func (ve *UserViolationError) Unwrap() error {
	return ve.Err
}

func NewUserViolationError(login string, err error) error {
	return &UserViolationError{
		Login: login,
		Err:   err,
	}
}
