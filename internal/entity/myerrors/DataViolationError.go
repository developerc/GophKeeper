package myerrors

import (
	"fmt"
)

type DataViolationError struct {
	Err  error
	Name string
}

func (dv *DataViolationError) Error() string {
	return fmt.Sprintf("data with name %s already exists", dv.Name)
}

func (dv *DataViolationError) Unwrap() error {
	return dv.Err
}

func NewDataViolationError(name string, err error) error {
	return &DataViolationError{
		Name: name,
		Err:  err,
	}
}
