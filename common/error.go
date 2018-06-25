package common

import "fmt"

// Error represents a generic anirip error
type Error struct {
	Msg string
	Err error
}

// NewError generates a new generic anirip error
func NewError(msg string, err error) *Error {
	return &Error{Msg: msg, Err: err}
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Error: %v - %v", e.Msg, e.Err)
	}
	return fmt.Sprintf("Error: %v", e.Msg)
}
