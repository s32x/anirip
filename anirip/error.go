package anirip

import "fmt"

type Error struct {
	Message string
	Err     error
}

func (e Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf(">>> Error : %v : %v", e.Message, e.Err)
	}
	return fmt.Sprintf(">>> Error : %v.", e.Message)
}
