package myerrors

import (
	"fmt"
)

type NotFoundError struct {
	Err  error
	Name string
}

func (nf *NotFoundError) Error() string {
	return fmt.Sprintf("data with name %s not found", nf.Name)
}

func (nf *NotFoundError) Unwrap() error {
	return nf.Err
}

func NewNotFoundError(name string, err error) error {
	return &NotFoundError{
		Name: name,
		Err:  err,
	}
}
